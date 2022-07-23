#ifndef _AST_H_
#define _AST_H_

#include "chaic.hpp"
#include "syntax/token.hpp"

namespace chai {
    // Annotation represents an annotation.
    class Annotation {
        // The annotation's name.
        std::string m_name;

        // The annotation's value.
        std::string m_value;

    public:
        // Creates a new annotation initializing fields to their given values.
        Annotation(std::string& name, const TextSpan& nameSpan, std::string& value, const TextSpan& valueSpan)
        : m_name(std::move(name))
        , nameSpan(nameSpan)
        , m_value(std::move(value))
        , valueSpan(valueSpan)
        {}

        // Create a new annotation with no value.
        Annotation(std::string& name, const TextSpan& nameSpan) 
        : m_name(std::move(name))
        , nameSpan(nameSpan)
        {}

        // The span where the name occurs in source text.
        TextSpan nameSpan;

        // The span where the value occurs in source text (if any).
        TextSpan valueSpan;

        // name returns a view to the annotation's name.
        inline std::string_view name() const { return m_name; }

        // value returns a view to the annotation's value.
        inline std::string_view value() const { return m_value; }
    };

    // AnnotationMap is the type used to represent a map of annotations.
    using AnnotationMap = std::unordered_map<std::string_view, std::unique_ptr<Annotation>>;

    // FuncDef represents a function definition.
    class FuncDef : public ASTNode {
        // The function's optional body: this may be a null pointer.
        std::unique_ptr<ASTNode> m_body;

    public:
        // Creates a new function definition.
        FuncDef(
            const TextSpan& span,
            AnnotationMap& annots, 
            Symbol* symbol, 
            std::vector<std::unique_ptr<Symbol>>& params, 
            ASTNode* body
        )
        : annotations(std::move(annots))
        , symbol(symbol)
        , params(std::move(params))
        , m_body(std::unique_ptr<ASTNode>(body))
        {
            m_span = span;
        }

        // The function's symbol.
        Symbol* symbol;

        // The function's parameters.
        std::vector<std::unique_ptr<Symbol>> params;

        // The function's annotations.
        AnnotationMap annotations;

        // body returns the function definition's body.  If there is no body, then
        // the a null pointer is returned.
        inline ASTNode* body() const { return m_body.get(); }
    };

    // OperDef represents an operator definition.
    class OperDef : public ASTNode {
        // The operator's optional body: this may be a null pointer.
        std::unique_ptr<ASTNode> m_body;

    public:
        // The string value of the operator used when displaying it.
        std::string_view opSym;

        // The operator overload corresponding to this operator definition.
        OperatorOverload* overload;

        // The operator's parameters.
        std::vector<std::unique_ptr<Symbol>> params;

        // The operator's annotations.
        AnnotationMap annotations;

        // body returns the operator definition's body.  If there is no body,
        // then the a null pointer is returned.
        inline ASTNode* body() const { return m_body.get(); }
    };

    /* ---------------------------------------------------------------------- */

    // Enumeration of possible Chai value categories.
    enum class ValueCategory {
        LVALUE,
        RVALUE
    };

    // ASTExpr is the base class for all AST expressions.
    class ASTExpr : public ASTNode {
    protected:
        // The resulting type of the expression.
        std::unique_ptr<Type> m_type;

    public:
        // constant returns whether this AST node is constant (immutable). 
        inline virtual bool constant() const { return true; }

        // category returns the value category of the result of the expression.
        inline virtual ValueCategory category() const { return ValueCategory::RVALUE; }

        // type returns the resulting type of the expression.
        inline virtual Type* type() const { return m_type.get(); }
    };

    // AppliedOperator represents a particular application of an operator.
    struct AppliedOperator {
        // The token corresponding to the operator.
        Token token;

        // The operator overload used for this particular application.
        OperatorOverload* overload;
    };

    /* ---------------------------------------------------------------------- */

