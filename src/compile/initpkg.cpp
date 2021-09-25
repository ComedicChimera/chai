#include "compiler.hpp"

#include <filesystem>
#include <format>
#include <stdexcept>

#include "depm/package.hpp"
#include "util.hpp"
#include "constants.hpp"
#include "tokenize/scanner.hpp"
#include "parse/parser.hpp"

namespace fs = std::filesystem;

namespace chai {
    // initPkg initializes a package based on a package path
    Package* Compiler::initPkg(Module* parentMod, const std::string& pkgAbsPath) {
        if (!fs::exists(pkgAbsPath))
            throw new std::logic_error(std::format("path does not exist: `{}`", pkgAbsPath));

        Package *pkg = new Package{.parentID=parentMod->id, .rootDir = pkgAbsPath};

        // add the package to the module before parsing (to make sure it isn't
        // imported multiple times/recursively).

        // check to see if this is the root package for the module
        if (fs::equivalent(parentMod->rootDir, pkgAbsPath)) {
            // set its name equal to the module name
            pkg->name = parentMod->name;

            // add it as the root package
            parentMod->rootPackage = pkg;
        } else {
            // otherwise, add it as a subpackage
            auto subPath = fs::relative(pkgAbsPath, parentMod->rootDir);
            parentMod->subPackages[subPath.string()] = pkg;

            // sets its name equal to the last component of its path
            pkg->name = subPath.filename().string();
        }

        // iterate over each file in the directory
        for (const auto& entry : fs::directory_iterator(pkgAbsPath)) {
            if (entry.is_regular_file() && entry.path().has_extension() && entry.path().extension() == CHAI_FILE_EXT) {
                // catch errors here so parse errors don't bubble/prevent other files from parsing
                try {
                    SrcFile file = {.parent = pkg, .filePath = entry.path().string()};

                    // create a scanner for the file
                    Scanner sc(file.filePath);

                    // parse the file and store it iff parsing succeeds
                    Parser p(this, file, sc);
                    
                    if (auto result = p.parse()) {
                        file.ast = result.value();
                        pkg->files.push_back(file);
                    }
                } catch (CompileMessage& e) {
                    reporter.reportCompileError(e);
                }              
            }          
        }

        // check for empty packages
        if (pkg->files.size() == 0)
            throw std::logic_error(std::format("package `{}` contains no source files", pkg->name));

        return pkg;
    }
}