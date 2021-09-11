#ifndef WARN_H_INCLUDED
#define WARN_H_INCLUDED

#include <string>

#include "message.hpp"

namespace chai {
    // Reporter is the error reporting module that is used to log errors in
    // the program
    class Reporter {
    public:
        // reportWarningMessage is used to emit a warning message to the user
        // that has no position within the code.
        void reportWarningMessage(const std::string&);

        // reportCompileError reports an positioned compile error as an error message
        // pointing to a specific location within the program.
        void reportCompileError(const CompileMessage&);
    }; 
}

#endif