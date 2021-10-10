#include "source.h"

#include <stdlib.h>
#include <string.h>

void pkg_dispose(package_t* pkg) {
    // free the package name
    free(pkg->name);

    // free all the files
    for (int i = 0; i < pkg->files_len; i++) {
        free(pkg->files[i]->file_path);
        free(pkg->files[i]);
    }

    // free the files block
    free(pkg->files);

    // free the package itself
    free(pkg);
}

/* -------------------------------------------------------------------------- */

source_file_t* mod_new_file(module_t* mod, package_t* pkg, const char* file_path) {
    

    // TODO: calculate the file path relative to the parent module

    // TODO: replace with file reference
    return NULL;
}

package_t* mod_new_pkg(module_t* mod, const char* pkg_path) {
    // use this counter to get sequential package IDs
    static uint32_t pkg_id_counter = 0;

    // determine the package name
    const char* last_backslash = strrchr(pkg_path, '/');
    size_t name_len = strlen(last_backslash)-1;
    char* pkg_name = (char*)malloc(name_len);
    memcpy(pkg_name, last_backslash+1, name_len);

    // create the package itself
    package_t* pkg = malloc(sizeof(package_t));
    pkg->name = pkg_name;
    pkg->id = pkg_id_counter++;
    pkg->parent_id = mod->id;

    // if the package is at the root directory of the module,
    // then it is the root package of the module
    if (!strcmp(mod->root_dir, pkg_path)) {
        mod->root_package = pkg;
    } else {
        // determine the sub-path of the directory


        // pkg_map_add(mod->sub_packages, pkg);
    }

    return pkg;
}

void mod_dispose(module_t* mod) {
    pkg_dispose(mod->root_package);
    pkg_map_dispose(mod->sub_packages);
}