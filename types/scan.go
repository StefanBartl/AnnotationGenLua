package types

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var classDefPattern = regexp.MustCompile(`^---@class\s+([\w_]+)`)

func ScanGlobalTypes(root string) map[string]bool {
	result := make(map[string]bool)

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".lua") {
			return nil
		}
		if !strings.Contains(path, "/types/") && !strings.Contains(path, "/@types/") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if classDefPattern.MatchString(line) {
				name := classDefPattern.FindStringSubmatch(line)[1]
				result[name] = true
			}
		}
		return nil
	})

	return result
}
