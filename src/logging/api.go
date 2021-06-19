package logging

import "os"

// logger is a global reference to a shared Logger (created/initialized with the
// compiler, but separated for general usage)
var logger Logger

// Initialize initializes the global logger with the provided log level
func Initialize(buildPath string, loglevelname string) {
	var loglevel int
	switch loglevelname {
	case "silent":
		loglevel = LogLevelSilent
	case "error":
		loglevel = LogLevelError
	case "warn":
		loglevel = LogLevelWarning
	// everything else (including invalid log levels) should default to verbose
	default:
		loglevel = LogLevelVerbose
	}

	logger = newLogger(buildPath, loglevel)
}

// ShouldProceed indicates whether or not the log module has encountered an errors.
// This is useful for sections of the compiler where multiple items are processed
// concurrently and having an error accumulator would be practical
func ShouldProceed() bool {
	return logger.errorCount == 0
}

// -----------------------------------------------------------------------------
// NOTE: All log functions will only display if the appropriate log level is
// set.  Most log functions will simply fail silently if below their appropriate
// log level.

// LogCompileError logs and a compilation error (user-induced, bad code)
func LogCompileError(lctx *LogContext, message string, kind int, pos *TextPosition) {
	logger.handleMsg(&CompileMessage{
		Message:  message,
		Kind:     kind,
		Position: pos,
		Context:  lctx,
		IsError:  true,
	})
}

// LogCompileWarning logs a compilation warning (user-induced, problematic code)
func LogCompileWarning(lctx *LogContext, message string, kind int, pos *TextPosition) {
	logger.handleMsg(&CompileMessage{
		Message:  message,
		Kind:     kind,
		Position: pos,
		Context:  lctx,
		IsError:  false,
	})
}

// LogConfigError logs an error related to project or compiler configuration
func LogConfigError(kind, message string) {
	logger.handleMsg(&ConfigError{Kind: kind, Message: message})
}

// LogBuildWarning logs a warning in the build process
func LogBuildWarning(kind, warning string) {
	// TODO
}

// LogFatal logs a fatal compilation error that was not expected: ie. the
// compiler did something it wasn't supposed to.
func LogFatal(message string) {
	displayFatalError(message)
	os.Exit(1)
}

// -----------------------------------------------------------------------------
// Below are all the "aesthetic" log functions that will only run if the log
// level is to verbose.  These provide additional information about the
// compilation process to the user so as to make the compiler more friendly.

// LogCompileHeader logs the pre-compilation header: information about the
// compiler's current configuration (version, target, caching, etc.)
func LogCompileHeader(target string, caching bool) {
	if logger.LogLevel == LogLevelVerbose {
		displayCompileHeader(target, caching)
	}
}

// LogBeginPhase logs the beginning of a compilation phase
func LogBeginPhase(phase string) {
	if logger.LogLevel == LogLevelVerbose {
		displayBeginPhase(phase)
	}
}

// LogEndPhase logs the conclusion of the current compilation phase
func LogEndPhase() {
	if logger.LogLevel == LogLevelVerbose {
		displayEndPhase(ShouldProceed())
	}
}

// LogCompilationFinished logs the concluding message for compilation. This
// displays information about the end of the compilation process.
func LogCompilationFinished() {
	// log all warnings
	if logger.LogLevel >= LogLevelWarning {
		for _, warning := range logger.warnings {
			warning.display()
		}
	}

	// log closing message
	if logger.LogLevel == LogLevelVerbose {
		displayCompilationFinished(ShouldProceed(), logger.errorCount, len(logger.warnings))
	}
}
