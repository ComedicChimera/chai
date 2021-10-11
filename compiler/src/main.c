#include <stdio.h>
#include <string.h>

#include "compile.h"
#include "depm/source.h"

int main(int argc, char* argv[]) {
    if (argc == 1) {
        printf("missing required argument: `fpath`");
        return 1;
    }

    // create the main compiler
    compiler_t c;
    init_compiler(&c, argv[1]);

    // analyze the user's code
    if (analyze(&c)) {
        // DEBUG: check to see if the module and package are there
        printf("%s\n", c.root_module->name);
        printf("%s\n", c.root_module->root_dir);
        printf("%s\n", c.root_module->root_package->name);
        printf("%s\n", c.root_module->root_package->rel_path);

        // if analysis succeeds, we can generate output
        generate(&c); 
    }

    return 0;
}