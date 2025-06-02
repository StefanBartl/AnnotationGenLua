// @module annotgen/fs/walker
// @brief Finds all relevant .lua files from a given path
// @description
// Walks a directory tree recursively and collects all .lua files,
// or directly returns the given file path if it's a single file.

package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// WalkLuaFiles returns all .lua files from a file or folder path
func WalkLuaFiles(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	// If a single .lua file
	if !info.IsDir() {
		if strings.HasSuffix(root, ".lua") {
			return []string{root}, nil
		}
		return nil, errors.New("given file is not a .lua file")
	}

	// Otherwise walk directory
	var result []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // can't access this file
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".lua") {
			result = append(result, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
