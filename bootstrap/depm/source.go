package depm

import (
	"chai/ast"
	"chai/report"
	"time"
)

// ChaiFile represents a Chai source file.
type ChaiFile struct {
	// Context is the module-relative compilation context of the file.
	Context *report.CompilationContext

	// Parent is the parent package to the file.
	Parent *ChaiPackage

	// Metadata is the map of all metadata keys specified for this file.
	// Metadata flags have an empty string as their value.
	Metadata map[string]string

	// Defs is the list of AST definitions that make up this source file.
	Defs []ast.Def

	// TODO: rest
}

// ChaiPackage represents a Chai source package.
type ChaiPackage struct {
	// ID is the unique ID of this package.
	ID uint

	// Name is the package name.
	Name string

	// Parent is the parent module to this package.
	Parent *ChaiModule

	// Files is a list of all the Chai source files that belong to this package.
	Files []*ChaiFile

	// GlobalTable is the global symbol table for this package.
	GlobalTable *SymbolTable

	// TODO: rest
}

// -----------------------------------------------------------------------------

// ChaiModule represents a Chai module.
type ChaiModule struct {
	// ID is the unique ID of this module.
	ID uint

	// Name is the module name.
	Name string

	// AbsPath is the absolute path to the root of the module.
	AbsPath string

	// RootPackage is the package at the module root.
	RootPackage *ChaiPackage

	// SubPackages is the map of sub-packages of this module organized by
	// sub-path: eg. the package `io.fs.path` would have a sub-path of
	// `.fs.path`.
	SubPackages map[string]*ChaiPackage

	// -----------------------------------------------------------------------------

	// ShouldCache indicates if compilation caching is enabled for this module.
	ShouldCache bool

	// LastBuildTime is a field used by modules to support compilation caching.
	LastBuildTime *time.Time
}

// Packages gets a list of the packages of this module.
func (m *ChaiModule) Packages() []*ChaiPackage {
	pkgs := make([]*ChaiPackage, len(m.SubPackages)+1)
	pkgs[0] = m.RootPackage

	n := 1
	for _, pkg := range m.SubPackages {
		pkgs[n] = pkg
		n++
	}

	return pkgs
}

// BuildProfile represents a loaded module build profile.  This profile only
// contains the information necessary to produce the binary and identify the
// correct build profiles of sub-modules according to the module schema.
type BuildProfile struct {
	Debug      bool
	OutputPath string
	TargetOS   string
	TargetArch string

	// OutputFormat should be one of the enumerated output formats.
	OutputFormat int

	// LinkObjects is the collection of path's to link with the final
	// executable.
	LinkObjects []string
}

// Enumeration of possible outform formats.
const (
	FormatBin = iota
	FormatObj
	FormatASM
	FormatLLVM
)
