package logging

import (
	"bufio"
	"chai/common"
	"errors"
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

// PrintErrorMessage prints a standard Go error to the console
func PrintErrorMessage(tag string, err error) {
	ErrorStyleBG.Print(tag)
	ErrorColorFG.Println(" " + err.Error())
}

// PrintWarningMessage prints a warning message to the console
func PrintWarningMessage(tag, msg string) {
	WarnStyleBG.Print(tag)
	WarnColorFG.Println(" " + msg)
}

// PrintInfoMessage prints an informational message to the user
func PrintInfoMessage(tag, msg string) {
	InfoStyleBG.Print(tag)
	InfoColorFG.Println(" " + msg)
}

// -----------------------------------------------------------------------------
// This section contains all the display functions for the different kinds of
// errors that can be logged -- these functions are called to print the error to
// the screen.

func (ce *ConfigError) display() {
	PrintErrorMessage(ce.Kind+" Error", errors.New(ce.Message))
}

var compileMsgStrings = map[int]string{
	LMKAnnot:    "Annotation",
	LMKArg:      "Argument",
	LMKTyping:   "Type",
	LMKDef:      "Definition",
	LMKGeneric:  "Generic",
	LMKImmut:    "Mutability",
	LMKImport:   "Import",
	LMKClass:    "Type Class",
	LMKMetadata: "Metadata",
	LMKName:     "Name",
	LMKProp:     "Property",
	LMKSyntax:   "Syntax",
	LMKToken:    "Token",
	LMKUsage:    "Usage",
	LMKUser:     "User",
	LMKOperApp:  "Operator",
	LMKPattern:  "Pattern",
}

func (cm *CompileMessage) display() {
	cm.displayBanner()
	fmt.Println(cm.Message)

	if cm.Position != nil {
		cm.displayCodeSelection()
	}
}

// displayBanner displays the banner on top of all compilation messages
func (cm *CompileMessage) displayBanner() {
	fmt.Print("\n\n-- ")
	kindStr := compileMsgStrings[cm.Kind]
	kindLen := len(kindStr)
	if cm.isError() {
		ErrorStyleBG.Print(kindStr + " Error")
		kindLen += 7
	} else {
		WarnStyleBG.Print(kindStr + " Warning")
		kindLen += 9
	}

	fmt.Print(" ")

	fileName := filepath.Base(cm.Context.FilePath)
	bannerLen := pterm.GetTerminalWidth() / 2
	if bannerLen > 50 {
		bannerLen = 50
	}
	dashCount := bannerLen - len(fileName) - kindLen - 1

	fmt.Print(strings.Repeat("-", dashCount) + " ")
	InfoColorFG.Println(fileName)
}

// displayCodeSelection displays the erroneous code (with line numbers) and
// highlights the appropriate sections
func (cm *CompileMessage) displayCodeSelection() {
	fmt.Println()

	// this open operation should always succeed
	f, err := os.Open(cm.Context.FilePath)
	if err != nil {
		LogFatal("failed to open file to display error message")
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
			lines[lineNumber-cm.Position.StartLn] = sc.Text()
		}
	}

	// calculate whitespace to trim
	minWhitespace := -1
	for _, line := range lines {
		leadingWhitespace := 0
		for _, c := range line {
			if c == ' ' {
				leadingWhitespace++
			} else if c == '\t' {
				leadingWhitespace += 4
			} else {
				break
			}
		}

		if minWhitespace == -1 {
			minWhitespace = leadingWhitespace
		} else if minWhitespace > leadingWhitespace {
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
		fmt.Println(strings.ReplaceAll(line, "\t", "    ")[minWhitespace:])

		// print the carrets
		fmt.Print(strings.Repeat(" ", maxLineNumberWidth), "|  ")
		if i == 0 {
			fmt.Print(strings.Repeat(" ", cm.Position.StartCol-minWhitespace))

			// if the selection is one line long then we don't print carrets to
			// the end of the line; if it isn't, then we print carrets to the
			// end
			if i == len(lines)-1 {
				ErrorColorFG.Print(strings.Repeat("^", cm.Position.EndCol-cm.Position.StartCol))
				fmt.Println()
			} else {
				ErrorColorFG.Println(strings.Repeat("^", len(line)-cm.Position.StartCol-minWhitespace))
			}
		} else if i == len(lines)-1 {
			// if we are at the last line, we print carrets until the end column
			// and then stop
			ErrorColorFG.Println(strings.Repeat("^", cm.Position.EndCol-minWhitespace))
		} else {
			// if we are in the middle of highlighting then we simply fill the
			// line with carrets
			ErrorColorFG.Println(strings.Repeat("^", len(line)-minWhitespace))
		}
	}

	fmt.Println()
}

const fatalErrorPostlude = `
This is likely a bug in the compiler.
Please open an issue on Github: github.com/ComedicChimera/chai`

func displayFatalError(msg string) {
	fmt.Print("\n\n")
	ErrorStyleBG.Print("Fatal Error ")
	ErrorColorFG.Println(msg)
	InfoColorFG.Println(fatalErrorPostlude)
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

const maxPhaseLength = len("Transforming")

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
