package mods

import (
	"chai/common"
	"chai/logging"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

// tomlModuleFile represents the module file as it is encoded in TOML
type tomlModuleFile struct {
	Module *tomlModule `toml:"module"`
}

// tomlModule represents a Chai module as it is encoded in TOML
type tomlModule struct {
	Name                string            `toml:"name"`
	LocalImportDirs     []string          `toml:"local-import-dirs,omitempty"`
	ShouldCache         bool              `toml:"caching"`
	CacheDirectory      string            `toml:"cache-directory,omitempty"`
	PathReplacements    map[string]string `toml:"path-replacements,omitempty"`
	BuildProfiles       []*tomlProfile    `toml:"profiles"`
	Dependencies        []*tomlDependency `toml:"dependencies"`
	Version             string            `toml:"chai-version"`
	AllowProfileElision bool              `toml:"allow-profile-elision"`
}

// tomlProfile represents a profile as it encoded in TOML
type tomlProfile struct {
	Name          string     `toml:"name"`
	TargetOS      string     `toml:"target-os"`
	TargetArch    string     `toml:"target-arch"`
	Debug         bool       `toml:"debug"`
	OutputPath    string     `toml:"output"`
	Format        string     `toml:"format"`
	DynamicLibs   []string   `toml:"dynamic-libs,omitempty"`
	StaticLibs    []string   `toml:"static-libs,omitempty"`
	Primary       bool       `toml:"primary"` // of profiles matching build config, choose this profile
	DefaultProf   bool       `toml:"default"` // in absence of build config, choose this profile
	LastBuildTime *time.Time `toml:"last-build"`
}

// tomlDependency represents a dependency as it is encoded in TOML
type tomlDependency struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Url     string `toml:"url"`
}

