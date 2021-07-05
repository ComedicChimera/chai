package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"errors"
)

// WalkIdentifierList walks an `identifier_list` node and returns a slice of the
// identifiers collected along with an error value indicating whether or not
// multiple of the same identifier value were encountered (the error message is
// the name encountered multiple times).  It also returns a map of all the
// identifiers and their corresponding positions.
func WalkIdentifierList(idBranch *syntax.ASTBranch) ([]string, map[string]*logging.TextPosition, error) {
	encountered := make(map[string]*logging.TextPosition)
	idList := make([]string, idBranch.Len()/2+1)

	for i, item := range idBranch.Content {
		if i%2 == 0 {
			name := item.(*syntax.ASTLeaf).Value

			if _, ok := encountered[name]; ok {
				return nil, nil, errors.New(name)
			}

			encountered[name] = item.Position()
			idList[i/2] = name
		}
	}

	return idList, encountered, nil
}

// -----------------------------------------------------------------------------

// nothingType returns a new nothing type
func nothingType() typing.DataType {
	return typing.PrimType(typing.PrimKindNothing)
}

// boolType returns a new boolean type
func boolType() typing.DataType {
	return typing.PrimType(typing.PrimKindBool)
}

// convertIncompleteToBranch converts a HIRExpr to an AST Branch assuming the
// HIRExpr is a HIRIncomplete
func convertIncompleteToBranch(expr sem.HIRExpr) *syntax.ASTBranch {
	return (*syntax.ASTBranch)(expr.(*sem.HIRIncomplete))
}
