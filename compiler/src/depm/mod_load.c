#include "source.h"

#include <string.h>
#include <stdio.h>

#ifdef _WIN32
    #include <direct.h>
    #define getcwd _getcwd
#else
    #include <unistd.h>
#endif

#include "toml.h"
#include "cwalk.h"

#include "report/report.h"
#include "hash.h"
#include "constants.h"



// report_module_error reports a fatal error processing a module.
static void report_module_error(const char* mod_path, const char* error_msg) {
    char buff[512];
    sprintf(buff, "module error in %s: %s", mod_path, error_msg);
    report_fatal(buff);
}

// report_module_warning reports a warning loading a module
static void report_module_warning(const char* mod_path, const char* warning_msg) {
    char buff[512];
    snprintf(buff, 512, "module error in %s: %s", mod_path, warning_msg);
    report_warning(buff);
}

// get_str_value gets a required string value from a TOML table. It throws a
// fatal error if the it is unable to retrieve the value.
static char* get_str_value(const char* mod_path, toml_table_t* table, const char* key) {
    toml_datum_t value = toml_string_in(table, key);

    if (value.ok)
        return value.u.s;

    char buff[128];
    snprintf(buff, 128, "missing required field: `%s`", key);
    report_module_error(mod_path, buff);

    // unreachable
    return NULL;
}


/* -------------------------------------------------------------------------- */

module_t* mod_load(const char* root_dir, build_profile_t* profile) {
    // calculate the module file path
    char* mod_path = malloc(strlen(root_dir) + strlen(CHAI_MOD_FILE_PATH) - 1);
    strcpy(mod_path, root_dir);
    strcpy(mod_path + strlen(root_dir), CHAI_MOD_FILE_PATH);

    // try to open the module file
    FILE* fp = fopen(mod_path, "r");
    if (!fp) {
        char buff[256];
        snprintf(buff, 256, "unable to open module file: %s", mod_path);
        report_fatal(buff);
    }

    // parse the toml module file
    char toml_error_buff[200];
    toml_table_t* toml_mod = toml_parse_file(fp, toml_error_buff, sizeof(toml_error_buff));
    fclose(fp);

    if (!toml_mod)
        report_module_error(mod_path, toml_error_buff);

    // initialize the module and read the required top level data
    module_t* mod = malloc(sizeof(module_t));

    mod->name = get_str_value(mod_path, toml_mod, "name");
    mod->id = string_hash(mod->name);
    mod->sub_packages = pkg_map_new();

    // get the absolute path to the module's root directory
    char cwd[256];
    if (getcwd(cwd, 256)) {
        char abs_path[512];
        size_t num_bytes = cwk_path_get_absolute(cwd, root_dir, abs_path, 512);
        mod->root_dir = malloc(num_bytes + 1);
        strcpy(mod->root_dir, abs_path);
    } else
        report_fatal("failed to get working directory");
        

    // check that the versions match up and emit errors & warnings as necessary
    switch (strcmp(get_str_value(mod_path, toml_mod, "chai-version"), CHAI_VERSION)) {
    case 0:
        // versions are equal -- nothing to do
        break;
    case -1:
        // module version is less than current Chai version -- therefore, we emit
        // a warning only since generally versions are backwards compatible
        report_module_warning(mod_path, "Chai version specific in module is behind the current Chai version");
        break;
    case 1:
        // module version is more than current Chai version => error
        report_module_error(mod_path, "Chai version specified in module is ahead of the current Chai version");
        break;
    }

    // TODO: load all optional top-level data

    // TODO: build profiles

    // return the loaded module
    return mod;
}