package report

import (
	"sync"
)

// Reporter is a type that is responsible for storing and logging output from
// the compiler as necessary: reporting information to the user.
type Reporter struct {
	errorCount int // Total encountered errors
	LogLevel   int

	// warnings is a list of all warnings to be logged at the end of compilation
	warnings []Message

	// m is the mutex used to synchonize the printing of error messages
	m *sync.Mutex
}

// Enumeration of the different log levels.
const (
	LogLevelSilent  = iota // no output at all
	LogLevelError          // only errors and closing compilation notification (success/fail)
	LogLevelWarning        // errors, warnings, and closing message
	LogLevelVerbose        // errors, warnings, compiler version and progress summary, closing message (DEFAULT)
)

// handleMsg prompts to reporter to process a message -- this message could be
// coming in concurrently and so we need to make sure we are not printing
// multiple things at the same time so we there is a mutex in place for this
// function.
func (r *Reporter) handleMsg(m Message) {
	r.m.Lock()

	if m.isError() {
		r.errorCount++

		if r.LogLevel > LogLevelSilent {
			displayEndPhase(false)
			m.display()
		}
	} else {
		r.warnings = append(r.warnings, m)
	}

	r.m.Unlock()
}
