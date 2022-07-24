#ifndef _CHAIC_H_
#define _CHAIC_H_

#include <string>
#include <string_view>
#include <vector>
#include <unordered_map>
#include <format>
#include <filesystem>
#include <memory>

// The full name of the compiler.
#define CHAI_NAME "chaic (x86_64-windows-msvc)"

// The current Chai version string.
#define CHAI_VERSION "0.1.0"

/* -------------------------------------------------------------------------- */

namespace chai {
     // Enumeration of different kinds of types.
    enum class TypeKind {
        UNIT,
        BOOL,
        INTEGER,
        FLOATING,
        POINTER,
        FUNCTION,
    };

    // Type represents a Chai data type.
    class Type {
    public:
        // kind returns type's kind.
        virtual TypeKind kind() const = 0;

        // size returns the size of the type in bytes.
        virtual size_t size() const = 0;

        // repr returns the string represents of the type.
        virtual std::string repr() const = 0;

        /* ---------------------------------------------------------------------- */

        // innerType returns the "inner type" of self.  For most types, this is just
        // an identity function; however, for types such as untyped and aliases
        // which essentially just wrap other types, this method unwraps them.
        inline virtual Type* innerType() { return this; }

        // align returns the alignment of the type in bytes.
        inline virtual size_t align() const { return size(); }

        // isNullable returns whether the type is nullable.
        inline virtual bool isNullable() const { return true; }

        /* ------------------------------------------------------------------ */

        // equals returns whether this type is equal the type `other`.
        inline bool equals(Type* other) { 
            return innerType()->internalEquals(other->innerType());
        }

        // castFrom returns whether the type `other` can be cast to this type.
        inline bool castFrom(Type* other) {
            return innerType()->internalCastFrom(other->innerType());
        }

        // bitSize returns the size of the type in bits.
        inline size_t bitSize() const { return size() * 8; }

        // bitAlign returns the alignment of the type in bits.
        inline size_t bitAlign() const { return align() * 8; }

    protected:
        // internalEquals returns whether this type is equal to the type `other`.
        // This method doesn't handle inner types and should not be called from
        // outside the implementation of `equals`.
        virtual bool internalEquals(Type* other) const = 0;

        /* ------------------------------------------------------------------ */

        // internalCastFrom returns whether you can cast the type `other` to
        // this type. This method need only return whether a cast is strictly
        // possible (ie. it doesn't need to check if two types are equal) and
        // doesn't handle inner types.  It should not be called from outside the
        // implementation of `castFrom`.
        inline virtual bool internalCastFrom(Type* other) const { return false; }
    };

    /* ---------------------------------------------------------------------- */

    // TextSpan represents a positional range in user source code.
    struct TextSpan {
        size_t startLine { 0 }, startCol { 0 };
        size_t endLine { 0 }, endCol { 0 };

        // Creates an empty Text Span.
        TextSpan() {}

        // Creates a new span which initializes all the fields to the
        // corresponding values.
        TextSpan(size_t startLine, size_t startCol, size_t endLine, size_t endCol)
        : startLine(startLine)
        , startCol(startCol)
        , endLine(endLine)
        , endCol(endCol)
        {}

        // Creates a new text span spanning from `start` to `end`.
        TextSpan(const TextSpan& start, const TextSpan& end)
        : startLine(start.startLine)
        , startCol(start.startCol)
        , endLine(end.endLine)
        , endCol(end.endCol)
        {}
    };

    /* ---------------------------------------------------------------------- */

    class ChaiFile;

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

        // The span where the symbol is defined.
        TextSpan m_defSpan;
    public:
        Symbol(
            ChaiFile* parent,
            std::string&& name,
            Type* type,
            const TextSpan& defSpan,
            SymbolKind kind,
            bool constant
        )
        : parent(parent)
        , m_name(name)
        , m_type(std::unique_ptr<Type>(type))
        , m_defSpan(defSpan)
        , kind(kind)
        , constant(constant)
        {}

        // The Chai source file containing the symbol's definition.
        ChaiFile* parent;

        // The kind of the symbol.
        SymbolKind kind;

        // Whether or not the symbol is constant.
        bool constant;

