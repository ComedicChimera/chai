package sem

import (
	"chai/common"
	"chai/logging"
	"chai/syntax"
	"errors"
	"fmt"
	"path/filepath"
)

// ChaiPackage represents a package in a Chai project
type ChaiPackage struct {
	// ID is a unique identifier for a package that is based on the package path
	// -- this is used for package look ups and namespacing in LLVM modules (to
	// prevent name collisions)
	ID uint

	// ParentID is the unique identifier of the module containing this package
	ParentID uint

	// Name is the short name of the package
	Name string

	// RootPath is the absolute path to the root directory of the package
	RootPath string

	// Files contains all the individual files in this package
	Files []*ChaiFile

	// GlobalTable is the table of globally declared symbols in this package
	GlobalTable map[string]*Symbol

	// ImportTable stores all the packages this package depends on
	ImportTable map[uint]*ChaiPackage

	// Initialized is used to indicate whether or not this package has been
	// initialized fully yet (ie. all its files have been initialized and all
	// their dependencies have been initialized).  We use this to detect modular
	// import cycles.
	Initialized bool
}

// NewPackage creates a new Chai package based on the given absolute, root path
// (does NOT perform file initialization)
func NewPackage(rootPath string) *ChaiPackage {
	return &ChaiPackage{
		ID:          common.GenerateIDFromPath(rootPath),
		Name:        filepath.Base(rootPath),
		RootPath:    rootPath,
		GlobalTable: make(map[string]*Symbol),
		ImportTable: make(map[uint]*ChaiPackage),
	}
}

// ImportSymbol attempts to import a symbol from the public namespace of a
// package.  It does not throw an error if the symbol could not imported.
func (pkg *ChaiPackage) ImportSymbol(name string) (*Symbol, bool) {
	if sym, ok := pkg.GlobalTable[name]; ok && sym.Public {
		return sym, true
	}

	return nil, false
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

	// Metadata is the map of metadata flags and arguments set for this file
	Metadata map[string]string

	// ImportedSymbols stores all the symbols this file imports
	ImportedSymbols map[string]*Symbol

	// VisiblePackages is a map of all the packages visible within the namespace
	// of the package arranged by the name by which they can be accessed
	VisiblePackages map[string]*ChaiPackage
}

// NewFile creates a new file inside a given package but does NOT add it to
// the list of files in that package
func NewFile(parent *ChaiPackage, fabspath string) *ChaiFile {
	return &ChaiFile{
		Parent:          parent,
		FilePath:        fabspath,
		LogContext:      &logging.LogContext{PackageID: parent.ID, FilePath: fabspath},
		ImportedSymbols: make(map[string]*Symbol),
		VisiblePackages: make(map[string]*ChaiPackage),
	}
}

// AddSymbolImports adds a list of names as imported symbols of this file. This
// function does NOT validate that those symbols are visible in the imported
// package.  It handles all errors and returns a flag indicating whether or not
// the attachment succeeded.
func (cf *ChaiFile) AddSymbolImports(importedPkg *ChaiPackage, importedSymbols map[string]*logging.TextPosition) bool {
	for name, pos := range importedSymbols {
		if name == "_" {
			logging.LogCompileError(
				cf.LogContext,
				"unable to import symbol by name `_`",
				logging.LMKImport,
				pos,
			)

			return false
		}

		if _, ok := cf.ImportedSymbols[name]; ok {
			logging.LogCompileError(
				cf.LogContext,
				fmt.Sprintf("symbol `%s` imported multiple times", name),
				logging.LMKImport,
				pos,
			)

			return false
		} else if _, ok := cf.VisiblePackages[name]; ok {
			logging.LogCompileError(
				cf.LogContext,
				fmt.Sprintf("symbol `%s` imported multiple times", name),
				logging.LMKImport,
				pos,
			)

			return false
		}

		cf.ImportedSymbols[name] = &Symbol{
			Name:       name,
			SrcPackage: importedPkg,
			Public:     true,
		}
	}

	if _, ok := cf.Parent.ImportTable[importedPkg.ID]; !ok {
		cf.Parent.ImportTable[importedPkg.ID] = importedPkg
	}
	return true
}

// AddPackageImport adds a package as an import of this file -- this is a
// package that is visible by name within the file (eg. `import pkg`).  This
// does NOT check for cross-module import cycles, but it will return an error if
// the imported symbols occur multiple times within the file (this error's
// message is the name of the duplicate imported symbol).
func (cf *ChaiFile) AddPackageImport(importedPkg *ChaiPackage, importedPkgName string) error {
	if importedPkgName != "_" {
		if _, ok := cf.ImportedSymbols[importedPkgName]; ok {
			return errors.New(importedPkgName)
		} else if _, ok := cf.VisiblePackages[importedPkgName]; ok {
			return errors.New(importedPkgName)
		}

		cf.VisiblePackages[importedPkgName] = importedPkg
	}

	if _, ok := cf.Parent.ImportTable[importedPkg.ID]; !ok {
		cf.Parent.ImportTable[importedPkg.ID] = importedPkg
	}

	return nil
}