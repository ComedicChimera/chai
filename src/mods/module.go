package mods

import (
	"chai/deps"
	"path/filepath"
	"strings"
	"time"
)

// ChaiModule represents a module -- specifically, the module configuration.
// NOTE: Profile information is not stored on the module but on the compiler
// object itself (as things like target, output, static libraries, etc are
// consistent and/or determined based on all loaded modules)
type ChaiModule struct {
	// ID is the module's unique ID
	ID uint

	// Name is the name of the module
	Name string

	// RootPackage is the package at the root of the module
	RootPackage *deps.ChaiPackage

	// SubPackages is a map of all the subpackages of this module organized by
	// their subpath (eg. `io.std` => `std`; `mod.b.c` => `b.c`)
	SubPackages map[string]*deps.ChaiPackage

	// DependsOn is a map of all the other modules this module depends on
	// organized by module name
	DependsOn map[string]*ChaiModule

	// -----------------------------------------------------------------------------

	// ModuleRoot is the path to the root directory of the current module
	ModuleRoot string

	// LocalImportDirs is a list of directories in which to check for imports
	// (outside of the current module and global import directories)
	LocalImportDirs []string

	// ShouldCache indicates whether or not compilation caching should be
	// performed for this module
	ShouldCache bool

	// CacheDirectory indicates which directory to search for cached module
	// versions for the purposes of compilation caching
	CacheDirectory string

	// LastBuildTime indicates when this module was last built for the purposes
	// of compilation caching.  It is stored as a field on the active profile.
	LastBuildTime *time.Time
}

// Packages returns a slice of all the packages in the module
func (cm *ChaiModule) Packages() []*deps.ChaiPackage {
	pkgs := []*deps.ChaiPackage{cm.RootPackage}
	for _, subpkg := range cm.SubPackages {
		pkgs = append(pkgs, subpkg)
	}

	return pkgs
}

// BuildPackagePathString builds a string indicating the full path to a package
// within a given module for use in producing informative error messages.
func (cm *ChaiModule) BuildPackagePathString(pkg *deps.ChaiPackage) string {
	if cm.RootPackage == pkg {
		return cm.Name
	}

	for subpath, subpkg := range cm.SubPackages {
		if subpkg == pkg {
			return cm.Name + "." + strings.ReplaceAll(subpath, string(filepath.Separator), ".")
		}
	}

	// unreachable
	return ""
}

// BuildProfile represents the profile that compiler will use to build -- it is
// returned from `LoadModule`.
type BuildProfile struct {
	// OutputPath is the path to the final output file(s)
	OutputPath string

	// OutputFormat is the type of output the compiler should produce.  This
	// should be one of the enumerated formats (prefixed `Format`).
	OutputFormat int

	// Debug indicates whether or not the compiler should build an output for
	// debug or for release
	Debug bool

	// TargetOS is the target operating system for compilation. This should be
	// one of the enumerated operating systems (prefixed `OS`)
	TargetOS int

	// TargetArch is the target architecure for compilation.  This should be one
	// of the enumerated architectures (prefixed `Arch`)
	TargetArch int

	// StaticLibraries is the list of static libraries to be linked into the
	// final build output.  These are absolute paths.
	StaticLibraries []string

	// DynamicLibraries is the list of the dynamic libraries that should be
	// marked as dependencies for the final output.  These are relative paths.
	DynamicLibraries []string
}

// Available Output Formats
const (
	FormatBin    = iota // Executable
	FormatLib           // Library
	FormatASM           // Assembly
	FormatLLVM          // LLVM
	FormatObject        // Unlinked Object Files
)

// Available Target OSs
const (
	OSWindows = iota
	// ...
)

// Available Target Archs
const (
	ArchI386 = iota
	ArchAmd64
	ArchArm
	// ...
)

// IsValidIdentifier returns whether or not a given string would be a valid
// identifier (module name, package name, etc.)
func IsValidIdentifier(idstr string) bool {
	if idstr[0] == '_' || ('a' <= idstr[0] && idstr[0] <= 'z') || ('A' <= idstr[0] && idstr[0] <= 'Z') {
		for _, c := range idstr[1:] {
			if c == '_' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
				continue
			}

			return false
		}

		return true
	}

	return false
}
