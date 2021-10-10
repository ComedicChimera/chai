#ifndef SOURCE_H_INCLUDED
#define SOURCE_H_INCLUDED

#include "util.h"

// source_file_t represents a single Chai source file
typedef struct {
    // parent_id is the id of the parent package
    uint64_t parent_id;

    // file_path is the module-relative path to the file
    char* file_path;

    // TODO: rest
} source_file_t;

// package_t represents a single Chai package
typedef struct {
    // id is the id of the package
    uint64_t id;

    // name is the name string of the package
    char* name;

    // parent_id is the id of the parent module
    uint64_t parent_id;

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
package_map_t* pkg_map_new();

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

// target_format_t is an enumeration of the target output formats of the compiler
typedef enum {
    TARGETF_BIN,
    TARGETF_LIB,
    TARGETF_ASM,
    TARGETF_LLVM,
    TARGETF_MIR
} target_format_t;

// build_profile_t represents a global Chai build profile -- this is generally
// created once as a singleton by the compiler, but is passed around and updated
// as modules are loaded.  Like the compiler, this doesn't really have a dispose
// function since it exists for the lifetime of the program.
typedef struct {
    // output_path is path to place the final outputted binary or library
    const char* output_path;

    // debug indicates whether the current profile is debug only
    bool debug;

    // target_os is the string name of target operating system
    const char* target_os;

    // target_arch is the string name of the target architecture
    const char* target_arch;

    // target_format is tha target output format
    target_format_t target_format;

    // static_libs is a list of libraries to be linked
    const char** static_libs;

    // link_objs is a list of additional objects to be linked 
    const char** link_objs;
} build_profile_t;

// module_t represents a single Chai module
typedef struct {
    // id is the id of the module
    uint64_t id;

    // name is the name of the module -- it is const because although it is
    // dynamically allocated, it exists for the lifetime for the program
    const char* name;

    // root_dir is the absolute root directory of the module
    const char* root_dir;

    // root_package is the package at the root of the module
    package_t* root_package;

    // sub_packages is a package map of the packages in the module organized by
    // sub-path; eg. the package at `a/b/c` where `a` is the parent module would
    // have a module sub-path of `.b.c`.  It is a pointer because the type is
    // incomplete.
    package_map_t* sub_packages;

    // TODO: rest
} module_t;

// loads a module at a given root directory
module_t* mod_load(const char* root_dir, build_profile_t* profile);

// mod_new_file creates a new file that is a child of package of this module
source_file_t* mod_new_file(module_t* mod, package_t* pkg, char* file_path);

// mod_new_pkg creates a new package that is a child of this module
package_t* mod_new_pkg(module_t* mod, const char* pkg_path);

// mod_dispose disposes of a module
void mod_dispose(module_t* mod);

#endif