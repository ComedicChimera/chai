#ifndef LOADER_H_INCLUDED
#define LOADER_H_INCLUDED

#include <optional>

#include "toml++/toml.h"

#include "module.hpp"

namespace chai {
    class ModuleLoader {
        Module mod;
        const std::string& modFilePath;
        const BuildProfile& globalProfile;

        void throwModuleError(const std::string&);

        template<typename T> 
        T getRequiredField(toml::table*, const std::string&);

        BuildProfile selectBuildProfile(toml::array*, std::optional<const std::string&>);
        BuildProfile loadProfile(toml::table*);

        BuildFormat convertBuildFormat(const std::string&);

    public:
        // ModuleLoader is constructed with the name of the module's root
        // directory and a global build profile
        ModuleLoader(const std::string&, const BuildProfile&);

        // load takes in an optional selected profile and attempts to load a
        // module from a configuration file in the root directory.  It also
        // selects a build profile from that module (if necessary) and returns a
        // build profile that should "supercede" the compiler's current one.
        std::pair<Module, BuildProfile> load(std::optional<const std::string&>);
    };
}

#endif