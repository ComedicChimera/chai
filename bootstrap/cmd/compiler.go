package cmd

import (
	"bufio"
	"chai/common"
	"chai/depm"
	"chai/report"
	"chai/syntax"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Compiler represents the global state of the compiler.
type Compiler struct {
	// rootAbsPath is the absolute path to the compilation root.
	rootAbsPath string

	// rootModule is the root module of the project being compiled.
	rootModule *depm.ChaiModule

	// baseProfile is the base profile of the project.
	baseProfile *depm.BuildProfile
}

// NewCompiler creates a new compiler.
func NewCompiler(rootRelPath string) *Compiler {
	// calculate the absolute path to the compilation root.
	rootAbsPath, err := filepath.Abs(rootRelPath)
	if err != nil {
		report.ReportFatal("error calculating absolute path: " + err.Error())
		return nil
	}

	return &Compiler{
		rootAbsPath: rootAbsPath,
		baseProfile: &depm.BuildProfile{},
	}
}

// Analyze runs the analysis phase of the compiler.
func (c *Compiler) Analyze() bool {
	// load the root module
	rootMod, ok := depm.LoadModule(c.rootAbsPath, "", c.baseProfile)
	if !ok {
		return false
	}
	c.rootModule = rootMod

	// now that the base profile is loading, we can display the compilation
	// header and report the start of analysis.
	report.ReportCompileHeader(
		fmt.Sprintf("%s/%s", c.baseProfile.TargetOS, c.baseProfile.TargetArch),
		rootMod.ShouldCache,
	)
	report.ReportBeginPhase("Analyzing...")

	// initialize the root package
	c.initPkg(rootMod, rootMod.AbsPath)

	// TODO: check for recursive types

	// TODO: type checking and generic evaluation

	// if we reach here, we can end report the end of the analysis phase.
	report.ReportEndPhase()
	return report.ShouldProceed()
}

// Generate runs the generation, LLVM, and linking phases of the compiler. The
// Analysis phase must be run before this.
func (c *Compiler) Generate() {
	// TODO
}

// -----------------------------------------------------------------------------

// initPkg initializes a package: the package is lexed, parsed, and added to the
// dependency graph.  Furthermore, since parsing implies partial semantic
// analysis, symbol resolution, and import resolution, these steps are also
// performed as part of initialization.  Type checking and generic evaluation
// are not performed in this step.
func (c *Compiler) initPkg(parentMod *depm.ChaiModule, pkgAbsPath string) {
	// determine and validate the package name
	pkgName := filepath.Base(pkgAbsPath)
	if !depm.IsValidIdentifier(pkgName) {
		report.ReportFatal(fmt.Sprintf("package at %s does not have a valid directory name: `%s`", pkgAbsPath, pkgName))
	}

	// create the package struct.
	pkg := &depm.ChaiPackage{
		ID:          depm.GenerateIDFromPath(pkgAbsPath),
		Name:        pkgName,
		Parent:      parentMod,
		GlobalTable: &depm.SymbolTable{},
	}

	// TODO: add package to dependency graph (before parsing to prevent import
	// errors)

	// list the elements of the package directory
	finfos, err := ioutil.ReadDir(pkgAbsPath)
	if err != nil {
		report.ReportFatal(fmt.Sprintf("[%s] failed to read directory of package `%s`: %s", parentMod.Name, pkgName, err.Error()))
	}

	// walk through the files and parse them all
	for _, finfo := range finfos {
		// select only files that are source files
		if !finfo.IsDir() && filepath.Ext(finfo.Name()) == common.ChaiFileExt {
			// calulate the file abs path
			fileAbsPath := filepath.Join(pkgAbsPath, finfo.Name())

			// calculate module relative file path
			fileRelPath, err := filepath.Rel(parentMod.AbsPath, fileAbsPath)
			if err != nil {
				report.ReportFatal(fmt.Sprintf("failed to calculate module relative path to file %s: %s", fileAbsPath, err.Error()))
			}

			// calcuate the file context
			ctx := &report.CompilationContext{
				ModName:     parentMod.Name,
				ModAbsPath:  parentMod.AbsPath,
				FileRelPath: fileRelPath,
			}

			// create the Chai source file
			chFile := &depm.ChaiFile{
				Context:  ctx,
				Parent:   pkg,
				Metadata: make(map[string]string),
			}

			// open the file and create the reader for it
			file, err := os.Open(fileAbsPath)
			if err != nil {
				report.ReportFatal(fmt.Sprintf("[%s] failed to open source file at %s: %s", parentMod.Name, fileRelPath, err.Error()))
			}
			defer file.Close()

			r := bufio.NewReader(file)

			// create the parser for the file
			p := syntax.NewParser(chFile, r)

			// parse the file and determine if it should be added
			if p.Parse() {
				pkg.Files = append(pkg.Files, chFile)
			}
		}
	}

	// make sure the package is not empty if there were no other errors
	if report.ShouldProceed() {
		if len(pkg.Files) == 0 {
			report.ReportFatal(fmt.Sprintf("[%s] package `%s` contains no compileable source files", parentMod.Name, pkg.Name))
		}
	}
}
