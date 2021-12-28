package cmd

import (
	"bufio"
	"bytes"
	"chai/common"
	"chai/depm"
	"chai/generate"
	"chai/report"
	"chai/resolve"
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

	// depGraph is the global package dependency graph.  It cannot be organized
	// by module name because module names are not guaranteed by unique within a
	// single compilation.
	depGraph map[uint]*depm.ChaiModule
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
		depGraph:    make(map[uint]*depm.ChaiModule),
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

	// add the root module to the dependency graph
	c.depGraph[rootMod.ID] = rootMod

	// now that the base profile is loading, we can display the compilation
	// header and report the start of analysis.
	report.ReportCompileHeader(
		fmt.Sprintf("%s/%s", c.profile.TargetOS, c.profile.TargetArch),
		rootMod.ShouldCache,
	)

	// initialize the root package (which will initialize all other packages
	// that this project depends on)
	_, ok = c.initPkg(rootMod, rootMod.AbsPath)
	if !ok {
		return false
	}

	// resolve global symbols and check for recursive types
	r := resolve.NewResolver(c.depGraph)
	if !r.Resolve() {
		return false
	}

	// check for operator collisions
	for _, mod := range c.depGraph {
		for _, pkg := range mod.Packages() {
			depm.CheckOperatorCollisions(pkg)
		}
	}

	if !report.ShouldProceed() {
		return false
	}

	// type check expressions and evaluate generics
	c.typeCheck()

	// TODO: prune unused functions

	return report.ShouldProceed()
}

// binPath the path to needed binaries relative to the Chai path.
const binPath = "tools/bin"

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

	// compute path to LLC
	llcPath := filepath.Join(common.ChaiPath, binPath, "llc.exe")

	// create temporary directory to store output files
	tempPath := filepath.Join(filepath.Dir(c.profile.OutputPath), ".chai")
	os.MkdirAll(tempPath, os.ModeDir)

	// write LLVM module(s?) to text file
	modFilePath := filepath.Join(tempPath, "mod.ll")
	writeOutputFile(modFilePath, mod.String())

	// compile LLVM module(s?) using LLC
	objFilePath := filepath.Join(tempPath, "mod.o")
	llc := exec.Command(llcPath, "-filetype", "obj", "-o", objFilePath, modFilePath)
	stderrBuff := bytes.Buffer{}
	llc.Stderr = &stderrBuff

	err := llc.Run()
	if err != nil {
		report.ReportFatal("failed to run llc:\n %s", stderrBuff.String())
	}

	// find the path to the VS developer prompt (to execute link.exe) using
	// `vswhere.exe` which will give us the path to the Visual Studio
	// installation.
	vswherePath := filepath.Join(common.ChaiPath, binPath, "vswhere.exe")
	vswhere := exec.Command(vswherePath, "-latest", "-products", "*", "-property", "installationPath")
	output, err := vswhere.Output()
	if err != nil {
		// TODO: make sure this accurately reports what happens when the tools
		// aren't located
		report.ReportFatal("error running locating VS build tools: %s", string(output))
	}
	vsDevCmdPath := filepath.Join(string(output[:len(output)-2]), "VC/Auxiliary/Build/vcvars64.bat")

	// determine all the objects that need to be linked.  We first link to
	// several import libraries that are used by all applications.  Then we add
	// in the user specified link objects followed by the generated object files
	// of the compiler.
	linkObjects := append([]string{"kernel32.lib", "libcmt.lib"}, c.profile.LinkObjects...)
	linkObjects = append(linkObjects, objFilePath)

	// link objects using `link.exe` executed from the VS developer prompt
	vsDevCmdArgs := []string{
		"&", "link.exe", "/entry:_start", "/subsystem:console", "/nologo",
		fmt.Sprintf("/out:%s", c.profile.OutputPath),
	}
	vsDevCmdArgs = append(vsDevCmdArgs, linkObjects...)
	link := exec.Command(vsDevCmdPath, vsDevCmdArgs...)

	output, err = link.Output()
	if err != nil {
		if output == nil {
			output = []byte(err.Error())
		}
		report.ReportFatal("failed to link program:\n%s", string(output))
	}

	// remove temporary directory
	if err := os.RemoveAll(tempPath); err != nil {
		report.ReportFatal("failed to clean up temporary directory: %s", err.Error())
	}
}

// -----------------------------------------------------------------------------

// initPkg initializes a package: the package is lexed, parsed, and added to the
// dependency graph.  Furthermore, all imports in the package are resolved, but
// symbols are NOT resolved at this stage -- only declared.  It returns the
// package it loads.  If this function fails, it returns false.
func (c *Compiler) initPkg(parentMod *depm.ChaiModule, pkgAbsPath string) (*depm.ChaiPackage, bool) {
	// calculate the package's module relative path
	modRelPath, err := filepath.Rel(parentMod.AbsPath, pkgAbsPath)
	if err != nil {
		report.ReportFatal("error calculated module-relative path to package: ", err.Error())
	}

	// determine and validate the package name
	pkgName := filepath.Base(pkgAbsPath)
	if !depm.IsValidIdentifier(pkgName) {
		report.ReportPackageError(
			parentMod.Name,
			fmt.Sprintf(".<%s>", modRelPath),
			fmt.Sprintf("package does not have a valid directory name: `%s`", pkgName),
		)
		return nil, false
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
		subPath := strings.ReplaceAll(modRelPath, string(filepath.Separator), ".")
		parentMod.SubPackages[subPath] = pkg
		pkg.ModSubPath = subPath
	}

	// list the elements of the package directory
	finfos, err := ioutil.ReadDir(pkgAbsPath)
	if err != nil {
		report.ReportPackageError(parentMod.Name, pkg.ModSubPath, "failed to read directory of package: "+err.Error())
		return nil, false
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
				report.ReportFatal("failed to calculate module relative path to file `%s`: %s", fileAbsPath, err.Error())
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
				report.ReportFatal("failed to open source file at `%s`: %s", fileRelPath, err.Error())
			}
			defer file.Close()

			r := bufio.NewReader(file)

			// create the parser for the file
			p := syntax.NewParser(c.importPackgage, chFile, r)

			// parse the file and determine if it should be added
			if p.Parse() {
				pkg.Files = append(pkg.Files, chFile)
			}
		}
	}

	// make sure the package is not empty if there were no other errors
	if report.ShouldProceed() {
		if len(pkg.Files) == 0 {
			report.ReportPackageError(parentMod.Name, pkg.ModSubPath, "package contains no compileable source files")
			return nil, false
		}
	}

	return pkg, true
}

// typeCheck types checks each file in each package of the project and
// determines all the generic instances to be used in generic evaluation.  It
// then evaluates all generics.
func (c *Compiler) typeCheck() {
	for _, mod := range c.depGraph {
		for _, pkg := range mod.Packages() {
			for _, file := range pkg.Files {
				w := walk.NewWalker(file)

				for _, def := range file.Defs {
					w.WalkDef(def)
				}
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
