#include "scanner.hpp"

#include <stdexcept>
#include <format>
#include <locale>
#include <codecvt>

#include "patterns.hpp"
#include "report/message.hpp"

namespace chai {
    Scanner::Scanner(const std::string& fpath) 
        : line(1)
        , col(1)
        , fpath(fpath)
    {
        file.open(fpath);
        if (!file)
            throw std::runtime_error(std::format("unable to open file at {}", fpath));        
    }

    Token Scanner::scanNext() {
        while (auto oc = peekChar()) {
            switch (oc.value()) {
                // skip all non-meaningful whitespace characters
                case '\t':
                case ' ':
                case '\r':
                case '\v': 
                case '\f':
                    skipChar();
                    break;
                // handle newlines
                case '\n':
                    readChar();
                    return makeToken(TokenKind::Newline);
                // handle comments
                case '#':
                    skipComment();
                    break;
                // handle split join
                case '\\':
                    skipChar();

                    if (auto ahead = peekChar()) {
                        if (ahead == '\n')
                            skipChar();
                        else {
                            throwScanError("expecting newline after split join");
                        }
                    }

                    break;
                // handle rune literals
                case '\'':
                    readChar();
                    return scanRuneLit();
                // handle std string literals
                case '\"':
                    readChar();
                    return scanStdStringLit();
                // handle raw string literals
                case '`':
                    readChar();
                    return scanRawStringLit();
                default:
                    // numeric constant
                    if (isdigit(oc.value()))
                        return scanNumberLit(oc.value());
                    // identifier or keyword
                    else if (iscsymf(oc.value())) {
                        // read in as long as the next character is valid
                        do {
                            readChar();
                        } while ((oc = peekChar()) && iscsym(oc.value()));

                        // check if the identifier matches a known keyword and
                        // return an appropriate token if so
                        if (auto kind = matchKeyword(tokBuff))
                            return makeToken(kind.value());

                        // if not, return it as an identifier
                        return makeToken(TokenKind::Identifier);
                    }
                    // operator or punctuation
                    else {
                        TokenKind kind = TokenKind::EndOfFile;
                        std::optional<TokenKind> okind;
                        while ((oc = peekChar()) && (okind = matchSymbol(tokBuff + oc.value()))) {
                            readChar();
                            kind = okind.value();
                        }                           

                        // kind == EOF => first token didn't match a symbol
                        if (kind == TokenKind::EndOfFile) {
                            readChar();
                            throwScanError("unexpected character");
                        }
                        
                        // otherwise it is a valid token :)
                        return makeToken(kind);
                    }
                    break;
            }
        }

        // if we reach here, this is always an EOF
        return {.kind = TokenKind::EndOfFile, .line=line, .col=col};
    }

    void Scanner::skipComment() {
        skipChar();

        if (auto ahead = peekChar()) {
            // multi-line comment
            if (ahead == '!') {
                skipChar();

                while (auto ahead = peekChar()) {
                    // discard the character
                    skipChar();

                    // if we hit an exclamation point, check if what
                    // follows is a `#`.  if so, exit the comment
                    if (ahead == '!') {
                        if ((ahead = peekChar()) && ahead == '#') {
                            skipChar();
                            break;
                        }
                    }
                }
            }
            // single-line comment 
            else {
                while ((ahead = peekChar()) && ahead != '\n') {
                    skipChar();
                }         
            }
        }
    }

    // scanRuneLit scans in a rune literal assuming the leading single quote has
    // already been read in.
    Token Scanner::scanRuneLit() {
        if (auto ahead = peekChar()) {
            if (ahead == '\\') {
                skipChar();

                readEscapeSequence();
            } else if (ahead == '\'')
                throwScanError("empty rune literal");
            else readChar();

            if ((ahead = peekChar()) && ahead == '\'') {
                readChar();

                return makeToken(TokenKind::RuneLiteral);
            }
        }

        // missing the closing quote
        throwScanError("expected closing quote for rune literal");
    }

