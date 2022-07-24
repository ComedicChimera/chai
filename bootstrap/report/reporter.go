package report

import "sync"

// Reporter is responsible for reporting errors, warnings, and other kinds of
// messages to the user during program execution.  The reporter respects the set
// log level and is synchronized: its methods can be safely called from multiple
// goroutines.
type Reporter struct {
	// The mutex used to synchonize different error method calls.
	m *sync.Mutex

	// The selected log level of the reporter.  This must be one of the
	// enumerated log levels below.
	logLevel int

	// Indicates whether or not an error has been detected.
	isErr bool
}

// Enumeration of the different possible log levels.
const (
	LogLevelSilent  = iota // Displays no output.
	LogLevelError          // Displays only errors to the user.
	LogLevelWarn           // Displays only warnings and errors to the user.
	LogLevelVerbose        // Displays all compilation messages to the user (default).
)

// rep is the global reporter instance.
var rep *Reporter

// InitReporter initializes the global error reporter to the given log level. If
// the reporter has already been initialized, this function does nothing.
func InitReporter(logLevel int) {
	if rep == nil {
		rep = &Reporter{
			m:        &sync.Mutex{},
			logLevel: logLevel,
			isErr:    false,
		}
	}
}
