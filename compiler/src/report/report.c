#include "report.h"

#include <stdio.h>

void report_fatal(const char* message) {
    printf("fatal error: %s\n", message);
    exit(1);
}

void report_warning(const char* message) {
    printf("warning: %s\n", message);
}

void report_compile_error(const char* fpath, text_pos_t position, const char* message) {
    printf("%s:%d:%d: %s\n", fpath, position.start_line, position.start_col, message);
}