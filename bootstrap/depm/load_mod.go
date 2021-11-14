package depm

import (
	"chai/common"
	"chai/report"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

// tomlModule represents a Chai module as it is encoded in TOML
type tomlModule struct {
	Name                string         `toml:"name"`
	ShouldCache         bool           `toml:"caching"`
	BuildProfiles       []*tomlProfile `toml:"profiles"`
	ChaiVersion         string         `toml:"chai-version"`
	AllowProfileElision bool           `toml:"allow-profile-elision"`
}

// tomlProfile represents a profile as it encoded in TOML
type tomlProfile struct {
	Name          string     `toml:"name"`
	TargetOS      string     `toml:"target-os"`
	TargetArch    string     `toml:"target-arch"`
	Debug         bool       `toml:"debug"`
	OutputPath    string     `toml:"output-path"`
	Format        string     `toml:"format"`
	LinkObjects   []string   `toml:"link-objects,omitempty"`
	DefaultProf   bool       `toml:"default"`
	BaseOnly      bool       `default:"false"`
	LastBuildTime *time.Time `toml:"last-build"`
}

// LoadModule loads and validates a module as well as determining the correct
// profile (if one exists).  `abspath` is the absolute path to the module
// directory. `selectedProfile` can be empty if these is no profile selected.
// The `baseProfile` argument is the build profile of the base module. This
// argument cannot be `nil` and will be populated with new link objects as
// profiles are loaded.  This can simply be an empty struct on initialization.
// Refer to the module schema for more information about the base profile.  This
// function returns the deserialized module and a successed boolean.
func LoadModule(abspath, selectedProfile string, baseProfile *BuildProfile) (*ChaiModule, bool) {
	// open file
	f, err := os.Open(filepath.Join(abspath, common.ChaiModuleFileName))
	if err != nil {
		report.ReportFatal(fmt.Sprintf("unable to open module file at `%s`: %s", abspath, err.Error()))
		return nil, false
	}
	defer f.Close()

	// unmarshal the contents
	buff, err := ioutil.ReadAll(f)
	if err != nil {
		report.ReportFatal(fmt.Sprintf("error reading module file at `%s`: %s", abspath, err.Error()))
		return nil, false
	}

	tomlMod := &tomlModule{}
	if err := toml.Unmarshal(buff, tomlMod); err != nil {
		report.ReportFatal(fmt.Sprintf("error parsing module file at `%s`: %s", abspath, err.Error()))
		return nil, false
	}

	// chaiMod is the final, extracted module that is returned
	chaiMod := &ChaiModule{
		// module root is the directory enclosing the module file
		AbsPath:     abspath,
		ID:          GenerateIDFromPath(abspath),
		SubPackages: make(map[string]*ChaiPackage),
	}

	// ensure that the base module is valid
	if !validateModule(chaiMod, tomlMod) {
		return nil, false
	}

	// select, validate, and merge an appropriate build profile
	if err := selectProfile(chaiMod, tomlMod, selectedProfile, baseProfile); err != nil {
		report.ReportModuleError(chaiMod.Name, err.Error())
		return nil, false
	}

	return chaiMod, true
}

// validateModule checks that the top level module contents are valid
func validateModule(chaiMod *ChaiModule, tomlMod *tomlModule) bool {
	if tomlMod.Name == "" {
		report.ReportModuleError(fmt.Sprintf("<module at `%s`>", chaiMod.AbsPath), "missing module name")
		return false
	}

	if !IsValidIdentifier(tomlMod.Name) {
		report.ReportModuleError(fmt.Sprintf("<module at `%s`>", chaiMod.AbsPath), "module name must be a valid identifier")
		return false
	}

	if tomlMod.ChaiVersion != common.ChaiVersion {
		report.ReportModuleWarning(tomlMod.Name, fmt.Sprintf("version of module `%s` (v%s) does not match current chai version (v%s)",
			tomlMod.Name,
			tomlMod.ChaiVersion,
			common.ChaiVersion,
		))
	}

	// move all the relevant TOML module attributes over to the Chai module
	chaiMod.Name = tomlMod.Name
	chaiMod.ShouldCache = tomlMod.ShouldCache

	return true
}

// selectProfile attempts to select a build profile based on the known base
// profile and a selected profile if one exists, validates this profile, and
// updates the base build profile accordingly.  More information about this
// process can be found in the module schema.
func selectProfile(chaiMod *ChaiModule, tomlMod *tomlModule, selectedProfile string, baseProfile *BuildProfile) error {
	// selectedProfile != nil => base profile
	if selectedProfile != "" {
		if len(tomlMod.BuildProfiles) == 0 {
			return errors.New("root module must provide at least one build profile")
		}

		for _, prof := range tomlMod.BuildProfiles {
			if prof.Name == selectedProfile {
				convProf, err := convertProfile(prof)

				if err != nil {
					return fmt.Errorf("error loading profile: %s", err.Error())
				}

				// selectedProfile can only be set on the root module so it is
				// not possible for the selected profile to conflict with the
				// root profile (as no root profile has been established)
				*baseProfile = *convProf

				// set the module last build time to the last build time of the
				// profile (may be `nil` if no caching is enabled)
				chaiMod.LastBuildTime = prof.LastBuildTime

				// found profile; exit
				return nil
			}
		}

		return fmt.Errorf("module has no profile named `%s`", selectedProfile)
	}

	var possibleProfiles []*BuildProfile
	var possibleTomlProfiles []*tomlProfile
	defaultProfile := -1

	for _, prof := range tomlMod.BuildProfiles {
		convProf, err := convertProfile(prof)
		if err != nil {
			return err
		}

		if convProf.Debug == baseProfile.Debug &&
			(baseProfile.OutputFormat == -1 || convProf.OutputFormat == baseProfile.OutputFormat) &&
			convProf.TargetArch == baseProfile.TargetArch &&
			convProf.TargetOS == baseProfile.TargetOS {

			if prof.DefaultProf {
				defaultProfile = len(possibleProfiles)
			}

			possibleProfiles = append(possibleProfiles, convProf)
			possibleTomlProfiles = append(possibleTomlProfiles, prof)
		}
	}

	switch len(possibleProfiles) {
	case 0:
		// no output path => root module, loading base profile.
		if baseProfile.OutputPath == "" {
			return errors.New("root module must specify a valid build profile")
		}

		if !tomlMod.AllowProfileElision {
			return errors.New("module does not specify a build profile compatible with the base profile")
		}
	case 1:
		updateProfile(baseProfile, possibleProfiles[0])
		chaiMod.LastBuildTime = possibleTomlProfiles[0].LastBuildTime
	default:
		if defaultProfile == -1 {
			report.ReportModuleWarning(
				chaiMod.Name,
				fmt.Sprintf("multiple possible profiles for module detected: building with profile `%s`", possibleTomlProfiles[0].Name),
			)

			defaultProfile = 0
		}

		updateProfile(baseProfile, possibleProfiles[defaultProfile])
		chaiMod.LastBuildTime = possibleTomlProfiles[defaultProfile].LastBuildTime
	}

	return nil
}

// OSNames lists the valid OS names
var OSNames = map[string]struct{}{
	"windows": {},
}

// ArchNames lists the valid arch names
var ArchNames = map[string]struct{}{
	"i386":  {},
	"amd64": {},
}

// formatNames maps TOML os name strings to enumerated format values
var formatNames = map[string]int{
	"bin":  FormatBin,
	"asm":  FormatASM,
	"llvm": FormatLLVM,
	"obj":  FormatObj,
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

	if _, ok := OSNames[tprof.TargetOS]; ok {
		newProfile.TargetOS = tprof.TargetOS
	} else {
		return nil, fmt.Errorf("%s is not a supported OS", tprof.TargetOS)
	}

	if _, ok := ArchNames[tprof.TargetArch]; ok {
		newProfile.TargetArch = tprof.TargetArch
	} else {
		return nil, fmt.Errorf("%s is not a supported architecture", tprof.TargetArch)
	}

	if formatVal, ok := formatNames[tprof.Format]; ok {
		newProfile.OutputFormat = formatVal
	} else {
		return nil, fmt.Errorf("%s is not a valid output format", tprof.Format)
	}

	newProfile.Debug = tprof.Debug
	newProfile.LinkObjects = tprof.LinkObjects
	newProfile.OutputPath = tprof.OutputPath

	return newProfile, nil
}

// updateProfile updates the base profile with any new configuration from a sub profile.
func updateProfile(baseProfile, subProfile *BuildProfile) {
	// base profile has not output path => subProfile is the base profile.
	if baseProfile.OutputPath == "" {
		*baseProfile = *subProfile
	} else {
		// just update link objects
		baseProfile.LinkObjects = append(baseProfile.LinkObjects, subProfile.LinkObjects...)
	}
}
