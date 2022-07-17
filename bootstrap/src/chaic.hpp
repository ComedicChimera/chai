#ifndef _CHAIC_H_
#define _CHAIC_H_

#include <string>
#include <string_view>
#include <vector>
#include <unordered_map>
#include <format>
#include <filesystem>

#include "report.hpp"
#include "types.hpp"

// The full name of the compiler.
#define CHAI_NAME "chaic (x86_64-windows-msvc)"

// The current Chai version string.
#define CHAI_VERSION "0.1.0"

namespace chai {
    class ChaiFile;

    /* ---------------------------------------------------------------------- */

    // Enumerates the various symbol kinds.
    enum class SymbolKind {
        VALUE, // A variable or constant.
        FUNC   // A function.
    };

    // Symbol represents a Chai symbol.
    class Symbol {
        // The name of the symbol.
        std::string m_name;

        // The type of the symbol.
        std::unique_ptr<Type> m_type;
    public:
        // The parent Chai package to this symbol.
        ChaiPackage* parent;

        // The kind of the symbol.
        SymbolKind kind;

        // The span where the symbol is defined.
        TextSpan defSpan;

        // Returns a view to the name of the symbol.
        inline std::string_view name() const { return m_name; }

        // Returns the type of the symbol.
        inline Type* type() const { return m_type.get(); }
    };

    /* ---------------------------------------------------------------------- */

    // ChaiPackage represents a Chai package: a collection of Chai source files
    // which share a common global namespace.  This is the minimum Chai
    // translation unit.
    class ChaiPackage {
        // The name of the Chai package.
        std::string m_name;

        // The absolute path to the root directory.
        std::string m_absPath;

    public:
        // The unique ID of the Chai package.
        size_t id;

        // The vector of files in the Chai package.
        std::vector<std::unique_ptr<ChaiFile>> files;

        // Creates a new Chai package.
        ChaiPackage(const std::filesystem::path& absPath)
        : m_name(absPath.filename().string())
        , m_absPath(absPath.string())
        , id { std::filesystem::hash_value(absPath) }
        {}

        // Return a view to the name of the Chai package.
        inline std::string_view name() const { return m_name; }

        // Returns a view to the absolute path of the Chai package.
        inline std::string_view absPath() const { return m_absPath; } 
    };

    // ChaiFile represents a Chai source file.
    class ChaiFile {
        // The absolute path to the Chai file.
        std::string m_absPath;

        // The display path of the Chai file.
        std::string m_displayPath;
        
        // The global symbol table for the package.
        std::unordered_map<std::string_view, std::unique_ptr<Symbol>> m_symbolTable;
    public:
        // The parent package to the Chai file.
        ChaiPackage* parent;

        // The unique number identifying the Chai source file within its package.
        size_t fileNumber;

        // Creates a new Chai file.
        ChaiFile(ChaiPackage* parent, size_t fileNumber, const std::filesystem::path& absPath)
        : parent(parent)
        , fileNumber(fileNumber)
        , m_absPath(absPath.string())
        , m_displayPath(std::format("({}) {}", parent->name(), absPath.filename().string()))
        {}

        // Returns a view to the absolute path to the Chai file.
        inline std::string_view absPath() const { return m_absPath; }

        // Returns a view to the display path of the Chai file. 
        inline std::string_view displayPath() const { return m_displayPath; }
    };
}

#endif