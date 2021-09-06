#include <iostream>
#include <stdexcept>

#include "tokenize/scanner.hpp"

int main() {
    try {
        chai::Scanner sc("tests/scan_test.chai");
        
        chai::Token tok;
        while ((tok = sc.scanNext()).kind != chai::TokenKind::EndOfFile) {
            std::cout << tok.value << '\n';
        }
    } catch (std::exception& e) {
        std::cout << e.what() << '\n';
    }
}