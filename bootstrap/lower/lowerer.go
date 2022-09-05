package lower

import (
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
	for _, chfile := range l.pkg.Files {
		l.chfile = chfile

		for _, def := range chfile.Definitions {
			l.lowerDef(def)
		}
	}
}

/* -------------------------------------------------------------------------- */

// appendStmt appends a new statement to the current block.
func (l *Lowerer) appendStmt(stmt mir.Statement) {
	*l.block = append(*l.block, stmt)
}
