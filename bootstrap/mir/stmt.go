package mir

import "chaic/report"

// Statement represents a MIR statement.
type Statement interface {
	// Span returns the source position of the statement.
	Span() *report.TextSpan
}
