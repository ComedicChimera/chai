#ifndef PACKAGE_H_INCLUDED
#define PACKAGE_H_INCLUDED

#include <vector>

#include "util.hpp"
#include "srcfile.hpp"

namespace chai {
    // Package represents a collection of Chai files that share a common
    // namespace.
    struct Package {
        // id is the unique ID of the package
        u32 id;

        // rootDir is the root directory of the package
        std::string rootDir;

        // files contains all the source files in the package
        std::vector<SrcFile> files;
    };
}

#endif 