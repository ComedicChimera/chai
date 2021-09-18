#include "depgraph.hpp"

#include <filesystem>

#define STD_LIB_PATH "modules/std/"
#define PUB_LIB_PATH "modules/pub/"

namespace fs = std::filesystem;

namespace chai {
    std::optional<Package*> DepGraph::importPackage(Module* parentMod, const std::string& modName, const std::string& pkgPath) {
        auto opModPath = findModule(parentMod, modName);
        if (opModPath) {
            auto modPath = opModPath.value();

            // retrieve the module itself
            Module* mod;
            auto modID = getModuleID(modPath);
            if (modMap.contains(modID)) {
                mod = modMap[modID];
            } else {
                // TODO: loadModule, handle build profiles
            }


        }

        return {};
    }

    // findModule searches for the absolute path to a module based on a parent
    // module and a module name
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
        modMap[getModuleID(mod->rootDir)] = mod;
    }
}