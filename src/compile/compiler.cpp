#include "compiler.hpp"

#include "depm/loader.hpp"

namespace chai {
    void Compiler::compile(const std::string& buildDir) {
        // load the root module
        rootMod = loadModule(buildDir);
        
        // TODO: ...
    }

    // loadModule takes in a module path (relative or absolute), loads it,
    // selects an appropriate build profile, updates the global profile, and
    // returns the module
    Module* Compiler::loadModule(const std::string& modPath, const std::string& selectedProfileName = "") {
        ModuleLoader loader(reporter, modPath, buildProfile);
        auto result = loader.load(selectedProfileName);
        buildProfile = result.second;

        depg.addModule(result.first);
        return result.first;
    }
}