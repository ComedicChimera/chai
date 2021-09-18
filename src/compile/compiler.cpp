#include "compiler.hpp"

#include "depm/loader.hpp"

namespace chai {
    void Compiler::compile(const std::string& buildDir) {
        // load the root module
        ModuleLoader loader(depGraph, reporter, buildDir, buildProfile);
        auto result = loader.load({});

        rootMod = result.first;
        
        buildProfile = result.second;
        
        // TODO: ...
    }
}