        // name returns a view to the name of the symbol.
        inline std::string_view name() const { return m_name; }

        // type returns the type of the symbol.
        inline Type* type() const { return m_type.get(); }

        // defSpan returns the span where the symbol is defined in source text.
        inline const TextSpan& defSpan() const { return m_defSpan; }
    };

    /* ---------------------------------------------------------------------- */

    // OperatorOverload represents a single overload of a particular operator.
    class OperatorOverload {
        // The counter used to generate new overload IDs.
        static inline size_t m_idCounter { 0 };

        // The span where the overload symbol is defined.
        TextSpan m_defSpan;

        // The intrinsic generator associated with this overload if any.
        std::string m_intrinsicName;

    public:
        // The unique ID of the operator overload.
        size_t id;

        // The Chai source file containing the overload's definition.
        ChaiFile* parent;

        // The signature of the operator overload.
        Type* signature;

        // defSpan returns the span where the operator symbol is defined.
        inline const TextSpan& defSpan() const { return m_defSpan; }

        // intrinsicName returns a view to the intrinsic generator associated
        // with this overload.
        inline std::string_view intrinsicName() const { return m_intrinsicName; }
    };

    // Operator represents all the visible definitions for an operator of a
    // given arity.
    class Operator {
        // The string value of the operator used when displaying it.
        std::string m_opSym;

    public:
        // The kind of the operator.
        int kind;

        // The arity of the overloads associated with this operator.
        int arity;

        // The overloads defined for this operator.
        std::vector<std::unique_ptr<OperatorOverload>> overloads;

        // opSym returns a view to the string value of the operator.
        inline std::string_view opSym() const { return m_opSym; }
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

        // The package's global symbol table.
        std::unordered_map<std::string_view, std::unique_ptr<Symbol>> m_symbolTable;

    public:
        // The unique ID of the Chai package.
        size_t id;

        // The vector of files in the Chai package.
        std::vector<std::unique_ptr<ChaiFile>> files;

        // Creates a new Chai package.
        ChaiPackage(const std::filesystem::path& absPath)
        : m_name(absPath.filename().string())
        , m_absPath(absPath.string())
        , id(std::filesystem::hash_value(absPath))
        {}

        // name returns a view to the name of the Chai package.
        inline std::string_view name() const { return m_name; }

        // absPath returns a view to the absolute path of the Chai package.
        inline std::string_view absPath() const { return m_absPath; } 

        // define attempts to define a new symbol in the package's global symbol
        // table.  If definition succeeds, the global symbol table takes
        // ownership of the passed in symbol.  If the definition fails, the
        // symbol is deleted and a compile error is thrown.  
        void define(Symbol* symbol);
    };

    /* ---------------------------------------------------------------------- */

    // ASTNode is base class of all AST nodes.
    class ASTNode {
    protected:
        // The span over which the node occurs in source text.
        TextSpan m_span;

    public:
        // span returns the span where the node occurs in source text.
        inline virtual const TextSpan& span() const { return m_span; } 

        // TODO: visiting methods.
    };

    /* ---------------------------------------------------------------------- */

    // ChaiFile represents a Chai source file.
    class ChaiFile {
        // The absolute path to the Chai file.
        std::string m_absPath;

        // The display path of the Chai file.
        std::string m_displayPath;
    public:
        // The parent package to the Chai file.
        ChaiPackage* parent;

        // The unique number identifying the Chai source file within its package.
        size_t fileNumber;

        // The AST definitions contained in the file.
        std::vector<std::unique_ptr<ASTNode>> definitions;

        // Creates a new Chai file.
        ChaiFile(ChaiPackage* parent, size_t fileNumber, const std::filesystem::path& absPath)
        : parent(parent)
        , fileNumber(fileNumber)
        , m_absPath(absPath.string())
        , m_displayPath(std::format("({}) {}", parent->name(), absPath.filename().string()))
        {}

        // absPath returns a view to the absolute path to the Chai file.
        inline std::string_view absPath() const { return m_absPath; }

        // displayPath returns a view to the display path of the Chai file. 
        inline std::string_view displayPath() const { return m_displayPath; }
    };
}

#endif