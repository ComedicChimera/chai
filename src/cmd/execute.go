package cmd

import (
	"chai/build"
	"chai/common"
	"chai/logging"
	"chai/mods"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ComedicChimera/olive"
)

// TODO: implement commands
// check      check packages and output errors
// clean      remove object files and cached data
// del        delete installed modules
// fetch      fetch and install a remote module
// make       compile intermediates (asm, object, etc.)
// new        create a new project
// run        compile and run packages and modules
// test       test packages and modules
// update     update or rollback whirl

// Execute runs the main `chai` application
func Execute() {
	// compilation cannot proceed without the chai_path
	if !initChaiPath() {
		return
	}

	// set up the argument parser and all its extended commands and arguments
	cli := olive.NewCLI("chai", "chai is a tool for managing Chai projects", true)
	logLvlArg := cli.AddSelectorArg("loglevel", "ll", "the compiler log level", false, []string{"silent", "error", "warn", "verbose"})
	logLvlArg.SetDefaultValue("verbose")

	buildCmd := cli.AddSubcommand("build", "compile source code", true)
	buildCmd.AddPrimaryArg("module-path", "the path to the module to build", true)
	buildCmd.AddStringArg("profile", "p", "the name of the profile to build", false)

	modCmd := cli.AddSubcommand("mod", "manage modules", true)
	modInitCmd := modCmd.AddSubcommand("init", "initialize a module", true)
	modInitCmd.AddFlag("no-profiles", "np", "indicates whether Chai should generate default profiles for this module")
	modInitCmd.AddFlag("caching", "ch", "indicate whether compilation caching should be enabled for this module")
	modInitCmd.AddPrimaryArg("module-path", "the path to the module directory", true)

	cli.AddSubcommand("version", "print the Chai version", false)

	// run the argument parser
	result, err := olive.ParseArgs(cli, os.Args)
	if err != nil {
		logging.PrintErrorMessage("CLI Usage Error", err)
		return
	}

	// process the inputed command line
	subcmdName, subResult, _ := result.Subcommand()
	switch subcmdName {
	case "build":
		execBuildCommand(subResult, result.Arguments["loglevel"].(string))
	case "mod":
		execModCommand(subResult)
	case "version":
		logging.PrintInfoMessage("Chai Version", common.ChaiVersion)
	}
}

// execBuildCommand executes the build subcommand and handles all errors
func execBuildCommand(result *olive.ArgParseResult, loglevel string) {
	// extract CLI data
	moduleRelPath, _ := result.PrimaryArg()

	modulePath, err := filepath.Abs(moduleRelPath)
	if err != nil {
		logging.PrintErrorMessage("Path Error", err)
		return
	}

	profArgVal, ok := result.Arguments["profile"]
	selectedProfile := ""
	if ok {
		selectedProfile = profArgVal.(string)
	}

	// attempt to load the module
	buildProfile := &mods.BuildProfile{}
	mod, err := mods.LoadModule(modulePath, selectedProfile, buildProfile)
	if err != nil {
		logging.PrintErrorMessage("Module Load Error", err)
		return
	}

	// initialize the logger
	logging.Initialize(mod.ModuleRoot, loglevel)

	// build the main project
	c := build.NewCompiler(mod, buildProfile)
	c.Compile()
}

// execModCommand executes the `mod` subcommand and its subcommands.  It handles
// all errors related to this command
func execModCommand(result *olive.ArgParseResult) {
	subcmdName, subResult, _ := result.Subcommand()

	workDir, err := os.Getwd()
	if err != nil {
		logging.PrintErrorMessage("Path Error", err)
		return
	}

	// TODO: mod tidy, mod update, mod install, mod del
	switch subcmdName {
	case "init":
		modNameValue, _ := subResult.PrimaryArg()
		if err := mods.InitModule(modNameValue, workDir, subResult.HasFlag("no-profiles"), subResult.HasFlag("caching")); err != nil {
			logging.PrintErrorMessage("Module Init Error", err)
		}
	}
}

// -----------------------------------------------------------------------------

// initChaiPath checks for a valid whirlpath and initializes its global value.
func initChaiPath() bool {
	if chaiPath, ok := os.LookupEnv("CHAI_PATH"); ok {
		finfo, err := os.Stat(chaiPath)

		if err != nil {
			logging.PrintErrorMessage("Config Error", fmt.Errorf("error loading chai_path: %s", err.Error()))
			return false
		}

		if !finfo.IsDir() {
			logging.PrintErrorMessage("Config Error", errors.New("error loading chai_path: must point to a directory"))
			return false
		}

		common.ChaiPath = chaiPath
		return true
	}

	logging.PrintErrorMessage("Config Error", errors.New("missing CHAI_PATH environment variable"))
	return false
}
