package wintool

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// VCToolPaths represents a `link.exe` instance with all its VC paths.
type VCToolPaths struct {
	ToolPath    string
	BinPath     string
	DyLibPath   string
	LibPath     string
	IncludePath string
}

// FindLink attempts to locate the MSVC linker using the Windows registry as a
// point to search from. If successful, a preconstructed `exec.Cmd` will be
// returned properly configured with the correct environment variables needed to
// run the linker properly. Otherwise, an error is returned indicating that the
// linker couldn't be found.
func FindLink(targetArch string) (*exec.Cmd, error) {
	// This API is only actually usable on Windows.
	if runtime.GOOS != "windows" {
		return nil, errors.New("wintools API only works on Windows")
	}

	// Find the list of available VS instances.
	instances, ok := findVS15PlusInstances(targetArch)
	if !ok {
		// TODO: try finding VS14 instance
		return nil, errors.New("missing MSVC build tools")
	}

	// Find all versions of the `link.exe` tool keyed by version.
	toolVersions := make(map[string]*VCToolPaths)
	for _, instance := range instances {
		if tool, ok := findToolInVS15PlusInstance(instance.InstallPath, targetArch); ok {
			toolVersions[instance.Version] = tool
		}
	}

	// Determine which vctool to use based on its version.
	var vctool *VCToolPaths
	switch len(toolVersions) {
	case 0: // No matching tools found.
		return nil, errors.New("unable to locate `link.exe`")
	case 1: // One matching tool found.
		for _, itool := range toolVersions {
			vctool = itool
			break
		}
	default: // Multiple matching tools found.
		{
			// Find the tool with the latest version.
			var latestVersionN uint64 = 0

			for version, itool := range toolVersions {
				versionN := getVersionInt(version)
				if versionN > latestVersionN {
					vctool = itool
					latestVersionN = versionN
				}
			}
		}
	}

	// Create the tool command builder.
	toolBuilder := newToolCmdBuilder(vctool)

	// Add all the Windows sdks.
	addSDKs(toolBuilder, targetArch)

	// Create and return the command.
	return toolBuilder.ToCommand(), nil
}

// findToolInVS15PlusInstance attempts to find a tool stored in the given VS 15+
// instance and the desired target architecture.
func findToolInVS15PlusInstance(instancePath, targetArch string) (*VCToolPaths, bool) {
	// Calculate the path to the MSVC tools version file and make sure it exists.
	versionFilePath := filepath.Join(instancePath, "VC/Auxiliary/Build/Microsoft.VCToolsVersion.default.txt")
	if _, err := os.Stat(versionFilePath); err != nil {
		return nil, false
	}

	// Read the MSVC version from version file.
	versionFile, err := os.Open(versionFilePath)
	if err != nil {
		return nil, false
	}
	defer versionFile.Close()

	versionB, err := ioutil.ReadAll(versionFile)
	if err != nil {
		return nil, false
	}

	version := strings.TrimSpace(string(versionB))

	// Get the base path to the installation.
	basePath := filepath.Join(instancePath, "VC/Tools/MSVC/", version)

	// Get the host and target architecture suffixes.
	hostArch := hostArchToVCHostSuffix[runtime.GOARCH]
	subDir := llvmArchToVS15PlusSubDir[targetArch]

	// Build the tool.
	tool := &VCToolPaths{}
	tool.BinPath = filepath.Join(basePath, fmt.Sprintf("bin/Host%s/%s", hostArch, subDir))
	tool.DyLibPath = tool.BinPath
	tool.LibPath = filepath.Join(basePath, "lib/"+subDir)
	tool.IncludePath = filepath.Join(basePath, "include")
	tool.ToolPath = filepath.Join(tool.BinPath, "link.exe")

	// Check if the tool itself exists.
	if _, err := os.Stat(tool.ToolPath); err != nil {
		return nil, false
	}

	// Return the located tool.
	return tool, true
}

// getVersionInt converts a VS version string to an integer so it can be
// compared.  Effectively, each of the "sub-versions" are encoded into
// corresponding bit positions in the integer.
func getVersionInt(versionString string) uint64 {
	versionComponents := make([]int, 4)
	for i, component := range strings.Split(versionString, ".") {
		v, err := strconv.Atoi(component)
		if err != nil {
			panic("failed to parse VS version string: " + err.Error())
		}

		versionComponents[i] = v
	}

	var version uint64
	version = uint64(versionComponents[0]) << 48
	version |= (uint64(versionComponents[1]) & 255) << 32
	version |= (uint64(versionComponents[2]) & 65535) << 16
	version |= (uint64(versionComponents[3]) & 65535)

	return version
}