    // readEscapeSequence is used to read in and interpret the sequence of
    // characters following a backslash in a string or rune.  It assumes the
    // leading backslash has already been skipped.  The correct "value" of the
    // escape sequence is inserted into the token buffer.
    void Scanner::readEscapeSequence() {
        if (auto c = peekChar()) {
            skipChar();
            switch (c.value()) {
                case 'a':
                    tokBuff.push_back('\a');
                    break;
                case 'b':
                    tokBuff.push_back('\b');
                    break;
                case 'f':
                    tokBuff.push_back('\f');
                    break;
                case 'n':
                    tokBuff.push_back('\n');
                    break;
                case 'r':
                    tokBuff.push_back('\r');
                    break;
                case 't':
                    tokBuff.push_back('\t');
                    break;
                case 'v':
                    tokBuff.push_back('\v');
                    break;
                case '0':
                    tokBuff.push_back('\0');
                    break;
                case '\'':
                    tokBuff.push_back('\'');
                    break;
                case '\"':
                    tokBuff.push_back('\"');
                    break;
                case '\\':
                    tokBuff.push_back('\\');
                    break;
                // special utf-8 encodings
                case 'x':
                    readUnicodeSequence(2);
                    break;
                case 'u':
                    readUnicodeSequence(4);
                    break;
                case 'U':
                    readUnicodeSequence(8);
                    break;
                default:
                    throwScanError("unknown escape code");
                    break;
            }
        } else
            throwScanError("expected escape sequence following backslash in literal");
    }

    // readUnicodeSequence reads a sequence of hex-encoded values and interprets
    // them as a single unicode code point.  This point is then converted into
    // UTF-8 bytes and written to the token buffer.  The input determines the
    // expected length of the unicode sequence.
    void Scanner::readUnicodeSequence(int count) {
        // read in the raw unicode string
        std::string unicodeStr;
        for (int i = 0; i < count; i++) {
            if (auto ahead = peekChar()) {
                if (isdigit(ahead.value()) || 'A' <= ahead && ahead < 'G' || 'a' <= ahead && ahead < 'g') {
                    unicodeStr.push_back(ahead.value());
                    skipChar();
                }
                else
                    throwScanError(std::format("expected a hexadecimal character (0-9, a-f, or A-F) not {}", ahead.value()));
            } else
                throwScanError("unexpected end of unicode escape sequence");
        }

        // convert the unicode string to an 32-bit unicode point
        char32_t codePoint = std::stoul(unicodeStr, nullptr, 16);

        // convert the unicode point to utf-8 bytes
        std::wstring_convert<std::codecvt_utf8<char32_t>, char32_t> converter;

        try {
            std::string u8bytes = converter.to_bytes(codePoint);

            // write the bytes to the token buffer
            tokBuff += u8bytes;
        } catch (std::exception) {
            throwScanError(std::format("{} is not a valid unicode character", unicodeStr));
        }      
    }

    // scanNumberLit scans in a numeric literal.
    Token Scanner::scanNumberLit(char c) {
        // read in the leading number
        readChar();

        int base = 10;

        // first, check for alternative bases -- note, we do include the leading
        // zero even in the case of a variable base so that backend knows what
        // kind of number literal to output to LLVM
        if (c == '0') {
            if (auto ahead = peekChar()) {
                // if the next character is a digit, it is obviously not a
                // variable base so we keep reading
                if (!isdigit(ahead.value())) {
                    switch (ahead.value()) {
                        // handle the three base codes
                        case 'b':
                            readChar();
                            base = 2;
                            break;
                        case 'o':
                            readChar();
                            base = 8;
                            break;
                        case 'x':
                            readChar();
                            base = 16;
                            break;
                        // handle all non-digit, but valid components of numeric
                        // literals
                        case 'e':
                        case 'E':
                        case '.':
                            break;
                        default:
                            // read in the character so we have something to error on
                            readChar();

                            // throw an error indicating that the base code is unknown
                            throwScanError(std::format("unknown interger base code: {}", ahead.value()));
                            break;
                    }
                }
            }
        }

        // these flags determine what kind of number literal is generated and
        // are used to maintain the state regardingly that literal as it is
        // being read in by the scanner.
        bool isFloating = false, encounteredExp = false, expectingNeg = false;
        bool isUnsigned = false, isLong = false;

        while (auto ahead = peekChar()) {
            // suffixes can only occur once per and at the end
            if (isUnsigned && isLong)
                break;
            else if (isUnsigned) {
                if (ahead == 'l') {
                    readChar();
                    isLong = true;
                    continue;
                }
                    
                break;
            } else if (isLong) {
                if (ahead == 'u') {
                    readChar();
                    isUnsigned = true;
                    continue;
                }
                
                break;
            }

            switch (base) {
                // binary literals
                case 2:
                    if (ahead == '0' || ahead == '1')
                        readChar();
                    else
                        goto loopexit;
                    break;
                // octal literals
                case 8:
                    if ('0' <= ahead && ahead < '8')
                        readChar();
                    else
                        goto loopexit;
                    break;
                // hex literals
                case 16:
                    if (isdigit(ahead.value()) || 'a' <= ahead && ahead <= 'f' || 'A' <= ahead && ahead <= 'F')
                        readChar();
                    else
                        goto loopexit;
                    break;
                // base 10 literals are the only literals that can also be
                // floating point numbers
                case 10:
                    if (isdigit(ahead.value()))
                        readChar();
                    else {
                        switch (ahead.value()) {
                            case 'e':
                            case 'E':
                                readChar();

                                if (encounteredExp) 
                                    throwScanError("floating literal cannot contain multiple exponents");

                                encounteredExp = true;
                                isFloating = true;
                                expectingNeg = true;
                                continue;
                            case '.':
                                readChar();

                                // if it is floating then either the decimal is misplaced after the exponent
                                // or multiple decimals occur in the literal: this is invalid in either case
                                if (isFloating)
                                    throwScanError("unexpected decimal point in float literal");

                                isFloating = true;
                                break;
                            case 'l':
                                readChar();

                                if (isFloating)
                                    throwScanError("float literal cannot be marked long");

                                isLong = true;
                                break;
                            case 'u':
                                readChar();

                                if (isFloating)
                                    throwScanError("float literal cannot be marked unsigned");

                                isUnsigned = true;   
                                break;
                            case '-':
                                // an unexpected negative is not necessarily an
                                // exit condition: eg. `5-2` is a valid Chai
                                // expression; erroring here would cause such an
                                // expression to be rejected
                                if (expectingNeg) {
                                    readChar();
                                    expectingNeg = false;
                                    continue;
                                }

                                // because using the same keyword to exit switch
                                // cases and loops was an *AMAZING* structured
                                // programming innovation /s
                                goto loopexit;  
                            default:
                                goto loopexit;                             
                        }
                    }

                    // clear expecting negative if we reach here
                    expectingNeg = false;
            }
        }

    loopexit:
        if (isFloating)
            return makeToken(TokenKind::FloatLiteral);
        else if (isUnsigned || isLong || base != 10)
            return makeToken(TokenKind::IntLiteral);
        else
            return makeToken(TokenKind::NumLiteral);
    }

