package cmd

import "chai/report"

func (c *Compiler) addPrelude() {
	// TEMPORARY: we are just going to add the core and core.runtime packages to
	// the package to get all the necessary intrinsics and make sure the program
	// entry point is properly defined. Later, we will actually import these in
	// the prelude (universe).

	// import core
	_, ok := c.importPackgage(c.rootModule, "core", "")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core`")
		return
	}

	// import core.runtime
	_, ok = c.importPackgage(c.rootModule, "core", ".runtime")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core.runtime`")
		return
	}
}
