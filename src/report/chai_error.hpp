#ifndef CHAI_ERROR_H_INCLUDED
#define CHAI_ERROR_H_INCLUDED

#include <stdexcept>
#include <string>

namespace chai {
    // TextPosition is used to indicate where in a given source file an error
    // occurred for the purposes of error reporting.
    struct TextPosition {
        int startLine, startCol;
        int endLine, endCol;
    };

    // ChaiCompileError is an error that occurred during the compilation of a
    // program -- it points to a specific, erroneous location in user source
    // text.
    class ChaiCompileError : public std::exception {
        std::string message;
        TextPosition position;
        std::string filePath; 

    public:
        ChaiCompileError(const std::string& msg, TextPosition pos, const std::string& fpath)
        : message(msg)
        , position(pos)
        , filePath(fpath)
        {}

        void display(); 
    };
}

#endif