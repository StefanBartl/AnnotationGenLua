// @module annotgen/types/model
// @brief Shared types used across annotgen
// @description
// This module defines core data structures like Field, ClassInfo, ModuleHeader
// and FileAnnotations used throughout the annotation and generation process.

package types

// Param represents a single parameter of a function field
type Param struct {
	Name string // parameter name
	Type string // parameter type
}

// Field represents a field in M.*
type Field struct {
	Name       string
	Type       string
	IsFunction bool
	Params     []Param
	ReturnType string
	Overloads  []string
	Line       int
}

// ClassInfo represents a @class block
type ClassInfo struct {
	ClassName string
	Extends   string
	Fields    []Field
}

// ModuleHeader holds file-level annotations
type ModuleHeader struct {
	ModulePath string
	Brief      string
	Desc       string
	Class      *ClassInfo
}

// FileAnnotations represents all extracted annotations from a file
type FileAnnotations struct {
	Path     string
	Header   ModuleHeader
	Fields []Field
	Globals map[string]bool
	Existing []string
	Updated  []string
}
