#ifndef WARN_H_INCLUDED
#define WARN_H_INCLUDED

#include <string>

namespace chai {
    // reportWarningMessage is used to emit a warning message to the user that
    // has no position within the code
    void reportWarningMessage(const std::string&);
}

#endif