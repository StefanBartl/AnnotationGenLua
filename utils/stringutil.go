// @module annotgen/utils/stringutil
// @brief Common string utilities for annotation formatting
// @description
// Contains helpers to clean up, normalize or dedent strings.

package utils

import "strings"

// Dedent removes common leading indentation from all lines
func Dedent(lines []string) []string {
	minIndent := -1
	for _, line := range lines {
		trim := strings.TrimLeft(line, " \t")
		if trim == "" {
			continue
		}
		indent := len(line) - len(trim)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return lines
	}

	var result []string
	for _, line := range lines {
		if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}
	return result
}

// TrimEmpty trims leading/trailing empty lines
func TrimEmpty(lines []string) []string {
	start, end := 0, len(lines)-1
	for start <= end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end >= start && strings.TrimSpace(lines[end]) == "" {
		end--
	}
	return lines[start : end+1]
}
