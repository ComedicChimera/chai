#include "compiler.hpp"

#include <filesystem>

#include "depm/package.hpp"
#include "util.hpp"
#include "constants.hpp"
#include "tokenize/scanner.hpp"

namespace fs = std::filesystem;

namespace chai {
    bool Compiler::initPkg(std::shared_ptr<Module> parentMod, const std::string& pkgAbsPath) {
        Package pkg = {.id = getID(), .rootDir = pkgAbsPath};

        // iterate over each file in the directory
        for (const auto& entry : fs::directory_iterator(pkgAbsPath)) {
            if (entry.is_regular_file() && entry.path().has_extension() && entry.path().extension() == CHAI_FILE_EXT) {
                SrcFile file = {.parent = pkg, .filepath = entry.path().string()};

                // create a scanner for the file
                Scanner sc(file.filepath);

                // TODO: parse the file and check for metadata

                pkg.files.push_back(file);
            }          
        }

        // TODO: check for empty packages

        // check to see if this is the root package for the module
        if (fs::equivalent(parentMod->rootDir, pkgAbsPath)) {
            // add it as the root package
            parentMod->rootPackage = pkg;
        } else {
            // otherwise, add it as a subpackage
            auto subPath = fs::relative(pkgAbsPath, parentMod->rootDir);
            parentMod->subPackages[subPath.string()] = pkg;
        }
    }
}