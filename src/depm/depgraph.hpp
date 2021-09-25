#ifndef DEPGRAPH_H_INCLUDED
#define DEPGRAPH_H_INCLUDED

#include <unordered_map>
#include <string>
#include <optional>

#include "module.hpp"
#include "util.hpp"

namespace chai {
    // DepGraph represents the modular dependency graph for Chai.  It is also
    // the owner of all modules and is a singleton.
    class DepGraph {
        std::hash<std::string> h;
        std::unordered_map<u64, Module*> modMap;
        std::string chaiPath;

        // genModuleID returns a module ID based on a module's root directory
        inline u64 genModuleID(const std::string& modDir) const { return h(modDir); };
    public:
        DepGraph();

        // findModule attempts to determine the path to a module based on a
        // parent module (the module that is importing it) and a module name
        std::optional<std::string> findModule(Module*, const std::string&);

        // addModule adds a new module to the dependency graph
        void addModule(Module*);

        // getModuleByID gets a module by its ID if it exists
        std::optional<Module*> getModuleByID(u64);

        // getModuleByPath gets a module by its module root path
        inline std::optional<Module*> getModuleByPath(const std::string& modPath) { 
            return getModuleByID(genModuleID(modPath));
        };

        ~DepGraph() {
            for (auto pair : modMap)
                delete pair.second;
        }
    };
}

#endif