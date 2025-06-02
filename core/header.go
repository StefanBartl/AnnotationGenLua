package core

import (
	"fmt"
	"strings"

	"annotgen/types"
)

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
		classBlock := RenderClassAnnotations(h.Class)
		lines = append(lines, classBlock...)
	}

	return lines
}

// Nur für `.annot.lua`
func BuildStandaloneHeader(fa *types.FileAnnotations) []string {
	return BuildHeaderBlock(fa)
}

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
	out = append(out, "")
	out = append(out, oldLines[start:]...)
	fa.Updated = out
	return out
}
