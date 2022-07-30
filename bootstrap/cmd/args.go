package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"chaic/depm"
	"chaic/report"
	"chaic/util"
)

const usage = `Usage: chaic [flags|options] <path to root directory or file>

Flags:
------
-h, --help      Displays usage information (ie. this text).    
-v, --version   Displays the current compiler version.
-d, --debug     Whether the compiler should output debug information. 

Options:
--------
-o,  --outpath    Sets the path for compilation output.  If a single output is
                  to be produced, this should be a file.  If multiple outputs
                  are to be produced, then this should be a directory.  Defaults
                  to out[.<platform extension>] if unspecified.
-ll, --loglevel   Sets the compiler's log-level.  Valid values are:
                    - "verbose" for outputting all messages (default)
                    - "warn" for outputting errors and warnings
                    - "error" for outputting errors only
                    - "silent" for no output
-m, --outmode     Sets compiler's output mode.  Valid values are:
                    - "exe" for producing an executable (default)
                    - "lib" for producing a static library
                    - "dyn" for producing a dynamic/shared library
                    - "obj" for producing object files for each package
                    - "asm" for producing assembly file for each package
                    - "llvm" for producing LLVM IR for each package
`

// Prints the usage message an exits the compiler with the given exit code.
func printUsage(exitCode int) {
	fmt.Print(usage, "\n")
	os.Exit(exitCode)
}

// argParser is a command-line argument parser.
type argParser struct {
	// The arguments being parsed.
	args []string

	// The argument parser's position within those arguments.
	ndx int
}

// Set containing all the argument names that correspond to options.
var options = map[string]struct{}{
	"o":         {},
	"m":         {},
	"ll":        {},
	"-outpath":  {},
	"-outmode":  {},
	"-loglevel": {},
}

// argumentError displays an argument error and exits the program.
func argumentError(message string, args ...interface{}) {
	fmt.Print("argument error: ", fmt.Sprintf(message, args...), "\n\n")
	printUsage(1)
}

// nextArg parses the next command-line argument if one exists.  The first value
// is the name of the argument.  If this argument is positional, this value is
// empty.  The second value is the value of argument. If this value is empty,
// the argument is a flag.  If an argument exists, at least one of the returned
// values will be non-empty.  The final value indicates whether or not there was
// an argument to parse.
func (ap *argParser) nextArg() (string, string, bool) {
	if ap.ndx < len(ap.args) {
		arg := ap.args[ap.ndx]
		ap.ndx++

		if strings.HasPrefix(arg, "-") { // flag or option
			name := arg[1:]

			if _, ok := options[name]; ok { // option
				// Make sure the option value exists.
				if ap.ndx < len(ap.args) && !strings.HasPrefix(ap.args[ap.ndx], "-") {
					value := ap.args[ap.ndx]
					ap.ndx++
					return name, value, true
				} else {
					argumentError("option %s requires an argument", strings.TrimLeft(name, "-"))
				}
			} else { // flag
				return name, "", true
			}

		} else { // positional
			return "", arg, true
		}
	}

	// No arguments to parse.
	return "", "", false
}

// useArg attempts to use a single command-line argument to initialize the
// compiler.  If the argument is invalid, the program will exit.
func useArg(c *Compiler, name, value string) {
	switch name {
	case "h", "-help":
		printUsage(0)
	case "v", "-version":
		fmt.Println(util.ChaiCompilerID)
		os.Exit(0)
	case "d", "-debug":
		c.debug = true
	case "ll", "-loglevel":
		{
			var logLevel int
			switch value {
			case "silent":
				logLevel = report.LogLevelSilent
			case "error":
				logLevel = report.LogLevelError
			case "warn":
				logLevel = report.LogLevelWarn
			case "verbose":
				logLevel = report.LogLevelVerbose
			default:
				argumentError("invalid log level")
			}

			report.InitReporter(logLevel)
		}
	case "o", "-outpath":
		{
			absPath, err := filepath.Abs(value)
			if err != nil {
				argumentError("invalid output path: %s", absPath)
			}

			c.outputPath = value
		}
	case "m", "-outmode":
		{
			var mode int
			switch value {
			case "exe":
				mode = OutModeExecutable
			case "asm":
				mode = OutModeASM
			case "llvm":
				mode = OutModeLLVM
			case "obj":
				mode = OutModeObj
			case "lib", "dyn":
				argumentError("output mode not implemented yet")
			default:
				argumentError("invalid output mode")
			}

			c.outputMode = mode
		}
	case "":
		if c.rootPath == "" {
			absPath, err := filepath.Abs(value)
			if err != nil {
				argumentError("invalid root path: %s", value)
			}

			c.rootPath = absPath
		} else {
			argumentError("root path specified multiple times")
		}
	default:
		argumentError("unknown flag: %s", name)
	}
}

// NewCompilerFromArgs creates a new compiler instance based on the given
// command line arguments if the arguments are valid and compilation should
// continue: ie. if the user requests the compiler version, then compilation
// should not continue.
func NewCompilerFromArgs() *Compiler {
	c := &Compiler{
		outputMode: OutModeExecutable,
		depGraph:   make(map[uint64]*depm.ChaiPackage),
	}

	ap := argParser{args: os.Args[1:], ndx: 0}

	// Parse all command line arguments.
	for {
		if name, value, ok := ap.nextArg(); ok {
			useArg(c, name, value)
		} else {
			break
		}
	}

	// Check to make sure a root path was specified.
	if c.rootPath == "" {
		argumentError("a root path must be specified")
	}

	// Set default values for any optional unspecified flags.
	report.InitReporter(report.LogLevelVerbose)

	if c.outputPath == "" {
		c.outputPath = filepath.Join(c.rootPath, "out")
	}

	// Validate and correct the output path as necessary.
	switch c.outputMode {
	case OutModeASM, OutModeLLVM, OutModeObj:
		{
			finfo, err := os.Stat(c.outputPath)
			if err == nil {
				if !finfo.IsDir() {
					argumentError("output path must be a directory in the current output mode")
				}
			} else if !os.IsNotExist(err) {
				argumentError("invalid output path: %s", err)
			}
		}
	default:
		{
			finfo, err := os.Stat(c.outputPath)
			if err == nil {
				if finfo.IsDir() {
					argumentError("output path must be a file in the current output mode")
				}
			} else if !os.IsNotExist(err) {
				argumentError("invalid output path: %s", err)
			}

			if runtime.GOOS == "windows" && !strings.HasSuffix(c.outputPath, ".exe") {
				c.outputPath += ".exe"
			}
		}
	}

	// Load the Chai path.
	if chaiPath := os.Getenv("CHAI_PATH"); chaiPath != "" {
		c.chaiPath = chaiPath
	} else {
		report.ReportFatal("missing CHAI_PATH")
	}

	return c
}
