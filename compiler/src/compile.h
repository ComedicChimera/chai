#ifndef COMPILE_H_INCLUDED
#define COMPILE_H_INCLUDED

#include "depm/source.h"

// init_pkg initializes a package.  More specifically, it gathers information
// about the package from the path, adds it to its parent module, parses the
// files in the package's directory, and resolves imports and symbol
// definitions.  After this function is called, the package is ready to be fed
// to the walker.
void init_pkg(module_t* mod, const char* pkg_path);

#endif