// @module annotgen/types/model
// @brief Shared types used across annotgen
// @description
// This module defines core data structures like Field, ClassInfo, ModuleHeader
// and FileAnnotations used throughout the annotation and generation process.

package types

// Field represents a single field in a Lua module (e.g. M.foo = …)
type Field struct {
	Name        string // e.g. "rate_limits"
	Type        string // e.g. "RateLimits"
	Line        int    // line number in source file
	HasDoc      bool   // whether this field has a comment annotation
	Comment     string // extracted docstring if available
	IsFunction  bool   // true if it's a function, false if a value/table
}

// ClassInfo represents a full @class block with associated fields
type ClassInfo struct {
	ClassName  string  // e.g. "ReposcopeMetrics"
	Extends    string  // optional: parent class
	Fields     []Field // list of fields in the class
}

// ModuleHeader holds top-of-file annotations like @module, @brief, @desc
type ModuleHeader struct {
	ModulePath string // e.g. "mygrep.types.aliases"
	Brief      string
	Desc       string
	Class      *ClassInfo // optional: attached class
}

// FileAnnotations bundles everything found/generated in a file
type FileAnnotations struct {
	Path     string       // file path
	Header   ModuleHeader // header info
	Existing []string     // lines of current file before modification
	Updated  []string     // output with new annotations
}
