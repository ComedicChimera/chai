package build

import (
	"chai/common"
	"chai/logging"
	"chai/mods"
	"chai/sem"
	"chai/syntax"
	"chai/walk"
	"fmt"
	"path/filepath"
)

// initDependencies walks the imports of a given package and recursively
// initializes all the dependencies determined from those imports.  It returns a
// boolean indicating whether or not it was successful in this initialization
// process.
func (c *Compiler) initDependencies(parentMod *mods.ChaiModule, pkg *sem.ChaiPackage) bool {
	// start by attaching the prelude import
	if !c.attachPrelude(parentMod, pkg) {
		return false
	}

	// walk each import appropriately
	for _, file := range pkg.Files {
		for i, ast := range file.AST.Content {
			if branch, ok := ast.(*syntax.ASTBranch); ok {
				if branch.Name == "import_stmt" {
					if !c.processImport(parentMod, file, branch) {
						return false
					}
				} else {
					// imports are only at the top so first non-import we
					// encounter, we exit, but also make sure to slice the
					// imports off the AST so future components of the compiler
					// won't have to worry with them
					file.AST.Content = file.AST.Content[i+1:]
					break
				}
			}
		}
	}

	return true
}

// processImport walks a single import statement and initializes the dependency
func (c *Compiler) processImport(parentMod *mods.ChaiModule, file *sem.ChaiFile, importStmt *syntax.ASTBranch) bool {
	var importedMod *mods.ChaiModule
	var importedPkg *sem.ChaiPackage

	importedSymbols := make(map[string]*logging.TextPosition)

	// name to declare the imported package as (support for renaming)
	var importedPkgName string
	var importedPkgPathPos *logging.TextPosition

	// walk the import AST and collect the data above
	var err error
	for _, item := range importStmt.Content {
		switch v := item.(type) {
		case *syntax.ASTBranch:
			if v.Name == "package_path" {
				importedMod, importedPkg, err = c.importPackage(parentMod, v)
				if err != nil {
					logging.LogConfigError("Import", err.Error())
					return false
				}
			} else /* `identifier_list` */ {
				_, importedSymbols, err = walk.WalkIdentifierList(v)
				if err != nil {
					logging.LogCompileError(
						file.LogContext,
						fmt.Sprintf("symbol `%s` imported multiple times", err.Error()),
						logging.LMKImport,
						importedSymbols[err.Error()],
					)
					return false
				}

				importedPkgName = importedPkg.Name
			}
		case *syntax.ASTLeaf:
			// only `IDENTIFIER` => rename
			if v.Kind == syntax.IDENTIFIER {
				// as clauses always come after main package name so this rename
				// will always override the package name
				importedPkgName = v.Value
				importedPkgPathPos = v.Position()
			}
		}
	}

	// check to make sure the package isn't importing itself
	if file.Parent != importedPkg {
		logging.LogCompileError(
			file.LogContext,
			fmt.Sprintf("package `%s` cannot import itself", parentMod.BuildPackagePathString(importedPkg)),
			logging.LMKImport,
			importedPkgPathPos,
		)

		return false
	}

	// add the imported module as a dependency of the parent module if it hasn't
	// already been added and is not equivalent to the parent module (ie.
	// importing another package within the same module)
	if parentMod != importedMod {
		if _, ok := parentMod.DependsOn[importedMod.Name]; !ok {
			// if the imported package is uninitialized, then there must be an
			// import cycle since it must still be processing its imports
			if !importedPkg.Initialized {
				// create the import paths to the two packages

				logging.LogCompileError(
					file.LogContext,
					fmt.Sprintf(
						"cross-module import cycle detected between `%s` and `%s`",
						parentMod.BuildPackagePathString(file.Parent),
						importedMod.BuildPackagePathString(importedPkg),
					),
					logging.LMKImport,
					importedPkgPathPos,
				)
			}

			parentMod.DependsOn[importedMod.Name] = importedMod
		}
	}

	// add imported symbols to file if they exist
	if len(importedSymbols) > 0 {
		return file.AddSymbolImports(importedPkg, importedSymbols)
	} else if err := file.AddPackageImport(importedPkg, importedPkgName); err != nil {
		logging.LogCompileError(
			file.LogContext,
			fmt.Sprintf("multiple symbols imported with name `%s`", err.Error()),
			logging.LMKImport,
			importedSymbols[err.Error()],
		)
		return false
	}

	// import public operators
	if err := file.ImportOperators(importedPkg); err != nil {
		logging.LogCompileError(
			file.LogContext,
			fmt.Sprintf("multiple conflicting overloads for `%s` operator", err.Error()),
			logging.LMKImport,
			importedPkgPathPos,
		)
	}

	return true
}

// importPackage attempts to import a package based on a given `package_path` AST
func (c *Compiler) importPackage(parentMod *mods.ChaiModule, pkgPathBranch *syntax.ASTBranch) (*mods.ChaiModule, *sem.ChaiPackage, error) {
	// extract the module path
	var modName string
	subPath := ""
	for i, item := range pkgPathBranch.Content {
		itemleaf := item.(*syntax.ASTLeaf)
		if itemleaf.Kind == syntax.IDENTIFIER {
			if modName == "" {
				modName = itemleaf.Value
			} else {
				subPath += itemleaf.Value
			}
		} else if itemleaf.Kind == syntax.DOT && i != 1 {
			subPath += string(filepath.Separator)
		}
	}

	// find the parent module
	importedMod, err := c.findModule(parentMod, modName)
	if err != nil {
		return nil, nil, err
	}

	// if there is no subpath, then we can safely just return the root package
	// and first-determined imported module as the correct package and module
	if subPath == "" {
		if importedMod.RootPackage == nil {
			if pkg, ok := c.initPackage(importedMod, importedMod.ModuleRoot); ok {
				return importedMod, pkg, nil
			} else {
				return nil, nil, fmt.Errorf("failed to load root package of module `%s`", importedMod.Name)
			}
		}

		return importedMod, importedMod.RootPackage, nil
	}

	// if we have a subpath, we first need to check our imported module already
	// has a listing for that subpackage; if it does, we simply return that
	// known package.
	if importedPkg, ok := importedMod.SubPackages[subPath]; ok {
		return importedMod, importedPkg, nil
	}

	// if the subpath is not listed, we need to initialize the package at that
	// subpath and return it
	if pkg, ok := c.initPackage(importedMod, filepath.Join(importedMod.ModuleRoot, subPath)); ok {
		return importedMod, pkg, nil
	}

	return nil, nil, fmt.Errorf("module `%s` has no package at sub-path `%s`", modName, subPath)
}

// findModule attempts to locate and load (if not already loaded) a module based
// on its module name. The `parentMod` provides the search context for finding
// the module (eg. local import directories)
func (c *Compiler) findModule(parentMod *mods.ChaiModule, modName string) (*mods.ChaiModule, error) {
	if subMod, ok := parentMod.DependsOn[modName]; ok {
		return subMod, nil
	}

	// first, determine the appropriate path to the new module
	if modAbsPath, ok := parentMod.ResolveModulePath(modName); ok {
		// check first to see if we have already loaded the module
		if loadedMod, ok := c.depGraph[common.GenerateIDFromPath(modAbsPath)]; ok {
			return loadedMod, nil
		}

		// if we haven't, load it, add it to the depG, and return it if possible
		mod, err := mods.LoadModule(modAbsPath, "", c.buildProfile)
		if err != nil {
			return nil, err
		}

		c.depGraph[mod.ID] = mod
		return mod, nil
	}

	return nil, fmt.Errorf("unable to locate module by name `%s`", modName)
}
