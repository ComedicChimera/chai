package mods

import (
	"chai/common"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

// ResolveModulePath takes in a module name and attempts to determine
// an appropriate module path based on that module name
func (m *ChaiModule) ResolveModulePath(name string) (string, bool) {
	// if we are importing with the same name as the current module, then we are
	// in the current module and importing from it
	if name == m.Name {
		return m.ModuleRoot, true
	}

	// search the enclosing directory of the module to find any adjacent modules
	if path, foundMod := searchPath(filepath.Dir(m.ModuleRoot), name); foundMod {
		return path, true
	}

	// local directories are next in priority
	for _, ldPath := range m.LocalImportDirs {
		if path, foundMod := searchPath(ldPath, name); foundMod {
			return path, true
		}
	}

	// check the public (global) path next
	if path, foundMod := searchPath(filepath.Join(common.ChaiPath, "lib/pub"), name); foundMod {
		return path, true
	}

	// finally check the standard path
	if path, foundMod := searchPath(filepath.Join(common.ChaiPath, "lib/std"), name); foundMod {
		return path, true
	}

	return "", false
}

// searchPath searches a directory for a module with a matching name, and
// returns the abspath to a module in that path if it exists
func searchPath(abspath, modName string) (string, bool) {
	// first we check a subpath with the same name as the module (before we
	// search the directory) since we know that it is likely that modules and
	// module paths will have the same name
	potentialModPath := filepath.Join(abspath, modName)
	if checkPath(potentialModPath, modName) {
		return potentialModPath, true
	}

	// otherwise, we perform linear search to attempt to find a matching module
	// in the directory (checking each possible path)
	finfos, err := ioutil.ReadDir(abspath)
	if err != nil {
		return "", false
	}

	for _, finfo := range finfos {
		potentialModPath = filepath.Join(abspath, finfo.Name())
		if finfo.IsDir() && checkPath(potentialModPath, modName) {
			return potentialModPath, true
		}
	}

	return "", false
}

// checkPath checks to see if a potential module path is valid -- accepts the
// path to the module root not the path to the module file
func checkPath(abspath, modName string) bool {
	// convert the abs path into a path to the module file
	mfPath := filepath.Join(abspath, common.ModuleFileName)

	// check to see if we can open the module file
	finfo, err := os.Stat(mfPath)
	if err != nil {
		return false
	}

	// make sure that it is a file
	if !finfo.IsDir() {
		// only really want to check the name here so we don't do the full
		// unmarshal -- LoadFile should be faster (kinda wish we could query
		// without loading the full file but ¯\_(ツ)_/¯).  Also, this may not be
		// a valid module at all in which case we shouldn't error since the user
		// didn't explicit specify that this path was a valid module.
		tree, err := toml.LoadFile(mfPath)
		if err != nil {
			return false
		}

		if tree.Has("name") {
			if nameField, ok := tree.Get("name").(string); ok {
				return nameField == modName
			}

			return false
		}
	}

	return false
}
