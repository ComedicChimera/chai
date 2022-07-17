#include "lexer.hpp"

#include <unordered_map>
#include <cctype>

namespace chai {
    Token Lexer::nextToken() {
        char c;
        while ((c = peek()) != EOF) {
            switch (c) {
                case ' ':
                case '\t':
                case '\n':
                case '\r':
                case '\v':
                case '\f':
                    // Skip all the whitespace characters.
                    skip();
                    break;
                case '/':
                    // Read in the `/` token in case we have a divide token.
                    mark();
                    read();

                    // Check for comment tokens.
                    c = peek();
                    if (c == '/' || c == '*') {
                        // Clear the read in token.
                        m_tokBuff.clear();

                        // Skip the comment and continue reading tokens.
                        skipComment(c);
                        continue;
                    }

                    // No comment => return a division token.
                    return makeToken(TokenKind::DIV);
                case '\'':
                    return lexRune();
                case '\"':
                    return lexStdString();
                case '`':
                    return lexRawString();
                default:
                    if (isdigit(c)) {
                        return lexNumeric();
                    } else if (isalpha(c) || c == '_') {
                        return lexKeywordOrIdent();
                    } else {
                        return lexPunctuationOrOperator();
                    }
            }
        }

        // No more characters to read => EOF.
        return { TokenKind::END_OF_FILE };
    }

    /* ---------------------------------------------------------------------- */

    // The table of operator/punctuation patterns matched to their token kinds.
    std::unordered_map<std::string, TokenKind> symbolPatterns {
        { "+", TokenKind::PLUS },
        { "-", TokenKind::MINUS },
        { "*", TokenKind::STAR },
        { "/", TokenKind::DIV },
        { "%", TokenKind::MOD },
        { "**", TokenKind::POWER },

        { "==", TokenKind::EQ },
        { "!=", TokenKind::NEQ },
        { "<", TokenKind::LT },
        { ">", TokenKind::GT },
        { "<=", TokenKind::LTEQ },
        { ">=", TokenKind::GTEQ },

        { "~", TokenKind::COMPL },
        { "<<", TokenKind::LSHIFT },
        { ">>", TokenKind::RSHIFT },
        { "&", TokenKind::AMP },
        { "^", TokenKind::CARRET },
        { "|", TokenKind::PIPE },

        { "!", TokenKind::NOT },
        { "&&", TokenKind::AND },
        { "||", TokenKind::OR },

        { "=", TokenKind::ASSIGN },
        { "++", TokenKind::INCREMENT },
        { "--", TokenKind::DECREMENT },

        { "(", TokenKind::LPAREN },
        { ")", TokenKind::RPAREN },
        { "{", TokenKind::LBRACE },
        { "}", TokenKind::RBRACE },
        { "[", TokenKind::LBRACKET },
        { "]", TokenKind::RBRACKET },
        { ",", TokenKind::COMMA },
        { ":", TokenKind::COLON },
        { ";", TokenKind::SEMICOLON },
        { "@", TokenKind::ATSIGN },
        { ".", TokenKind::DOT },
    };

    Token Lexer::lexPunctuationOrOperator() {
        mark();
        char c = read();

        // Assert that the first character of the operator/punctuation is a
        // known lexical symbol: if it isn't, we need to error.
        if (symbolPatterns.find(m_tokBuff) == symbolPatterns.end()) {
            error("unknown symbol");
        }

        // Read until we don't match an operators/punctuations.
        while ((c = peek()) != EOF && symbolPatterns.find(m_tokBuff + c) != symbolPatterns.end()) {
            read();
        }

        return makeToken(symbolPatterns[m_tokBuff]);
    }

    /* ---------------------------------------------------------------------- */

    // The table of keyword patterns matched to their token kinds.
    std::unordered_map<std::string, TokenKind> keywordPatterns {
        { "package", TokenKind::PACKAGE },

        { "func", TokenKind::FUNC },
        { "oper", TokenKind::OPER },

        { "let", TokenKind::LET },
        { "const", TokenKind::CONST },

        { "if", TokenKind::IF },
        { "elif", TokenKind::ELIF },
        { "else", TokenKind::ELSE },
        { "while", TokenKind::WHILE },
        { "for", TokenKind::FOR },

        { "break", TokenKind::BREAK },
        { "continue", TokenKind::CONTINUE },
        { "return", TokenKind::RETURN },

        { "as", TokenKind::AS },  

        { "unit", TokenKind::UNIT },
        { "bool", TokenKind::BOOL },
        { "i8", TokenKind::I8 },
        { "u8", TokenKind::U8 },
        { "i16", TokenKind::I16 },
        { "u16", TokenKind::U16 },
        { "i32", TokenKind::I32 },
        { "u32", TokenKind::U32 },
        { "i64", TokenKind::I64 },
        { "u64", TokenKind::U64 },
        { "f32", TokenKind::F32 },
        { "f64", TokenKind::F64 },  

        { "null", TokenKind::NULLLIT },
        { "true", TokenKind::BOOLLIT },
        { "false", TokenKind::BOOLLIT },
    };

