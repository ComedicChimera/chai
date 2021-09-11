#include <iostream>
#include <stdexcept>

#include "tokenize/scanner.hpp"
#include "report/reporter.hpp"

int main() {
    chai::Reporter reporter;

    try {
        chai::Scanner sc("tests/scan_test.chai");
        
        chai::Token tok;
        while ((tok = sc.scanNext()).kind != chai::TokenKind::EndOfFile) {
            std::cout << "(" << (int)tok.kind << ", " << tok.value << ")" << '\n';
        }
    } catch (chai::CompileMessage& cm) {
        reporter.reportCompileError(cm);
    } catch (std::exception& e) {
        std::cout << e.what() << '\n';
    }
}