#include "lexer.h"

#include <stdio.h>
#include <ctype.h>
#include <string.h>

#define TOK_BUFF_BLOCK_SIZE 16

typedef struct lexer_t {
    // fpath is the path to the Lexer's current file
    const char* fpath;

    // file the current file being read by the lexer
    FILE* file;

    // tok_buff stores the token as it is being built
    char* tok_buff;

    // tok_buff_len is the length of token buffer
    uint32_t tok_buff_len;

    // tok_buff_cap is the capacity of the token buffer
    uint32_t tok_buff_cap;

    // line indicates the current line, starting from 1
    uint32_t line;

    // col indicates the current column, starting from 1
    uint32_t col;

    // start_line indicates which line token construction began on
    uint32_t start_line;

    // start_col indicates which column token construction began on
    uint32_t start_col;
} lexer_t;

// lexer_get_pos generates a text_pos_t based on the Lexer's current position.
static text_pos_t lexer_get_pos(lexer_t* lexer) {
    return (text_pos_t){
        .start_line = lexer->start_line,
        .start_col = lexer->start_col,
        .end_line = lexer->line,
        .end_col = lexer->col
        };
}

// lexer_fail reports an error lexing the user source file
static void lexer_fail(lexer_t* lexer, const char* message) {
    report_compile_error(lexer->fpath, lexer_get_pos(lexer), message);

    // clear the token buffer as its contents are no longer useful
    if (lexer->tok_buff != NULL) {
        free(lexer->tok_buff);

        // make sure dispose doesn't delete the buffer twice
        lexer->tok_buff = NULL;
    }      
}

// lexer_mark_start marks the beginning of a token as it is being read in
static void lexer_mark_start(lexer_t* lexer) {
    lexer->start_line = lexer->line;
    lexer->start_col = lexer->col;
}

// lexer_peek peeks ahead one character in the token stream
static char lexer_peek(lexer_t* lexer) {
    if (feof(lexer->file))
        return EOF;

    char ahead = fgetc(lexer->file);
    ungetc(ahead, lexer->file);
    return ahead;
}

// lexer_update_pos updates the lexer's position based on a character
static void lexer_update_pos(lexer_t* lexer, char c) {
    switch (c) {
        case '\n':
            lexer->line++;
            lexer->col = 1;
            break;
        case '\t':
            lexer->col += 4;
            break;
        default:
            lexer->col++;
            break;
    }
}

// lexer_skip skips a character in the input stream and returns the skipped
// character
static char lexer_skip(lexer_t* lexer) {
    // handle EOF case
    if (feof(lexer->file))
        return EOF;

    char c = fgetc(lexer->file);
    if (c == EOF)
        // something went wrong reading the file
        return EOF;

    // still want to update the position
    lexer_update_pos(lexer, c);
    return c;
}

// lexer_write_char writes a character to the token buffer
static void lexer_write_char(lexer_t* lexer, char c) {
    // check if the length is equal to the capacity => if so, we need to grow
    // the capacity to fit more chars
    if (lexer->tok_buff_len == lexer->tok_buff_cap) {
        lexer->tok_buff_cap += TOK_BUFF_BLOCK_SIZE;
        lexer->tok_buff = (char*)realloc(lexer->tok_buff, lexer->tok_buff_cap);
    }

    // insert at the current length, then increment the length
    lexer->tok_buff[lexer->tok_buff_len++] = c;
}

// lexer_read reads a character from the input stream into the token buffer and
// returns the read-in character
static char lexer_read(lexer_t* lexer) {
    // handle EOF case
    if (feof(lexer->file))
        return EOF;

    char c = fgetc(lexer->file);
    if (c == EOF)
        // something went wrong reading the file
        return EOF;

    // update lexer state
    lexer_update_pos(lexer, c);
    lexer_write_char(lexer, c);

    return c;
}

// lexer_make_token produces a new token from the lexer's current state
token_t lexer_make_token(lexer_t* lexer, token_kind_t kind) {
    // add null terminator to end the token value
    lexer_write_char(lexer, '\0');

    // build the token
    token_t tok = (token_t){
        .kind = kind, 
        .value = lexer->tok_buff, 
        .position = lexer_get_pos(lexer)
    };

    // mark the token buffer as "empty" -- make sure it isn't freed when the
    // lexer is disposed (don't want the last token to be destroyed)
    lexer->tok_buff = NULL;

    return tok;
}

