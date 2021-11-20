package cmd

import (
	"chai/common"
	"chai/report"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ComedicChimera/olive"
)

// Execute is the main entry point for the `chai` CLI utility
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

	buildCmd.AddStringArg("output-path", "o", "the output path for the file", false)
	tarchArg := buildCmd.AddSelectorArg("target-arch", "ta", "the target architecture", false, []string{"amd64"})
	tarchArg.SetDefaultValue("amd64")
	tosArg := buildCmd.AddSelectorArg("target-os", "tos", "the target OS", false, []string{"windows"})
	tosArg.SetDefaultValue("windows")
	buildCmd.AddFlag("release", "rm", "build in release mode")

	modCmd := cli.AddSubcommand("mod", "manage modules", true)
	modInitCmd := modCmd.AddSubcommand("init", "initialize a module", true)
	modInitCmd.AddFlag("no-profiles", "np", "indicates whether Chai should generate default profiles for this module")
	modInitCmd.AddFlag("caching", "ch", "indicate whether compilation caching should be enabled for this module")
	modInitCmd.AddPrimaryArg("module-path", "the path to the module directory", true)

	cli.AddSubcommand("version", "print the Chai version", false)

	// run the argument parser
	result, err := olive.ParseArgs(cli, os.Args)
	if err != nil {
		report.ReportFatal(err.Error())
	}

	// process the inputed command line
	subcmdName, subResult, _ := result.Subcommand()
	switch subcmdName {
	case "build":
		execBuildCommand(subResult, result.Arguments["loglevel"].(string))
	case "mod":
		execModCommand(subResult)
	case "version":
		report.DisplayInfoMessage("Chai Version", common.ChaiVersion)
	}
}

// execBuildCommand executes the build subcommand and handles all errors
func execBuildCommand(result *olive.ArgParseResult, loglevel string) {
	// initialize the reporter
	report.InitReporter(report.LogLevelVerbose)

	// get the primary argument: the root path
	rootPath, _ := result.PrimaryArg()

	// construct the build profile
	profile := &BuildProfile{
		Debug:        !result.HasFlag("release"),
		TargetOS:     result.Arguments["target-os"].(string),
		TargetArch:   result.Arguments["target-arch"].(string),
		OutputFormat: FormatBin,
	}

	if outputPath, ok := result.Arguments["output-path"]; ok {
		profile.OutputPath = outputPath.(string)
	} else {
		// NOTE: we assume EXE since this bootstrapped compiler will only
		// compile to windows
		profile.OutputPath = filepath.Join(rootPath, "out.exe")
	}

	if filepath.Ext(profile.OutputPath) == ".o" {
		profile.OutputFormat = FormatObj
	}

	// TODO: support other extensions

	// create the compiler
	c := NewCompiler(rootPath, profile)

	// run analysis
	if c.Analyze() {
		// if analysis succeeds, run generation
		c.Generate()
	}

	// end whatever the final compilation phase was and display the concluding
	// message of compilation.
	report.ReportEndPhase()
	report.ReportCompilationFinished(c.profile.OutputPath)
}

// execModCommand executes the `mod` subcommand and its subcommands.  It handles
// all errors related to this command
func execModCommand(result *olive.ArgParseResult) {
	// TODO
}

// -----------------------------------------------------------------------------

// initChaiPath checks for a valid whirlpath and initializes its global value.
func initChaiPath() bool {
	if chaiPath, ok := os.LookupEnv("CHAI_PATH"); ok {
		finfo, err := os.Stat(chaiPath)

		if err != nil {
			report.ReportFatal(fmt.Sprintf("error loading chai_path: %s", err.Error()))
		}

		if !finfo.IsDir() {
			report.ReportFatal("error loading chai_path: must point to a directory")
		}

		common.ChaiPath = chaiPath
		return true
	}

	report.ReportFatal("missing CHAI_PATH environment variable")

	// unreachable
	return false
}
