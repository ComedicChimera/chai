#include "source.h"

#include <string.h>
#include <stdio.h>

#include "report/report.h"

#define INITIAL_TABLE_SIZE 16
#define TABLE_GROWTH_FACTOR 2

// kv_pair_t is a simple struct used to group key-value pairs in the package map
// table
typedef struct {
    char* key;
    package_t* value;
} kv_pair_t;

typedef struct package_map_t {
    // hash_table is table containing pointers to key-value pairs. Any empty
    // space in the hash table will be marked by a `NULL` pointer.
    kv_pair_t** hash_table;

    // num_pairs store the number of pairs currently in the hash table
    uint32_t num_pairs;

    // hash_table_cap stores the current capacity of the table
    uint32_t hash_table_cap;
} package_map_t;

// string_hash calculates the un-moduloed hash value for a string. This uses the
// simple algorithm presented in K&R version 2 -- it is effective enough for our
// purposes.
static uint64_t string_hash(const char* string) {
    uint64_t hash_val = 0;

    for (; *string != '\0'; string++) {
        hash_val = *string + 31 * hash_val;
    }

    return hash_val;
}

// pkg_map_lookup attempts to locate a key-value pair within the hash table
// based on a given key.  It returns `NULL` if it doesn't find a match
static kv_pair_t* pkg_map_lookup(package_map_t* map, const char* key) {
    // get the key hash to perform the lookup
    uint64_t hash = string_hash(key);
    uint64_t index = hash % map->hash_table_cap;

    // first check the initial hash position with the table
    kv_pair_t* pair = map->hash_table[index];

    // if this pair is `NULL`, we know there is no matching entry
    if (pair == NULL)
        return NULL;

    // otherwise, we need to check to make sure the keys match
    if (!strcmp(pair->key, key))
        return pair;

    // key's don't match => it is linear probing time
    int j = index+1;
    for (; j != index; j++) {
        // wrap around
        if (j == map->hash_table_cap)
            j = 0;

        pair = map->hash_table[j];

        // we hit a NULL, no match
        if (pair == NULL)
            return NULL;

        // check to see if our keys match; if they do, we found a match
        if (!strcmp(pair->key, key))
            return pair;
    }

    // we searched entire table (ouch) => no match
    return NULL;
}

// pkg_map_insert adds a new key-value pair into the hash table.  It returns a
// boolean indicating whether or not the insertion was successful -- ie. whether
// or not a key value pair was already in place.
static bool pkg_map_insert(package_map_t* map, kv_pair_t* pair) {
    // TODO
}

/* -------------------------------------------------------------------------- */

package_map_t pkg_map_new() {
    return (package_map_t) {
        // calloc will zero out our hash table for us :)
        .hash_table = (kv_pair_t**)calloc(sizeof(kv_pair_t*), INITIAL_TABLE_SIZE),
        .num_pairs = 0,
        .hash_table_cap = INITIAL_TABLE_SIZE
    };
}

void pkg_map_add(package_map_t* map, char* key, package_t* pkg) {
    // create the new pair
    kv_pair_t* pair = (kv_pair_t*)malloc(sizeof(kv_pair_t));
    pair->key = key;
    pair->value = pkg;

    // attempt to insert it into the map
    if (!pkg_map_insert(map, pair)) {
        // fatal errors exit the program immediately so no need to bother
        // freeing anything
        char buff[256];
        sprintf(buff, "package with name `%s` inserted into package map multiple times", pkg->name);
        report_fatal(buff);
    }
}

package_t* pkg_map_get(package_map_t* map, const char* key) {
    kv_pair_t* pair = pkg_map_lookup(map, key);
    if (pair == NULL)
        return NULL;

    return pair->value;
}

void pkg_map_dispose(package_map_t* map) {
    for (int i = 0; i < map->hash_table_cap; i++) {
        kv_pair_t* pair = map->hash_table[i];
        free(pair->key);
        pkg_dispose(pair->value);

        free(pair);
    }

    free(map->hash_table);
}