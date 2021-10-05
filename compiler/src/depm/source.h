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

    // parent_id is the id of the parent module
    uint32_t parent_id;

    // files a buffer of the source files in this package
    source_file_t** files;

    // files_len is the number of files in this package
    uint32_t files_len;

    // TODO: rest
} package_t;

typedef struct {
    // id is the id of the module
    uint32_t id;

    // TODO: rest
} module_t;



#endif