    // Identifier represents an named symbol used in source code.
    class Identifier : public ASTExpr {
    public:
        // The symbol associated with this identifier.
        Symbol* symbol;

        inline Type* type() const override { return symbol->type(); }

        inline bool constant() const override { return symbol->constant; }

        inline ValueCategory category() const override { return ValueCategory::LVALUE; }
    };

    // Literal represents a literal value.
    class Literal : public ASTExpr {
        // The text value of the literal.
        std::string m_text;

    public:
        // The token kind that generated the literal.
        TokenKind kind;

        // text returns a view to the text value of the literal.
        inline std::string_view text() const { return m_text; }
    };

    // Indirect represents a pointer indirection operation.
    class Indirect : public ASTExpr {
        // The element being indirected.
        std::unique_ptr<ASTExpr> m_elem;

    public:
        // Whether or not the pointer indirection is constant.
        bool isConstIndirect;

        // elem returns a pointer to the element being indirected.
        inline ASTExpr* elem() const { return m_elem.get(); }
    };

    // Dereference represents a pointer dereference.
    class Dereference : public ASTExpr {
        // The pointer being dereferenced.
        std::unique_ptr<ASTExpr> m_pointer;

    public:
        // pointer returns a pointer to the pointer being dereferenced.
        inline ASTExpr* pointer() const { return m_pointer.get(); }

        inline bool constant() const override { 
            // TODO: dynamic cast to a pointer type
            return false;
        }
    };

    // UnaryOperApp represents a unary operator application.
    class UnaryOperApp : public ASTExpr {
        // The operand of the unary operator.
        std::unique_ptr<ASTExpr> m_operand;

    public:
        // The operator being applied.
        AppliedOperator appliedOp;

        // operand returns a pointer to the operand.
        inline ASTExpr* operand() const { return m_operand.get(); }
    };

    // BinaryOperApp represents a binary operator application.
    class BinaryOperApp : public ASTExpr {
        // The left-hand operand of the binary operator.
        std::unique_ptr<ASTExpr> m_lhs;

        // The right-hand operand of the binary operator.
        std::unique_ptr<ASTExpr> m_rhs;

    public:
        // The operator being applied.
        AppliedOperator appliedOp;

        // lhs returns a pointer to the left-hand operand.
        inline ASTExpr* lhs() const { return m_lhs.get(); }

        // rhs returns a pointer to the right-hand operand.
        inline ASTExpr* rhs() const { return m_rhs.get(); }
    };

    // TypeCast represents a type cast.  The destination type is returned by 
    // the `type()` method.
    class TypeCast : public ASTExpr {
        // The value being type cast.
        std::unique_ptr<ASTExpr> m_src;

    public:
        // src returns a pointer to the value being cast.
        inline ASTExpr* src() const { return m_src.get(); }
    };

    /* ---------------------------------------------------------------------- */

    // VarList represents a list of variables with a common initializer.
    class VarList {
        // The optional initializer.
        std::unique_ptr<ASTExpr> m_initializer;

    public:
        // The variable symbol's being initialized.
        std::vector<std::unique_ptr<Symbol>> vars;

        // initializer returns a pointer to the initializer if it exists.
        inline ASTExpr* initializer() const { return m_initializer.get(); }
    };

    // VarDecl represents a variable declaration.
    class VarDecl : public ASTNode {
    public:
        // The variable's being declaration.
        std::vector<VarList> varLists;
    };

    // Assignment represents an assignment statement.
    class Assignment : public ASTNode {
    public:
        // The LHS expressions.
        std::vector<std::unique_ptr<ASTExpr>> lhsExprs;

        // The RHS expressions.
        std::vector<std::unique_ptr<ASTExpr>> rhsExprs;

        // The compound operator being applied if any.
        std::optional<AppliedOperator> compoundOp;
    };

    // IncDecStmt represents an increment/decrement statement.
    class IncDecStmt : public ASTNode {
        // The operand of the increment/decrement statement.
        std::unique_ptr<ASTExpr> m_operand;

