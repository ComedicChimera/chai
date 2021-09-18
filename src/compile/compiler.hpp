#ifndef COMPILER_H_INCLUDED
#define COMPILER_H_INCLUDED

#include <unordered_map>
#include <string>

#include "util.hpp"
#include "depm/module.hpp"
#include "report/reporter.hpp"
#include "depm/depgraph.hpp"

namespace chai {
    // Compiler holds the high-level state and acts as the main controller for
    // the compilation process
    class Compiler {
        // depGraph is the compiler's modular dependency graph
        DepGraph depGraph;

        // buildProfile is the global build configuration used in module loading
        BuildProfile buildProfile;

        // rootMod is the root/main module for whole project
        Module* rootMod;

        // reporter is the global reporter used for all of compilation
        Reporter reporter;

        bool initPkg(Module*, const std::string&);

    public:
        // compile takes a build directory and runs the main compilation
        // algorithm taking it to be the root
        void compile(const std::string&);        
    };
}


#endif