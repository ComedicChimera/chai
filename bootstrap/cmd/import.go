package cmd

import "chai/depm"

// importPackage retrieves a reference to an initialized Chai package based on
// the module path and package subpath.  This may return an already-initialized
// package or it may load a new package depending on whether or not it located a
// suitable reference in the dependency graph.
func (c *Compiler) importPackgage(parentMod *depm.ChaiModule, modulePath, pkgSubPath string) (*depm.ChaiPackage, bool) {
	// TODO
	return nil, false
}
