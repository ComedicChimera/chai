package logging

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
)

// PrintBlockMessage prints a message with text in a block before it
func PrintBlockMessage(message, tag string, fgColor, bgColor pterm.Color) {
	bgColor.Print(tag)
	fgColor.Println(" ", message)
}

// PrintCLIError prints an error in using the CLI
func PrintCLIError(errMsg string) {
	PrintBlockMessage(errMsg, "Error", pterm.FgLightRed, pterm.BgLightRed)
}

// -----------------------------------------------------------------------------
// This section contains all the display functions for the different kinds of
// errors that can be logged -- these functions are called to print the error to
// the screen.

func (ce *ConfigError) display() {
	PrintBlockMessage(ce.Message, ce.Kind+" Error", pterm.FgLightRed, pterm.BgLightRed)
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
		pterm.BgLightRed.Print(kindStr + " Error")
		kindLen += 7
	} else {
		pterm.BgLightYellow.Print(kindStr + " Warning")
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
	pterm.FgCyan.Println(fileName)
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
		pterm.FgLightGreen.Print(fmt.Sprintf(lineNumberFmtStr, i+cm.Position.StartLn))
		fmt.Print("|  ")
		fmt.Println(strings.ReplaceAll(line, "\t", "    ")[minWhitespace:])

		// print the carrets
		fmt.Print(strings.Repeat(" ", maxLineNumberWidth), "|  ")
		if i == 0 {
			fmt.Print(strings.Repeat(" ", cm.Position.StartCol-minWhitespace))

			// if the selection is one line long then we don't print carrots to
			// the end of the line; if it isn't, then we print carrets to the
			// end
			if i == len(lines)-1 {
				pterm.FgRed.Print(strings.Repeat("^", cm.Position.EndCol-cm.Position.StartCol))
				fmt.Println()
			} else {
				pterm.FgRed.Println(strings.Repeat("^", len(line)-cm.Position.StartCol-minWhitespace))
			}
		} else if i == len(lines)-1 {
			// if we are at the last line, we print carrets until the end column
			// and then stop
			pterm.FgRed.Println(strings.Repeat("^", cm.Position.EndCol-minWhitespace))
		} else {
			// if we are in the middle of highlighting then we simply fill the
			// line with carrets
			pterm.FgRed.Println(strings.Repeat("^", len(line)-minWhitespace))
		}
	}

	fmt.Println()
}
