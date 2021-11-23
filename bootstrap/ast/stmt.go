package ast

import (
	"chai/report"
	"chai/typing"
)

// VarDecl is a variable declaration statement.
type VarDecl struct {
	ExprBase
	DefBase

	VarLists []*VarList
	Pos      *report.TextPosition
}

func (vd *VarDecl) Names() []string {
	var names []string
	for _, vlist := range vd.VarLists {
		names = append(names, vlist.Names...)
	}

	return names
}

func (vd *VarDecl) Position() *report.TextPosition {
	return vd.Pos
}

// VarList is a set of variables declared with the same type extension or
// initializer in a variable declaration.
type VarList struct {
	Names         []string
	NamePositions []*report.TextPosition
	Mutabilities  []int
	Type          typing.DataType
	Initializer   Expr
}

// Assign denotes an assignment.
type Assign struct {
	ExprBase

	LHSExprs []Expr
	RHSExprs []Expr

	// Oper is `nil` if no compound operator is being applied.
	Oper *Oper
}

func (a *Assign) Position() *report.TextPosition {
	return report.TextPositionFromRange(
		a.LHSExprs[0].Position(),
		a.RHSExprs[len(a.RHSExprs)-1].Position(),
	)
}

// UnaryUpdate denotes an increment or a decrement operation.
type UnaryUpdate struct {
	ExprBase

	Operand Expr
	Oper    *Oper
}

func (uu *UnaryUpdate) Position() *report.TextPosition {
	return report.TextPositionFromRange(
		uu.Operand.Position(),
		uu.Oper.Pos,
	)
}
