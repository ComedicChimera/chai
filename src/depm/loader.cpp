#include "loader.hpp"

#include <filesystem>
#include <stdexcept>

#include "constants.hpp"
#include "report/warn.hpp"

#define MODULE_FILENAME "chai-mod.toml"

namespace fs = std::filesystem;

namespace chai {
    ModuleLoader::ModuleLoader(const std::string& modDir, const BuildProfile& globalProfile) 
    : modFilePath((fs::path(modDir) / fs::path(MODULE_FILENAME)).string())
    , mod{.rootDir = modDir}
    , globalProfile(globalProfile)
    {}

    void ModuleLoader::throwModuleError(const std::string& message) {
        throw std::logic_error(std::format("[module error]\n{}: {}", modFilePath, message));
    }

    template<typename T>
    T ModuleLoader::getRequiredField(toml::table& tbl, const std::string& fieldName) {
        std::optional<T> field = tbl[fieldName].value();
        if (!field)
            throwModuleError(std::format("missing or malformed required field: `{}`", fieldName));
    }

    std::pair<Module, BuildProfile> ModuleLoader::load(std::optional<const std::string&> selectedProfileName) {
        // load the full module file table
        toml::table tbl;
        try {  
            tbl = toml::parse_file(modFilePath);
        } catch (toml::parse_error& err) {
            throw std::logic_error(std::format("[module error]\n{}:{}:{}: {}", 
                modFilePath, 
                err.source().begin.line, 
                err.source().begin.column,
                err.description()
                ));
        }

        // load the module table `[module]`
        std::optional<toml::table> optTomlMod = tbl["module"].value<toml::table>();
        if (!optTomlMod)
            throwModuleError("missing required table `module`");

        auto tomlMod = optTomlMod.value();

        // load all the high level module data
        mod.name = getRequiredField<std::string>(tomlMod, "name");
        mod.shouldCache = tomlMod.value_or(false);
        mod.cacheDirectory = tomlMod.value_or(".chai/cache");
        
        if (tomlMod.find("local-import-dirs") != tomlMod.end()) {
            auto tomlLocalImportDirs = tomlMod["local-import-dirs"];
            if (auto* tomlLocalImportDirs = tomlMod["local-import-dirs"].as_array()) {
                for (auto& elem : *tomlLocalImportDirs) {
                    if (elem.is_string())
                        mod.localImportDirs.push_back(elem.value<std::string>().value());
                    else
                        throwModuleError("elements of `local-import-dirs` must be strings");
                }
            } else {
                throwModuleError("`local-import-dirs` must be an array");
            }
        }

        // check that the Chai versions match up
        auto modChaiVersion = getRequiredField<std::string>(tomlMod, "chai-version");
        int modChaiVersionMajor, modChaiVersionMinor, modChaiVersionBuild;
        if (!sscanf(modChaiVersion.c_str(), "%d.%d.%d", &modChaiVersionMajor, &modChaiVersionMinor, &modChaiVersionBuild))
            throwModuleError("chai version number must be of the following form: `{int}.{int}.{int}`");

        if (modChaiVersionMajor != CHAI_VERSION_MAJOR || modChaiVersionMinor != CHAI_VERSION_MINOR || modChaiVersionBuild != CHAI_VERSION_BUILD) {
            reportWarningMessage(std::format(
                "chai version of `v{}.{}.{}` doesn't match module chai version of `v{}`", 
                CHAI_VERSION_MAJOR, 
                CHAI_VERSION_MINOR,
                CHAI_VERSION_BUILD,
                modChaiVersion
                )
            );
        }

        // TODO: load and retrieve dependencies

        // select and load a build profile if necessary
        if (tomlMod.find("profiles") == tomlMod.end()) {
            // if the globalProfile has no name, then there is no current
            // profile and we are building the main module
            if (globalProfile.name == "")
                throwModuleError("main module must specify build profiles");

            return std::make_pair(mod, globalProfile);
        }
        
        if (auto* tomlProfiles = tomlMod["profiles"].as_array()) {
            auto profile = selectBuildProfile(tomlProfiles, selectedProfileName);

            // return the built module and profile
            return std::make_pair(mod, profile); 
        } else
            throwModuleError("field `profiles` must be an array");   
    }

    // selectBuildProfile takes the list of TOML build profiles as well as a
    // possible profile name and walks through each profile and attempts to
    // select the one that best matches the current configuration.
    BuildProfile ModuleLoader::selectBuildProfile(toml::array* tomlProfiles, std::optional<const std::string&> selectedProfileName) {
        if (selectedProfileName) {
            
        } else {

        }
    }
}