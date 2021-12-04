package cmd

import (
	"bufio"
	"chai/common"
	"chai/depm"
	"chai/generate"
	"chai/report"
	"chai/syntax"
	"chai/walk"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compiler represents the global state of the compiler.
type Compiler struct {
	// rootAbsPath is the absolute path to the compilation root.
	rootAbsPath string

	// rootModule is the root module of the project being compiled.
	rootModule *depm.ChaiModule

	// profile is current build profile of the compiler.
	profile *BuildProfile
}

// NewCompiler creates a new compiler.
func NewCompiler(rootRelPath string, profile *BuildProfile) *Compiler {
	// calculate the absolute path to the compilation root.
	rootAbsPath, err := filepath.Abs(rootRelPath)
	if err != nil {
		report.ReportFatal("error calculating absolute path: %s", err.Error())
		return nil
	}

	return &Compiler{
		rootAbsPath: rootAbsPath,
		profile:     profile,
	}
}

// Analyze runs the analysis phase of the compiler.
func (c *Compiler) Analyze() bool {
	// load the root module
	rootMod, ok := depm.LoadModule(c.rootAbsPath)
	if !ok {
		return false
	}
	c.rootModule = rootMod

	// now that the base profile is loading, we can display the compilation
	// header and report the start of analysis.
	report.ReportCompileHeader(
		fmt.Sprintf("%s/%s", c.profile.TargetOS, c.profile.TargetArch),
		rootMod.ShouldCache,
	)

	// initialize the root package (which will initialize all other packages
	// that this project depends on)
	c.initPkg(rootMod, rootMod.AbsPath)

	// TODO: resolve global symbols and check for recursive types

	// check for operator collisions
	// TODO: check every module in the dep graph
	for _, pkg := range rootMod.Packages() {
		depm.CheckOperatorCollisions(pkg)
	}

	if !report.ShouldProceed() {
		return false
	}

	// type check expressions and evaluate generics
	c.typeCheck()

	// TODO: prune unused functions

	return report.ShouldProceed()
}

// llvmBinPath the path to needed LLVM binaries relative to the Chai path.
const llvmBinPath = "vendor/bin"

// Generate runs the generation, LLVM, and linking phases of the compiler. The
// Analysis phase must be run before this.
func (c *Compiler) Generate() {
	// TODO: decide whether or not refactor to produce a single LLVM module
	// instead of multiple: don't want to have to make more calls to `llc` or
	// create more work for the linker than necessary.  Or is it better to
	// produce multiple LLVM modules, compile them, and link them all at the
	// end?

	// generate the main LLVM module
	// TODO: generate whole dependency graph
	g := generate.NewGenerator(c.rootModule.RootPackage)
	mod := g.Generate()

	// compute path to LLC and LLD
	llcPath := filepath.Join(common.ChaiPath, llvmBinPath, "llc.exe")
	lldPath := filepath.Join(common.ChaiPath, llvmBinPath, "lld-link.exe")

	// create temporary directory to store output files
	tempPath := filepath.Join(filepath.Dir(c.profile.OutputPath), ".chai")
	os.MkdirAll(tempPath, os.ModeDir)

	// write LLVM module(s?) to text file
	modFilePath := filepath.Join(tempPath, "mod.ll")
	writeOutputFile(modFilePath, mod.String())

	// compile LLVM module(s?) using LLC
	objFilePath := filepath.Join(tempPath, "mod.o")
	llc := exec.Command(llcPath, "-filetype", "obj", "-o", objFilePath, modFilePath)
	llc.Stderr = os.Stdout

	err := llc.Run()
	if err != nil {
		report.ReportFatal("failed to run llc: %s", err.Error())
	}

	// because Windows is a pain, we have to get the path to the Windows SDK
	// in order to be able to generate target code

	// link objects using `lld`
	// TODO: add in mechanism to link in other objects (passed to profile)
	lld := exec.Command(
		lldPath, "/entry:_start", "/subsystem:console",
		fmt.Sprintf("/out:%s", c.profile.OutputPath), objFilePath,
	)
	lld.Stderr = os.Stdout

	err = lld.Run()
	if err != nil {
		report.ReportFatal("failed to link program: %s", err.Error())
	}

	// remove temporary directory
	if err := os.RemoveAll(tempPath); err != nil {
		report.ReportFatal("failed to clean up temporary directory: %s", err.Error())
	}
}

// -----------------------------------------------------------------------------

// initPkg initializes a package: the package is lexed, parsed, and added to the
// dependency graph.  Furthermore, all imports in the package are resolved, but
// symbols are NOT resolved at this stage -- only declared.
func (c *Compiler) initPkg(parentMod *depm.ChaiModule, pkgAbsPath string) {
	// determine and validate the package name
	pkgName := filepath.Base(pkgAbsPath)
	if !depm.IsValidIdentifier(pkgName) {
		report.ReportFatal("package at %s does not have a valid directory name: `%s`", pkgAbsPath, pkgName)
	}

	// create the package struct.
	pkgID := depm.GenerateIDFromPath(pkgAbsPath)
	pkg := &depm.ChaiPackage{
		ID:            pkgID,
		Name:          pkgName,
		Parent:        parentMod,
		SymbolTable:   make(map[string]*depm.Symbol),
		OperatorTable: make(map[int]*depm.Operator),
	}

	// add it to its parent module
	if pkgAbsPath == parentMod.AbsPath {
		// root package
		parentMod.RootPackage = pkg
	} else {
		// sub package
		pkgRelPath, err := filepath.Rel(parentMod.AbsPath, pkgAbsPath)
		if err != nil {
			report.ReportFatal("error computing package relative path: %s", err.Error())
		}

		subPath := strings.ReplaceAll(pkgRelPath, string(filepath.Separator), ".")
		parentMod.SubPackages[subPath] = pkg
		pkg.ModSubPath = subPath
	}

	// TODO: add package to dependency graph (before parsing to prevent import
	// errors)

	// list the elements of the package directory
	finfos, err := ioutil.ReadDir(pkgAbsPath)
	if err != nil {
		report.ReportFatal("[%s] failed to read directory of package `%s`: %s", parentMod.Name, pkgName, err.Error())
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
				report.ReportFatal("failed to calculate module relative path to file %s: %s", fileAbsPath, err.Error())
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
				report.ReportFatal("[%s] failed to open source file at %s: %s", parentMod.Name, fileRelPath, err.Error())
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
			report.ReportFatal("[%s] package `%s` contains no compileable source files", parentMod.Name, pkg.Name)
		}
	}
}

// typeCheck types checks each file in each package of the project and
// determines all the generic instances to be used in generic evaluation.  It
// then evaluates all generics.
func (c *Compiler) typeCheck() {
	// TODO: traverse dep graph
	for _, pkg := range c.rootModule.Packages() {
		for _, file := range pkg.Files {
			w := walk.NewWalker(file)

			for _, def := range file.Defs {
				w.WalkDef(def)
			}
		}
	}
}

// -----------------------------------------------------------------------------

// writeOutputFile is used to quickly write an output file for the compiler.
func writeOutputFile(fpath, content string) {
	// open or create the file
	file, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		report.ReportFatal("failed to open output file `%s`: %s", fpath, err.Error())
	}
	defer file.Close()

	// write the data
	_, err = file.WriteString(content)
	if err != nil {
		report.ReportFatal("failed to write output to file `%s`: %s", fpath, err.Error())
	}
}
