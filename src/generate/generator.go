package generate

import (
	"chai/logging"
	"chai/sem"
	"fmt"
	"io/ioutil"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// NOTE: Because the LLVM bindings for Go were unable to even be installed.
// Worse still, once installed, after over a month of banging my head into the
// wall, I was unable to get a single build using them to compile on Windows:
// missing libraries, bad configs, and a whole other nightmare of seemingly
// unsolveable build errors plagued the process.  I am unsure as to whether or
// not this was the fault of the bindings themselves or CGO, but regardless, an
// alternate strategy was needed.
// So, I chose to use the library `llir/llvm` which provides all the
// infrastructure necessary to generate LLVM IR source text.  I can then pass
// these source files to `opt` and `llc` to generate optimized assembler and
// feed that resulting assembly into the native assembler. It is not as
// efficient, but it works -- I am sick of trying to figure out/fix other
// people's broken sh!t so this is where we are now.

// Generator is responsible for converting each LLVM package into an LLVM
// module.  It can be run concurrently; however, it should be denoted that
// running it does open a file handle briefly (at the end of its execution) so
// caution should be exercised with respect to how many handles are being
// opened.
type Generator struct {
	// srcPkg is the Chai source package being compiled
	srcPkg *sem.ChaiPackage

	// llModule is the LLVM module being built by this generator
	llModule *ir.Module

	// globalValues is a map of all the globally defined values
	globalValues map[string]*value.Value

	// globalTypes is a map of all the globally defined types
	globalTypes map[string]*types.Type
}

// LLVMSymbol represents a symbol declared in an LLVM module
type LLVMSymbol struct {
}

// NewGenerator creates a new generator
func NewGenerator(pkg *sem.ChaiPackage) *Generator {
	return &Generator{
		srcPkg:       pkg,
		llModule:     ir.NewModule(),
		globalValues: make(map[string]*value.Value),
		globalTypes:  make(map[string]*types.Type),
	}
}

// Generate generates a new LLVM module from the generator's package and writes
// the source text to a temporary file, returns the path to that temporary file,
// and a boolean indicating success or failure
func (g *Generator) Generate(targetOS, targetArch string) (string, bool) {
	// TODO: generate the package

	// write the module to a temp file
	file, err := ioutil.TempFile("chai-build", "chai-llout.*.ll")
	if err != nil {
		logging.LogFatal(fmt.Sprintf("failed to create temporary file for LLVM IR generation: %s", err.Error()))
		return "", false
	}

	// write the module source to the file
	if _, err = g.llModule.WriteTo(file); err != nil {
		logging.LogFatal(fmt.Sprintf("failed to write to temporary file during LLVM IR generation: %s", err.Error()))
		return "", false
	}

	// return the name of the temporary file
	return file.Name(), true
}
