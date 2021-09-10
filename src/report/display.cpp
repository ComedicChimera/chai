#include "chai_error.hpp"
#include "warn.hpp"

#include <iostream>

namespace chai {
    void ChaiCompileError::display() {
        std::cout << filePath << ':' << position.startLine << ':' << position.startCol << ": " << message << '\n';
    }

    void reportWarningMessage(const std::string& message) {
        std::cout << "[warning]\n" << message << '\n';
    }
}