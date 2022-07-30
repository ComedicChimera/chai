package wintool

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VSInstance represents an installed Visual Studio Instance.
type VSInstance struct {
	// The path to the installation.
	InstallPath string

	// The version string.
	Version string
}

// findVS15PlusInstances finds all of the installed VS 15+ instances. This
// principally done using the COM interface as specified here:
// https://devblogs.microsoft.com/setup/changes-to-visual-studio-15-setup/
// However, the COM interface doesn't always work and so this function will
// fallback to using VSwhere if the COM interface fails to locate the tools.
func findVS15PlusInstances(targetArch string) ([]VSInstance, bool) {
	if instances, ok := findVS15PlusInstancesUsingCOM(); ok {
		return instances, true
	}

	return findVS15PlusInstancesUsingVSWhere(targetArch)
}

// findVS15PlusInstancesUsingCOM attempts to find the VS 15+
// instances using the Windows COM API.
func findVS15PlusInstancesUsingCOM() ([]VSInstance, bool) {
	// TODO: I simply do not know how to implement this properly. Let some
	// Windiws Dev with actual knowledge of the COM APIs try to do this.
	return nil, false
}

// findVS15PlusInstancesUsingVSWhere attempts to find VS 15+ instances using
// `vswhere.exe`.
func findVS15PlusInstancesUsingVSWhere(targetArch string) ([]VSInstance, bool) {
	// Calculate the path to `vswhere.exe`.
	vswherePath := filepath.Join(
		os.Getenv("ProgramFiles(x86)"),
		"Microsoft Visual Studio/Installer/vswhere.exe",
	)

	// Validate that `vswhere.exe` exists.
	if _, err := os.Stat(vswherePath); err != nil {
		return nil, false
	}

	// Call `vswhere` with the appropriate arguments to find the location of the
	// installed MSVC build tools.
	output, err := exec.Command(
		vswherePath,
		"-latest",
		"-products", "*",
		"-requires", "Microsoft.VisualStudio.Component.VC.Tools."+llvmArchToVSArch[targetArch],
		"-format", "text",
		"-nologo",
	).Output()

	if err != nil {
		return nil, false
	}

	lines := strings.Split(string(output), "\n")

	var instance VSInstance
	for _, line := range lines {
		content := strings.Split(strings.TrimSpace(line), ": ")
		if len(content) != 2 {
			continue
		}

		key, value := content[0], content[1]

		switch key {
		case "installationPath":
			instance.InstallPath = value

			if len(instance.Version) > 0 {
				return []VSInstance{instance}, true
			}
		case "installationVersion":
			instance.Version = value

			if len(instance.InstallPath) > 0 {
				return []VSInstance{instance}, true
			}
		}
	}

	return nil, false
}
