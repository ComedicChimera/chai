package report

import (
	"os"
	"sync"
)

// reporter is a global reference to a shared reporter.
var reporter Reporter

// InitReporter initializes the global reporter with the provided log level.
func InitReporter(loglevel int) {
	reporter = Reporter{
		LogLevel: loglevel,
		m:        &sync.Mutex{},
	}
}

// ShouldProceed indicates whether or not there have been any non-fatal errors
// that should cause compilation to stop at the current phase/sub-phase.
func ShouldProceed() bool {
	return reporter.errorCount == 0
}

// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// NOTE: All report functions will only display if the appropriate log level is
// set.  Most report functions will simply fail silently if below their
// appropriate log level.

// ReportCompileError reports a compilation error.
func ReportCompileError(ctx *CompilationContext, pos *TextPosition, msg string) {
	reporter.handleMsg(&CompileMessage{
		Context:  ctx,
		Position: pos,
		Message:  msg,
		IsError:  true,
	})
}

// ReportCompileWarning reports a compilation warning.
func ReportCompileWarning(ctx *CompilationContext, pos *TextPosition, msg string) {
	reporter.handleMsg(&CompileMessage{
		Context:  ctx,
		Position: pos,
		Message:  msg,
		IsError:  false,
	})
}

// ReportModuleError reports an error loading a module.
func ReportModuleError(modName string, msg string) {
	reporter.handleMsg(&ModuleMessage{
		ModName: modName,
		Message: msg,
		IsError: true,
	})
}

// ReportModuleWarning reports a warning from loading a module.
func ReportModuleWarning(modName string, msg string) {
	reporter.handleMsg(&ModuleMessage{
		ModName: modName,
		Message: msg,
		IsError: false,
	})
}

// ReportFatal reports a fatal error and exits the program.
func ReportFatal(msg string) {
	displayFatalError(msg)
	os.Exit(1)
}

// -----------------------------------------------------------------------------
// Below are all the "aesthetic" reporting functions that will only run if the
// log level is to verbose.  These provide additional information about the
// compilation process to the user so as to make the compiler more friendly.

// LogCompileHeader reports the pre-compilation header: information about the
// compiler's current configuration (version, target, caching, etc.).
func ReportCompileHeader(target string, caching bool) {
	if reporter.LogLevel == LogLevelVerbose {
		displayCompileHeader(target, caching)
	}
}

// ReportBeginPhase reports the beginning of a compilation phase.
func ReportBeginPhase(phase string) {
	if reporter.LogLevel == LogLevelVerbose {
		displayBeginPhase(phase)
	}
}

// ReportEndPhase reports the conclusion of the current compilation phase.
func ReportEndPhase() {
	if reporter.LogLevel == LogLevelVerbose {
		displayEndPhase(ShouldProceed())
	}
}

// ReportCompilationFinished reports the concluding message for compilation.
// This displays information about the end of the compilation process.
func ReportCompilationFinished() {
	// log all warnings
	if reporter.LogLevel >= LogLevelWarning {
		for _, warning := range reporter.warnings {
			warning.display()
		}
	}

	// log closing message
	if reporter.LogLevel == LogLevelVerbose {
		displayCompilationFinished(ShouldProceed(), reporter.errorCount, len(reporter.warnings))
	}
}
