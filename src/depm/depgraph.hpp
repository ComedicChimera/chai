#ifndef DEPGRAPH_H_INCLUDED
#define DEPGRAPH_H_INCLUDED

#include <unordered_map>
#include <string>
#include <optional>

#include "package.hpp"
#include "util.hpp"

namespace chai {
    // DepGraph represents the modular dependency graph for Chai
    class DepGraph {
        std::hash<std::string> h;
        std::unordered_map<u64, Module*> modMap;
        std::string chaiPath;

        std::optional<std::string> findModule(Module*, const std::string&);

    public:
        // getModuleID returns a module ID based on a module's root directory
        inline u64 getModuleID(const std::string& modDir) const { return h(modDir); };

        // importPackage attempts to import package and add its relevant modules
        // to the dependency graph.  It accepts a parent module, a module name,
        // and a package path as input
        std::optional<Package*> importPackage(Module*, const std::string&, const std::string&);

        // addModule adds a new module to the dependency graph
        void addModule(Module*);

        ~DepGraph() {
            for (auto pair : modMap)
                delete pair.second;
        }
    };
}

#endif