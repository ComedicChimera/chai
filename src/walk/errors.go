package walk

import (
	"chai/logging"
	"fmt"
)

// logError logs a compile error in the current file
func (w *Walker) logError(msg string, kind int, pos *logging.TextPosition) {
	logging.LogCompileError(
		w.SrcFile.LogContext,
		msg,
		kind,
		pos,
	)
}

// logWarning logs a compile warning in the current file
func (w *Walker) logWarning(msg string, kind int, pos *logging.TextPosition) {
	logging.LogCompileWarning(
		w.SrcFile.LogContext,
		msg,
		kind,
		pos,
	)
}

// -----------------------------------------------------------------------------

func (w *Walker) logRepeatDef(name string, pos *logging.TextPosition) {
	w.logError(
		fmt.Sprintf("symbol `%s` declared multiple times in scope", name),
		logging.LMKName,
		pos,
	)
}
