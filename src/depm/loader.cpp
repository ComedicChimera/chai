#include "loader.hpp"

#include <filesystem>
#include <stdexcept>

#include "constants.hpp"

#define MODULE_FILENAME "chai-mod.toml"

namespace fs = std::filesystem;

namespace chai {
    ModuleLoader::ModuleLoader(DepGraph& depg, Reporter& re, const std::string& modDir, const BuildProfile& globalProfile) 
    : modFilePath((fs::path(modDir) / fs::path(MODULE_FILENAME)).string())
    , depg(depg)
    , globalProfile(globalProfile)
    , reporter(re)
    {
        mod = new Module {.id=depg.getModuleID(modDir), .rootDir = fs::absolute(modDir).string()};
    }

    void ModuleLoader::throwModuleError(const std::string& message) {
        throw std::logic_error(std::format("[module error]\n{}: {}", modFilePath, message));
    }

    template<typename T>
    T ModuleLoader::getRequiredField(toml::table* tbl, const std::string& fieldName) {
        std::optional<T> field = tbl[fieldName].value();
        if (!field)
            throwModuleError(std::format("missing or malformed required field: `{}` for table at {}:{}", fieldName, tbl.source().begin().line, tbl.source().begin().column));
    }

    std::pair<Module*, BuildProfile> ModuleLoader::load(std::optional<const std::string&> selectedProfileName) {
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
        mod->name = getRequiredField<std::string>(&tomlMod, "name");
        mod->shouldCache = tomlMod.value_or(false);
        mod->cacheDirectory = tomlMod.value_or(".chai/cache");
        
        if (tomlMod.contains("local-import-dirs")) {
            auto tomlLocalImportDirs = tomlMod["local-import-dirs"];
            if (auto* tomlLocalImportDirs = tomlMod["local-import-dirs"].as_array()) {
                for (auto& elem : *tomlLocalImportDirs) {
                    if (elem.is_string())
                        mod->localImportDirs.push_back(elem.value<std::string>().value());
                    else
                        throwModuleError("elements of `local-import-dirs` must be strings");
                }
            } else {
                throwModuleError("`local-import-dirs` must be an array");
            }
        }

        // check that the Chai versions match up
        auto modChaiVersion = getRequiredField<std::string>(&tomlMod, "chai-version");
        int modChaiVersionMajor, modChaiVersionMinor, modChaiVersionBuild;
        if (!sscanf(modChaiVersion.c_str(), "%d.%d.%d", &modChaiVersionMajor, &modChaiVersionMinor, &modChaiVersionBuild))
            throwModuleError("chai version number must be of the following form: `{int}.{int}.{int}`");

        if (modChaiVersionMajor != CHAI_VERSION_MAJOR || modChaiVersionMinor != CHAI_VERSION_MINOR || modChaiVersionBuild != CHAI_VERSION_BUILD) {
            reporter.reportWarningMessage(std::format(
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
        if (!tomlMod.contains("profiles")) {
            // if the globalProfile has no name, then there is no current
            // profile and we are building the main module
            if (globalProfile.name == "")
                throwModuleError("main module must specify build profiles");

            depg.addModule(mod);
            return std::make_pair(mod, globalProfile);
        }
        
        if (auto* tomlProfiles = tomlMod["profiles"].as_array()) {
            auto profile = selectBuildProfile(tomlProfiles, selectedProfileName);

            // return the built module and profile
            depg.addModule(mod);
            return std::make_pair(mod, profile); 
        } else
            throwModuleError("field `profiles` must be an array");   
    }

    // selectBuildProfile takes the list of TOML build profiles as well as a
    // possible profile name and walks through each profile and attempts to
    // select the one that best matches the current configuration.
    BuildProfile ModuleLoader::selectBuildProfile(toml::array* tomlProfiles, std::optional<const std::string&> selectedProfileName) {
        // user specified a profile
        if (selectedProfileName) {
            for (auto& item : *tomlProfiles) {
                if (auto tomlProfile = item.as_table()) {
                    auto name = getRequiredField<std::string>(tomlProfile, "name");

                    if (name == selectedProfileName)
                        return loadProfile(tomlProfile);
                } else
                    throwModuleError("profile must be a table");          
            }

            throwModuleError(std::format("no profile by name: {}", selectedProfileName.value()));
        } 
        // these is no global/overriding profile
        else if (globalProfile.name == "") {
            for (auto& item : *tomlProfiles) {
                if (auto tomlProfile = item.as_table()) {
                    if (tomlProfile->contains("default")) {
                        if (getRequiredField<bool>(tomlProfile, "default"))
                            return loadProfile(tomlProfile);
                    }
                } else
                    throwModuleError("profile must be a table");
            }

            throwModuleError("no default profile");
        }
        // there is a global profile
        else {
            std::vector<toml::table*> matchingProfiles;

            for (auto& item : *tomlProfiles) {
                if (auto tomlProfile = item.as_table()) {
                    if (getRequiredField<std::string>(tomlProfile, "target-os") == globalProfile.targetOS 
                        && getRequiredField<std::string>(tomlProfile, "target-arch") == globalProfile.targetArch) {
                            matchingProfiles.push_back(tomlProfile);
                        }
                } else
                    throwModuleError("profile must be a table");
            }

            switch (matchingProfiles.size()) {
            case 0:
                throwModuleError("no matching profiles for module build configuration");
                break;
            case 1:
                return loadProfile(matchingProfiles[0]);
            default:
                for (auto tomlProfile : matchingProfiles) {
                    if (tomlProfile->contains("primary")) {
                        if (getRequiredField<bool>(tomlProfile, "primary"))
                            return loadProfile(tomlProfile);
                    }
                }
            }
        }
    }

    // loadProfile converts a TOML profile into a usable BuildProfile
    BuildProfile ModuleLoader::loadProfile(toml::table* tbl) {
        // load standard profile information
        BuildProfile profile = {
            .name = getRequiredField<std::string>(tbl, "name"),
            .targetOS = getRequiredField<std::string>(tbl, "target-os"),
            .targetArch = getRequiredField<std::string>(tbl, "target-arch"),          
            .targetFormat = convertBuildFormat(getRequiredField<std::string>(tbl, "format")),
            .debug = getRequiredField<bool>(tbl, "debug"),
            .outputPath = getRequiredField<std::string>(tbl, "output")  
        };

        // check that the OS is supported
        if (profile.targetOS != "windows")
            throwModuleError(std::format("unsupported OS: `{}`", profile.targetOS));

        // check that the architecture is supported
        if (profile.targetArch != "amd64" && profile.targetArch != "i386")
            throwModuleError(std::format("unsupported architecture: `{}`", profile.targetArch));

        // check for static libraries
        if (tbl->contains("static-libs")) {
            if (auto tomlStaticLibs = tbl->get("static-libs")->as_array()) {
                for (auto& item : *tomlStaticLibs) {
                    if (auto str = item.as_string())
                        profile.staticLibraries.push_back(str->get());
                    else
                        throwModuleError("elements of `static-libs` field must be string paths to static libraries");
                }
            } else
                throwModuleError("field `static-libs` must be an array");
        }
        
        // check for link objects
        if (tbl->contains("link-objects")) {
            if (auto tomlLinkObjs = tbl->get("link-objects")->as_array()) {
                for (auto& item : *tomlLinkObjs) {
                    if (auto str = item.as_string())
                        profile.linkObjects.push_back(str->get());
                    else
                        throwModuleError("elements of `link-objects` field must be string paths to object files");
                }
            } else
                throwModuleError("field `link-objects` must be an array");
        }

        return profile;
    }

    BuildFormat ModuleLoader::convertBuildFormat(const std::string& formatString) {
        if (formatString == "bin")
            return BuildFormat::Bin;
        else if (formatString == "llvm")
            return BuildFormat::LLVM;
        else if (formatString == "asm")
            return BuildFormat::Asm;
        else
            throwModuleError(std::format("unsupported build format `{}`", formatString));
    }
}