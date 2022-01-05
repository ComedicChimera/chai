package depm

import (
	"chai/ast"
	"chai/report"
	"fmt"
	"log"
	"path/filepath"
	"strings"
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

	// ------------------------------------------------------------------------

	// ImportedSymbols is the table of symbols specifically imported by this
	// file.  This table is separated from that storing imported operators.
	ImportedSymbols map[string]*Symbol

	// ImportedOperators is the table of operators imported by this file.
	ImportedOperators map[int]*Operator

	// VisiblePackages is the a table of the package that this file imported by
	// name (ie. no imported symbols).
	VisiblePackages map[string]*ChaiPackage
}

// NewFile creates a new Chai file.
func NewFile(parent *ChaiPackage, abspath string) *ChaiFile {
	// calculate module relative file path
	fileRelPath, err := filepath.Rel(parent.Parent.AbsPath, abspath)
	if err != nil {
		log.Fatalf("failed to calculate module relative path to file `%s`: %s\n", abspath, err.Error())
	}

	// calcuate the file context
	ctx := &report.CompilationContext{
		ModName:     parent.Parent.Name,
		ModAbsPath:  parent.Parent.AbsPath,
		FileRelPath: fileRelPath,
	}

	// create the Chai source file
	return &ChaiFile{
		Context:           ctx,
		Parent:            parent,
		Metadata:          make(map[string]string),
		ImportedSymbols:   make(map[string]*Symbol),
		ImportedOperators: make(map[int]*Operator),
		VisiblePackages:   make(map[string]*ChaiPackage),
	}
}

// ImportCollides returns true if an imported name collides with another
// imported name.
func (chFile *ChaiFile) ImportCollides(name string) bool {
	if _, ok := chFile.VisiblePackages[name]; ok {
		return true
	}

	if _, ok := chFile.ImportedSymbols[name]; ok {
		return true
	}

	return false
}

// -----------------------------------------------------------------------------

// ChaiPackage represents a Chai source package.
type ChaiPackage struct {
	// ID is the unique ID of this package.
	ID uint64

	// Name is the package name.
	Name string

	// Parent is the parent module to this package.
	Parent *ChaiModule

	// ModSubPath is sub path to the package (the key that is used to store it
	// in the `SubPackages` dictionary).  This may an empty string if the
	// package is the root package of its parent module.
	ModSubPath string

	// Files is a list of all the Chai source files that belong to this package.
	Files []*ChaiFile

	// SymbolTable is the global symbol table for this package.
	SymbolTable map[string]*Symbol

	// OperatorTable is the global table of operator definitions.
	OperatorTable map[int]*Operator

	// -------------------------------------------------------------------------

	// ImportedPackages is the table of packages imported by files of this
	// package along with a record of which symbols were imported.
	ImportedPackages map[uint64]ChaiPackageImport
}

// ChaiPackageImport details a package that was imported by another package.
type ChaiPackageImport struct {
	Pkg       *ChaiPackage
	Symbols   map[string]*Symbol
	Operators map[int]*Operator
}

// NewPackage creates a new Chai package and adds it to its parent module. It
// will throw an appropriate error if it fails to create the package.
func NewPackage(parentMod *ChaiModule, abspath string) (*ChaiPackage, bool) {
	// calculate the package's module relative path
	modRelPath, err := filepath.Rel(parentMod.AbsPath, abspath)
	if err != nil {
		log.Fatalln("error calculated module-relative path to package: ", err.Error())
	}

	// determine and validate the package name
	pkgName := filepath.Base(abspath)
	if !IsValidIdentifier(pkgName) {
		report.ReportPackageError(
			parentMod.Name,
			fmt.Sprintf(".<%s>", modRelPath),
			fmt.Sprintf("package does not have a valid directory name: `%s`", pkgName),
		)
		return nil, false
	}

	// make the package
	pkg := &ChaiPackage{
		ID:               GenerateIDFromPath(abspath),
		Name:             pkgName,
		Parent:           parentMod,
		SymbolTable:      make(map[string]*Symbol),
		OperatorTable:    make(map[int]*Operator),
		ImportedPackages: make(map[uint64]ChaiPackageImport),
	}

	// add it to its parent module
	if abspath == parentMod.AbsPath {
		// root package
		parentMod.RootPackage = pkg
	} else {
		// sub package
		subPath := "." + strings.ReplaceAll(modRelPath, string(filepath.Separator), ".")
		parentMod.SubPackages[subPath] = pkg
		pkg.ModSubPath = subPath
	}

	return pkg, true
}

// Path returns the full package path string for the package for the purposes
// of error reporting: eg. `io.std` or `core.runtime`.
func (pkg *ChaiPackage) Path() string {
	return pkg.Parent.Name + pkg.ModSubPath
}

// -----------------------------------------------------------------------------

// ChaiModule represents a Chai module.
type ChaiModule struct {
	// ID is the unique ID of this module.
	ID uint64

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
