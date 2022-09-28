package lower

import (
	"chaic/ast"
	"chaic/depm"
	"chaic/mir"
)

// Lowerer is responsible for converting the AST to MIR.
type Lowerer struct {
	// The package being lowered.
	pkg *depm.ChaiPackage

	// The MIR bundle being generated from the package.
	bundle *mir.Bundle

	// The current file being lowered into the bundle.
	chfile *depm.ChaiFile

	// The current block being generated in.
	block *[]mir.Statement

	// The list of deferred predicates to be lowered after definitions.
	deferred []deferredPredicate
}

// deferredPredicate is a predicate waiting to be lowered.
type deferredPredicate struct {
	// The function containing the block.
	fn *mir.Function

	// The AST node of the predicate.
	pred ast.ASTNode
}

// Lower converts a Chai package into a MIR bundle.
func Lower(pkg *depm.ChaiPackage) *mir.Bundle {
	l := &Lowerer{pkg: pkg, bundle: &mir.Bundle{
		ID:         pkg.ID,
		PkgAbsPath: pkg.AbsPath,
	}}

	l.lower()

	return l.bundle
}

/* -------------------------------------------------------------------------- */

// lower converts the Lowerer's Chai package into its MIR bundle.
func (l *Lowerer) lower() {
	// Lower all the definitions.
	for _, chfile := range l.pkg.Files {
		l.chfile = chfile

		for _, def := range chfile.Definitions {
			l.lowerDef(def)
		}
	}

	// Lower all the deferred predicates.
	for _, dp := range l.deferred {
		l.block = &dp.fn.Body
		if bodyExpr, ok := dp.pred.(ast.ASTExpr); ok {
			l.appendStmt(l.lowerExpr(bodyExpr))
		} else {
			l.lowerBlock(dp.pred.(*ast.Block))
		}
	}
}

/* -------------------------------------------------------------------------- */

// appendStmt appends a new statement to the current block.
func (l *Lowerer) appendStmt(stmt mir.Statement) {
	*l.block = append(*l.block, stmt)
}
