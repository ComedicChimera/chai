package wintool

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// addSDKs adds the Windows SDK paths to a tool builder.
func addSDKs(toolBuilder *ToolCmdBuilder, targetArch string) error {
	// Add the UCRT installation to the tool builder.
	ucrtDir, ucrtVersion, ok := getUCRTDir()

	if !ok {
		return errors.New("unable to find any UCRT installations")
	}

	hostDir := llvmArchToUCRTHostDir[targetArch]
	targetDir := llvmArchToVS15PlusSubDir[targetArch]

	toolBuilder.BinPaths = append(toolBuilder.BinPaths, filepath.Join(ucrtDir, "bin", ucrtVersion, hostDir))
	toolBuilder.IncludePaths = append(toolBuilder.IncludePaths, filepath.Join(ucrtDir, "include", "ucrt"))
	toolBuilder.LibPaths = append(toolBuilder.LibPaths, filepath.Join(ucrtDir, "lib", ucrtVersion, "ucrt", targetDir))

	// Add the Windows 10 SDK installation to the tool builder.
	sdkDir, sdkVersion, ok := getSDK10Dir()
	if !ok {
		// TODO: support 8.1 SDK
		return errors.New("unable to find any Windows 10 SDK installation")
	}

	toolBuilder.BinPaths = append(toolBuilder.BinPaths, filepath.Join(sdkDir, "lib", hostDir))
	toolBuilder.LibPaths = append(toolBuilder.LibPaths, filepath.Join(sdkDir, "lib", sdkVersion, "um", targetDir))

	sdkInclude := filepath.Join(sdkDir, "include", sdkVersion)
	toolBuilder.IncludePaths = append(toolBuilder.IncludePaths, filepath.Join(sdkInclude, "um"))
	toolBuilder.IncludePaths = append(toolBuilder.IncludePaths, filepath.Join(sdkInclude, "cppwinrt"))
	toolBuilder.IncludePaths = append(toolBuilder.IncludePaths, filepath.Join(sdkInclude, "winrt"))
	toolBuilder.IncludePaths = append(toolBuilder.IncludePaths, filepath.Join(sdkInclude, "shared"))

	return nil
}

// getUCRTDir attempts to find the installed UCRT directory and version.
func getUCRTDir() (string, string, bool) {
	// Find the root directory of the UCRT using the registry.
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows Kits\Installed Roots`, registry.QUERY_VALUE)
	if err != nil {
		return "", "", false
	}
	defer k.Close()

	rootDir, _, err := k.GetStringValue("KitsRoot10")
	if err != nil {
		return "", "", false
	}

	// Determine the directory containing the UCRT libaries.
	libDir := filepath.Join(rootDir, "lib")

	// Search the directory to find the maximum UCRT version installed.
	// Technically, we could try to lookup the current Windows version, but that
	// is kind of nightmare.  This is basically just what vcvars does so it
	// should be suitable for our purposes.
	finfos, err := ioutil.ReadDir(libDir)
	if err != nil {
		return "", "", false
	}

	var maxVersion string
	for _, finfo := range finfos {
		if finfo.IsDir() && strings.HasPrefix(finfo.Name(), "10.") {
			// Make sure the directory actually contains a UCRT installation.
			if _, err := os.Stat(filepath.Join(libDir, finfo.Name(), "ucrt")); err != nil {
				continue
			}

			if maxVersion == "" || finfo.Name() > maxVersion {
				maxVersion = finfo.Name()
			}
		}
	}

	if maxVersion == "" {
		return "", "", false
	}

	return rootDir, maxVersion, true
}

// getSDK10Dir finds the installed Windows 10 SDK if it exists.
func getSDK10Dir() (string, string, bool) {
	// Find the root directory for the Windows SDK using the registry.
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Microsoft SDKs\Windows\v10.0`, registry.QUERY_VALUE)
	if err != nil {
		return "", "", false
	}
	defer k.Close()

	rootDir, _, err := k.GetStringValue("InstallationFolder")
	if err != nil {
		return "", "", false
	}

	// Determine the directory containing the Windows libraries.
	libDir := filepath.Join(rootDir, "lib")

	// Search the directory to find the maximum SDK version installed. The logic
	// here is similar to finding the UCRT: trying to lookup the current Windows
	// version is very annoying and just finding the maximum installed version
	// by searching is easier.
	finfos, err := ioutil.ReadDir(libDir)
	if err != nil {
		return "", "", false
	}

	var maxVersion string
	for _, finfo := range finfos {
		if finfo.IsDir() {
			// Since x64 and x86 libraries are always installed together, and we
			// only actually need the libraries, we can just check for
			// `um/x64/kernel32.lib` to validate an installation directory.
			if finfo, err := os.Stat(filepath.Join(libDir, finfo.Name(), `um\x64\kernel32.lib`)); err == nil && !finfo.IsDir() {
				if maxVersion == "" || finfo.Name() > maxVersion {
					maxVersion = finfo.Name()
				}
			}
		}
	}

	if maxVersion == "" {
		return "", "", false
	}

	return rootDir, maxVersion, true
}
