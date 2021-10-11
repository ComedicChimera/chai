#include "source.h"

#include <stdlib.h>
#include <string.h>

#include "cwalk.h"

void src_file_dispose(source_file_t* file) {
    free(file->file_rel_path);

    free(file);
}

void pkg_dispose(package_t* pkg) {
    // free the package name
    free(pkg->name);

    // free the package relative path
    free(pkg->rel_path);

    // free all the files
    for (int i = 0; i < pkg->files_len; i++) {
        src_file_dispose(pkg->files[i]);
    }

    // free the files block
    free(pkg->files);

    // free the package itself
    free(pkg);
}

/* -------------------------------------------------------------------------- */

source_file_t* mod_new_file(module_t* mod, package_t* pkg, const char* file_abs_path) {
    // allocate a new source file
    source_file_t* file = malloc(sizeof(source_file_t));
    file->parent_id = pkg->id;

    // calculate the path of the file relative to its parent module
    char file_rel_path[128];
    size_t file_rel_path_len = cwk_path_get_relative(mod->root_dir, file_abs_path, file_rel_path, 128);
    file->file_rel_path = malloc(file_rel_path_len + 1);
    strcpy(file->file_rel_path, file_rel_path);

    return file;
}

package_t* mod_new_pkg(module_t* mod, const char* pkg_abs_path) {
    // use this counter to get sequential package IDs
    static uint32_t pkg_id_counter = 0;

    // create the package itself
    package_t* pkg = malloc(sizeof(package_t));

    // determine the package name
    char *pkg_name;
    size_t pkg_name_len;
    cwk_path_get_basename(pkg_abs_path, &pkg_name, &pkg_name_len);
    pkg->name = malloc(pkg_name_len + 1);
    strcpy(pkg->name, pkg_name);

    // determine the package path relative path
    char pkg_rel_path[256];
    size_t pkg_rel_path_len = cwk_path_get_relative(mod->root_dir, pkg_abs_path, pkg_rel_path, 256);
    pkg->rel_path = malloc(pkg_rel_path_len + 1);
    strcpy(pkg->rel_path, pkg_rel_path);    

    // store the various IDs of the package
    pkg->id = pkg_id_counter++;
    pkg->parent_id = mod->id;

    // if the package is at the root directory of the module,
    // then it is the root package of the module
    if (!strcmp(mod->root_dir, pkg_abs_path)) {
        mod->root_package = pkg;
    } else {
        pkg_map_add(mod->sub_packages, pkg->rel_path, pkg);
    }

    return pkg;
}

void mod_dispose(module_t* mod) {
    free(mod->name);
    free(mod->root_dir);

    pkg_dispose(mod->root_package);
    pkg_map_dispose(mod->sub_packages);
}