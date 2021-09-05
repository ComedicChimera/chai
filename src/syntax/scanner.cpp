#include "scanner.hpp"

#include <stdexcept>
#include <format>
#include <locale>
#include <codecvt>

#include "logging/chai_error.hpp"

namespace chai {
    Scanner::Scanner(const std::string& fpath) 
        : line(1)
        , col(0)
        , fpath(fpath)
    {
        file.open(fpath);
        if (!file)
            throw std::runtime_error("unable to open file");        
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

                        // TODO: check to see if the identifier matches a keyword

                        // if not, return it as an identifier
                        return makeToken(TokenKind::Identifier);
                    }
                    // operator or punctuation
                    else {
                        // TODO
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
                while ((ahead == peekChar()) && ahead != '\n') {
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
            } else readChar();

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
                if (isdigit(ahead.value()) || '@' < ahead && ahead < 'G' || '`' < ahead && ahead < 'g') {
                    unicodeStr.push_back(ahead.value());
                    skipChar();
                }

                throwScanError(std::format("expected a hexadecimal character (0-9, a-f, or A-F) not %c", ahead.value()));
            } else
                throwScanError("unexpected end of unicode escape sequence");
        }

        // convert the unicode string to an 32-bit unicode point
        char32_t codePoint = std::stoul(unicodeStr, nullptr, 0);

        // convert the unicode point to utf-8 bytes
        std::wstring_convert<std::codecvt_utf8<char32_t>, char32_t> converter;
        std::string u8bytes = converter.to_bytes(codePoint);

        // write the bytes to the token buffer
        tokBuff += u8bytes;
    }

    // scanNumberLit scans in a numeric literal.
    Token Scanner::scanNumberLit(char c) {
        // read in the leading number
        readChar();

        int base = 10;

        // first, check for variable bases -- note, we do include the leading
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
                            skipChar();
                            base = 2;
                            break;
                        case 'o':
                            skipChar();
                            base = 8;
                            break;
                        case 'x':
                            skipChar();
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
                            throwScanError(std::format("unknown interger base code: %c", ahead.value()));
                            break;
                    }
                }
            }
        }

        // these flags determine what kind of number literal is generated and
        // are used to maintain the state regardingly that literal as it is
        // being read in by the scanner.
        bool isFloating = false, encounteredExp = false;
        bool isUnsigned = false, isLong = false;

        // TODO: main number reading loop
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
        throw ChaiCompileError(msg, {.startLine=line, .startCol=col-(int)tokBuff.length(), .endLine=line, .endCol=col}, fpath);
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
        switch (c) {
            case '\n':
                line++;
                col = 0;
                break;
            case '\t':
                col += 4;
                break;
            default:
                col += 1;
                break;
        }
    }
}