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
	overloadPattern = regexp.MustCompile(`^---@overload\s+fun\((.*)\)`)
	modulePattern   = regexp.MustCompile(`^---@module\s+['"](.+)['"]`)
	briefPattern    = regexp.MustCompile(`^---@brief\s*(.*)`)
	descPattern     = regexp.MustCompile(`^---@desc\s*(.*)`)
	paramPattern    = regexp.MustCompile(`^---@param\s+(\w+)\s+(\w+)`)
	returnPattern   = regexp.MustCompile(`^---@return\s+(.+)`)
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
		lines          []string
		fields         []types.Field
		foundNames     = map[string]bool{}
		module         = ""
		brief          = ""
		desc           = ""
		classNameMap   = map[string]string{}
		declaredTypes  = map[string]bool{}
		prevClass      string
		docComment     string
		hasTagged      bool
		overloads      []string
		paramBuffer    []types.Param
		returnBuffer   string
	)

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
			declaredTypes[prevClass] = true

		case paramPattern.MatchString(trim):
			m := paramPattern.FindStringSubmatch(trim)
			paramBuffer = append(paramBuffer, types.Param{Name: m[1], Type: m[2]})
			hasTagged = true

		case returnPattern.MatchString(trim):
			r := strings.TrimSpace(returnPattern.FindStringSubmatch(trim)[1])
			hasTagged = true
			if strings.HasPrefix(r, "{") && strings.Contains(r, "}") {
				i := strings.Index(r, "}") + 1
				if i > 0 && i <= len(r) {
					r = r[:i]
				}
			} else if i := strings.Index(r, " "); i > 0 {
				r = r[:i]
			}
			returnBuffer = r

		case overloadPattern.MatchString(trim):
			overloads = append(overloads, "---@overload fun("+overloadPattern.FindStringSubmatch(trim)[1]+")")

		case strings.HasPrefix(trim, "---") && !strings.HasPrefix(trim, "---@"):
			if docComment == "" && !hasTagged {
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
				s := strings.Split(paramList, ",")
				for _, p := range s {
					n := strings.TrimSpace(p)
					if n != "" {
						params = append(params, types.Param{Name: n, Type: "any"})
					}
				}
			} else {
				params = paramBuffer
			}
			ret := returnBuffer
			if ret == "" {
				ret = "any"
			}
			var paramSig []string
			for _, p := range params {
				paramSig = append(paramSig, fmt.Sprintf("%s: %s", p.Name, p.Type))
			}
			f := types.Field{
				Name:       name,
				IsFunction: true,
				Params:     params,
				ReturnType: ret,
				Overloads:  overloads,
				Type:       fmt.Sprintf("fun(%s): %s", strings.Join(paramSig, ", "), ret),
			}
			if docComment != "" {
				f.Type += " " + docComment
			}
			fields = append(fields, f)
			foundNames[name] = true
			paramBuffer = nil
			returnBuffer = ""
			docComment = ""
			hasTagged = false
			overloads = nil

		case assignPattern.MatchString(trim):
			m := assignPattern.FindStringSubmatch(trim)
			name := m[1]
			if prevClass != "" && m[2] == "{" {
				classNameMap[name] = prevClass
			}
			prevClass = ""

		default:
			if !strings.HasPrefix(trim, "--") {
				paramBuffer = nil
				returnBuffer = ""
				docComment = ""
				hasTagged = false
			}
		}
	}

	for idx, line := range lines {
		if m := assignPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if foundNames[name] {
				continue
			}
			guess := strings.Title(name)
			typ := "any"
			if declaredTypes[guess] {
				typ = guess
			} else if cls, ok := classNameMap[name]; ok {
				typ = cls
			}
			fields = append(fields, types.Field{
				Name:       name,
				Type:       typ,
				Line:       idx + 1,
				IsFunction: false,
			})
			foundNames[name] = true
		}
	}

	className := inferModuleClassName(path)
	class := &types.ClassInfo{
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

func AnalyzeFileWithOptions(path string, resolveGlobals bool) (*types.FileAnnotations, error) {
	ann, err := AnalyzeFile(path)
	if err != nil {
		return nil, err
	}

	if resolveGlobals {
		projectRoot := findProjectRoot(path)
		if projectRoot != "" {
			globals := types.ScanGlobalTypes(projectRoot)
			ann.Globals = globals
		}
	}

	return ann, nil
}

func findProjectRoot(path string) string {
	dir := filepath.Dir(path)
	for i := 0; i < 10; i++ {
		if hasMarker(dir) {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func hasMarker(dir string) bool {
	files := []string{".git", "init.lua", "lua"}
	for _, name := range files {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}
