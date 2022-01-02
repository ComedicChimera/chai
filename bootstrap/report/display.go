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

func (mm *ModuleMessage) display() {
	if mm.isError() {
		ErrorStyleBG.Print("Module Error")
		ErrorColorFG.Printf(" [%s]: %s\n", mm.ModName, mm.Message)
	} else {
		WarnStyleBG.Print("Module Warning")
		WarnColorFG.Printf(" [%s]: %s\n", mm.ModName, mm.Message)
	}
}

func (cm *CompileMessage) display() {
	cm.displayHeader()

	if cm.Position != nil {
		cm.displayCodeSelection()
	}
}

func (pe *PackageError) display() {
	ErrorStyleBG.Print("Package Error")
	ErrorColorFG.Printf(" [%s]: %s\n", pe.ModRelPath, pe.Message)
}

func displayFatalError(msg string) {
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

	// end of file
	if cm.Position == nil {
		ErrorColorFG.Printf(" [%s] %s: %s\n", cm.Context.ModName, cm.Context.FileRelPath, msg)
	} else {
		ErrorColorFG.Printf(" [%s] %s:(%d, %d): %s\n", cm.Context.ModName, cm.Context.FileRelPath, cm.Position.StartLn, cm.Position.StartCol, msg)
	}

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
		if i == len(lines)-1 && cm.Position.EndCol == 1 {
			break
		}

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
			} else if cm.Position.StartCol == len(line)+1 {
				ErrorColorFG.Println("^")
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

	fmt.Println()
}

// displayCompilationFinished displays a compilation finished message
func displayCompilationFinished(success bool, outputPath string) {
	fmt.Print("\n")

	if success {
		SuccessColorFG.Print("All Done! ")
	} else {
		ErrorColorFG.Print("Oh no! ")
	}

	// display compilation time if successful
	if success {
		fmt.Print("[")
		InfoColorFG.Printf("%.3fms", float64(time.Since(rep.startTime).Nanoseconds())/1e6)
		fmt.Print("] ")
	}

	// display errors and warnings
	fmt.Print("(")

	switch rep.errorCount {
	case 0:
		SuccessColorFG.Print(0)
		fmt.Print(" errors, ")
	case 1:
		ErrorColorFG.Print(1)
		fmt.Print(" error, ")
	default:
		ErrorColorFG.Print(rep.errorCount)
		fmt.Print(" errors, ")
	}

	switch len(rep.warnings) {
	case 0:
		SuccessColorFG.Print(0)
		fmt.Println(" warnings)")
	case 1:
		WarnColorFG.Print(1)
		fmt.Println(" warning)")
	default:
		WarnColorFG.Print(len(rep.warnings))
		fmt.Println(" warnings)")
	}

	// display output path
	if success {
		fmt.Print("Output written to: ")
		wd, err := os.Getwd()
		if err != nil {
			ReportFatal("error getting working directory: %s", err.Error())
		}

		relpath, err := filepath.Rel(wd, outputPath)
		if err != nil {
			ReportFatal("error calculating relative path to output dir: %s", err.Error())
		}
		InfoColorFG.Println(relpath)
	}
}
