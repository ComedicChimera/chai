#ifndef HASH_H_INCLUDED
#define HASH_H_INCLUDED

#include "util.h"

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

#endif