/* -------------------------------------------------------------------------- */

// lexer comment skips a line or block comment and returns whether a newline
// should be emitted (ie. did it encountered a line comment)
static bool lexer_comment(lexer_t* lexer) {
    // skip the leading `#`
    lexer_skip(lexer);

    // check to see if a `!` is next => multiline or not
    char ahead = lexer_skip(lexer);
    if (ahead == '!') {
        // multiline comment

        // read until a `!#` is encountered (or the file ends)
        while ((lexer_skip(lexer) != '!' || lexer_skip(lexer) != '#') && lexer_peek(lexer) != EOF);

        // no newline necessary
        return false;
    } else if (ahead == '\n') {
        // newline immediately after `#`
        return true;
    } else {
        // singleline comment

        // read until newline or end of file
        while (lexer_peek(lexer) != '\n' && lexer_peek(lexer) != EOF)
            lexer_skip(lexer);

        // should emit a newline
        return true;
    }
}

// lexer_match_keyword attempts to match an identifier to a reserved keyword. It
// returns TOK_IDENTIFIER if no match occurs.
static token_kind_t lexer_match_keyword(lexer_t* lexer) {
    // we need to null-terminate the token buffer before we can compare it, but
    // we also know that Chai's longest keyword is less than 15 bytes so we only
    // need at most 15 bytes of the token buffer to tell if it matches a keyword
    // or not (one extra byte to see if there is meaningful character data after
    // the match to preclude matching identifiers that start with keywords).  So
    // we just copy <=15 bytes from the token buffer to a 16 byte buffer and add
    // in a null terminator at the end of the meaningful data.
    char keywordStr[16];
    memcpy(keywordStr, lexer->tok_buff, min(lexer->tok_buff_len, 15));
    keywordStr[min(lexer->tok_buff_len, 15)] = '\0';

    // next, we just compare to all the keywords we know
    if (!strcmp(keywordStr, "def")) return TOK_DEF;
    else if (!strcmp(keywordStr, "end")) return TOK_END;

    else if (!strcmp(keywordStr, "type")) return TOK_TYPE;

    else if (!strcmp(keywordStr, "if")) return TOK_IF;
    else if (!strcmp(keywordStr, "elif")) return TOK_ELIF;
    else if (!strcmp(keywordStr, "else")) return TOK_ELSE;
    else if (!strcmp(keywordStr, "match")) return TOK_MATCH;
    else if (!strcmp(keywordStr, "case")) return TOK_CASE;
    else if (!strcmp(keywordStr, "while")) return TOK_WHILE;

    else if (!strcmp(keywordStr, "import")) return TOK_IMPORT;
    else if (!strcmp(keywordStr, "from")) return TOK_FROM;
    else if (!strcmp(keywordStr, "as")) return TOK_AS;
    else if (!strcmp(keywordStr, "pub")) return TOK_PUB;

    else if (!strcmp(keywordStr, "true") || !strcmp(keywordStr, "false")) return TOK_BOOLLIT;

    else return TOK_IDENTIFIER;
}

