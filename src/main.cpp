#include <iostream>
#include <stdexcept>

#include "tokenize/scanner.hpp"
#include "report/chai_error.hpp"

int main() {
    try {
        chai::Scanner sc("tests/scan_test.chai");
        
        chai::Token tok;
        while ((tok = sc.scanNext()).kind != chai::TokenKind::EndOfFile) {
            std::cout << "(" << (int)tok.kind << ", " << tok.value << ")" << '\n';
        }
    } catch (chai::ChaiCompileError& ce) {
        ce.display();
    } catch (std::exception& e) {
        std::cout << e.what() << '\n';
    }
}