package report

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// displayICE displays an internal compiler error message.
func displayICE(message string) {
	fmt.Printf("internal compiler error: %s\n", message)
	fmt.Print("This error was not supposed to happen: please open an issue on GitHub at [insert link]\n\n")
}

// displayFatal displays a fatal error message.
func displayFatal(message string) {
	fmt.Printf("fatal error: %s\n\n", message)
}

// displayCompileMessage displays a compilation error or warning.  The label is
// the string to prefix the message with: eg. if we want to display an error,
// the label is "error".
func displayCompileMessage(label, absPath, reprPath string, span *TextSpan, message string) {
	if span == nil {
		fmt.Printf("%s: %s: %s\n\n", reprPath, label, message)
	} else {
		fmt.Printf("%s:%d:%d: %s: %s\n\n", reprPath, span.StartLine+1, span.StartCol+1, label, message)
		displaySourceText(absPath, span)
	}
}

// displayStdError displays a standard Go error.
func displayStdError(reprPath string, err error) {
	fmt.Printf("%s: error: %s\n\n", reprPath, err)
}

// -----------------------------------------------------------------------------

// displaySourceText displays a segment of source text defined by a text span.
func displaySourceText(absPath string, span *TextSpan) {
	// Open the file so we can read the desired source text.
	file, err := os.Open(absPath)
	if err != nil {
		displayICE(fmt.Sprintf("failed to open file %s for reporting: %s\n", absPath, err))
		os.Exit(-1)
	}
	defer file.Close()

	// Collect all the source lines containing the given source text.
	var lines []string
	sc := bufio.NewScanner(file)
	for ln := 0; sc.Scan(); ln++ {
		if span.StartLine <= ln && ln <= span.EndLine {
			lines = append(lines, strings.ReplaceAll(sc.Text(), "\t", "    "))
		}
	}

	if err := sc.Err(); err != nil {
		displayICE(fmt.Sprintf("failed to read file %s for reporting: %s\n", absPath, err))
		os.Exit(-1)
	}

	// Calculate the minimum line indentation.
	minIndent := math.MaxInt
	for _, line := range lines {
		lineIndent := 0
		for _, c := range line {
			if c == ' ' {
				lineIndent++
			} else {
				break
			}
		}

		if lineIndent < minIndent {
			minIndent = lineIndent
		}
	}

	// Calculate the maximum line number length.
	maxLineNumLen := len(strconv.Itoa(span.EndLine + 1))

	// Generate the format string for line numbers.
	lineNumFmtStr := "%-" + strconv.Itoa(maxLineNumLen) + "v | "

	for i, line := range lines {
		// Print the line number and separator bar.
		fmt.Printf(lineNumFmtStr, i+span.StartLine+1)

		// Print the source text with the leading indent trimmed off.
		fmt.Println(line[minIndent:])

		// Print the line and bar used for the line for carret underlining.
		fmt.Print(strings.Repeat(" ", maxLineNumLen), " | ")

		// Calculate the number of spaces before carret underlining begins. For
		// any line which is not the starting line, this is always zero since
		// the underlining is always continuing from the previous line. For all
		// other lines, it is start column - the minimum indent.
		var carretPrefixCount int
		if i == 0 {
			carretPrefixCount = span.StartCol - minIndent
		} else {
			carretPrefixCount = 0
		}

		// Calculate the number of characters at the end of the source line that
		// should not be highlighted.  For all lines except the last line, this
		// is zero, since underlining should span until the end of the line and
		// over onto the next line.  For the last line, it is length of the line
		// - the end column of the errorenous source text.
		var carretSuffixCount int
		if i == len(lines)-1 {
			carretSuffixCount = len(line) - span.EndCol
		} else {
			carretSuffixCount = 0
		}

		// Print the number of spaces that come before the carret (ie. skip
		// underlining until the start column).
		fmt.Print(strings.Repeat(" ", carretPrefixCount))

		// Print the underlining carrets for the given line.  This number is
		// always the length of the line - (the number of carret that are
		// skipped at the start + the number of carrets that are skipped at the
		// end).  We also subtract off the minimum indentation to account for
		// the fact that that part of the line is neglected by the prefix count
		// (ie. to cancel the - minimum indent inside the calculate for carret
		// prefix count).
		fmt.Println(strings.Repeat("^", len(line)-carretSuffixCount-carretPrefixCount-minIndent))
	}

	// Print newlines after the error message.
	fmt.Println()
}
