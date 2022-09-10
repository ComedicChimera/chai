package depm

import (
	"chaic/ast"
	"chaic/common"
	"chaic/types"
)

/*
Infinite Type Checking
----------------------

Here follows the algorithm used to perform infinite type checking:

This algorithm is a slight modification of the Three-Color DFS algorithm.

All named types have one of three colors associated with them:

1. White
2. Grey
3. Black

All named types start with the color white.  A depth-first search of the type
dependency graph is then performed by following all opaque references contained
inside named types if the reference is not indirect.

When each node is visited, it is initially assigned the color grey.  Then, all
its child nodes are visited in a depth-first fashion.  Once this has completed,
the color the node is then updated to black.

Before visiting a child node, the algorithm checks the nodes color.  If the node
is White, the node will be visited.  If the node is grey, then a cycle has been
found: the node is marked Black and false is returned.  If the node is black,
the node is not visited.

The algorithm returns false if the recursive calls on any of the child nodes
return false.  The algorithm will be run with each declared named type as the
start node provided that node is not already colored black.  If the node is
colored black, then it has already been visited and all of its children have as
well.

Note that the algorithm will stop traversal if any black node is detected.
*/

// CheckForInfiniteTypes checks all declared types to make sure none are
// infinite.  It returns false if any infinite types are detected.
func CheckForInfiniteTypes(depGraph map[uint64]*ChaiPackage) bool {
	noInfTypes := true

	for _, pkg := range depGraph {
		for _, file := range pkg.Files {
			for _, def := range file.Definitions {
				var declSym *common.Symbol
				switch v := def.(type) {
				case *ast.StructDef:
					declSym = v.Symbol
				}

				if !searchFrom(declSym.Type.(types.NamedType)) {
					// TODO: figure out how to get the position out
					// report.ReportCompileError(
					// 	file.AbsPath,
					// 	file.ReprPath,
					// 	declSym.DefSpan,
					// )
					noInfTypes = false
				}
			}
		}
	}

	return noInfTypes
}

// searchFrom performs the main infinite type detection algorithm as described
// above starting from the node start.  This function does NOT assume that the
// node is white or grey.
func searchFrom(nt types.NamedType) bool {
	switch nt.Color() {
	case types.ColorBlack:
		return true
	case types.ColorGrey:
		nt.SetColor(types.ColorBlack)
		return false
	default: // White
		nt.SetColor(types.ColorGrey)
		result := searchChildren(nt)
		nt.SetColor(types.ColorBlack)
		return result
	}
}

// searchChildren searches all the child named types of a type.
func searchChildren(typ types.Type) bool {
	// Pointers and primitives aren't searched because they can't count as
	// direct references.
	switch v := typ.(type) {
	case *types.StructType:
		for _, field := range v.Fields {
			if !searchChildren(types.InnerType(field.Type)) {
				return false
			}
		}
	case *types.TupleType:
		for _, elemType := range v.ElementTypes {
			if !searchChildren(types.InnerType(elemType)) {
				return false
			}
		}
	}

	return true
}
