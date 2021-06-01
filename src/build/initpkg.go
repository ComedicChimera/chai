package build

import (
	"chai/common"
	"chai/deps"
	"chai/logging"
	"chai/mods"
	"chai/syntax"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// initPackage attempts to initialize a directory as a Chai Package. It loads
// and parses all Chai files in the directory but does not process dependencies.
// The package is appropriately added as a subpackage of its module. The package
// is returned after it is initialized along with a boolean flag indicating
// success or failure.
func (c *Compiler) initPackage(parentMod *mods.ChaiModule, abspath string) (*deps.ChaiPackage, bool) {
	// validate package path
	finfo, err := os.Stat(abspath)
	if err != nil {
		logging.LogConfigError("Package", fmt.Sprintf("unable to load package at %s: %s", abspath, err.Error()))
		return nil, false
	}

	if !finfo.IsDir() {
		logging.LogConfigError("Package", "a package must be directory not a file")
		return nil, false
	}

	// create a new package
	newpkg := deps.NewPackage(abspath)

	// load and parse package files concurrently
	finfos, err := ioutil.ReadDir(abspath)
	if err != nil {
		logging.LogConfigError("Package", fmt.Sprintf("error walking directory %s: %s", abspath, err.Error()))
		return nil, false
	}

	fchan := make(chan *deps.ChaiFile)
	fcount := 0
	for _, finfo := range finfos {
		// we only want to parse Chai files (not directories or other files)
		if !finfo.IsDir() && filepath.Ext(finfo.Name()) == common.SrcFileExtension {
			go c.initFile(fchan, newpkg, filepath.Join(abspath, finfo.Name()))
			fcount++
		}
	}

	if fcount == 0 {
		logging.LogConfigError("Package", "unable to load a package that contains no Chai source files")
		return nil, false
	}

	for i := 0; i < fcount; i++ {
		newfile := <-fchan
		if newfile != nil {
			newpkg.Files = append(newpkg.Files, newfile)
		}
	}

	// add the new package to its parent module
	modPkgPath, err := filepath.Rel(parentMod.ModuleRoot, abspath)
	if err != nil {
		logging.LogFatal("failed to compute package path inside module")
	}

	if modPkgPath == "." {
		parentMod.RootPackage = newpkg
	} else {
		parentMod.SubPackages[modPkgPath] = newpkg
	}

	return newpkg, logging.ShouldProceed()
}

// initFile attempts to load and parse a file concurrently.  It takes in a
// channel to write to if the file is initialized successfully as well as a path
// to the file.  Note that if the file fails to initialize, an appropriate error
// will be logged and `nil` will be written to the channel.
func (c *Compiler) initFile(fchan chan *deps.ChaiFile, parentpkg *deps.ChaiPackage, fabspath string) {
	// create the file struct
	newfile := &deps.ChaiFile{
		Parent:      parentpkg,
		FilePath:    fabspath,
		LogContext:  &logging.LogContext{PackageID: parentpkg.ID, FilePath: fabspath},
		GlobalTable: make(map[string]*deps.Symbol),
	}

	// create the scanner for the file
	if sc, ok := syntax.NewScanner(fabspath, newfile.LogContext); ok {
		// TODO: process metadata

		// create and runs the parser
		p := syntax.NewParser(c.parsingTable, sc)
		if ast, ok := p.Parse(); ok {
			// this is never NOT an ast branch
			newfile.AST = ast.(*syntax.ASTBranch)

			// parsing succeeds => write file to channel and return
			fchan <- newfile
			return
		}
	}

	// if we reach this point, we know initialization failed to we just write
	// `nil` to the channel
	fchan <- nil
}
