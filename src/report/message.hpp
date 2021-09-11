#ifndef MESSAGE_H_INCLUDED
#define MESSAGE_H_INCLUDED

#include <stdexcept>
#include <string>

#include "util.hpp"
#include "position.hpp"

namespace chai {
    // CompileMessage is a warning or error that occurred during the compilation
    // of a program -- it points to a specific, problematic location in user
    // source text.
    struct CompileMessage : public std::exception {
        std::string message;
        TextPosition position;
        std::string filePath; 

        CompileMessage(const std::string& msg, const TextPosition& pos, const std::string& fpath)
        : message(msg)
        , position(pos)
        , filePath(fpath)
        {}
    };
}

#endif