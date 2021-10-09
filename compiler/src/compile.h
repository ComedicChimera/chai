#ifndef COMPILE_H_INCLUDED
#define COMPILE_H_INCLUDED

#include "depm/source.h"

// compiler_t is a globally defined type that contains all the state values that
// are shared during compilation: eg. the dependency graph.  There is no real
// dispose method since it exists until program execution completes so there is
// no real need to explicitly free any of its resources.
typedef struct compiler_t {
    // root_module is the root module of the project being compiled.
    module_t* root_module;

} compiler_t;

// init_compiler initializes a new compiler in a given root module directory. It
// takes the compiler to initialize as a reference to allow the compiler to be
// stack allocated.  It returns a boolean indicating if initialization was
// successful -- ie. could it find a root module to load.
void init_compiler(compiler_t* c, const char* root_dir);

// analyze runs the analysis phase of compilation: producing fully typed and
// validated ASTs for all loaded dependencies.  It returns whether or not
// analysis succeeded.
bool analyze(compiler_t* c);

// generate runs the generation phase of compilation: producing an output binary
// (or other target as described in the build config) in the directory specified
// by the build config.
bool generate(compiler_t* c);

#endif