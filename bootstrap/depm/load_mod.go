package depm

import (
	"chai/common"
	"chai/report"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

// tomlModule represents a Chai module as it is encoded in TOML
type tomlModule struct {
	Name        string `toml:"name"`
	ShouldCache bool   `toml:"caching"`
	ChaiVersion string `toml:"chai-version"`
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

// LoadModule loads and validates a module.  `abspath` is the absolute path to
// the module directory.  This function returns the deserialized module and a
// success boolean.
func LoadModule(abspath string) (*ChaiModule, bool) {
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

// // OSNames lists the valid OS names
// var OSNames = map[string]struct{}{
// 	"windows": {},
// }

// // ArchNames lists the valid arch names
// var ArchNames = map[string]struct{}{
// 	"i386":  {},
// 	"amd64": {},
// }

// // formatNames maps TOML os name strings to enumerated format values
// var formatNames = map[string]int{
// 	"bin":  FormatBin,
// 	"asm":  FormatASM,
// 	"llvm": FormatLLVM,
// 	"obj":  FormatObj,
// }