    public:
        // The operator being applied.
        AppliedOperator appliedOp;

        // operand returns a pointer to the operand of inc/dec operator.
        inline ASTExpr* operand() const { return m_operand.get(); }
    };

    // KeywordStmt represents a statement comprised of a single keyword such as
    // `break` or `continue`.
    class KeywordStmt : public ASTNode {
    public:
        // The keyword kind of the statement.
        TokenKind keywordKind;
    };

    // ReturnStmt represents a `return` statement.
    class ReturnStmt : public ASTNode {
    public:
        // The values being returned (if any).
        std::vector<std::unique_ptr<ASTExpr>> returnValues;
    };

    /* ---------------------------------------------------------------------- */

    // Block represents a collection of statements.
    class Block : public ASTNode {
    public:
        // The statements comprising the block.
        std::vector<std::unique_ptr<ASTNode>> statements;
    };

    // CondBranch represents a single conditional branch of statement.
    class CondBranch {
        // The optional header variable declaration.
        std::unique_ptr<VarDecl> m_headerVarDecl;
        
        // The condition of the branch.
        std::unique_ptr<ASTExpr> m_condition;

        // The body of the conditional branch.
        std::unique_ptr<Block> m_block;

    public:
        // headerVarDecl returns a pointer to header variable declaration if it
        // exists.  Otherwise, `nullptr` is returned.
        inline ASTNode* headerVarDecl() const { return m_headerVarDecl.get(); }

        // condition returns a pointer to the conditional branch's condition.
        inline ASTExpr* condition() const { return m_condition.get(); }  

        // block returns a pointer to the block.
        inline Block* block() const { return m_block.get(); }
    };

    // IfTree represents a complex if statement tree.
    class IfTree : public ASTNode {
        // The optional else block.
        std::unique_ptr<Block> m_elseBlock;

    public:
        // The conditional branches
        std::vector<CondBranch> branches;

        // elseBlock returns a pointer to the `else` block if it exists.
        // Otherwise, `nullptr` is returned.
        inline Block* elseBlock() const { return m_elseBlock.get(); }
    };

    // WhileLoop represents a while loop.
    class WhileLoop : public ASTNode {
        // The optional else block.
        std::unique_ptr<Block> m_elseBlock;

    public:
        // The main conditional branch of the while loop.
        CondBranch branch;

        // elseBlock returns a pointer to the `else` block if it exists.
        // Otherwise, `nullptr` is returned.
        inline Block* elseBlock() const { return m_elseBlock.get(); }
    };

    // CForLoop represents a C-style for loop.
    class CForLoop : public ASTNode {
        // The optional header variable declaration.
        std::unique_ptr<VarDecl> m_headerVarDecl;
        
        // The condition of the branch.
        std::unique_ptr<ASTExpr> m_condition;

        // The body of the conditional branch.
        std::unique_ptr<Block> m_block;

        // The optional update statement.
        std::unique_ptr<ASTNode> m_updateStmt;

        // The optional else block.
        std::unique_ptr<Block> m_elseBlock;

    public:
        // headerVarDecl returns a pointer to header variable declaration if it
        // exists.  Otherwise, `nullptr` is returned.
        inline ASTNode* headerVarDecl() const { return m_headerVarDecl.get(); }

        // condition returns a pointer to the conditional branch's condition.
        inline ASTExpr* condition() const { return m_condition.get(); }  

        // block returns a pointer to the block.
        inline Block* block() const { return m_block.get(); }

        // updateStmt returns a pointer to the loop update statement if it exists.
        // Otherwise, `nullptr` is returned.
        inline ASTNode* updateStmt() const { return m_updateStmt.get(); }
        
        // elseBlock returns a pointer to the `else` block if it exists.
        // Otherwise, `nullptr` is returned.
        inline Block* elseBlock() const { return m_elseBlock.get(); }
    };
}

#endif