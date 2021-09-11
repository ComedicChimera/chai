#ifndef COMPILER_H_INCLUDED
#define COMPILER_H_INCLUDED

#include <unordered_map>
#include <string>
#include <memory>

#include "util.hpp"
#include "depm/module.hpp"

namespace chai {
    // Compiler holds the high-level state and acts as the main controller for
    // the compilation process
    class Compiler {
        // depGraph represents the dependency graph of the compiler.  It stores
        // all the modules (and therefore packages, files, etc.) being compiled.
        std::unordered_map<u32, std::shared_ptr<Module>> depGraph;

        // buildProfile is the global build configuration used in module loading
        BuildProfile buildProfile;

        // rootMod is the root/main module for whole project
        std::shared_ptr<Module> rootMod;

    public:
        // compile takes a build directory and runs the main compilation
        // algorithm taking it to be the root
        void compile(const std::string&);        
    };
}


#endif