#ifndef POSITION_H_INCLUDED
#define POSITION_H_INCLUDED

#include "util.hpp"

namespace chai {
    // TextPosition is used to indicate where in a given source file an error
    // occurred for the purposes of error reporting.
    struct TextPosition {
        u32 startLine, startCol;
        u32 endLine, endCol;
    };

    // positionOfSpan takes two text positions and returns a position that spans
    // between and included them
    inline TextPosition positionOfSpan(const TextPosition& start, const TextPosition& end) {
        return TextPosition{
            .startLine = start.startLine,
            .startCol = start.startCol,
            .endLine = end.endLine,
            .endCol = end.endCol
        };
    }
}

#endif