// lexer_oper_or_punct matches an operator or a punctuation character. It
// returns the token_kind of the operator it matched so the caller can build the
// full token.  If no match occurs, it returns `TOK_EOF`.
static token_kind_t lexer_oper_or_punct(lexer_t* lexer) {
    // mark the beginning of the token
    lexer_mark_start(lexer);

    // read the leading/primary operator component
    char next = lexer_read(lexer);

    // kind is the kind the initial kind that matched for secondary matches
    token_kind_t kind = TOK_EOF;

    // check if it matches any of the possible token starts; some branches
    // return if they match since that token is of length one
    switch (next) {
        case '+':
            kind = TOK_PLUS;
            break;
        case '-':
            kind = TOK_MINUS;
            break;
        case '*':
            kind = TOK_STAR;
            break;
        case '/':
            kind = TOK_DIV;
            break;
        case '%': return TOK_MODULO;
        case '<':
            kind = TOK_LT;
            break;
        case '>':
            kind = TOK_GT;
            break;
        case '=':
            kind = TOK_ASSIGN;
            break;
        case '&':
            kind = TOK_AMP;
            break;
        case '|':
            kind = TOK_PIPE;
            break;
        case '^': return TOK_CARRET;
        case '!': 
            kind = TOK_NOT;
            break;
        case '.':
            kind = TOK_DOT;
            break;
        case '(': return TOK_LPAREN;
        case ')': return TOK_RPAREN;
        case '[': return TOK_LBRACKET;
        case ']': return TOK_RBRACKET;
        case '{': return TOK_LBRACE;
        case '}': return TOK_RBRACE;
        case '@': return TOK_ANNOTAT;
        case ',': return TOK_COMMA;
        case ':': return TOK_COLON;
        case ';': return TOK_SEMICOLON;
        default: return TOK_EOF;
    }

    // match tokens that might be multi-character
    char ahead = lexer_peek(lexer);
    switch (kind) {
        case TOK_PLUS:
            if (ahead == '+') {
                lexer_read(lexer);
                return TOK_INCREM;
            } else
                return kind;
        case TOK_MINUS:
            if (ahead == '-') {
                lexer_read(lexer);
                return TOK_DECREM;
            } else if (ahead == '>') {
                lexer_read(lexer);
                return TOK_ARROW;  
            } else
                return kind;
        case TOK_STAR:
            if (ahead == '*') {
                lexer_read(lexer);
                return TOK_POWER;
            } else
                return kind;
        case TOK_DIV:
            if (ahead == '/') {
                lexer_read(lexer);
                return TOK_FLOORDIV;
            } else
                return kind;
        case TOK_AMP:
            if (ahead == '&') {
                lexer_read(lexer);
                return TOK_AND;
            } else
                return kind;
        case TOK_PIPE:
            if (ahead == '|') {
                lexer_read(lexer);
                return TOK_OR;
            } else
                return kind;
        case TOK_GT:
            if (ahead == '>') {
                lexer_read(lexer);
                return TOK_RSHIFT;
            } else if (ahead == '=') {
                lexer_read(lexer);
                return TOK_GTEQ;  
            } else
                return kind;
        case TOK_LT:
            if (ahead == '<') {
                lexer_read(lexer);
                return TOK_LSHIFT;
            } else if (ahead == '=') {
                lexer_read(lexer);
                return TOK_LTEQ;  
            } else
                return kind;
        case TOK_ASSIGN:
            if (ahead == '=') {
                lexer_read(lexer);
                return TOK_EQ;
            } else
                return kind;
        case TOK_NOT:
            if (ahead == '=') {
                lexer_read(lexer);
                return TOK_NEQ;
            } else
                return kind;
        case TOK_DOT:
            if (ahead == '.') {
                lexer_read(lexer);
                ahead = lexer_peek(lexer);

                if (ahead == '.') {
                    lexer_read(lexer);
                    return TOK_ELLIPSIS;
                } else
                    return TOK_RANGETO;
            } else
                return kind;
        default:
            // just a singular character operator :)
            return kind;
    }
}

// lexer_unicode_escape reads in the escape sequence that follows a unicode
// escape prefix (`\x`, `\u`, `\U`)
static bool lexer_unicode_escape(lexer_t* lexer, int count) {
    for (int i = 0; i < count; i++) {
        char next = lexer_read(lexer);

        if (next == EOF) {
            lexer_fail(lexer, "expected rest of escape code not end of file");
            return false;
        } else if ('0' <= next && next <= '9' || 'a' <= next && next <= 'f' || 'A' <= next && next <= 'F')
            continue;
        else {
            char buff[64];
            sprintf(buff, "unexpected character: `%c`", next);
            lexer_fail(lexer, buff);
            return false;
        }
    }

    return true;
}

// lexer_escape_code reads in an escape code encountered in a string or rune
// assuming the leading `\` hasn't been read in yet
static bool lexer_escape_code(lexer_t* lexer) {
    // note: we actually write the escape code lexical value to the token stream
    // because some escape codes may cause problems with the compiler's string
    // handling (eg. if it contains a null terminator, we don't want to write
    // that so the compiler processes the whole string)

    // read the leading backslash
    lexer_read(lexer);

    // read in the escape code (it should br written to the stream)
    char ahead = lexer_read(lexer);
    if (ahead == EOF) {
        lexer_fail(lexer, "expected escape code not end of file");
        return false;
    }

    // match it against the valid escape codes; if the escape code is not a
    // unicode int value (eg. not `\xb2`), then we just write the actual
    // character value to the token buffer
    switch (ahead) {
        case 'a':
        case 'b':
        case 'f':
        case 'n':
        case 'r':
        case 't':
        case 'v':
        case '0':
        case '\"':
        case '\'':
        case '\\':
            // standard acceptable escape codes
            break;
        // unicode escape codes
        case 'x':
            return lexer_unicode_escape(lexer, 2);
        case 'u':
            return lexer_unicode_escape(lexer, 4);
        case 'U':
            return lexer_unicode_escape(lexer, 8);
        default:
            // invalid escape code
            char buff[32];
            sprintf(buff, "unknown escape code: `\\%c`", ahead);
            lexer_fail(lexer, buff);
            return false;

    }

    // if we reach here, all good
    return true;
}

