#ifndef SYMBOL_H_INCLUDED
#define SYMBOL_H_INCLUDED

#include <string>

#include "report/position.hpp"
#include "util.hpp"

namespace chai {
    // Mutability enumerates the different kinds of mutability statuses a symbol
    // can have
    enum class Mutability {
        Immutable,
        NeverMutated,
        Mutable
    };

    // DefKind is an enumation of the different ways a symbol can be defined
    enum class DefKind {
        Function,
        Variable,
        TypeDef,
        Unknown
    };

    // Symbol represents a defined/declared symbol in Chai
    struct Symbol {
        // name is the declared name of the symbol
        std::string name;

        // parentID the package ID of the pkg the symbol is defined in
        u64 parentID;

        // isPublic indicates whether or not this symbol is externally visible
        bool isPublic;

        // TODO: data type

        // mutability indicates how/if this symbol can be mutated.  It is used
        // to track both immutability and symbol usage for optimization purposes
        // (implicit constancy)
        Mutability mutability;

        // defKind indicates the source for this symbol's definition (function,
        // type def, etc.).  This is used for usage checking.
        DefKind defKind;

        // position is the text position of this symbol's definition (eg. the
        // `func` in `def func()`)
        TextPosition position;
    };
}

#endif