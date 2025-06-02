// @module annotgen/core/annotator
// @brief Renders structured class and field annotations for Lua files
// @description
// This module transforms parsed ClassInfo into a list of strings that
// represent `---@class` and `---@field` annotations, ready to be inserted
// into a Lua source file.

package core

import (
	"fmt"
	"annotgen/types"
)

// RenderClassAnnotations builds `---@class` + `---@field` blocks from ClassInfo
func RenderClassAnnotations(class *types.ClassInfo) []string {
	if class == nil {
		return nil
	}

	var lines []string

	// Header line: @class ClassName [: Parent]
	if class.Extends != "" {
		lines = append(lines, fmt.Sprintf("---@class %s : %s", class.ClassName, class.Extends))
	} else {
		lines = append(lines, fmt.Sprintf("---@class %s", class.ClassName))
	}

	// Fields: sorted in order of appearance
	for _, field := range class.Fields {
		fieldType := field.Type
		if fieldType == "" {
			if field.IsFunction {
				fieldType = "fun(...)"
			} else {
				fieldType = "any"
			}
		}
		lines = append(lines, fmt.Sprintf("---@field %s %s", field.Name, fieldType))
	}

	return lines
}
