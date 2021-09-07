#include "chai_error.hpp"

#include <iostream>

namespace chai {
    void ChaiCompileError::display() {
        std::cout << filePath << ':' << position.startLine << ':' << position.startCol << ": " << message << '\n';
    }
}