    Token Lexer::lexKeywordOrIdent() {
        mark();
        read();

        char c;
        while (isalnum(c = peek()) || c == '_') {
            read();
        }

        // Check if identifier is a keyword.
        if (keywordPatterns.find(m_tokBuff) == keywordPatterns.end()) {
            return makeToken(TokenKind::IDENTIFIER);
        } else {
            return makeToken(keywordPatterns[m_tokBuff]);
        }
    }

    /* ---------------------------------------------------------------------- */

    Token Lexer::lexNumeric() {
        mark();
        char c = read();

        // Whether the lexer requires the next character to be a digit.
        bool requiresDigit = false;

        // Read the base prefix if it exists.
        int base = 10;
        if (c == '0') {
            switch (peek()) {
            case 'b':
                read();

                base = 2;
                requiresDigit = true;
                break;
            case 'o':
                read();

                base = 8;
                requiresDigit = true;
                break;
            case 'x':
                read();

                base = 16;
                requiresDigit = true;
                break;
            }
        }

        // Whether the number is a floating-point literal.
        bool isFloat = false;

        // Whether the floating-point literal has an exponent.
        bool hasExp = false;

        // Whether the lexer is expecting a `-` sign for a negative exponent.
        bool expectNeg = false;

        // Whether the integer literal is 64 bit; is unsigned.
        bool isLong = false, isUnsigned = false;

        while ((c = peek()) != EOF) {
            // Skip any `_` used for readability.
            if (c == '_') {
                skip();
                continue;
            }

            // Handle integer suffixes.
            if (!isFloat) {
                if (isUnsigned) {
                    if (c == 'l') {
                        read();

                        isLong = true;
                    }

                    break;
                } else if (isLong) {
                    if (c == 'u') {
                        read();

                        isUnsigned = true;
                    }

                    break;
                } else if (c == 'u') {
                    read();

                    isUnsigned = true;
                    continue;
                } else if (c == 'l') {
                    read();

                    isLong = true;
                    continue;
                }
            }
            
            // Handle each base individually.
            switch (base) {
            case 2: // Base 2 can only be number or int: no float logic.
                if (c == '0' || c == '1')
                    read();
                else 
                    goto exitLoop;

                break;
            case 8: // Base 8 can only be number or int: no float logic.
                if ('0' <= c && c < '9')
                    read();
                else
                    goto exitLoop;

                break;
            case 10: // Base 10 can be a number, int, or float.
                switch (c) {
                case '.':
                    if (requiresDigit || isFloat)
                        goto exitLoop;
                    
                    read();

                    isFloat = true;
                    requiresDigit = true;
                    continue;
                case 'e':
                case 'E':
                    if (requiresDigit || hasExp)
                        goto exitLoop;

                    read();

                    isFloat = true;
                    hasExp = true;
                    expectNeg = true;
                    requiresDigit = true;
                    continue;
                case '-':
                    if (!expectNeg)
                        goto exitLoop;

                    read();

                    expectNeg = false;
                    continue;
                default:
                    if (isdigit(c)) {
                        read();
                        expectNeg = false;
                    } else
                        goto exitLoop;
                }
                
                break;
            case 16: // Base 16 can be a number, int, or float.
                switch (c) {
                case '.':
                    if (requiresDigit || isFloat)
                        goto exitLoop;
                    
                    read();

                    isFloat = true;
                    requiresDigit = true;
                    continue;
                case 'p':
                case 'P':
                    if (requiresDigit || hasExp)
                        goto exitLoop;

                    read();

                    isFloat = true;
                    hasExp = true;
                    expectNeg = true;
                    requiresDigit = true;
                    continue;
                case '-':
                    if (!expectNeg)
                        goto exitLoop;

                    read();

                    expectNeg = false;
                    continue;
                default:
                    if (isxdigit(c)) {
                        read();
                        expectNeg = false;
                    } else
                        goto exitLoop;
                
                break;
                }
            }

            // Indicate that a digit was received.
            requiresDigit = false;
        }

    exitLoop:
        // Ensure that the literal is not malformed.
        if (requiresDigit)
            error("incomplete numeric literal");

        // Determine the type of literal to return.
        if (isFloat)
            return makeToken(TokenKind::FLOATLIT);
        else if (isUnsigned || isLong)
            return makeToken(TokenKind::INTLIT);
        else
            return makeToken(TokenKind::NUMLIT);
    }

