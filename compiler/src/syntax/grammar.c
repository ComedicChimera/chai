#include "parser.h"

// This file implements the actual syntactic structure of the Chai programming
// language: it is where all of the recursive-descent methods of the parser
// live.  The comment above each function is its EBNF production.

// NOTE: This file assumes that the reader is already familiar with the workings
// and design of Chai's symbol table and symbol resolution algorithm.  If you
// are not, please read the comment in `depm/symbol_table.h` for more
// information. In general, this file does not explain the actual ordering of
// the parsing actions or any of the semantic actions performed both for sake of
// brevity and because they should be obvious given the EBNF grammar above each
// function.

/* -------------------------------------------------------------------------- */

// TODO: parsing routines

/* -------------------------------------------------------------------------- */

// file = {import_statement} {pub_block | top_level}
// This function is main entry point for the parsing algorithm (ie. the start
// symbol of Chai's grammar).  It is the only parsing function that is
// externally visible.
bool parse_file(parser_t* p, build_profile_t* profile) {
    // TODO
    return true;
}


