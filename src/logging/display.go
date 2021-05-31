package logging

import "github.com/pterm/pterm"

// PrintBlockMessage prints a message with text in a block before it
func PrintBlockMessage(message, tag string, fgColor, bgColor pterm.Color) {
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(bgColor)).Print(tag)
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(fgColor)).Println(" ", message)
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
