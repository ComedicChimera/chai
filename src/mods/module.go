package mods

import "time"

// ChaiModule represents a module -- specifically, the module configuration.
// NOTE: Profile information is not stored on the module but on the compiler
// object itself (as things like target, output, static libraries, etc are
// consistent and/or determined based on all loaded modules)
type ChaiModule struct {
	// Name is the name of the module
	Name string

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

	// PathReplacements maps certain import paths to different directories. The
	// key is the original path and the value is the path to replace it with.
	// Note that these paths override explicit import paths and import
	// semantics.
	PathReplacements map[string]string
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

// ResolveImportPath attempts to convert a relative import path into an absolute
// path to a package or module.  This fails if the path cannot be resolved but
// does not check for Chai code in those directories.
func (m *ChaiModule) ResolveImportPath(path string) (string, bool) {
	return "", false
}

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
