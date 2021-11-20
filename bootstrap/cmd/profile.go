package cmd

// BuildProfile represents the current build profile.
type BuildProfile struct {
	Debug      bool
	OutputPath string
	TargetOS   string
	TargetArch string

	// OutputFormat should be one of the enumerated output formats.
	OutputFormat int

	// LinkObjects is the collection of path's to link with the final
	// executable.
	LinkObjects []string
}

// Enumeration of possible outform formats.
const (
	FormatBin = iota
	FormatObj

	// These may be supported in Alpha chai, but TBD
	FormatLib
	FormatDLL
)
