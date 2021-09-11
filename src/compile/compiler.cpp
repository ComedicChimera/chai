#include "compiler.hpp"

#include "depm/loader.hpp"

namespace chai {
    void Compiler::compile(const std::string& buildDir) {
        // load the root module
        ModuleLoader loader(reporter, buildDir, buildProfile);
        auto result = loader.load({});

        rootMod = new Module;
        *rootMod = result.first;
        depGraph[rootMod->id] = rootMod;
        
        buildProfile = result.second;
        
    }
}