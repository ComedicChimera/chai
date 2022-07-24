package ast

import "chaic/report"

// The abstract interface for all AST nodes.
type ASTNode interface {
	// The text span of the AST.
	Span() *report.TextSpan
}

// A utility base struct for all AST nodes.
type ASTBase struct {
	// The span over which the AST node occurs.
	span *report.TextSpan
}

// NewASTBaseOn creates a new AST base with the given span.
func NewASTBaseOn(span *report.TextSpan) ASTBase {
	return ASTBase{span: span}
}

// NewASTBaseOver creates a new AST base spanning over two spans.
func NewASTBaseOver(start, end *report.TextSpan) ASTBase {
	return ASTBase{span: report.NewSpanOver(start, end)}
}

func (ab ASTBase) Span() *report.TextSpan {
	return ab.span
}
