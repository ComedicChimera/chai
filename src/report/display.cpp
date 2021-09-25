#include "reporter.hpp"

#include <iostream>

namespace chai {
    void Reporter::reportCompileError(const CompileMessage& msg) {
        std::cout << msg.position.filePath << ':' << msg.position.startLine << ':' << msg.position.startCol << ": " << msg.message << '\n';
    }

    void Reporter::reportWarningMessage(const std::string& message) {
        std::cout << "[warning]\n" << message << '\n';
    }
}