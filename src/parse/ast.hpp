#ifndef AST_H_INCLUDED
#define AST_H_INCLUDED

#include <vector>
#include <memory>

#include "report/position.hpp"

namespace chai {
    // ASTKind enumerates the different kinds of AST nodes
    enum class ASTKind {
        Root, // Root node of the AST for a file
    };

    // ASTNode is the parent class for all Chai AST nodes
    class ASTNode {
    public:
        virtual ASTKind kind() const = 0;
        virtual TextPosition position() const = 0;
    };

    // ASTRoot is root node for a file
    struct ASTRoot : public ASTNode {
        std::vector<ASTNode*> declarations;

        ASTKind kind() const override { return ASTKind::Root; };
        TextPosition position() const override { 
            return positionOfSpan(declarations[0]->position(), declarations.back()->position()); 
        };

        ~ASTRoot() {
            for (auto* item : declarations)
                delete item;
        }
    };
}


#endif