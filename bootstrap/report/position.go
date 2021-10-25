package report

// TextPosition represents a positional range in the source text
type TextPosition struct {
	StartLn, StartCol int // starting line, starting 0-indexed column
	EndLn, EndCol     int // ending Line, column trailing token (one over)
}

// TextPositionFromRange takes two positions and computes the text position
// spanning them.
func TextPositionFromRange(start, end *TextPosition) *TextPosition {
	return &TextPosition{
		StartLn:  start.StartLn,
		StartCol: start.StartCol,
		EndLn:    end.EndLn,
		EndCol:   end.EndCol,
	}
}
