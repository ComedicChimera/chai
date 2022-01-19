package cmd

import (
	"bufio"
	"bytes"
	"chai/common"
	"chai/depm"
	"chai/generate"
	"chai/report"
	"chai/syntax"
	"chai/typing"
	"chai/walk"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/llir/llvm/ir"
)

// Compiler represents the global state of the compiler.
type Compiler struct {
	// rootAbsPath is the absolute path to the compilation root.
	rootAbsPath string

	// rootModule is the root module of the project being compiled.
	rootModule *depm.ChaiModule

	// profile is current build profile of the compiler.
	profile *BuildProfile

	// depGraph is the global dependency graph: a table of modules organized by
	// ID.  It is used to locate packages by their package path (ie. resolve
	// imports).  It cannot be organized by module name because module names are
	// not guaranteed by unique within a single project.
	depGraph map[uint64]*depm.ChaiModule

	// pkgList is an unorganized list of all the individual packages to be
	// compiled into the resulting binary output.  While it is possible to
	// determine a listing a packages using the dependency graph, doing so
	// requires a non-trivial amount of computation.  Since this operation is
	// performed many times during compilation, the admittedly slight
	// performance boost and added convenience of this auxilliary list is well
	// worth the small memory cost.
	pkgList []*depm.ChaiPackage

	// uni is the shared universe for the project.
	uni *depm.Universe
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
		depGraph:    make(map[uint64]*depm.ChaiModule),
		uni:         depm.NewUniverse(),
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

	// add the prelude packages and imports to the project
	c.addPrelude()

	// resolve global symbols and check for recursive types
	r := depm.NewResolver(c.pkgList)
	if !r.Resolve() {
		return false
	}

	// check for operator collisions
	for _, pkg := range c.pkgList {
		depm.CheckOperatorCollisions(pkg)
	}

	if !report.ShouldProceed() {
		return false
	}

	// type check expressions and evaluate generics
	c.typeCheck()

	// TODO: prune unused functions

	// check for main function
	if mainSym, ok := c.rootModule.RootPackage.SymbolTable["main"]; ok && mainSym.DefKind == depm.DKFuncDef {
		if mainFt, ok := mainSym.Type.(*typing.FuncType); ok {
			if mainFt.ReturnType.Equiv(typing.NothingType()) && len(mainFt.Args) == 0 && mainFt.IntrinsicName == "" {
				return report.ShouldProceed()
			}
		}

		report.ReportFatal("main function signature must be of the form: `() -> ()`")
	} else {
		report.ReportFatal("main package missing main function")
	}

	return false
}

// binPath the path to needed binaries relative to the Chai path.
const binPath = "tools/bin"

// Generate runs the generation, LLVM, and linking phases of the compiler. The
// Analysis phase must be run before this.
func (c *Compiler) Generate() {
	// compute path to LLC
	llcPath := filepath.Join(common.ChaiPath, binPath, "llc.exe")

	// create temporary directory to store output files
	tempPath := filepath.Join(filepath.Dir(c.profile.OutputPath), ".chai")
	os.MkdirAll(tempPath, os.ModeDir)

	// create the global start builder
	sb := generate.NewStartBuilder(len(c.pkgList))

	// compile each package in the dependency graph into an object file
	// concurrently and write the object file paths to a channel to they can be
	// linked together
	objFilePathCh := make(chan string)
	nObjFiles := 0

	for _, pkg := range c.pkgList {
		nObjFiles += 1

		go func(pkg *depm.ChaiPackage) {
			// generate the LLVM module
			g := generate.NewGenerator(sb, pkg, pkg.ID == c.rootModule.RootPackage.ID)
			llMod := g.Generate()

			// generate the object file
			objFilePath := filepath.Join(tempPath, fmt.Sprintf("pkg%d.o", pkg.ID))
			if err := compileLLVMModule(llcPath, llMod, objFilePath); err != nil {
				report.ReportFatal("failed to run llc on `%s`:\n %s", pkg.Path(), err.Error())
			}

			// write the object file path and mark the goroutine as finished
			objFilePathCh <- objFilePath
		}(pkg)
	}

	// find the path to the VS developer prompt (to execute link.exe) using
	// `vswhere.exe` which will give us the path to the Visual Studio
	// installation while other packages are compileable
	vswherePath := filepath.Join(common.ChaiPath, binPath, "vswhere.exe")
	vswhere := exec.Command(vswherePath, "-latest", "-products", "*", "-property", "installationPath")
	output, err := vswhere.Output()
	if err != nil {
		// TODO: make sure this accurately reports what happens when the tools
		// aren't located
		report.ReportFatal("error running locating VS build tools: %s", string(output))
	}
	vsDevCmdPath := filepath.Join(string(output[:len(output)-2]), "VC/Auxiliary/Build/vcvars64.bat")

	// collect the object file paths into slice (this also ensures we don't
	// proceed until compilation is done)

	var objFilePaths []string
	for i := 0; i < nObjFiles; i++ {
		objFilePaths = append(objFilePaths, <-objFilePathCh)
	}

	close(objFilePathCh)

	// generate and compile the initialization module
	sb.BuildMainInitFunc()
	initObjPath := filepath.Join(tempPath, "init.o")
	if err := compileLLVMModule(llcPath, sb.GetInitMod(), initObjPath); err != nil {
		report.ReportFatal("failed to run llc on the init module:\n %s", err.Error())
	}
	objFilePaths = append(objFilePaths, initObjPath)

	// determine all the objects that need to be linked.  We first link to
	// several import libraries that are used by all applications.  Then we add
	// in the user specified link objects followed by the generated object files
	// of the compiler.
	linkObjects := append([]string{"kernel32.lib", "libcmt.lib"}, c.profile.LinkObjects...)
	linkObjects = append(linkObjects, objFilePaths...)

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
	// create a new package in the parent module
	pkg, ok := depm.NewPackage(parentMod, pkgAbsPath)
	if !ok {
		return nil, false
	}

	// add the package to the package list
	c.pkgList = append(c.pkgList, pkg)

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

			// create the Chai source file
			chFile := depm.NewFile(pkg, fileAbsPath)

			// open the file and create the reader for it
			file, err := os.Open(fileAbsPath)
			if err != nil {
				report.ReportFatal("failed to open source file at `%s`: %s", chFile.Context.FileRelPath, err.Error())
			}
			defer file.Close()

			r := bufio.NewReader(file)

			// create the parser for the file
			p := syntax.NewParser(c.uni, c.importPackgage, chFile, r)

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
	for _, pkg := range c.pkgList {
		for _, file := range pkg.Files {
			w := walk.NewWalker(c.uni, file)

			for _, def := range file.Defs {
				w.WalkDef(def)
			}
		}
	}
}

// -----------------------------------------------------------------------------

// compileLLVMModule takes an LLVM module and an output path and attempts to
// compile it to an object file using LLC.  It returns an error if it fails.
func compileLLVMModule(llcPath string, mod *ir.Module, objFilePath string) error {
	// write LLVM module to text file
	modFilePath := objFilePath[:len(objFilePath)-2] + ".ll"
	writeOutputFile(modFilePath, mod.String())

	// compile LLVM module using LLC
	llc := exec.Command(llcPath, "-filetype", "obj", "-o", objFilePath, modFilePath)
	stderrBuff := bytes.Buffer{}
	llc.Stderr = &stderrBuff

	err := llc.Run()
	if err != nil {
		return errors.New(stderrBuff.String())
	}

	return nil
}

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
