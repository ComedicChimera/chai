#ifndef COMPILER_H_INCLUDED
#define COMPILER_H_INCLUDED

#include <unordered_map>
#include <string>
#include <optional>

#include "util.hpp"
#include "depm/module.hpp"
#include "report/reporter.hpp"
#include "depm/depgraph.hpp"

namespace chai {
    // Compiler holds the high-level state and acts as the main controller for
    // the compilation process
    class Compiler {
        // depg is the compiler's modular dependency graph
        DepGraph depg;

        // buildProfile is the global build configuration used in module loading
        BuildProfile buildProfile;

        // rootMod is the root/main module for whole project
        Module* rootMod;

        // reporter is the global reporter used for all of compilation
        Reporter reporter;

        Package* initPkg(Module*, const std::string&);
        Module* loadModule(const std::string&, const std::string& = "");

    public:
        // compile takes a build directory and runs the main compilation
        // algorithm taking it to be the root
        void compile(const std::string&);     

        // importPackage imports a package belonging to a named module at at a
        // named import path and add its relevant modules to the dependency
        // graph.  It accepts a parent module, a module name, and a package path
        // as input.  It returns a pointer to the initialized package
        std::optional<Package*> importPackage(Module*, const std::string&, const std::string&);

        // getProfile returns a const reference to the current global build profile
        inline const BuildProfile& getProfile() const { return buildProfile; }
    };
}


#endif