// lexer_std_string_lit reads a standard string literal assuming the leading
// double quote hasn't already been read in.  It returns a boolean indicating if
// an error occurred while reading the string (eg. invalid escape code)
static bool lexer_std_string_lit(lexer_t* lexer, token_t* tok) {
    // mark the beginning of the string
    lexer_mark_start(lexer);

    // skip the leading double quote (we don't want it in the literal value)
    lexer_skip(lexer);

    // read until we encounter a closing quote
    for (char ahead = lexer_peek(lexer); ahead != '\"'; ahead = lexer_peek(lexer)) {
        // handle escape codes
        if (ahead == '\\') {
            if (!lexer_escape_code(lexer))
                return false;
        }     
        // handle newlines in strings
        else if (ahead == '\n') {
            lexer_fail(lexer, "standard string literals cannot contain newlines");
            return false;
        }
        // handle EOF in the middle of a string
        else if (ahead == EOF) {
            lexer_fail(lexer, "expected a closing quote before end of file");
            return false;
        }
        else
            lexer_read(lexer);
    }

    // skip the closing quote
    lexer_skip(lexer);

    // make the the string token
    *tok = lexer_make_token(lexer, TOK_STRINGLIT);
    return true;
}

// lexer_rune_lit scans in a rune literal assuming the leading `'` hasn't been
// read in.  It returns a boolean indicating if an error occurred while reading
// the rune (eg. invalid escape code)
static bool lexer_rune_lit(lexer_t* lexer, token_t* tok) {
    // mark the beginning of the rune
    lexer_mark_start(lexer);

    // skip the leading `'` so it doesn't end up in the literal value
    lexer_skip(lexer);

    // peek the main character of the rune literal
    char body = lexer_peek(lexer);

    switch (body) {
        case EOF:
            // if it is an EOF, error
            lexer_fail(lexer, "expected a closing quote before end of file");
            return false;
        case '\n':
            // newlines aren't allowed in runes
            lexer_fail(lexer, "rune literals cannot contain newlines");
            return false;
        case '\\':
            // handle escapes codes
            if (!lexer_escape_code(lexer))
                return false;

            break;
        case '\'':
            // empty rune literal
            
            // read in the `'` to build an appropriate error message
            lexer_skip(lexer);

            lexer_fail(lexer, "empty rune literal");
            return false;
        default:
            // just a regular character
            lexer_read(lexer);
            break;
    }

    // skip the closing `'`
    char closer = lexer_skip(lexer);
    if (closer != '\'') {
        // handle invalid closer
        char buff[64];
        sprintf(buff, "expected a closing quote not `%c`", closer);
        lexer_fail(lexer, buff);
        return false;
    }

    // make the token and return
    *tok = lexer_make_token(lexer, TOK_RUNELIT);
    return true;
}

// lexer_raw_string_lit scans in a raw string literal assuming the leading
// backtick hasn't been read in
static bool lexer_raw_string_lit(lexer_t* lexer, token_t* tok) {
    // mark the start of the string
    lexer_mark_start(lexer);

    // skip the leading back tick
    lexer_skip(lexer);

    // read in the string literal
    for (char ahead = lexer_peek(lexer); ahead != '`'; ahead = lexer_peek(lexer)) {
        if (ahead == EOF) {
            lexer_fail(lexer, "expected closing backtick before end of file");
            return false;
        }

        // read the string content
        lexer_read(lexer);
    }
    
    // skip the closing back tick
    lexer_skip(lexer);

    // make the token and return
    *tok = lexer_make_token(lexer, TOK_STRINGLIT);
    return true;
}

