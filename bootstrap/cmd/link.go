package cmd

import (
	"chaic/report"
	"chaic/wintool"
	"os"
	"os/exec"
	"runtime"
)

// linkExecutable creates a Chai executable from `objFilePaths` and an
// additional link objects specified by the user.
func (c *Compiler) linkExecutable(objFilePaths []string) {
	// TODO: add the ability to specify a linker by command-line arguments.

	// Determine the base link command to run.
	var linkCommand *exec.Cmd
	if runtime.GOOS == "windows" {
		// Get the target architecture.
		// targetArch := strings.Split(llvm.HostTriple(), "-")[0]
		targetArch := ""

		// Find the Windows linker if possible.
		lc, err := wintool.FindLink(targetArch)
		if err == nil {
			linkCommand = lc
		} else {
			report.ReportFatal(err.Error())
		}

		// Add the default Windows-specific link options to the linker.
		linkCommand.Args = append(
			linkCommand.Args,
			"/entry:_start",      // Set the entry point.
			"/subsystem:console", // Set the executable to be a console app.
			"/nologo",            // Turn off the logo banner for error reporting.
			"/out:"+c.outputPath, // Add the executable output path.

			// Requires System Libraries:
			"kernel32.lib",
		)
	} else {
		// By default, we will use `ld` on other platforms.  We don't need to do
		// anything to find `ld` since it should be in the PATH on all Unix
		// versions that want to compile Chai.
		linkCommand = exec.Command(
			"ld",
			"-e _start",        // Set the entry point.
			"-o "+c.outputPath, // Set the output path.
		)
	}

	// Add all our object files to the linker arguments.
	linkCommand.Args = append(linkCommand.Args, objFilePaths...)

	// Run the linker.
	out, err := linkCommand.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Exit error => we were able to find the linker, but there were
			// link errors.  We can just output those to the user.
			report.ReportFatal("link error:\n%s", string(out))
		} else {
			// Some other error: probably couldn't find the linker.
			report.ReportFatal("failed to run linker: %s", err)
		}
	}

	// Once linking is finished, we can remove all the object files: avoid
	// making a mess in the user's working directory.
	for _, objFilePath := range objFilePaths {
		if err := os.Remove(objFilePath); err != nil {
			report.ReportFatal("failed to delete produced object files: %s", err)
		}
	}
}
