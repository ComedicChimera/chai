#ifndef SOURCE_H_INCLUDED
#define SOURCE_H_INCLUDED

#include "util.h"

typedef struct {
    // parent_id is the id of the parent package
    uint32_t parent_id;

    // file_path is the module-relative path to the file
    char* file_path;

    // TODO: rest
} source_file_t;

typedef struct {
    // id is the id of the package
    uint32_t id;

    // name is the name string of the package
    char* name;

    // parent_id is the id of the parent module
    uint32_t parent_id;

    // files a buffer of the source files in this package
    source_file_t** files;

    // files_len is the number of files in this package
    uint32_t files_len;

    // TODO: rest
} package_t;

// pkg_dispose disposes of a package and its associated resources
void pkg_dispose(package_t* pkg);

/* -------------------------------------------------------------------------- */

// package_map_t is a map of packages organized by a string value such as a path
// or a name.  It is generally used in modules to allow for sub-package
// organization.  Note that the package map owns both the strings and packages.
typedef struct package_map_t package_map_t;

// pkg_map_new creates a new package map
package_map_t pkg_map_new();

// pkg_map_add adds a new package the key value map.  It will log a fatal error
// if a package is inserted into the map multiple times. 
void pkg_map_add(package_map_t* map, char* key, package_t* package);

// pkg_map_get looks up a given key in the package map and returns a package
// pointer if it finds a match and `NULL` if it doesn't.
package_t* pkg_map_get(package_map_t* map, const char* key);

// pkg_map_dispose disposes of all the resources used by the package map,
// including the keys and values
void pkg_map_dispose(package_map_t* map);

/* -------------------------------------------------------------------------- */

typedef struct {
    // id is the id of the module
    uint32_t id;

    // root_package is the package at the root of the module
    package_t* root_package;

    // sub_packages is a package map of the packages in the module organized by
    // sub-path; eg. the package at `a/b/c` where `a` is the parent module would
    // have a module sub-path of `.b.c`
    package_map_t sub_packages;

    // TODO: rest
} module_t;



#endif