    /* ---------------------------------------------------------------------- */

    void Lexer::readEscapeCode() {
        switch (read()) {
            case 'a':
            case 'b':
            case 'f':
            case 'n':
            case 'r':
            case 't':
            case 'v':
            case '0':
            case '\'':
            case '\"':
            case '\\':
                // Read standard escape codes.
                break;
            case EOF:
                error("expected escape code not end of file");
            default:
                error("unknown escape code");
        }
    }

    Token Lexer::lexRune() {
        mark();
        skip();

        // Validate the content of the rune literal.
        char c = read();
        switch (c) {
            case '\n':
                error("rune literal cannot contain newlines");
                break;
            case EOF:
                error("expected closing single quote not end of file");
                break;
            case '\'':
                error("rune literal cannot be empty");
                break;
            case '\\':
                readEscapeCode();
                break;
        }

        // Ensure that the rune literal ends properly.
        c = skip();
        if (c == EOF)
            error("expected closing single quote not end of file");
        else if (c != '\'')
            error("rune literal cannot contain multiple characters");

        return makeToken(TokenKind::RUNELIT);
    }

    Token Lexer::lexStdString() {
        mark();
        skip();

        char c;
        while ((c = peek()) != '\"') {
            switch (c) {
                case EOF:
                    error("expected closing double quote not end of file");
                    break;
                case '\n':
                    error("standard string literal cannot contain newlines");
                    break;
                case '\\':
                    read();
                    readEscapeCode();
                    break;
                default:
                    read();
                    break;
            }
        }

        skip();

        return makeToken(TokenKind::STRINGLIT);
    }

    Token Lexer::lexRawString() {
        mark();
        skip();

        char c;
        while ((c = peek()) != '`') {
            switch (c) {
                case EOF:
                    error("expected closing backtick not end of file");
                    break;
                case '\\':
                    // Test for the `\`` and `\\` escape codes.
                    skip();

                    c = peek();
                    if (c == '\\' || c == '`')
                        read();
                    else {
                        // Add the `\` back into the token buffer.
                        m_tokBuff.push_back('\\');

                        read();
                    }
                    break;
                default:
                    read();
                    break;  
            }
        }

        skip();

        return makeToken(TokenKind::STRINGLIT);
    }

    /* ---------------------------------------------------------------------- */

    void Lexer::skipComment(char c) {
        if (c == '/') {
            // Skip the second `/` of the comment.
            skip();
            
            // Skip the line comment: newline or EOF will end it.
            while ((c = peek()) != '\n' && c != EOF)
                skip();
        } else {
            // Skip the `*` of the comment.
            skip();

            // If we encounter an EOF, then we want to stop reading regardless
            // of if we have token or not. 
            while ((c = peek()) != EOF) {
                // We always want to skip the token we are on even if it ends
                // the comment.
                skip();

                // Check for the `*/` which ends the token.
                if (c == '*' || (c = peek()) == '/') {
                    skip();
                    break;
                }
            }
        }
    }

    /* ---------------------------------------------------------------------- */

    Token Lexer::makeToken(TokenKind kind) { 
        Token tok { 
            .kind { kind }, 
            .value { m_tokBuff }, 
            .span { m_startLine, m_startCol, m_line, m_col },
        };

        m_tokBuff.clear();

        return tok;
    }

    void Lexer::mark() {
        m_startLine = m_line;
        m_startCol = m_col;
    }


    void Lexer::error(const std::string& message) {
        throw CompileError(
            m_chFile->displayPath(), 
            m_chFile->absPath(), 
            message, 
            TextSpan { m_startLine, m_startCol, m_line, m_col }
        );
    }

    /* ---------------------------------------------------------------------- */

    char Lexer::read() {
        char c = m_file.get();
        if (c == EOF)
            return -1;

        m_tokBuff.push_back(c);
        updatePosition(c);

        return c;
    }

    char Lexer::skip() {
        char c = m_file.get();
        if (c == EOF)
            return -1;

        updatePosition(c);

        return c;
    }

    char Lexer::peek() {
        return m_file.peek();
    }

    void Lexer::updatePosition(char c) {
        switch (c) {
        case '\n':
            m_line++;
            m_col = 0;
            break;
        case '\t':
            m_col += 4;
            break;
        default:
            m_col++;
            break;
        }
    }
}