// lexer_num_lit scans in a numeric literal assuming the first char hasn't been
// read in yet.
static bool lexer_num_lit(lexer_t* lexer, token_t* tok) {
    // mark the start of the numeric literal
    lexer_mark_start(lexer);

    // read in the leading char to check for base prefixes
    char leading = lexer_read(lexer);

    // if it is zero, then we may have a base prefix
    int base = 10;
    if (leading == '0') {
        char ahead = lexer_peek(lexer);
        switch (ahead) {
            case 'b':
                base = 2;
                lexer_read(lexer);
                break;
            case 'o':
                base = 8;
                lexer_read(lexer);
                break;
            case 'x':
                base = 16;
                lexer_read(lexer);
                break;
        }
    }

    // data to determine and validate literal
    bool is_floating = false, has_exponent = false, expecting_minus = false;
    bool is_long = false, is_unsigned = false;

    // greedily consume tokens until we can't anymore -- the loop will break
    // itself as necessary (breaking logic is too complex for loop header);
    // note: at all times in this loop if we encounter an "unexpected"
    // character, we break out not error: numbers can be followed by all manner
    // symbols that may appear related to them -- we have no cause to fail
    for (char ahead = lexer_peek(lexer); ahead != EOF; ahead = lexer_peek(lexer)) {
        // handle unsigned and long suffixes
        if (is_unsigned) {
            // only thing remaining can be long suffix
            if (ahead == 'l') {
                lexer_read(lexer);
                is_long = true;
            }
            
            break;
        } else if (is_long) {
            // only thing remaining can be unsigned suffix
            if (ahead == 'u') {
                lexer_read(lexer);
                is_unsigned = true;
            }

            break;
        }

        // handle the content that differs by base
        switch (base) {
            // all bases that are not 10 must be integral bases so their logic
            // is fairly simple -- we use continue to avoid the break at the end
            // if the character's match
            case 2:
                if ('0' == ahead || '1' == ahead) {
                    lexer_read(lexer);
                    continue;
                }

                break;
            case 8:
                if ('0' <= ahead && ahead <= '7') {
                    lexer_read(lexer);
                    continue;
                }

                break;
            case 16:
                if (isdigit(ahead) || 'a' <= ahead && ahead <= 'f' || 'A' <= ahead && ahead <= 'F') {
                    lexer_read(lexer);
                    continue;
                }

                break;
            case 10:
                if (isdigit(ahead)) {
                    lexer_read(lexer);

                    // clear expecting minus if we encounter a digit
                    expecting_minus = false;
                    continue;
                }

                switch (ahead) {
                    case '.':
                        // we want to fail here if the dot is unexpected to
                        // avoid ambiguity -- numbers always consume dots so
                        // having it ignore simply because decimal exponents or
                        // duplicate dots are illegal is unwise
                        if (has_exponent) {
                            lexer_fail(lexer, "decimal exponents are not allowed");
                            return false;
                        } else if (is_floating) {
                            lexer_fail(lexer, "floating point number cannot contain multiple decimals");
                            return false;
                        }

                        lexer_read(lexer);
                        is_floating = true;
                        continue;
                    case 'e':
                    case 'E':
                        // by same logic as decimals, we also want to fail here
                        // on duplicate exponents
                        if (has_exponent) {
                            lexer_fail(lexer, "floating point number cannot contain multiple exponents");
                            return false;
                        }

                        lexer_read(lexer);
                        is_floating = true;
                        has_exponent = true;
                        expecting_minus = true;
                        continue;
                    case '-':
                        // obviously, we do NOT want to fail here since `5-3`
                        // should be a valid Chai expression :)
                        if (expecting_minus) {
                            lexer_read(lexer);
                            expecting_minus = false;
                            continue;
                        }

                        break;
                }

                break;
        }

        // check for first suffix
        if (!is_floating) {
            if (ahead == 'u') {
                lexer_read(lexer);
                is_unsigned = true;
                continue;
            } else if (ahead == 'l') {
                lexer_read(lexer);
                is_unsigned = true;
                continue;
            }
        }

        // if we reach here, we always want to exit the loop -- found a
        // non-matching character
        break;
    }

    // check to make sure at least one digit is provided for base-prefixed
    // literals (more than just `0p` in the token buffer)
    if (base != 10 && lexer->tok_buff_len < 3) {
        lexer_fail(lexer, "at least one digit expected after base prefix");
        return false;
    }

    // make the literal based on collected info
    if (base != 10 || is_long || is_unsigned)
        *tok = lexer_make_token(lexer, TOK_INTLIT);
    else if (is_floating)
        *tok = lexer_make_token(lexer, TOK_FLOATLIT);
    else
        *tok = lexer_make_token(lexer, TOK_NUMLIT);

    // all cases succeed :)
    return true;
}


