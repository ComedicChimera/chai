#ifndef REPORT_H_INCLUDED
#define REPORT_H_INCLUDED

#include "util.h"

// text_pos_t represents a position in source code
typedef struct {
    uint32_t start_line;
    uint32_t start_col;
    uint32_t end_line;
    uint32_t end_col;
} text_pos_t;

// report_fatal reports a fatal error that doesn't correspond with some source
// code position -- these can be errors opening files, reading directories, etc.
void report_fatal(const char* message);

// report_compile_error reports an error compiling a specific source file
void report_compile_error(const char* fpath, text_pos_t position, const char* message);

#endif