#include "depgraph.hpp"

#include <filesystem>
#include <cstdlib>
#include <stdexcept>

#define STD_LIB_PATH "modules/std/"
#define PUB_LIB_PATH "modules/pub/"

namespace fs = std::filesystem;

namespace chai {
    DepGraph::DepGraph() {
        auto chaiPathEnvVal = std::getenv("CHAI_PATH");
        if (chaiPathEnvVal == NULL)
            throw std::logic_error("missing environment variable: `CHAI_PATH`");

        chaiPath = chaiPathEnvVal;
    }

    std::optional<std::string> DepGraph::findModule(Module* parentMod, const std::string& modName) {
        // first check to see if we are importing the current module
        if (parentMod->name == modName)
            return parentMod->rootDir;

        // then, see if we are importing from the modules local import directories
        for (auto& localDir : parentMod->localImportDirs) {
            auto subPath = fs::path(localDir) / fs::path(modName);
            if (fs::exists(subPath))
                return subPath.string();
        }

        // then, check for any public/global imports
        auto globPath = fs::path(chaiPath) / fs::path(PUB_LIB_PATH) / fs::path(modName);
        if (fs::exists(globPath))
            return globPath.string();

        // finally, check for standard imports
        auto stdPath = fs::path(chaiPath) / fs::path(STD_LIB_PATH) / fs::path(modName);
        if (fs::exists(stdPath))
            return stdPath.string();

        // no matching path was found
        return {};
    }

    void DepGraph::addModule(Module* mod) {
        mod->id = genModuleID(mod->rootDir);
        modMap[mod->id] = mod;
    }

    std::optional<Module*> DepGraph::getModuleByID(u64 id) {
        if (modMap.contains(id))
            return modMap[id];

        return {};
    }
}