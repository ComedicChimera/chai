package lower

import (
	"chai/depm"
	"chai/ir"
)

// Lowerer converts a package into an IR bundle.
type Lowerer struct {
	pkg *depm.ChaiPackage
	b   *ir.Bundle

	// globalPrefix is the string prefix prepended to all global symbols of this
	// package to prevent linking errors and preserve package namespaces.
	globalPrefix string

	// currfile is the current file being lowered.
	currfile *depm.ChaiFile
}

// NewLowerer creates a new lowerer for the given package.
func NewLowerer(pkg *depm.ChaiPackage) *Lowerer {
	return &Lowerer{
		pkg:          pkg,
		globalPrefix: pkg.Parent.Name + pkg.ModSubPath + ".",
		b:            ir.NewBundle(),
	}
}

// Lower runs the lowerer.
func (l *Lowerer) Lower() *ir.Bundle {
	for _, file := range l.pkg.Files {
		l.currfile = file

		for _, def := range file.Defs {
			l.lowerDef(def)
		}
	}

	return l.b
}
