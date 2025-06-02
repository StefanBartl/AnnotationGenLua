package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"annotgen/types"
)

var (
	assignPattern   = regexp.MustCompile(`^M\.(\w+)\s*=\s*(\{)?`)
	funcPattern     = regexp.MustCompile(`^function\s+M\.(\w+)\s*\(([^)]*)\)`)
	classPattern    = regexp.MustCompile(`^---@class\s+([^\s:]+)`)
	modulePattern   = regexp.MustCompile(`^---@module\s+['"](.+)['"]`)
	briefPattern    = regexp.MustCompile(`^---@brief\s*(.*)`)
	descPattern     = regexp.MustCompile(`^---@desc\s*(.*)`)
	paramPattern    = regexp.MustCompile(`^---@param\s+(\w+)\s+(\w+)`)
	returnPattern   = regexp.MustCompile(`^---@return\s+(.+)`)
	localMPattern   = regexp.MustCompile(`^local\s+M\s*=\s*\{\s*\}`)
)

func camelize(name string) string {
	base := strings.TrimSuffix(name, ".lua")
	parts := strings.Split(base, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func inferModuleClassName(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	project := "Project"
	file := "Unknown"
	for i, part := range parts {
		if part == "lua" && i < len(parts)-1 {
			if i+1 < len(parts) {
				project = strings.Title(parts[i+1])
			}
			if i+2 < len(parts) {
				file = camelize(parts[len(parts)-1])
			}
			break
		}
	}
	return project + file
}

func inferModulePath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "lua" && i < len(parts)-1 {
			moduleParts := parts[i+1:]
			moduleParts[len(moduleParts)-1] = strings.TrimSuffix(moduleParts[len(moduleParts)-1], ".lua")
			return strings.Join(moduleParts, ".")
		}
	}
	return strings.TrimSuffix(filepath.Base(path), ".lua")
}

func AnalyzeFile(path string) (*types.FileAnnotations, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		lines        []string
		fields       []types.Field
		foundNames   = map[string]bool{}
		module       = ""
		brief        = ""
		desc         = ""
		classNameMap = map[string]string{}
	)

	var class *types.ClassInfo
	var prevClass string
	var paramBuffer []types.Param
	var returnBuffer string
	var docComment string
	var hasTaggedComment bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		trim := strings.TrimSpace(line)

		switch {
		case modulePattern.MatchString(trim):
			module = modulePattern.FindStringSubmatch(trim)[1]

		case briefPattern.MatchString(trim):
			brief = briefPattern.FindStringSubmatch(trim)[1]

		case descPattern.MatchString(trim):
			desc = descPattern.FindStringSubmatch(trim)[1]

		case classPattern.MatchString(trim):
			prevClass = classPattern.FindStringSubmatch(trim)[1]

		case paramPattern.MatchString(trim):
			m := paramPattern.FindStringSubmatch(trim)
			paramBuffer = append(paramBuffer, types.Param{Name: m[1], Type: m[2]})
			hasTaggedComment = true

		case returnPattern.MatchString(trim):
			r := strings.TrimSpace(returnPattern.FindStringSubmatch(trim)[1])
			hasTaggedComment = true
			// extract type only, ignoring description
			if strings.HasPrefix(r, "{") && strings.Contains(r, "}") {
				i := strings.Index(r, "}") + 1
				if i > 0 && i <= len(r) {
					r = r[:i]
				}
			} else if i := strings.Index(r, " "); i > 0 {
				r = r[:i]
			}
			returnBuffer = r

		case strings.HasPrefix(trim, "---") && !strings.HasPrefix(trim, "---@"):
			if docComment == "" && !hasTaggedComment {
				docComment = strings.TrimSpace(strings.TrimPrefix(trim, "---"))
			}

		case funcPattern.MatchString(trim):
			m := funcPattern.FindStringSubmatch(trim)
			name := m[1]
			paramList := strings.TrimSpace(m[2])
			if foundNames[name] {
				continue
			}

			var params []types.Param
			if len(paramBuffer) == 0 && paramList != "" {
				rawNames := strings.Split(paramList, ",")
				for _, r := range rawNames {
					n := strings.TrimSpace(r)
					if n != "" {
						params = append(params, types.Param{Name: n, Type: "any"})
					}
				}
			} else {
				params = paramBuffer
			}

			field := types.Field{
				Name:       name,
				IsFunction: true,
				Params:     params,
				ReturnType: returnBuffer,
			}

			var paramSig []string
			for _, p := range params {
				paramSig = append(paramSig, fmt.Sprintf("%s: %s", p.Name, p.Type))
			}

			ret := returnBuffer
			if ret == "" {
				ret = "any"
			}

			field.Type = fmt.Sprintf("fun(%s): %s", strings.Join(paramSig, ", "), ret)
			if docComment != "" {
				field.Type += " " + strings.TrimSpace(docComment)
			}

			fields = append(fields, field)
			foundNames[name] = true

			paramBuffer = nil
			returnBuffer = ""
			docComment = ""
			hasTaggedComment = false

		case assignPattern.MatchString(trim):
			m := assignPattern.FindStringSubmatch(trim)
			name := m[1]
			isTable := m[2] == "{"
			if prevClass != "" && isTable {
				classNameMap[name] = prevClass
			}
			prevClass = ""

		default:
			if !strings.HasPrefix(trim, "--") {
				paramBuffer = nil
				returnBuffer = ""
				docComment = ""
				hasTaggedComment = false
			}
		}
	}

	for idx, line := range lines {
		if matches := assignPattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if foundNames[name] {
				continue
			}

			fieldType := "any"
			if cls, ok := classNameMap[name]; ok {
				fieldType = cls
			}

			field := types.Field{
				Name:       name,
				Type:       fieldType,
				Line:       idx + 1,
				IsFunction: false,
			}
			fields = append(fields, field)
			foundNames[name] = true
		}
	}

	className := inferModuleClassName(path)
	class = &types.ClassInfo{
		ClassName: className,
		Extends:   className + "Def",
		Fields:    fields,
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
