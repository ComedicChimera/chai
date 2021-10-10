#include "source.h"

#include <string.h>
#include <stdio.h>

#include "report/report.h"
#include "hash.h"

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

static bool pkg_map_insert(package_map_t* map, kv_pair_t* pair);

// pkg_map_grow grows the hash table of the package map
static void pkg_map_grow(package_map_t* map) {
    // first, we want to save the old hash table temporarily so we can insert
    // its pairs into the new table
    kv_pair_t** old_table = map->hash_table;
    uint32_t old_cap = map->hash_table_cap;
    
    // then, we want to allocate a new block of memory for the new table. This
    // is because we need to rebalance the hash table and so realloc simply
    // isn't sufficient.
    map->hash_table_cap *= TABLE_GROWTH_FACTOR;
    map->hash_table = (kv_pair_t**)calloc(sizeof(kv_pair_t*), map->hash_table_cap);

    // just insert all non-null members of the old table into the new table
    for (uint32_t i = 0; i < old_cap; i++) {
        if (old_table[i] != NULL)
            pkg_map_insert(map, old_table[i]);
    }

    // get rid of the old table now that we no longer need it
    free(old_table);
}

// pkg_map_insert adds a new key-value pair into the hash table.  It returns a
// boolean indicating whether or not the insertion was successful -- ie. whether
// or not a key value pair was already in place.
static bool pkg_map_insert(package_map_t* map, kv_pair_t* pair) {
    // I chose to grow the hash table whenever the number of pairs is equal to
    // 1/2 of the hash table capacity.  This number is chosen because as the
    // table fills, the lookup and insertion time increases dramatically so we
    // want to grow the table sooner rather than later to preserve performance.
    // This does waste memory, but also leads to much more performant lookups --
    // plus, it is very uncommon that a module will contain more than 7
    // packages, so this should happen fairly rarely.
    if (map->num_pairs++ >= map->hash_table_cap / 2)
        pkg_map_grow(map);

    // get the hash and expected index of the pair
    uint64_t hash = string_hash(pair->key);
    uint64_t index = hash % map->hash_table_cap;

    // check to see if the key-value pair in at the given position in the hash
    // table; if there is no key there, then we insert the pair into the table
    pair = map->hash_table[index];
    if (pair == NULL) {
        map->hash_table[index] = pair;
        return true;
    }

    // perform a linear probe to find the insertion position
    int j = index + 1;
    for (; j != index; j++) {
        if (j == map->hash_table_cap)
            j = 0;

        if (map->hash_table[j] == NULL) {
            map->hash_table[j] = pair;
            return true;
        }
    }
    
    // unreachable -- there should always be a location to insert since the
    // table is grown before the function is called.
    return false;
}

/* -------------------------------------------------------------------------- */

package_map_t* pkg_map_new() {
    package_map_t* map = malloc(sizeof(package_map_t));
     // calloc will zero out our hash table for us :)
    map->hash_table = calloc(sizeof(kv_pair_t*), INITIAL_TABLE_SIZE);
    map->num_pairs = 0;
    map->hash_table_cap = INITIAL_TABLE_SIZE;

    return map;
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