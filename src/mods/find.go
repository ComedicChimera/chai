package mods

import (
	"chai/common"
	"os"
	"path/filepath"
)

// ResolveModulePath takes in a module name and attempts to determine
// an appropriate module path based on that module name
func (m *ChaiModule) ResolveModulePath(name string) (string, bool) {
	// if we are importing with the same name as the current module, then we are
	// in the current module and importing from it
	if name == m.Name {
		return m.ModuleRoot, true
	}

	// local directories are next in priority
	for _, ldPath := range m.LocalImportDirs {
		mpath := filepath.Join(ldPath, name, common.ModuleFileName)
		if checkPath(mpath) {
			return filepath.Dir(mpath), true
		}
	}

	// check the public (global) path next
	pubPath := filepath.Join(common.ChaiPath, "lib/pub", name, common.ModuleFileName)
	if checkPath(pubPath) {
		return filepath.Dir(pubPath), true
	}

	// finally check the standard path
	stdPath := filepath.Join(common.ChaiPath, "lib/std", name, common.ModuleFileName)
	if checkPath(stdPath) {
		return filepath.Dir(stdPath), true
	}

	return "", false
}

// checkPath checks to see if a potential module path is valid
func checkPath(abspath string) bool {
	// note that we are looking for a module file here, not just a directory
	finfo, err := os.Stat(abspath)
	if err != nil {
		return false
	}

	return !finfo.IsDir()
}