// LoadModule loads and validates a module as well as determining the correct
// profile (if one exists).  `path` is the path to the module directory.
// `selectedProfile` can be empty if these is no profile selected.  The
// `rootProfile` argument is the build profile of the root module -- the module
// being compiled initially.  This argument cannot be `nil` and will be
// populated with new profile data as profiles are loaded (eg. new static
// libraries). This can simply be an empty struct on initialization (with the
// exception of the version field which must be populated)  This function
// returns the deserialized module and an error value.
func LoadModule(path, selectedProfile string, rootProfile *BuildProfile) (*ChaiModule, error) {
	// open file
	f, err := os.Open(filepath.Join(path, common.ModuleFileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// unmarshal the contents
	buff, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	tmf := &tomlModuleFile{}
	if err := toml.Unmarshal(buff, tmf); err != nil {
		return nil, err
	}

	// chaiMod is the final, extracted module that is returned
	chaiMod := &ChaiModule{
		// module root is the directory enclosing the module file
		ModuleRoot: path,
	}

	// ensure that the base module is valid
	if err := validateModule(chaiMod, tmf.Module); err != nil {
		return nil, err
	}

	// fetch any dependencies of the module (so that they can be imported as
	// necessary later)
	if err := fetchDependencies(tmf.Module); err != nil {
		return nil, err
	}

	// select, validate, and merge an appropriate build profile
	if err := selectProfile(chaiMod, tmf.Module, selectedProfile, rootProfile); err != nil {
		return nil, err
	}

	// move all the relevant TOML module attributes over to the Chai module
	chaiMod.Name = tmf.Module.Name
	chaiMod.CacheDirectory = tmf.Module.CacheDirectory
	chaiMod.ShouldCache = tmf.Module.ShouldCache
	chaiMod.LocalImportDirs = tmf.Module.LocalImportDirs
	chaiMod.PathReplacements = tmf.Module.PathReplacements

	return chaiMod, nil
}

// validateModule checks that the top level module contents are valid
func validateModule(cmod *ChaiModule, mod *tomlModule) error {
	if mod.Name == "" {
		return fmt.Errorf("missing module name for module at %s", cmod.ModuleRoot)
	}

	if !IsValidIdentifier(mod.Name) {
		return errors.New("module name must be a valid identifier")
	}

	if mod.ShouldCache && mod.CacheDirectory == "" {
		return fmt.Errorf("a cache directory must be specified in module %s since caching is enabled", mod.Name)
	}

	if mod.Version != common.ChaiVersion {
		logging.LogBuildWarning(
			"module",
			fmt.Sprintf("version of module `%s` (v%s) does not match current chai version (v%s)", mod.Name, mod.Version, common.ChaiVersion),
		)
	}

	return nil
}

// selectProfile attempts to select a build profile based on the known root
// profile and a selected profile if one exists, validates this profile, and
// updates the root build profile accordingly
func selectProfile(cmod *ChaiModule, mod *tomlModule, selectedProfile string, rootProfile *BuildProfile) error {
	if selectedProfile != "" {
		if len(mod.BuildProfiles) == 0 {
			return fmt.Errorf("module %s must provide at least one build profile", mod.Name)
		}

		for _, prof := range mod.BuildProfiles {
			if prof.Name == selectedProfile {
				convProf, err := convertProfile(prof)

				if err != nil {
					return fmt.Errorf("%s in module %s", mod.Name, err.Error())
				}

				// selectedProfile can only be set on the root module so it is
				// not possible for the selected profile to conflict with the
				// root profile (as no root profile has been established)
				*rootProfile = *convProf

				// set the module last build time to the last build time of the
				// profile (may be `nil` if no caching is enabled)
				cmod.LastBuildTime = prof.LastBuildTime

				// found profile; exit
				return nil
			}
		}

		return fmt.Errorf("module `%s` has no profile `%s`", mod.Name, selectedProfile)
	}

	// if there is no output path, then we know the root profile hasn't been initialized yet
	// because all valid profiles must have an output path
	if rootProfile.OutputPath == "" {
		if len(mod.BuildProfiles) == 0 {
			return fmt.Errorf("module %s must provide at least one build profile", mod.Name)
		}

		for _, prof := range mod.BuildProfiles {
			if prof.DefaultProf {
				convProf, err := convertProfile(prof)
				if err != nil {
					return err
				}

				*rootProfile = *convProf
				cmod.LastBuildTime = prof.LastBuildTime

				// found profile; exit
				return nil
			}
		}

		return fmt.Errorf("module `%s` does not specify a default profile; `--profile` argument is required", mod.Name)
	}

	var possibleProfiles []*BuildProfile
	var possibleTomlProfiles []*tomlProfile
	primaryProfile := -1

	for _, prof := range mod.BuildProfiles {
		convProf, err := convertProfile(prof)
		if err != nil {
			return err
		}

		if convProf.Debug == rootProfile.Debug &&
			convProf.OutputFormat == rootProfile.OutputFormat &&
			convProf.TargetArch == rootProfile.TargetArch &&
			convProf.TargetOS == rootProfile.TargetOS {

			if prof.Primary {
				primaryProfile = len(possibleProfiles)
			}

			possibleProfiles = append(possibleProfiles, convProf)
			possibleTomlProfiles = append(possibleTomlProfiles, prof)
		}
	}

	switch len(possibleProfiles) {
	case 0:
		if !mod.AllowProfileElision {
			return fmt.Errorf("module `%s` does not specify a build profile compatible with that of root module; building with root profile", mod.Name)
		}
	case 1:
		updateProfile(rootProfile, possibleProfiles[0])
		cmod.LastBuildTime = possibleTomlProfiles[0].LastBuildTime
	default:
		if primaryProfile == -1 {
			logging.LogBuildWarning(
				"module",
				fmt.Sprintf("multiple possible profiles for module `%s` detected; building with profile `%s`", mod.Name, possibleTomlProfiles[0].Name),
			)

			primaryProfile = 0
		}

		updateProfile(rootProfile, possibleProfiles[primaryProfile])
		cmod.LastBuildTime = possibleTomlProfiles[primaryProfile].LastBuildTime
	}

	return nil
}

// osNames maps TOML os name strings to enumerated OS values
var osNames = map[string]int{
	"windows": OSWindows,
}

// archNames maps TOML os name strings to enumerated arch values
var archNames = map[string]int{
	"i386":  ArchI386,
	"amd64": ArchAmd64,
	"arm":   ArchArm,
}

// formatNames maps TOML os name strings to enumerated format values
var formatNames = map[string]int{
	"bin":  FormatBin,
	"asm":  FormatASM,
	"llvm": FormatLLVM,
	"lib":  FormatLib,
	"obj":  FormatObject,
}

// convertProfile converts a TOML build profile into a `*BuildProfile`
func convertProfile(tprof *tomlProfile) (*BuildProfile, error) {
	if tprof.Name == "" {
		return nil, errors.New("profile must specify a name")
	}

	if tprof.OutputPath == "" {
		return nil, errors.New("profile must specify an output path")
	}

	if tprof.TargetOS == "" {
		return nil, errors.New("profile must specify a target operating system")
	}

	if tprof.TargetArch == "" {
		return nil, errors.New("profile must specify a target architecture")
	}

	if tprof.Format == "" {
		return nil, errors.New("profile must specify an output format")
	}

	newProfile := &BuildProfile{}

	if osVal, ok := osNames[tprof.TargetOS]; ok {
		newProfile.TargetOS = osVal
	} else {
		return nil, fmt.Errorf("%s is not a supported OS", tprof.TargetOS)
	}

	if archVal, ok := archNames[tprof.TargetArch]; ok {
		newProfile.TargetArch = archVal
	} else {
		return nil, fmt.Errorf("%s is not a supported architecture", tprof.TargetArch)
	}

	if formatVal, ok := formatNames[tprof.Format]; ok {
		newProfile.OutputFormat = formatVal
	} else {
		return nil, fmt.Errorf("%s is not a valid output format", tprof.Format)
	}

	newProfile.Debug = tprof.Debug
	newProfile.DynamicLibraries = tprof.DynamicLibs
	newProfile.StaticLibraries = tprof.StaticLibs
	newProfile.OutputPath = tprof.OutputPath

	return newProfile, nil
}

// updateProfile updates the root profile with any new configuration from a sub profile
func updateProfile(rootProfile, subProfile *BuildProfile) {
	rootProfile.DynamicLibraries = append(rootProfile.DynamicLibraries, subProfile.DynamicLibraries...)
	rootProfile.StaticLibraries = append(rootProfile.StaticLibraries, subProfile.StaticLibraries...)
}

// fetchDependencies fetches any dependencies of a module that are not already
// installed so that they can be loaded later
func fetchDependencies(mod *tomlModule) error {
	// TODO
	return nil
}
