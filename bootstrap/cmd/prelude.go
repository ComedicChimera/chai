package cmd

import "chai/report"

func (c *Compiler) addPrelude() {
	// Import `core`.
	corePkg, ok := c.importPackgage(c.rootModule, "core", "")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core`")
		return
	}

	// Add `core` to the Universe.
	c.uni.CorePkg = corePkg

	// TODO: reserve prelude names

	// Import `core.runtime`.
	_, ok = c.importPackgage(c.rootModule, "core", ".runtime")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core.runtime`")
		return
	}

	// TODO: import `core.types`

	// TODO: add prelude imports
}
