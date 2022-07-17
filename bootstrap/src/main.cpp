#include <iostream>

#include "compiler.hpp"

namespace chai {
    // run is the main entry point for the compiler.
    void run(int argc, char* argv[]) {
        // Create and initialize the compiler.
        Compiler c;
        c.initFromArgs(argc, argv);

        // Initialize the root package.
        c.initPackage(c.rootDir());
    }
}

int main(int argc, char* argv[]) {
    chai::run(argc, argv);
    return 0;
}