#ifndef _AST_H_
#define _AST_H_

#include "chaic.hpp"

namespace chai {
    // ASTNode is base class of all AST nodes.
    class ASTNode {
        // The span over which the node occurs in source text.
        TextSpan m_span;
    public:
        // span returns the span where the node occurs in source text.
        inline virtual const TextSpan& span() const { return m_span; } 

        // TODO: visiting methods.
    };

    /* ---------------------------------------------------------------------- */

    // Annotation represents an annotation.
    class Annotation {
        // The annotation's name.
        std::string m_name;

        // The annotation's value.
        std::string m_value;

    public:
        // Creates a new annotation initializing fields to their given values.
        Annotation(const std::string& name, const TextSpan& nameSpan, const std::string& value, const TextSpan& valueSpan)
        : m_name(name)
        , nameSpan(nameSpan)
        , m_value(value)
        , valueSpan(valueSpan)
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
        // The resulting type of the expression.
        std::unique_ptr<Type> m_type;

    public:
        // constant returns whether this AST node is constant (immutable). 
        inline virtual bool constant() const = 0;

        // category returns the value category of the result of the expression.
        inline virtual ValueCategory category() const = 0;

        /* ------------------------------------------------------------------ */

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

    // TODO: block statements

    /* ---------------------------------------------------------------------- */

    // TODO: simple statements

    /* ---------------------------------------------------------------------- */

    // TODO: expressions
}

#endif