#include <stdio.h>
#include <string.h>

#include "depm/source.h"

int main(int argc, char* argv[]) {
    if (argc == 1) {
        printf("missing required argument: `fpath`");
        return 1;
    }

    module_t* mod = mod_load(argv[1], NULL);
    return 0;
}