/* -------------------------------------------------------------------------- */

lexer_t* lexer_new(const char* fpath) {
    lexer_t* lexer = (lexer_t*)malloc(sizeof(lexer_t));

    // store the fpath in the lexer for error reporting purposes
    lexer->fpath = fpath;

    // open the file for reading
    lexer->file = fopen(fpath, "r");
    if (lexer->file == NULL) {
        char buff[256];
        sprintf(buff, "failed to open file at `%s`", fpath);
        report_fatal(buff);
    }

    // initialize the Lexer's position
    lexer->line = 1;
    lexer->start_line = 1;
    lexer->col = 1;
    lexer->start_col = 1;

    // mark the token buffer as empty
    lexer->tok_buff = NULL;

    return lexer;
}

bool lexer_next(lexer_t* lexer, token_t* tok) {
    // initialize the token buffer
    lexer->tok_buff = (char*)malloc(TOK_BUFF_BLOCK_SIZE);
    lexer->tok_buff_len = 0;
    lexer->tok_buff_cap = TOK_BUFF_BLOCK_SIZE;

    // read as long as there are more tokens to be consumed
    for (char ahead = lexer_peek(lexer); ahead != EOF; ahead = lexer_peek(lexer)) {
        switch (ahead) {
            // skip whitespace
            case ' ':
            case '\t':
            case '\v':
            case '\f':
            case '\r':
                lexer_skip(lexer);
                break;
            // handle newlines
            case '\n':
                // mark the start of the newline
                lexer_mark_start(lexer);

                // read it in
                lexer_read(lexer);

                // make and return the token
                *tok = lexer_make_token(lexer, TOK_NEWLINE);
                return true;
            // handle split joins
            case '\\':
                // mark the beginning for error reporting purposes
                lexer_mark_start(lexer);

                // skip the `\` 
                lexer_skip(lexer);

                // expect newline next
                char next = lexer_skip(lexer);
                if (ahead != '\n') {
                    lexer_fail(lexer, "expected newline immediately after backslash");
                    return false;
                }            

                break;
            // handle comments
            case '#':
                // if `lexer_comment` returns true, we should emit a newline
                // after the comment (it was a line comment)
                if (lexer_comment(lexer)) {
                    // mark the start of the newline
                    lexer_mark_start(lexer);

                    // read it in
                    lexer_read(lexer);

                    // make and return the token
                    *tok = lexer_make_token(lexer, TOK_NEWLINE);
                    return true;
                }
                break;
            // handle standard strings
            case '\"':
                return lexer_std_string_lit(lexer, tok);
            // handle runes
            case '\'':
                return lexer_rune_lit(lexer, tok);
            // handle raw/multiline strings
            case '`':
                return lexer_raw_string_lit(lexer, tok);
            default:
                // handle identifiers and keyword
                if (iscsymf(ahead)) {
                    // mark the beginning of the identifier
                    lexer_mark_start(lexer);

                    // keep reading until we encountered a character that can't
                    // be in an identifier
                    do {
                        lexer_read(lexer);
                        ahead = lexer_peek(lexer);
                    } while (iscsym(ahead));

                    // make the identifier or keyword token
                    *tok = lexer_make_token(lexer, lexer_match_keyword(lexer));
                    return true;
                } else if (isdigit(ahead))
                    return lexer_num_lit(lexer, tok);
                else {
                    token_kind_t kind = lexer_oper_or_punct(lexer);

                    // unknown character
                    if (kind == TOK_EOF) {
                        // should never be an actual end of file because
                        // `lexer_oper_or_punct` only returns TOK_EOF if the
                        // first character doesn't match which we know from the
                        // loop condition *can't* be an EOF

                        // build the message and fail
                        char buff[128];
                        sprintf(buff, "unknown character: `%c`", lexer_peek(lexer));
                        lexer_fail(lexer, buff);
                        return false;
                    } else {
                        *tok = lexer_make_token(lexer, kind);
                        return true;
                    }
                }
                break;
        }
    }

    // if the loop breaks, and we reach here, we have found an EOF
    tok->kind = TOK_EOF;
    return true;
}

void lexer_dispose(lexer_t* lexer) {
    // close the opened file
    fclose(lexer->file);

    // if there is still content remaining in the token buffer, free it
    if (lexer->tok_buff != NULL) {
        free(lexer->tok_buff);
    }

    // free the Lexer itself
    free(lexer);
}