package depm

import (
	"chaic/ast"
	"chaic/common"
	"chaic/report"
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

				if span, ok := findInfinites(declSym.Type.(types.NamedType)); !ok {
					report.ReportCompileError(
						file.AbsPath,
						file.ReprPath,
						span,
						"infinite type detected",
					)
					noInfTypes = false
				}
			}
		}
	}

	return noInfTypes
}

// findInfinites performs the main infinite type detection algorithm as
// described above starting from the node start.  It returns the text span of
// the root opaque type if an infinite type cycle is detected.
func findInfinites(nt types.NamedType) (*report.TextSpan, bool) {
	switch nt.Color() {
	case types.ColorBlack:
		return nil, true
	case types.ColorGrey:
		nt.SetColor(types.ColorBlack)
		return nil, false
	default: // White
		nt.SetColor(types.ColorGrey)
		span, ok := searchChildren(nt)
		nt.SetColor(types.ColorBlack)
		return span, ok
	}
}

// searchChildren searches the child types of a named type for infinite types.
func searchChildren(nt types.NamedType) (*report.TextSpan, bool) {
	switch v := nt.(type) {
	case *types.StructType:
		for _, field := range v.Fields {
			if span, ok := searchChild(field.Type); !ok {
				return span, false
			}
		}
	}

	return nil, true
}

// searchChild searches a child type of a named type for infinite types.
func searchChild(typ types.Type) (*report.TextSpan, bool) {
	switch v := typ.(type) {
	case *types.OpaqueType:
		if _, ok := findInfinites(v.Value); !ok {
			return v.Span, false
		}
	case *types.TupleType:
		for _, elemType := range v.ElementTypes {
			if span, ok := searchChild(elemType); !ok {
				return span, false
			}
		}
	}

	return nil, true
}
