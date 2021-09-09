#include "loader.hpp"

#include <filesystem>
#include <stdexcept>

#define MODULE_FILENAME "chai-mod.toml"

namespace fs = std::filesystem;

namespace chai {
    ModuleLoader::ModuleLoader(const std::string& modDir) 
    : modFilePath((fs::path(modDir) / fs::path(MODULE_FILENAME)).string())
    , mod{.rootDir = modDir}
    {
    }

    void ModuleLoader::throwModuleError(const std::string& message) {
        throw std::logic_error(std::format("[module error]\n{}: {}", modFilePath, message));
    }

    template<typename T>
    T ModuleLoader::getRequiredField(toml::table& tbl, const std::string& fieldName) {
        std::optional<T> field = tbl[fieldName].value();
        if (!field)
            throwModuleError(std::format("missing required field: `{}`", fieldName));
    }

    std::pair<Module, BuildProfile> ModuleLoader::load(std::optional<const std::string&> profile) {
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
        std::optional<toml::table> tomlMod = tbl["module"].value<toml::table>();
        if (!tomlMod)
            throwModuleError("missing required table `module`");


        // TODO: begin accessing the fields in the module

        return std::make_pair(mod, BuildProfile{.debug = true});    
    }
}