#include "module.hpp"

#include <filesystem>
#include <stdexcept>

#include "toml++/toml.h"

#define MODULE_FILENAME "chai-mod.toml"

namespace fs = std::filesystem;

namespace chai {
    std::pair<Module, BuildProfile> loadModule(const std::string& moduleDir) {
        auto modFilePath = fs::path(moduleDir) / fs::path(MODULE_FILENAME);

        toml::table tbl;
        try {  
            tbl = toml::parse_file(modFilePath.string());
        } catch (toml::parse_error& err) {
            throw std::logic_error(std::format("[module error]\n{}:{}:{}: {}", 
                modFilePath, 
                err.source().begin.line, 
                err.source().begin.column,
                err.description()
                ));
        }

        return std::make_pair(Module{.rootDir = moduleDir}, BuildProfile{.debug = true});
    }
}