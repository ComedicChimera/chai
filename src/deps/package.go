package deps

import (
	"chai/common"
	"chai/logging"
	"chai/syntax"
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
		ID:       common.GenerateIDFromPath(rootPath),
		Name:     filepath.Base(rootPath),
		RootPath: rootPath,
	}
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
	AST *syntax.ASTBranch

	// GlobalTable is the table of globally declared symbols in this package
	GlobalTable map[string]*Symbol
}

// AddSymbolImports adds a list of names as imported symbols of this file. This
// function does NOT validate that those symbols are visible in the imported
// package.
func (cf *ChaiFile) AddSymbolImports(importedPkg *ChaiPackage, importedSymbolNames []string) {
	// TODO
}

// AddPackageImport adds a package as an import of this file -- this is a
// package that is visible by name within the file (eg. `import pkg`).  This
// does NOT check for cross-module import cycles, but it will return an error if
// the imported symbols occur multiple times within the file (this error's
// message is the name of the duplicate imported symbol).
func (cf *ChaiFile) AddPackageImport(importedPkg *ChaiPackage, importedPkgName string) error {
	// TODO
	return nil
}
