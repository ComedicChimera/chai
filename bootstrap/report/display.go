package report

import (
	"bufio"
	"chai/common"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

var (
	SuccessColorFG = pterm.FgLightGreen
	SuccessStyleBG = pterm.NewStyle(pterm.BgLightGreen, pterm.FgBlack)
	WarnColorFG    = pterm.FgYellow
	WarnStyleBG    = pterm.NewStyle(pterm.BgYellow, pterm.FgBlack)
	ErrorColorFG   = pterm.FgRed
	ErrorStyleBG   = pterm.NewStyle(pterm.BgRed, pterm.FgWhite)
	InfoColorFG    = SuccessColorFG
	InfoStyleBG    = SuccessStyleBG
)

// ReportInfoMessage displays a message about some information the user
// requested (eg. the current Chai version)
func DisplayInfoMessage(dataName, msg string) {
	InfoStyleBG.Print(dataName)
	InfoColorFG.Println(msg)
}

// -----------------------------------------------------------------------------
// This section contains all the display functions for the different kinds of
// errors that can be logged -- these functions are called to print the error to
// the screen.

func (me *ModuleMessage) display() {
	ErrorStyleBG.Print("Module Error")
	ErrorColorFG.Printf("in module `%s`: %s\n", me.ModName, me.Message)
}

func (cm *CompileMessage) display() {
	cm.displayHeader()

	if cm.Position != nil {
		cm.displayCodeSelection()
	}
}

func displayFatalError(msg string) {
	fmt.Print("\n\n")
	ErrorStyleBG.Print("Fatal Error")
	ErrorColorFG.Println(" " + msg)
}

// -----------------------------------------------------------------------------

// displayBanner displays the header of the compilation message.
func (cm *CompileMessage) displayHeader() {
	if cm.isError() {
		ErrorStyleBG.Print("Compile Error")
	} else {
		WarnStyleBG.Print("Compile Warning")
	}

	// sanitize message
	msg := strings.ReplaceAll(cm.Message, "\r", "\\r")
	msg = strings.ReplaceAll(msg, "\n", "\\n")

	ErrorColorFG.Printf(" [%s] %s:(%d, %d): %s\n", cm.Context.ModName, cm.Context.FileRelPath, cm.Position.StartLn, cm.Position.StartCol, msg)
}

// displayCodeSelection displays the erroneous code (with line numbers) and
// highlights the appropriate sections
func (cm *CompileMessage) displayCodeSelection() {
	fmt.Println()

	// this open operation should always succeed
	f, err := os.Open(filepath.Join(cm.Context.ModAbsPath, cm.Context.FileRelPath))
	if err != nil {
		ReportFatal("failed to open file to display error message")
	}
	defer f.Close()

	// read the file line by line until we encounter the selected lines; capture
	// the lines first so we can determine how much whitespace to trim before
	// printing
	sc := bufio.NewScanner(f)
	sc.Split(bufio.ScanLines)
	lines := make([]string, cm.Position.EndLn-cm.Position.StartLn+1)
	for lineNumber := 1; sc.Scan(); lineNumber++ {
		if lineNumber >= cm.Position.StartLn && lineNumber <= cm.Position.EndLn {
			// replace all tabs with spaces so we can easily determine how much
			// whitespace to print and how many carrets to add
			lines[lineNumber-cm.Position.StartLn] = strings.ReplaceAll(sc.Text(), "\t", "    ")
		}
	}

	// calculate whitespace to trim
	minWhitespace := -1
	for _, line := range lines {
		leadingWhitespace := 0
		for _, c := range line {
			if c == ' ' {
				leadingWhitespace++
			} else {
				break
			}
		}

		if minWhitespace == -1 || minWhitespace > leadingWhitespace {
			minWhitespace = leadingWhitespace
		}
	}

	// calculate the amount to pad line numbers by and use it to build a padding
	// format string (so we can use it to print out line numbers neatly)
	maxLineNumberWidth := len(strconv.Itoa(cm.Position.EndLn)) + 1
	lineNumberFmtStr := "%-" + strconv.Itoa(maxLineNumberWidth) + "v"

	// print each line followed by the line of selecting underscores
	for i, line := range lines {
		// print the line number and the line itself (trimmed)
		InfoColorFG.Print(fmt.Sprintf(lineNumberFmtStr, i+cm.Position.StartLn))
		fmt.Print("|  ")
		fmt.Println(line[minWhitespace:])

		// print the carrets
		fmt.Print(strings.Repeat(" ", maxLineNumberWidth), "|  ")
		if i == 0 {
			fmt.Print(strings.Repeat(" ", cm.Position.StartCol-minWhitespace-1))

			// if the selection is one line long then we don't print carrets to
			// the end of the line; if it isn't, then we print carrets to the
			// end
			if i == len(lines)-1 {
				ErrorColorFG.Print(strings.Repeat("^", cm.Position.EndCol-cm.Position.StartCol))
				fmt.Println()
			} else {
				ErrorColorFG.Println(strings.Repeat("^", len(line)-cm.Position.StartCol+1))
			}
		} else if i == len(lines)-1 {
			// if we are at the last line, we print carrets until the end column
			// and then stop
			ErrorColorFG.Println(strings.Repeat("^", cm.Position.EndCol-minWhitespace-1))
		} else {
			// if we are in the middle of highlighting then we simply fill the
			// line with carrets
			ErrorColorFG.Println(strings.Repeat("^", len(line)-minWhitespace))
		}
	}

	fmt.Println()
}

// -----------------------------------------------------------------------------

// displayCompileHeader displays all the compiler information before starting compilation
func displayCompileHeader(target string, caching bool) {
	fmt.Print("chai ")
	InfoColorFG.Print("v" + common.ChaiVersion)
	fmt.Print(" -- target: ")
	InfoColorFG.Println(target)

	if caching {
		fmt.Println("compiling using cache")
	}
}

// phaseSpinner stores the current phase spinner
var phaseSpinner *pterm.SpinnerPrinter
var currentPhase string
var phaseStartTime time.Time

const maxPhaseLength = len("Generating")

// displayBeginPhase displays the beginning of a compilation phase
func displayBeginPhase(phase string) {
	currentPhase = phase
	phaseText := phase + "..." + strings.Repeat(" ", maxPhaseLength-len(phase)+2)
	phaseSpinner = pterm.DefaultSpinner.WithStyle(pterm.NewStyle(InfoColorFG))

	phaseSpinner.SuccessPrinter = &pterm.PrefixPrinter{
		MessageStyle: pterm.NewStyle(pterm.FgDefault),
		Prefix: pterm.Prefix{
			Style: SuccessStyleBG,
			Text:  "Done",
		},
	}

	phaseSpinner.FailPrinter = &pterm.PrefixPrinter{
		MessageStyle: pterm.NewStyle(pterm.FgDefault),
		Prefix: pterm.Prefix{
			Style: ErrorStyleBG,
			Text:  "Fail",
		},
	}

	phaseSpinner.Start(phaseText)
	phaseStartTime = time.Now()
}

// displayEndPhase displays the end of a compilation phase
func displayEndPhase(success bool) {
	if phaseSpinner != nil {
		if success {
			phaseSpinner.Success(
				currentPhase+strings.Repeat(" ", maxPhaseLength-len(currentPhase)+2),
				fmt.Sprintf("(%.3fs)", time.Since(phaseStartTime).Seconds()),
			)
		} else {
			phaseSpinner.Fail(currentPhase + strings.Repeat(" ", maxPhaseLength-len(currentPhase)+2))
		}

		phaseSpinner = nil
	}
}

// displayCompilationFinished displays a compilation finished message
func displayCompilationFinished(success bool, errorCount, warningCount int) {
	fmt.Print("\n")

	if success {
		SuccessColorFG.Print("All done! ")
	} else {
		ErrorColorFG.Print("Oh no! ")
	}

	fmt.Print("(")

	switch errorCount {
	case 0:
		SuccessColorFG.Print(0)
		fmt.Print(" errors, ")
	case 1:
		ErrorColorFG.Print(1)
		fmt.Print(" error, ")
	default:
		ErrorColorFG.Print(errorCount)
		fmt.Print(" errors, ")
	}

	switch warningCount {
	case 0:
		SuccessColorFG.Print(0)
		fmt.Println(" warnings)")
	case 1:
		WarnColorFG.Print(1)
		fmt.Println(" warning)")
	default:
		WarnColorFG.Print(warningCount)
		fmt.Println(" warnings)")
	}
}
