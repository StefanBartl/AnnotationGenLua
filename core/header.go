package core

import (
	"fmt"
	"strings"

	"annotgen/types"
)

// Renders both the definition class (with fields) and the derived class
func RenderSplitClassAnnotations(class *types.ClassInfo) ([]string, string) {
	if class == nil {
		return nil, ""
	}

	defName := class.Extends
	if defName == "" {
		defName = class.ClassName + "Def"
	}

	var defLines []string
	defLines = append(defLines, fmt.Sprintf("---@class %s", defName))

	for _, field := range class.Fields {
		fieldType := field.Type
		if fieldType == "" {
			if field.IsFunction {
				fieldType = "fun(...)"
			} else {
				fieldType = "any"
			}
		}
		defLines = append(defLines, fmt.Sprintf("---@field %s %s", field.Name, fieldType))
	}

	instLine := fmt.Sprintf("---@class %s : %s", class.ClassName, defName)
	return defLines, instLine
}

// Builds the complete top annotation block for merging into the original file
func BuildHeaderBlock(fa *types.FileAnnotations) []string {
	h := fa.Header
	var lines []string

	lines = append(lines, fmt.Sprintf("---@module '%s'", h.ModulePath))
	if h.Brief != "" {
		lines = append(lines, fmt.Sprintf("---@brief %s", h.Brief))
	} else {
		lines = append(lines, "---@brief LEER")
	}
	if h.Desc != "" {
		lines = append(lines, fmt.Sprintf("---@desc %s", h.Desc))
	} else {
		lines = append(lines, "---@desc LEER")
	}

	if h.Class != nil {
		defBlock, instLine := RenderSplitClassAnnotations(h.Class)
		lines = append(lines, defBlock...)
		lines = append(lines, "") // spacer
		lines = append(lines, instLine)
		lines = append(lines, "local M = {}")
	}

	return lines
}

// Used when merging annotation block into full file content
func MergeWithFileContent(fa *types.FileAnnotations) []string {
	oldLines := fa.Existing
	var out []string

	start := 0
	for i, line := range oldLines {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "--@") && trim != "" {
			start = i
			break
		}
	}

	out = append(out, BuildHeaderBlock(fa)...)
	out = append(out, "") // spacer
	out = append(out, oldLines[start:]...)
	fa.Updated = out
	return out
}

// Used for .annot.lua output: annotations only
func BuildStandaloneHeader(fa *types.FileAnnotations) []string {
	return BuildHeaderBlock(fa)
}