    // scanStdStringLit scans in a standard (double quoted) string literal
    // assuming the leading double quote has already been read in.
    Token Scanner::scanStdStringLit() {
        while (auto ahead = readChar()) {
            switch (ahead.value()) {
                case '\n':
                    throwScanError("standard string cannot contain newlines");
                    break;
                case '\"':
                    return makeToken(TokenKind::StringLiteral);
                case '\\':
                    readEscapeSequence();
                    break;
            }
        }

        throwScanError("expected closing double quote for standard string literal");
    }

    // scanRawStringLit scans in a raw/multiline (backtick quoted) string
    // literal assuming the leading backtick has already been read in.
    Token Scanner::scanRawStringLit() {
        while (auto ahead = readChar()) {
            if (ahead == '`')
                return makeToken(TokenKind::StringLiteral);
        }

        throwScanError("expected closing backtick for raw string literal");
    }

    // -------------------------------------------------------------------------- //

    // makeToken creates a new token based on the contents of the token buffer
    Token Scanner::makeToken(TokenKind kind) {
        Token token = {.kind = kind, .value=tokBuff, .line=line, .col=col};
        tokBuff.clear();

        return token;
    }

    // throwScanError throws a new error regarding the current token being scanned
    void Scanner::throwScanError(const std::string& msg) {
        throw CompileMessage(msg, {.startLine=tokStartLine, .startCol=tokStartCol, .endLine=line, .endCol=col}, fpath);
    }

    // -------------------------------------------------------------------------- //

    // readChar attempts to get a new character from the input stream.  It also
    // updates the position of the scanner and adds the character to the current
    // token buffer.
    std::optional<char> Scanner::readChar() {
        char c = 0;
        if (!file.get(c))
            return {};

        updatePosition(c);
        tokBuff.push_back(c);
        return c;
    }

    // skipChar attempts to get a new character from the input stream.  If it
    // succeeds, the scanner's position is updated and the character is
    // discarded.
    bool Scanner::skipChar() {
        char c = 0;
        if (!file.get(c))
            return false;

        updatePosition(c);
        return true;
    }

    // peekChar attempts to look one char ahead in the input stream.  If it
    // succeeds the char is returned.  In any case, this function has no effect
    // on the scanner's state -- it acts as a lookahead.
    std::optional<char> Scanner::peekChar() {
        int c = file.peek();
        if (c == EOF)
            return {};

        return (char)c;
    }

    // updatePosition updates the position of the scanner based on a char that
    // was just read in.
    void Scanner::updatePosition(char c) {
        // whenever the token buffer is of length zero and a character is
        // written to it, the start position is reset
        if (tokBuff.length() == 0) {
            tokStartLine = line;
            tokStartCol = col;
        }

        switch (c) {
            case '\n':
                line++;
                col = 1;
                break;
            case '\t':
                col += 4;
                break;
            default:
                col++;
                break;
        }
    }
}