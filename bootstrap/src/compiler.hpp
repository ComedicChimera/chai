#ifndef _COMPILER_H_
#define _COMPILER_H_

#include <string>
#include <string_view>
#include <unordered_map>
#include <filesystem>

#include "chaic.hpp"

namespace chai {
    // OutputMode represents a type of output the compiler can produce.
    enum class OutputMode {
        EXE, // Output an executable (default).
        OBJ, // Output object files.
        ASM, // Output ASM files.
        LLVM, // Output LLVM IR source files.
    };

    // Compiler represents the global state for the Chai compiler: this is
    // effectively the compiler's "driver" class.
    class Compiler {
        // The absolute path to the compilation root directory.
        std::string m_rootDir;

        // The absolute path to the compilation output path.
        std::string m_outputPath;

        // The compilation mode.
        OutputMode m_mode { OutputMode::EXE };

        // Whether to compile in debug mode.
        bool m_debug { false };

        // The main Chai dependency graph: contains all the packages in the
        // current compilation project/unit.
        std::unordered_map<size_t, std::unique_ptr<ChaiPackage>> m_depGraph;

    public:
        // initFromArgs initializes the compiler based on the command-line
        // arguments.  It returns whether compilation should continue.
        bool initFromArgs(int argc, char* argv[]);

        // initPackage initializes a Chai package at the given absolute path. If
        // package initialization fails, then `nullptr` is returned.
        ChaiPackage* initPackage(const std::filesystem::path& absPath);

        // rootDir returns a view to the root directory of the compiler.
        inline std::string_view rootDir() const { return m_rootDir; }
    };
}

#endif