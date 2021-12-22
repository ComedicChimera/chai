package cmd

import (
	"chai/common"
	"chai/depm"
	"os"
	"path/filepath"
	"strings"
)

// importPackage retrieves a reference to an initialized Chai package based on
// the module name and package subpath.  This may return an already-initialized
// package or it may load a new package depending on whether or not it located a
// suitable reference in the dependency graph.
func (c *Compiler) importPackgage(parentMod *depm.ChaiModule, moduleName, pkgSubPath string) (*depm.ChaiPackage, bool) {
	// get the absolute path to the module
	modAbsPath, ok := c.getModuleAbsPath(parentMod, moduleName)
	if !ok {
		return nil, false
	}

	// generate a module ID based on the absolute path and use it either
	// retrieve the already loaded module from the dependency graph or load a
	// new module based on the path and add it to the dependency graph.
	modID := depm.GenerateIDFromPath(modAbsPath)
	var mod *depm.ChaiModule
	if mod, ok = c.depGraph[modID]; !ok {
		mod, ok = depm.LoadModule(modAbsPath)
		if !ok {
			return nil, false
		}

		// add the new mod the dependency graph
		c.depGraph[modID] = mod
	}

	// if the package has already been initialized in the loaded module, just
	// return that loaded package.  NOTE: we will check recursive imports later.
	if pkgSubPath == "" {
		if mod.RootPackage != nil {
			return mod.RootPackage, true
		}

		return c.initPkg(mod, mod.AbsPath)
	} else {
		if pkg, ok := mod.SubPackages[pkgSubPath]; ok {
			return pkg, true
		}

		// convert the pkgSubPath into an actual usable file relative path
		return c.initPkg(mod, filepath.Join(mod.AbsPath, strings.ReplaceAll(pkgSubPath[1:], ".", "/")))
	}
}

// getModuleAbsPath takes a module name and a parent module attempts to generate
// an absolute path to the module.
func (c *Compiler) getModuleAbsPath(parentMod *depm.ChaiModule, moduleName string) (string, bool) {
	// TODO: non-standard imports

	stdModPath := filepath.Join(common.ChaiPath, "modules/std/", moduleName)
	if finfo, err := os.Stat(stdModPath); err == nil && finfo.IsDir() {
		return stdModPath, true
	}

	return "", false
}
