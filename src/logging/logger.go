package logging

import (
	"sync"
)

// Logger is a type that is responsible for storing and logging output from the
// compiler as necessary
type Logger struct {
	errorCount int // Total encountered errors
	LogLevel   int

	// warnings is a list of all warnings to be logged at the end of compilation
	warnings []LogMessage

	// buildPath is used to shorten display paths in errors
	buildPath string

	// m is the mutex used to synchonize the printing of error messages
	m *sync.Mutex
}

// Enumeration of the different log levels
const (
	LogLevelSilent  = iota // no output at all
	LogLevelError          // only errors and closing compilation notification (success/fail)
	LogLevelWarning        // errors, warnings, and closing message
	LogLevelVerbose        // errors, warnings, compiler version and progress summary, closing message (DEFAULT)
)

// newLogger creates a new logger struct
func newLogger(buildPath string, loglevel int) Logger {
	return Logger{
		buildPath: buildPath,
		LogLevel:  loglevel,
		m:         &sync.Mutex{},
	}
}

// handleMsg prompts to logger to process a message -- this message could be
// coming in concurrently and so we need to make sure we are not printing multiple
// things at the same time so we there is a mutex in place for this function
func (l *Logger) handleMsg(lm LogMessage) {
	l.m.Lock()

	if lm.isError() {
		l.errorCount++

		if l.LogLevel > LogLevelSilent {
			displayEndPhase(false)
			lm.display()
		}
	} else {
		l.warnings = append(l.warnings, lm)
	}

	l.m.Unlock()
}
