package cmd

import (
	"chaic/common"
	"chaic/depm"
	"chaic/report"
	"chaic/syntax"
	"chaic/types"
	"chaic/walk"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"unicode"
)

// InitPackage initializes the package with given absolute path.  Initialization
// includes parsing the contents of the package and initializing all of the
// packages that this package imports.  The initialized package is returned if
// initialization is successful.
func (c *Compiler) InitPackage(pkgAbsPath string) (*depm.ChaiPackage, bool) {
	// Create the package and add it to the dependency graph.
	pkg := &depm.ChaiPackage{
		ID:          depm.GetPackageIDFromAbsPath(pkgAbsPath),
		Name:        filepath.Base(pkgAbsPath),
		AbsPath:     pkgAbsPath,
		SymbolTable: make(map[string]*common.Symbol),
		TableMutex:  &sync.Mutex{},
	}
	c.depGraph[pkg.ID] = pkg

	// Validate the package name.
	if !isValidIdentifier(pkg.Name) {
		report.ReportFatal("%s is not a valid package name", pkg.Name)
	}

	// Open the directory.
	finfos, err := ioutil.ReadDir(pkg.AbsPath)
	if err != nil {
		report.ReportFatal("failed to read directory of package %s: %s", pkg.Name, err)
	}

	// Parse all the source files in the package concurrently.
	wg := &sync.WaitGroup{}
	for _, finfo := range finfos {
		// We only want to try to load source files.
		if finfo.IsDir() || filepath.Ext(finfo.Name()) != ".chai" {
			continue
		}

		// Create the Chai source file.
		chFile := &depm.ChaiFile{
			Parent:     pkg,
			FileNumber: len(pkg.Files),
			AbsPath:    filepath.Join(pkg.AbsPath, finfo.Name()),
			ReprPath:   fmt.Sprintf("[%s] %s", pkg.Name, finfo.Name()),
			OpaqueRefs: make(map[string][]*types.OpaqueType),
		}

		// Add it to its parent package.
		pkg.Files = append(pkg.Files, chFile)

		// Parse the file concurrently.
		wg.Add(1)
		go func(chFile *depm.ChaiFile) {
			syntax.ParseFile(chFile)
			wg.Done()
		}(chFile)
	}

	// Wait for parsing to finish.
	wg.Wait()

	// Make sure the package is non-empty.
	if len(pkg.Files) == 0 {
		report.ReportFatal("package must contain source files")
	}

	return pkg, !report.AnyErrors()
}

// ResolveSymbols performs symbol resolution, import resolution, and infinite
// type checking.
func (c *Compiler) ResolveSymbols() bool {
	// Concurrently resolve all opaque types.
	wg := &sync.WaitGroup{}

	allResolved := true
	for _, pkg := range c.depGraph {
		wg.Add(1)

		go func(pkg *depm.ChaiPackage) {
			for _, file := range pkg.Files {
				allResolved = depm.ResolveOpaques(file) && allResolved
			}
		}(pkg)
	}

	// We can't check for infinite types if symbols haven't resolved.
	if !allResolved {
		return false
	}

	// Check for infinite types.
	return depm.CheckForInfiniteTypes(c.depGraph)
}

// WalkPackages performs semantic analysis on packages in the dependency graph.
func (c *Compiler) WalkPackages() bool {
	// Check packages concurrently.
	wg := &sync.WaitGroup{}

	for _, pkg := range c.depGraph {
		wg.Add(1)

		go func(pkg *depm.ChaiPackage) {
			for _, file := range pkg.Files {
				walk.WalkFile(file)
			}

			wg.Done()
		}(pkg)
	}

	wg.Wait()

	return !report.AnyErrors()
}

// -----------------------------------------------------------------------------

// isValidIdentifier returns whether the given name is a valid Chai identifier.
func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	for i, c := range name {
		if i == 0 {
			if unicode.IsLetter(c) || c == '_' {
				continue
			}
		} else if unicode.IsLetter(c) || '0' <= c && c <= '9' || c == '_' {
			continue
		}

		return false
	}

	return true
}
