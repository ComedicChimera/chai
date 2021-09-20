#include "compiler.hpp"

#include <filesystem>

namespace fs = std::filesystem;

namespace chai {
    std::optional<Package*> Compiler::importPackage(Module* parentMod, const std::string& modName, const std::string& pkgPath) {
        auto opModPath = depg.findModule(parentMod, modName);
        if (opModPath) {
            auto modPath = opModPath.value();

            // retrieve the module itself
            Module* mod;
            if (auto omod = depg.getModuleByPath(modPath))
                mod = omod.value();
            else
                mod = loadModule(modPath);

            // check if the package already exists; return the preloaded version
            // if it does or return a newly initialize version if it does not
            if (pkgPath == "" && mod->rootPackage != NULL)
                return mod->rootPackage;
            else if (pkgPath != "" && mod->subPackages.contains(pkgPath))
                return mod->subPackages[pkgPath];
            else
                return initPkg(mod, (fs::path(mod->rootDir) / pkgPath).string());
        }

        return {};
    }
}