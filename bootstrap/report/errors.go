package report

import (
	"fmt"
	"os"
)

// TextSpan represents a range or "span" of source text. It is used to specify
// erroneous or otherwise significant source text in a Chai program.  Text spans
// are inclusive on both sides: the starting position is the position of the
// first character in the span and the ending position is the position of the
// last character in the span.  The line and column numbers are zero-indexed.
type TextSpan struct {
	// The line and column beginning the text span.
	StartLine, StartCol int

	// The line and column ending the text span.
	EndLine, EndCol int
}

// NewSpanOver returns a new text span which spans over and between the two
// given text spans.
func NewSpanOver(start, end *TextSpan) *TextSpan {
	return &TextSpan{
		StartLine: start.StartLine,
		StartCol:  start.StartCol,
		EndLine:   end.EndLine,
		EndCol:    end.EndCol,
	}
}

// -----------------------------------------------------------------------------

// LocalCompileError is a compilation error that occurs in a context in which
// the file is known by the error handler and thus doesn't need to be passed
// along with the error.
type LocalCompileError struct {
	// The error message.
	Message string

	// The span over which the error occurs.
	Span *TextSpan
}

func (lce *LocalCompileError) Error() string {
	return lce.Message
}

// Raise creates a new local compile error.
func Raise(span *TextSpan, msg string, args ...interface{}) *LocalCompileError {
	return &LocalCompileError{Message: fmt.Sprintf(msg, args...), Span: span}
}

// -----------------------------------------------------------------------------

// ReportICE reports an internal compiler error.  These are errors that
// specifically result for a bug or unexpected condition occurring with the
// compiler: they are not intended to ever happen.  These errors are always
// displayed regardless of log level.
func ReportICE(message string, args ...interface{}) {
	rep.m.Lock()
	defer rep.m.Unlock()

	displayICE(fmt.Sprintf(message, args...))

	os.Exit(-1)
}

// ReportFatal reports a fatal error.  These are errors that should cause all
// compilation to stop immediately.  However, they are expected errors that
// generally result from invalid configuration of some form: missing CHAI_PATH,
// can't find requisite tools (eg. `link.exe`), etc.
func ReportFatal(message string, args ...interface{}) {
	if rep.logLevel > LogLevelSilent {
		rep.m.Lock()
		defer rep.m.Unlock()

		displayFatal(fmt.Sprintf(message, args...))
	}

	os.Exit(1)
}

// ReportCompileError reports a compilation error: ie. erroneous input code. The
// absPath is the absolute path to the erroneous source file. The reprPath is
// the representative path to the erroneous source file: it is the file's
// ReprPath field.  The span may be nil in which case no position information
// will be printed.
func ReportCompileError(absPath, reprPath string, span *TextSpan, message string, args ...interface{}) {
	if rep.logLevel > LogLevelSilent {
		rep.m.Lock()
		defer rep.m.Unlock()

		rep.isErr = true

		displayCompileMessage("error", absPath, reprPath, span, fmt.Sprintf(message, args...))
	}
}

// ReportCompileWarning reports a compilation warning.  The arguments are of the
// same form as those to ReportCompileError.
func ReportCompileWarning(absPath, reprPath string, span *TextSpan, message string, args ...interface{}) {
	if rep.logLevel > LogLevelWarn {
		rep.m.Lock()
		defer rep.m.Unlock()

		displayCompileMessage("warning", absPath, reprPath, span, fmt.Sprintf(message, args...))
	}
}

// ReportStdError reports a non-fatal, standard Go error.
func ReportStdError(reprPath string, err error) {
	if rep.logLevel > LogLevelError {
		rep.m.Lock()
		defer rep.m.Unlock()

		rep.isErr = true

		displayStdError(reprPath, err)
	}
}

// -----------------------------------------------------------------------------

// AnyErrors returns whether or not any errors were detected.
func AnyErrors() bool {
	return rep.isErr
}

// -----------------------------------------------------------------------------

// CatchErrors catches any errors thrown by a `panic` during a stage of
// compilation. In effect, this handler determines when any errors
// "unrecoverable" within a given subsection of the compiler should stop
// bubbling.
// NB: This function must ALWAYS be deferred.
func CatchErrors(absPath, reprPath string) {
	if x := recover(); x != nil {
		if cerr, ok := x.(*LocalCompileError); ok {
			ReportCompileError(
				absPath,
				reprPath,
				cerr.Span,
				cerr.Message,
			)
		} else if serr, ok := x.(error); ok {
			ReportStdError(reprPath, serr)
		} else {
			ReportFatal("%s", x)
		}
	}
}
