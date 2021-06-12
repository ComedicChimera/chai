package walk

import "chai/logging"

// logError logs a compile error
func (w *Walker) logError(msg string, kind int, pos *logging.TextPosition) {
	logging.LogCompileError(
		w.SrcFile.LogContext,
		msg,
		kind,
		pos,
	)
}
