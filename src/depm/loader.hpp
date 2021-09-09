#ifndef LOADER_H_INCLUDED
#define LOADER_H_INCLUDED

#include <optional>

#include "toml++/toml.h"

#include "module.hpp"

namespace chai {
    class ModuleLoader {
        Module mod;
        const std::string& modFilePath;

        void throwModuleError(const std::string&);

        template<typename T> 
        T getRequiredField(toml::table&, const std::string&);

    public:
        ModuleLoader(const std::string&);

        std::pair<Module, BuildProfile> load(std::optional<const std::string&>);
    };
}

#endif