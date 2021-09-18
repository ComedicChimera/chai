#include "depgraph.hpp"

namespace chai {
    std::optional<Package*> DepGraph::importPackage(const std::string& modName, const std::string& pkgPath) {
        // TODO
        return {};
    }

    void DepGraph::addModule(Module* mod) {
        modMap[getModuleID(mod->rootDir)] = mod;
    }
}