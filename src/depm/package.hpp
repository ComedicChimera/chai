#ifndef PACKAGE_H_INCLUDED
#define PACKAGE_H_INCLUDED

#include <vector>

#include "util.hpp"
#include "srcfile.hpp"
#include "module.hpp"
#include "symbols/symbol_table.hpp"

namespace chai {
    // Package represents a collection of Chai files that share a common
    // namespace.
    struct Package {
        // id is the unique ID of the package
        u64 id = getPkgID();

        // parent is the parent module of this package
        Module* parent;

        // name is the name of the package based on its directory
        std::string name;

        // rootDir is the root directory of the package
        std::string rootDir;

        // files contains all the source files in the package
        std::vector<SrcFile> files;

        // globalTable is the global symbol table for this package
        SymbolTable globalTable;

        // importedPackages is the map of packages this one depends on
        std::unordered_map<u64, Package*> importedPackages;

    private:
        // getPkgID generates a new ID for use by a given package
        static u64 getPkgID() {
            static u64 idCounter = 0;
            return idCounter++;
        }
    };
}

#endif 