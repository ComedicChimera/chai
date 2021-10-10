#include "compile.h"

#include <stdio.h>
#include <stdlib.h>

#include "dirent.h"
#include "cwalk.h"

#include "constants.h"
#include "report/report.h"

// init_pkg initializes a package.  More specifically, it gathers information
// about the package from the path, adds it to its parent module, parses the
// files in the package's directory, and resolves imports and symbol
// definitions.  After this function is called, the package is ready to be fed
// to the walker.
static void init_pkg(compiler_t* c, module_t* mod, const char* pkg_path) {
    // create new package of the given module
    package_t* pkg = mod_new_pkg(mod, pkg_path);

    // open the directory of the package
    struct dirent** files;
    int n = scandir(pkg_path, &files, NULL, alphasort);
    if (n < 0) {
        char buff[256];
        snprintf(buff, 256, "failed to open package directory: `%s`", pkg_path);
        report_fatal(buff);
    }

    // allocate a buffer to store all the source file pointers -- zero out all
    // uninitialized memory.  This will be used to actually denote when the
    // number of source files ends since there will likely be fewer source files
    // per directory than actual total directory entries.
    pkg->files = calloc(sizeof(source_file_t*), n);

    // walk through the entries the directory. using the `p` pointer to keep
    // track of where to put the next source file as we load it.
    source_file_t** p = pkg->files;
    for (int i = 0; i < n; i++) {
        struct dirent* item = files[i]; 

        // check that directory entry is a file
        if (item->d_type == DT_REG) {
            // check that the file extension matches
            const char* extension;
            size_t ext_length;
            if (!cwk_path_get_extension(item->d_name, &extension, &ext_length))
                continue;

            if (!strcmp(extension, CHAI_FILE_EXT)) {
                // create a new file
                source_file_t* curr_file = mod_new_file(mod, pkg, item->d_name);

                // TODO: parse the file and determine if it should be compiled
            }
        }

        // free the direntry after we are done processing it
        // free(item);
    }

    // free the directory entries
    free(files);

    // validate that the package is not empty
    if (pkg->files_len == 0) {
        char buff[256];
        snprintf(buff, 256, "package `%s` contains no source files", pkg->name);
        report_fatal(buff);
    }
}

/* -------------------------------------------------------------------------- */

void init_compiler(compiler_t* c, const char* root_dir) {
    // load the root module
    c->root_module = mod_load(root_dir, &c->profile);
}

bool analyze(compiler_t* c) {
    // begin by initializing the root package -- this should also initialize all
    // the sub-packages as well
    init_pkg(c, c->root_module, c->root_module->root_dir);

    // TODO: semantic analysis

    return false;
}

bool generate(compiler_t* c) {
    // TODO
    return false;
}