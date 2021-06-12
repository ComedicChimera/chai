package build

import (
	"chai/logging"
	"chai/mods"
	"chai/sem"
)

// preludeImports is a table of each prelude package and what to import from it
var preludeImports = map[string][]string{
	".": {"print_int"},
	// TODO
}

// attachPrelude adds the prelude imports to a standard package
func (c *Compiler) attachPrelude(parentMod *mods.ChaiModule, pkg *sem.ChaiPackage) bool {
	// make sure not to attach core as dependency of itself
	if parentMod != c.coreMod {
		// note that we are adding the core module by name here so that it is
		// always accessible -- this will override any other modules named
		// `core` -- every module HAS to be able to access the core module
		if _, ok := parentMod.DependsOn[c.coreMod.Name]; !ok {
			parentMod.DependsOn[c.coreMod.Name] = c.coreMod
		}
	}

	// go through each file in the package and perform all necessary attachments
	for _, file := range pkg.Files {
		// don't attach the util package (core root package) if the file
		// requests it not be attached (via. `!! no_util`)
		if _, ok := file.Metadata["no_util"]; !ok && !c.attachPreludePackage(file, c.coreMod.RootPackage) {
			return false
		}

		// TODO: attach other critical prelude packages
	}

	return true
}

// attachPreludePackage attaches a single package to a prelude file
func (c *Compiler) attachPreludePackage(file *sem.ChaiFile, preludePkg *sem.ChaiPackage) bool {
	// obviously, we don't want to attach a prelude package to itself
	if file.Parent != preludePkg {
		var preludeImportedSymbolNames []string

		// check if root package
		if preludePkg == c.coreMod.RootPackage {
			preludeImportedSymbolNames = preludeImports["."]
		} else {
			preludeImportedSymbolNames = preludeImports[preludePkg.Name]
		}

		preludeImportedSymbols := make(map[string]*logging.TextPosition)
		for _, name := range preludeImportedSymbolNames {
			preludeImportedSymbols[name] = nil
		}

		return file.AddSymbolImports(preludePkg, preludeImportedSymbols)
	}

	return true
}
