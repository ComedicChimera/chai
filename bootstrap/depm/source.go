package depm

import (
	"chaic/ast"
	"chaic/common"
	"hash/fnv"
	"sync"
)

// ChaiFile represents a Chai source file.
type ChaiFile struct {
	// The parent package to this Chai file.
	Parent *ChaiPackage

	// The identifying number of this file within its parent package.
	FileNumber int

	// The absolute path to this Chai file.
	AbsPath string

	// The representative path to this Chai file.  This is the path
	// that should be displayed to the user to identify the file.
	ReprPath string

	// The list of definitions contained in this Chai file.
	Definitions []ast.ASTNode
}

// -----------------------------------------------------------------------------

// ChaiPackage represents a Chai package: a collection of source files which
// share a common global namespace and are compiled together a single,
// standalone translation unit.
type ChaiPackage struct {
	// The unique ID of the package.
	ID int

	// The name of the package.
	Name string

	// The absolute path to the package directory.
	AbsPath string

	// The root package relative to this package: defined by its package path.
	RootPkg *ChaiPackage

	// The list of source files contained in the package.
	Files []*ChaiFile

	// The global symbol table shared between the package's source files.
	SymbolTable map[string]*common.Symbol

	// The global operator table shared between the package's source files.
	OperatorTable map[int][]*common.Operator

	// The global mutex used to synchronize access to the package's tables.
	TableMutex *sync.Mutex
}

// GetPackageIDFromAbsPath returns a package ID based on an absolute path.
func GetPackageIDFromAbsPath(absPath string) int {
	a := fnv.New64a()
	a.Write([]byte(absPath))
	return int(a.Sum64())
}
