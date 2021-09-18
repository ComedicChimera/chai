#include "parser.hpp"

#include <format>

#include "report/message.hpp"

namespace chai {
    Parser::Parser(SrcFile& file, Scanner& sc, BuildProfile& profile, DepGraph& depg)
    : sc(sc)
    , file(file)
    , globalProfile(profile)
    , depGraph(depg)
    {}

    Token Parser::next() {
        if (lookahead) {
            auto l = lookahead;
            lookahead = {};
            return l.value();
        }

        return sc.scanNext();
    }

    Token Parser::expect(TokenKind kind) {
        auto tok = next();

        if (tok.kind != kind) {
            // end of file is substitutable for a newline
            if (kind == TokenKind::Newline && tok.kind == TokenKind::EndOfFile)
                return tok;

            reject(tok);
        }     

        return tok;    
    }

    void Parser::reject(const Token& tok) {
        if (tok.kind == TokenKind::EndOfFile)
            throw CompileMessage("unexpected end of file", tok.position, file.parent.parent->getErrorPath(sc.getFilePath()));

        throw CompileMessage(std::format("unexpected token: `{}`", tok.value), tok.position, file.parent.parent->getErrorPath(sc.getFilePath()));
    }

    Token Parser::peek() {
        if (!lookahead)
            lookahead = sc.scanNext();

        return lookahead.value();
    }

    // -------------------------------------------------------------------------- //

    std::optional<ASTRoot*> Parser::parse() {
        ASTRoot* root = new ASTRoot;

        bool more = true;
        while (more) {
            auto curr = next();

            switch (curr.kind) {
                // metadata
                case TokenKind::Not:
                    // expecting `!!`
                    expect(TokenKind::Not);

                    if (!parseMetadata())
                        return {};
                    break;
                // import statement
                case TokenKind::Import:
                    // TODO
                    break;
                // leading newlines get skipped
                case TokenKind::Newline:
                    continue;
                // end of file breaks out of the loop
                case TokenKind::EndOfFile:
                    more = false;
                    break;
                // otherwise, it must be a top level declaration
                default:
                    // TODO

                    // top level will consume all else
                    more = false;
                    break;
            }   
        }             

        return root;
    }

    // parseMetadata parses file metadata (assuming the first two `!!` have been
    // read in). It returns a boolean indicating whether or not compilation
    // should continue based on the metadata and the build configuration.
    bool Parser::parseMetadata() {
        auto state = 0;

        bool requiresParen = false;
        std::string currentKey;
        while (true) {
            switch (state) {
            case 0:
                {
                    auto tok = next();

                    switch (tok.kind) {
                        case TokenKind::Identifier:
                            currentKey = tok.value;
                            state = 2;
                            break;
                        case TokenKind::LParen:
                            requiresParen = true;
                            state = 1;
                            break;
                        default:
                            reject(tok);
                    }
                }
                break;
            case 1:
                currentKey = expect(TokenKind::Identifier).value;
                state = 2;
                break;
            case 2:
                {
                    auto tok = next();

                    switch (tok.kind) {
                        case TokenKind::Comma:
                            state = 1;
                            file.metadata[currentKey] = "";
                            
                            if (currentKey == "no_compile")
                                return false;

                            currentKey = "";
                            break;
                        case TokenKind::Assign:
                            state = 3;
                            break;
                        case TokenKind::Newline:
                            if (!requiresParen)
                                return true;

                            break;
                        case TokenKind::EndOfFile:
                            if (requiresParen)
                                reject(tok);

                            return true;
                        case TokenKind::RParen:
                            if (!requiresParen)
                                reject(tok);

                            return true;
                        default:
                            reject(tok);
                    }
                }
                break;
            case 3:
                {
                    auto currentValue = expect(TokenKind::StringLiteral).value;

                    // erase the quotes
                    currentValue.erase(0);
                    currentValue.pop_back();

                    if (currentKey == "os" && globalProfile.targetOS != currentValue)
                        return false;
                    else if (currentKey == "arch" && globalProfile.targetArch != currentValue)
                        return false;

                    file.metadata[currentKey] = currentValue;
                    currentKey = "";

                    state = 4;
                }
                break;
            case 4:
                {
                    auto tok = next();

                    switch (tok.kind) {
                        case TokenKind::Comma:
                            state = 1;
                            break;
                        case TokenKind::RParen:
                            if (!requiresParen)
                                reject(tok);

                            return true;
                        case TokenKind::Newline:
                            if (!requiresParen)
                                return true;

                            break;
                        case TokenKind::EndOfFile:
                            if (requiresParen)
                                reject(tok);

                            return true;
                        default:
                            reject(tok);
                    }
                }
                break;
            }
        }
    }

    // parseImport parses an import statement assuming the leading `import` has
    // been read.
    void Parser::parseImport() {
        auto firstIdent = expect(TokenKind::Identifier);
        
        // parsing data collected from the import statement
        std::vector<Token> importedSymbols;
        std::string moduleName;
        std::string packagePath;
        TextPosition packagePathPos;
        std::string packageRename;
        TextPosition packageRenamePos;

        while (true) {
            auto punct = next();

            switch (punct.kind) {
                // list of symbol imports (`import id {, id} ...`)
                case TokenKind::Comma:
                    importedSymbols = { firstIdent };
                    concatVec(importedSymbols, parseIdentList(TokenKind::Comma));
                    expect(TokenKind::From);
                    // fallthrough to case after `from`
                // single symbol import (`import id {, id} from id {. id}`)
                case TokenKind::From:
                    {
                        auto packagePathToks = parseIdentList(TokenKind::Dot);
                        packagePathPos = positionOfSpan(packagePathToks[0].position, packagePathToks.back().position);
                        expect(TokenKind::Newline);

                        moduleName = packagePathToks[0].value;
                        packagePathToks.erase(packagePathToks.begin());

                        for (auto& item : packagePathToks) {
                            packagePath.push_back('/');
                            packagePath += item.value;                     
                        }

                        goto loopexit;
                    }
                // import path (`import id {. id} ...`)
                case TokenKind::Dot:             
                    {
                        moduleName = firstIdent.value;
                        auto subPathToks = parseIdentList(TokenKind::Dot);
                        for (auto& pkgName : subPathToks) {
                            packagePath.push_back('/');
                            packagePath += pkgName.value;
                        }

                        packagePathPos = positionOfSpan(firstIdent.position, subPathToks.back().position);

                        auto closingPunct = next();

                        if (closingPunct.kind == TokenKind::Newline)
                            goto loopexit;
                        else if (closingPunct.kind != TokenKind::As)
                            reject(closingPunct);

                        // fallthrough to `as` case
                    }
                // as named import (`import id {. id} as id`)
                case TokenKind::As:
                    {
                        auto ident = expect(TokenKind::Newline);
                        packageRename = ident.value;
                        packageRenamePos = ident.position;
                        goto loopexit;
                    }
                // end of import statement
                case TokenKind::Newline:
                    goto loopexit;
                default:
                    reject(punct);
                    break;
            }
        }        

        loopexit:
            // TODO
    }
}