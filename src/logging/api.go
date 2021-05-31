package logging

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
	case "warning":
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
	return logger.ErrorCount == 0
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

}
