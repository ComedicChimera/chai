#ifndef SRCFILE_H_INCLUDED
#define SRCFILE_H_INCLUDED

#include <string>
#include <unordered_map>

#include "parse/ast.hpp"
#include "symbols/symbol.hpp"

namespace chai {
    // SrcFile represents a Chai source file and local namespace. It contains
    // the AST for that source file
    struct SrcFile {
        // parent is the parent package to the file
        Package* parent;

        // filePath is the absolute path to the file
        std::string filePath;

        // ast is the root AST node for the file
        ASTRoot* ast;

        // metadata is a dictionary of all the metadata for this file
        std::unordered_map<std::string, std::string> metadata;

        // importedSymbols is the map of all imported defined symbols for this
        // file -- organized by name (not a symbol table since these symbols are
        // not managed by this file)
        std::unordered_map<std::string, Symbol*> importedSymbols;

        // visiblePackages is the map of all packages that are visible as
        // symbols within this file.  It is organized by the alias that the
        // package can be referred to within this source file
        std::unordered_map<std::string, Package*> visiblePackages;
    };
}



#endif 