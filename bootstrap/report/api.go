package report

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// rep is a global reference to a shared rep.
var rep reporter

// InitReporter initializes the global reporter with the provided log level.
func InitReporter(loglevel int) {
	rep = reporter{
		LogLevel:  loglevel,
		m:         &sync.Mutex{},
		startTime: time.Now(),
	}
}

// ShouldProceed indicates whether or not there have been any non-fatal errors
// that should cause compilation to stop at the current phase/sub-phase.
func ShouldProceed() bool {
	return rep.errorCount == 0
}

// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// NOTE: All report functions will only display if the appropriate log level is
// set.  Most report functions will simply fail silently if below their
// appropriate log level.

// ReportCompileError reports a compilation error.
func ReportCompileError(ctx *CompilationContext, pos *TextPosition, msg string) {
	rep.handleMsg(&CompileMessage{
		Context:  ctx,
		Position: pos,
		Message:  msg,
		IsError:  true,
	})
}

// ReportCompileWarning reports a compilation warning.
func ReportCompileWarning(ctx *CompilationContext, pos *TextPosition, msg string) {
	rep.handleMsg(&CompileMessage{
		Context:  ctx,
		Position: pos,
		Message:  msg,
		IsError:  false,
	})
}

// ReportModuleError reports an error loading a module.
func ReportModuleError(modName string, msg string) {
	rep.handleMsg(&ModuleMessage{
		ModName: modName,
		Message: msg,
		IsError: true,
	})
}

// ReportModuleWarning reports a warning from loading a module.
func ReportModuleWarning(modName string, msg string) {
	rep.handleMsg(&ModuleMessage{
		ModName: modName,
		Message: msg,
		IsError: false,
	})
}

// ReportPackageError reports an error initializing a package.
func ReportPackageError(modName, pkgSubPath, msg string) {
	rep.handleMsg(&PackageError{
		ModRelPath: modName + pkgSubPath,
		Message:    msg,
	})
}

// ReportFatal reports a fatal error and exits the program.  It also
// automatically formats error messages as necessary.
func ReportFatal(msg string, args ...interface{}) {
	rep.errorCount++

	displayFatalError(fmt.Sprintf(msg, args...))

	os.Exit(1)
}

// -----------------------------------------------------------------------------
// Below are all the "aesthetic" reporting functions that will only run if the
// log level is to verbose.  These provide additional information about the
// compilation process to the user so as to make the compiler more friendly.

// LogCompileHeader reports the pre-compilation header: information about the
// compiler's current configuration (version, target, caching, etc.).
func ReportCompileHeader(target string, caching bool) {
	if rep.LogLevel == LogLevelVerbose {
		displayCompileHeader(target, caching)
	}
}

// ReportCompilationFinished reports the concluding message for compilation.
// This displays information about the end of the compilation process.
func ReportCompilationFinished(outputPath string) {
	// log all warnings
	if rep.LogLevel >= LogLevelWarning {
		for _, warning := range rep.warnings {
			warning.display()
		}
	}

	// log closing message
	if rep.LogLevel == LogLevelVerbose {
		displayCompilationFinished(ShouldProceed(), outputPath)
	}
}
