#include "compile.h"

#include <stdio.h>

#include "dirent.h"

// init_pkg initializes a package.  More specifically, it gathers information
// about the package from the path, adds it to its parent module, parses the
// files in the package's directory, and resolves imports and symbol
// definitions.  After this function is called, the package is ready to be fed
// to the walker.
static void init_pkg(compiler_t* c, module_t* mod, const char* pkg_path) {
    // TODO
}

/* -------------------------------------------------------------------------- */

void init_compiler(compiler_t* c, const char* root_dir) {
    // TODO
}

bool analyze(compiler_t* c) {
    // TODO
    return false;
}

bool generate(compiler_t* c) {
    // TODO
    return false;
}