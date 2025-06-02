// @module annotgen/utils/log
// @brief Optional logging helper
// @description
// Provides basic debug logging for verbose CLI output.

package utils

import (
	"fmt"
	"os"
)

var verbose bool = false

// SetVerbose activates logging
func SetVerbose(v bool) {
	verbose = v
}

// Debug prints only if verbosity is enabled
func Debug(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[annotgen] "+format+"\n", args...)
	}
}
