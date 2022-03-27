package cmd

import "chai/report"

// preludeImportTable is a table of symbols to import from each of the prelude
// packages.
var preludeImportTable = map[string][]string{
	".types": {"string", "rune", "usize"},
}

// addPrelude adds the prelude to the Universe.  This should be called before an
// user packages are imported.
func (c *Compiler) addPrelude() {
	// import `core`
	corePkg, ok := c.importPackgage(c.rootModule, "core", "")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core`")
		return
	}

	// add `core` to the Universe
	c.uni.CorePkg = corePkg

	// reserve prelude names
	for _, names := range preludeImportTable {
		c.uni.ReserveNames(names...)
	}

	// import `core.runtime`
	_, ok = c.importPackgage(c.rootModule, "core", ".runtime")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core.runtime`")
		return
	}

	// import `core.types`
	typesPkg, ok := c.importPackgage(c.rootModule, "core", ".types")
	if !ok {
		report.ReportFatal("failed to import required prelude package: `core.types`")
	}

	// add `core.types` prelude imports
	c.uni.AddSymbolImports(typesPkg, preludeImportTable[".types"]...)
}
