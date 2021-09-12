#ifndef SRCFILE_H_INCLUDED
#define SRCFILE_H_INCLUDED

#include <string>
#include <unordered_map>

#include "parse/ast.hpp"

namespace chai {
    // SrcFile represents a Chai source file and local namespace. It contains
    // the AST for that source file
    struct SrcFile {
        // parent is the parent package to the file
        Package& parent;

        // filePath is the absolute path to the file
        std::string filePath;

        // ast is the root AST node for the file
        ASTRoot* ast;

        // metadata is a dictionary of all the metadata for this file
        std::unordered_map<std::string, std::string> metadata;
    };
}



#endif 