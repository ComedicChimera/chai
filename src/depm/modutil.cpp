#include "module.hpp"
#include "depgraph.hpp"

#include <filesystem>

namespace fs = std::filesystem;

namespace chai {
    std::string Module::getErrorPath(const std::string& path) {
        return fs::relative(path, rootDir).string();
    }
}