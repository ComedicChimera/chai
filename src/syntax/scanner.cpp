#include "scanner.hpp"

#include <stdexcept>

namespace chai {
    Scanner::Scanner(const std::string& fpath) 
        : line(1)
        , col(0)
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
                            // TODO: throw token error
                        }
                    }
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

    // -------------------------------------------------------------------------- //

    // makeToken creates a new token based on the contents of the token buffer
    Token Scanner::makeToken(TokenKind kind) {
        Token token = {.kind = kind, .value=tokBuff, .line=line, .col=col};
        tokBuff.clear();

        return token;
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