#ifndef MODULE_H_INCLUDED
#define MODULE_H_INCLUDED

#include <string>
#include <unordered_map>
#include <vector>

#include "util.hpp"
#include "package.hpp"

namespace chai {
    // BuildFormat represents the kind of output the compiler can produce
    enum class BuildFormat {
        Bin,
        LLVM,
        Asm
    };

    // BuildProfile represents a specific build configuration
    struct BuildProfile {
        // name is the user-defined name of this profile
        std::string name;

        // targetOS is the name of the target operating system.  It must be one
        // of the supported operating system names.
        std::string targetOS;

        // targetArch is the name of the target architecture.  It must be one of
        // the supported architecture names.
        std::string targetArch;

        // targetFormat indicates the target output format.
        BuildFormat targetFormat;

        // debug indicates whether or not to build in debug module
        bool debug;

        // outputPath is where Chai should dump the executable (or output files)
        std::string outputPath;

        // staticLibraries is the list of static libraries to be linked into
        // the final build output.  These are absolute paths.
        std::vector<std::string> staticLibraries;

        // linkObjects is the list of additional objects to be linked into the
        // final output
        std::vector<std::string> linkObjects;
    };

    // Module represents a Chai module after it processed by the module loader
    struct Module {
        // id is the unique identifier of the module
        u64 id;

        // name is the name of module
        std::string name;

        // rootDir is the absolute directory of the module
        std::string rootDir;

        // localImportDirs is a list of directories the compiler should check
        // for modules only in this specific module
        std::vector<std::string> localImportDirs;

        // shouldCache indicates whether or not compilation caching should be
        // performed for this module
        bool shouldCache;

        // cacheDirectory is the directory where the module should store its
        // cached version.  Defaults to `.chai/cache`
        std::string cacheDirectory;

        // dependsOn enumerates the other modules that this module depends on
        std::unordered_map<u64, Module*> dependsOn;

        // rootPackage is the package at the root of the module directory
        Package* rootPackage = NULL;

        // subPackages is a list of all the packages contained in subdirectories
        // of this module; ie. they are subordinate to this module.  They are
        // organized by their subpath which is of the form `/pkg1/pkg2` (with any
        // extra slashes for lower levels of depth)
        std::unordered_map<std::string, Package*> subPackages;

        // TODO: lastBuildTime

        // getErrorPath converts an absolute path to a file into a relative path
        // with respect to its parent module that can be used for error messages
        std::string getErrorPath(const std::string&);

        ~Module() {
            delete rootPackage;

            for (auto& pair : subPackages) {
                delete pair.second;
            }
        }
    };
}

#endif 