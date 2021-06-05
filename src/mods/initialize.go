package mods

import (
	"chai/common"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml"
)

// InitModule creates a new module with the given name at the given path
func InitModule(name, path string, noProfiles, enableCaching bool) error {
	// convert the module directory to the path to module file
	modFilePath := filepath.Join(path, common.ModuleFileName)

	// check to see if a module already exists
	_, err := os.Stat(modFilePath)
	if err == nil {
		return errors.New("module file already exists")
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("module file error: %s", err.Error())
	}

	// validate module name
	if !IsValidIdentifier(name) {
		return errors.New("module name must be a valid identifier")
	}

	// create module
	mod := &tomlModule{
		Name:                name,
		AllowProfileElision: true,
		Version:             common.ChaiVersion,
	}

	if enableCaching {
		mod.CacheDirectory = filepath.Join(path, ".chai")
		mod.ShouldCache = true
	}

	if !noProfiles {
		mod.BuildProfiles = []*tomlProfile{newInitProfile(name, path, true), newInitProfile(name, path, false)}
	}

	// encode and save module to file
	f, err := os.Create(modFilePath)
	if err != nil {
		return fmt.Errorf("error creating module file: %s", err.Error())
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(&tomlModuleFile{Module: mod}); err != nil {
		return fmt.Errorf("error encoding TOML %s", err.Error())
	}

	return nil
}

// newInitProfile creates a new initial profile for a module
func newInitProfile(modName, modDir string, debug bool) *tomlProfile {
	prof := &tomlProfile{
		TargetOS:    runtime.GOOS,
		TargetArch:  runtime.GOARCH,
		OutputPath:  filepath.Clean(filepath.Join(filepath.Dir(modDir), "bin", modName+"_debug")),
		Format:      "bin",
		Debug:       debug,
		DefaultProf: debug, // debug profile is the default
	}

	if strings.Contains(runtime.GOOS, "windows") {
		prof.OutputPath += ".exe"
	}

	if debug {
		prof.Name = "debug"
	} else {
		prof.Name = "release"
	}

	return prof
}
