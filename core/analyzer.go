package core

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"annotgen/types"
)

var (
	assignPattern   = regexp.MustCompile(`^M\.(\w+)\s*=\s*`)
	funcPattern     = regexp.MustCompile(`^function\s+M\.(\w+)\s*\(`)
	classPattern    = regexp.MustCompile(`^---@class\s+([^\s:]+)(?:\s*:\s*([^\s]+))?`)
	modulePattern   = regexp.MustCompile(`^---@module\s+['"](.+)['"]`)
	briefPattern    = regexp.MustCompile(`^---@brief\s*(.*)`)
	descPattern     = regexp.MustCompile(`^---@desc\s*(.*)`)
	fieldPattern    = regexp.MustCompile(`^---@field\s+(\w+)\s+([\w\[\]\|]+)`)
	localMPattern   = regexp.MustCompile(`^local\s+M\s*=\s*\{\s*\}`)
)

// PascalCase aus Dateinamen
func extractClassNameFromPath(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".lua")
	parts := strings.Split(name, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

// Modulpfad z. B. reposcope.utils.metrics aus ./lua/reposcope/utils/metrics.lua
func inferModulePath(path string) string {
	parts := strings.Split(path, string(filepath.Separator))

	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "lua" && i < len(parts)-1 {
			moduleParts := parts[i+1:]
			moduleFile := strings.TrimSuffix(moduleParts[len(moduleParts)-1], ".lua")
			moduleParts[len(moduleParts)-1] = moduleFile
			return strings.Join(moduleParts, ".")
		}
	}

	// Fallback: nur Dateiname
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".lua")
}

func AnalyzeFile(path string) (*types.FileAnnotations, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		lines      []string
		class      *types.ClassInfo
		module     string
		brief      string
		desc       string
		fields     []types.Field
		foundNames = map[string]bool{}
	)

	scanner := bufio.NewScanner(file)
	ln := 0
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		ln++

		trim := strings.TrimSpace(line)

		switch {
		case modulePattern.MatchString(trim):
			module = modulePattern.FindStringSubmatch(trim)[1]

		case briefPattern.MatchString(trim):
			brief = briefPattern.FindStringSubmatch(trim)[1]

		case descPattern.MatchString(trim):
			desc = descPattern.FindStringSubmatch(trim)[1]

		case classPattern.MatchString(trim):
			matches := classPattern.FindStringSubmatch(trim)
			class = &types.ClassInfo{
				ClassName: matches[1],
				Extends:   matches[2],
			}

		case fieldPattern.MatchString(trim) && class != nil:
			matches := fieldPattern.FindStringSubmatch(trim)
			field := types.Field{
				Name:   matches[1],
				Type:   matches[2],
				HasDoc: true,
			}
			fields = append(fields, field)
			foundNames[field.Name] = true
		}
	}

	if class == nil {
		for _, line := range lines {
			if localMPattern.MatchString(line) {
				name := extractClassNameFromPath(path)
				class = &types.ClassInfo{
					ClassName: name,
					Extends:   name + "Def",
				}
				break
			}
		}
	}

	for idx, line := range lines {
		if matches := assignPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if foundNames[name] {
				continue
			}
			field := types.Field{
				Name:       name,
				Type:       "",
				Line:       idx + 1,
				HasDoc:     false,
				IsFunction: strings.Contains(line, "function("),
			}
			fields = append(fields, field)
			foundNames[name] = true
		}
		if matches := funcPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if foundNames[name] {
				continue
			}
			field := types.Field{
				Name:       name,
				Type:       "fun(...)",
				Line:       idx + 1,
				HasDoc:     false,
				IsFunction: true,
			}
			fields = append(fields, field)
			foundNames[name] = true
		}
	}

	if class != nil {
		class.Fields = fields
	}

	if module == "" {
		module = inferModulePath(path)
	}

	return &types.FileAnnotations{
		Path: path,
		Header: types.ModuleHeader{
			ModulePath: module,
			Brief:      brief,
			Desc:       desc,
			Class:      class,
		},
		Existing: lines,
	}, nil
}
