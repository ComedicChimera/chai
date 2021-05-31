package deps

import (
	"chai/logging"
	"chai/syntax"
	"hash/fnv"
	"path/filepath"
)

// ChaiPackage represents a package in a Chai project
type ChaiPackage struct {
	// ID is a unique identifier for a package that is based on the package path
	// -- this is used for package look ups and namespacing in LLVM modules (to
	// prevent name collisions)
	ID uint

	// Name is the short name of the package
	Name string

	// RootPath is the absolute path to the root directory of the package
	RootPath string

	// Files contains all the individual files in this package
	Files []*ChaiFile
}

// NewPackage creates a new Chai package based on the given absolute, root path
// (does NOT perform file initialization)
func NewPackage(rootPath string) *ChaiPackage {
	return &ChaiPackage{
		ID:       generatePkgID(rootPath),
		Name:     filepath.Base(rootPath),
		RootPath: rootPath,
	}
}

// generatePkgID converts a package path into a package ID
func generatePkgID(abspath string) uint {
	h := fnv.New32a()
	h.Write([]byte(abspath))
	return uint(h.Sum32())
}

// ChaiFile represents a file of Chai source code
type ChaiFile struct {
	// Parent is a reference to this file's parent package
	Parent *ChaiPackage

	// FilePath is the absolute path to the file
	FilePath string

	// LogContext is the log context for this file
	LogContext *logging.LogContext

	// AST is the abstract syntax tree representing the contents of this file
	AST syntax.ASTNode

	// GlobalTable is the table of globally declared symbols in this package
	GlobalTable map[string]*Symbol
}
