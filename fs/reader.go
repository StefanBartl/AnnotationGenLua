// @module annotgen/fs/reader
// @brief Loads Lua file contents as lines
// @description
// This utility reads a file and returns its content as a string slice,
// useful for dry-run or preprocessing extensions.

package fs

import (
	"bufio"
	"os"
)

// ReadLines returns the content of a file as a slice of strings
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
