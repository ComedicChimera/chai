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

	// tempPath is the path to output temporary files to during compilation
	// (typically `.chai`).
	tempPath string
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
		tempPath:    filepath.Join(filepath.Dir(profile.OutputPath), ".chai"),
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

// vsDevPrompt is the path to the script to execute commands from the Windows developer prompt
const vsDevPrompt = "C:\\Program Files (x86)\\Microsoft Visual Studio\\2019\\BuildTools\\VC\\Auxiliary\\Build\\vcvars64.bat"

// Generate runs the generation, LLVM, and linking phases of the compiler. The
// Analysis phase must be run before this.
func (c *Compiler) Generate() {
	// TODO: refactor to produce a single LLVM module instead of multiple: don't
	// want to have to make more calls to `llc` or create more work for the
	// linker than necessary.

	// generate LLVM modules
	// TODO: use depgraph
	var modules []*ir.Module
	for _, pkg := range c.rootModule.Packages() {
		g := generate.NewGenerator(pkg)
		modules = append(modules, g.Generate())
	}

	// DEBUG: write LLVM modules to text file
	for i, mod := range modules {
		c.writeOutputFile(fmt.Sprintf("mod%d.ll", i), mod.String())
	}

	// TEMPORARY: compile LLVM modules
	var objFilePaths []string
	for i := range modules {
		objFilePath := filepath.Join(c.tempPath, fmt.Sprintf("mod%d.o", i))
		llc := exec.Command("llc", "-filetype", "obj", "-o",
			objFilePath,
			filepath.Join(c.tempPath, fmt.Sprintf("mod%d.ll", i)),
		)
		output, err := llc.Output()
		if err != nil {
			fmt.Println("[LLC]:\n", string(output))
			report.ReportFatal("failed to run llc: %s", err.Error())
		}

		objFilePaths = append(objFilePaths, objFilePath)
	}

	// TEMPORARY: link output executable
	args := append([]string{"&", "link", "/entry:_start", "/subsystem:console", fmt.Sprintf("/out:%s", c.profile.OutputPath), "kernel32.lib"}, objFilePaths...)
	link := exec.Command(vsDevPrompt, args...)
	output, err := link.Output()
	if err != nil {
		fmt.Println("[LINK]:\n", string(output))
		report.ReportFatal("failed to run link.exe: %s", err.Error())
	}

	// delete the .chai directory
	if err := os.RemoveAll(c.tempPath); err != nil {
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

// writeOutputFile is used to write a temporary compiler output to the file
// system. The output name of the file being written is provided relative to the
// the temporary output directory (normally .chai).
func (c *Compiler) writeOutputFile(fileOutRelPath string, fileText string) {
	// determine actual output path and create all enclosing directories
	fileOutPath := filepath.Join(c.tempPath, fileOutRelPath)

	err := os.MkdirAll(filepath.Dir(fileOutPath), os.ModeDir)
	if err != nil {
		report.ReportFatal("failed to create directories: %s\n", err.Error())
	}

	// open or create the file
	file, err := os.OpenFile(fileOutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		report.ReportFatal("failed to open output file: %s\n", err.Error())
	}
	defer file.Close()

	// write the data
	_, err = file.WriteString(fileText)
	if err != nil {
		report.ReportFatal(fmt.Sprintf("failed to write output: %s\n", err.Error()))
	}
}
