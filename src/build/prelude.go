package build

import (
	"chai/logging"
	"chai/mods"
	"chai/sem"
	"path/filepath"
)

// preludeImports is a table of each prelude package and what to import from it
var preludeImports = map[string][]string{
	".":     {"min"},
	"types": {},
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
		// go through an attach each of the prelude imports
		for pkgName, preludeImport := range preludeImports {
			if pkgName == "." {
				// don't attach the util package (core root package) if the file
				// requests it not be attached (via. `!! no_util`)
				if _, ok := file.Metadata["no_util"]; ok {
					continue
				}

				if !c.attachPreludePackage(file, c.coreMod.RootPackage, preludeImport) {
					return false
				}
			} else if _, ok := c.coreMod.SubPackages[pkgName]; ok {
				// if the package has already been initialized, attach it
				if !c.attachPreludePackage(file, c.coreMod.SubPackages[pkgName], preludeImport) {
					return false
				}
			} else if preludePkg, ok := c.initPackage(c.coreMod, filepath.Join(c.coreMod.ModuleRoot, pkgName)); ok {
				// if the package has not been initialized, initialize it and
				// then attach it
				if !c.attachPreludePackage(file, preludePkg, preludeImport) {
					return false
				}
			} else {
				logging.LogFatal("failed to initialize required prelude package")
				return false
			}
		}
	}

	return true
}

// attachPreludePackage attaches a single package to a prelude file.  The final
// argument is the list of symbol names to import from the package
func (c *Compiler) attachPreludePackage(file *sem.ChaiFile, preludePkg *sem.ChaiPackage, preludeImport []string) bool {
	// obviously, we don't want to attach a prelude package to itself
	if file.Parent != preludePkg {
		preludeImportedSymbols := make(map[string]*logging.TextPosition)
		for _, name := range preludeImport {
			preludeImportedSymbols[name] = nil
		}

		return file.AddSymbolImports(preludePkg, preludeImportedSymbols)
	}

	return true
}
