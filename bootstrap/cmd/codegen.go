package cmd

import (
	"chaic/codegen"
	"chaic/depm"
	"chaic/llc"
	"chaic/report"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// CodeGen generates the appropriate compilation output.
func (c *Compiler) CodeGen() {
	if c.outputMode == OutModeLLVM {
		c.generateLLVMIRModules()
	} else {
		outputFiles := c.generateASMOrObjectFiles()

		switch c.outputMode {
		case OutModeExecutable:
			c.linkExecutable(outputFiles)
		case OutModeStaticLib:
		case OutModeDynamicLib:
		}
	}
}

// generateLLVMIRModules generates all packages to the LLVM IR modules.
func (c *Compiler) generateLLVMIRModules() {
	// // Create the output path for the LLVM modules.
	if err := os.MkdirAll(c.outputPath, fs.ModeDir); err != nil {
		report.ReportFatal("failed to create output directory: %s", err)
	}

	// Generate and output the modules concurrently.
	wg := &sync.WaitGroup{}
	for _, pkg := range c.depGraph {
		ctx := llc.NewContext()
		wg.Add(1)

		go func(ctx *llc.Context, pkg *depm.ChaiPackage) {
			defer wg.Done()

			// Generate the LLVM module.
			mod := codegen.Generate(ctx, pkg)

			// Output it to a file.
			if err := mod.WriteToFile(filepath.Join(c.outputPath, fmt.Sprintf("%s_%d.ll", pkg.Name, pkg.ID))); err != nil {
				report.ReportFatal("failed to output module to file: %s", err)
			}
		}(ctx, pkg)
	}

	wg.Wait()
}

// generateASMOrObjectFiles generates all packages to ASM files or object files
// depending on the output mode.  It returns the paths to the output files it
// produces.
func (c *Compiler) generateASMOrObjectFiles() []string {
	// Determine and if necessary create the output path that should be used for
	// this stage of compilation: if this is the final stage, this path is the
	// specified/ default output path.  Otherwise it is the root directory. This
	// is also when we indicate to the compiler to produce ASM or object files
	// at the LLVM toolchain level.
	var stageOutputPath string
	stageOutputExt := ".o"
	codegenFileType := llc.ObjectFile
	if c.outputMode == OutModeASM || c.outputMode == OutModeObj {
		if err := os.Mkdir(c.outputPath, fs.ModeDir); err != nil {
			report.ReportFatal("failed to create output directory: %s", err)
		}

		if c.outputMode == OutModeASM {
			codegenFileType = llc.AssemblyFile
			stageOutputExt = ".s"
		}

		stageOutputPath = c.outputPath
	} else {
		stageOutputPath = c.rootPath
	}

	// Get the target to use for code generation.
	triple := llc.HostTriple()
	target, ok := llc.GetTargetFromTriple(triple)
	if !ok {
		report.ReportFatal("unable to find code generator target: %s", triple)
	}

	// Create the target machine for code generation.
	gctx := llc.NewContext()
	defer gctx.Dispose()

	tm := gctx.NewHostMachine(target, llc.CodeGenLevelDefault, llc.RelocDefault, llc.CodeModelDefault)

	// Generate all of the packages to object files concurrently.
	objFileChan := make(chan string)
	for _, pkg := range c.depGraph {
		// Create the context for the generator.
		ctx := llc.NewContext()

		// Create the object file.
		go func(ctx *llc.Context, pkg *depm.ChaiPackage) {
			// Generate the module.
			mod := codegen.Generate(ctx, pkg)

			// Calculate the object file output path.
			outputPath := filepath.Join(stageOutputPath, fmt.Sprintf("%s_%d%s", pkg.Name, pkg.ID, stageOutputExt))

			// Emit the object file.
			if err := tm.CompileModule(mod, outputPath, codegenFileType); err != nil {
				report.ReportFatal("failed to compile module: %s", err)
			}
		}(ctx, pkg)
	}

	// Collect the object files into an array.
	objFilePaths := make([]string, len(c.depGraph))
	for i := 0; i < len(c.depGraph); i++ {
		objFilePaths[i] = <-objFileChan
	}

	return objFilePaths
}
