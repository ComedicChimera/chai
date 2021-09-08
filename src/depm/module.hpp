#ifndef MODULE_H_INCLUDED
#define MODULE_H_INCLUDED

#include <string>
#include <unordered_map>
#include <vector>

namespace chai {
    // BuildFormat represents the kind of output the compiler can produce
    enum class BuildFormat {
        Bin,
        LLVM,
        Asm
    };

    // BuildProfile represents a specific build configuration
    struct BuildProfile {
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

        // dynamicLibraries is the list of dynamic libraries that should be
        // marked as dependencies for the final output
        std::vector<std::string> dynamicLibraries;

        // linkObjects is the list of additional objects to be linked into the
        // final output
        std::vector<std::string> linkObjects;
    };

    // Module represents a Chai module after it processed by the module loader
    struct Module {
        // id is the unique identifier of the module
        unsigned int id;

        // moduleName is the name of imported module
        std::string moduleName;

        // rootDir is the directory of the module
        std::string rootDir;

        // localImportDirs is a list of directories the compiler should check
        // for modules only in this specific module
        std::vector<std::string> localImportDirs;

        // pathReplacements is used to map one module lookup path to another
        std::unordered_map<std::string, std::string> pathReplacements;

        // shouldCache indicates whether or not compilation caching should be
        // performed for this module
        bool shouldCache;

        // TODO: rootPackage, subPackages, lastBuildTime
    };

    // loadModule takes in a module path, loads it, selects an appropriate build
    // profile, and returns the module and build profile.
    std::pair<Module, BuildProfile> loadModule(const std::string&);
}

#endif 