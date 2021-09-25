#include "parser.hpp"

namespace chai {
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

                    if (currentKey == "os" && compiler->getProfile().targetOS != currentValue)
                        return false;
                    else if (currentKey == "arch" && compiler->getProfile().targetArch != currentValue)
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
        std::vector<Token> importedSymbolToks;
        std::string moduleName;
        std::string packagePath;
        TextPosition packagePathPos;
        std::string packageName;
        TextPosition packageNamePos;

        while (true) {
            auto punct = next();

            switch (punct.kind) {
                // list of symbol imports (`import id {, id} ...`)
                case TokenKind::Comma:
                    importedSymbolToks = { firstIdent };
                    concatVec(importedSymbolToks, parseIdentList(TokenKind::Comma));
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
                            packagePath.push_back('.');
                            packagePath += item.value;                     
                        }

                        // we don't update the name/name position here since it
                        // will never be used

                        goto loopexit;
                    }
                // import path (`import id {. id} ...`)
                case TokenKind::Dot:             
                    {
                        moduleName = firstIdent.value;
                        auto subPathToks = parseIdentList(TokenKind::Dot);
                        for (auto& pkgName : subPathToks) {
                            packagePath.push_back('.');
                            packagePath += pkgName.value;
                        }

                        packagePathPos = positionOfSpan(firstIdent.position, subPathToks.back().position);

                        auto closingPunct = next();

                        if (closingPunct.kind == TokenKind::Newline)
                            goto loopexit;
                        else if (closingPunct.kind != TokenKind::As)
                            reject(closingPunct);

                        packageName = subPathToks.back().value;
                        packageNamePos = subPathToks.back().position;

                        // fallthrough to `as` case
                    }
                // as named import (`import id {. id} as id`)
                case TokenKind::As:
                    {
                        auto ident = expect(TokenKind::Newline);
                        packageName = ident.value;
                        packageNamePos = ident.position;
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
        // start by attempting to import that package
        auto opkg = compiler->importPackage(file.parent->parent, moduleName, packagePath);

        // throw an appropriate error if the package was unable to be imported
        if (!opkg) {
            auto fullPathStr = moduleName;
            if (packagePath != "")
                fullPathStr += packagePath;

            throwError(packagePathPos, "unable to locate package: `{}`", fullPathStr);
        }

        // add that package to the parent package
        auto pkg = opkg.value();       
        file.parent->importedPackages[pkg->id] = pkg;

        // no imported symbols => package imported by name
        if (importedSymbolToks.size() == 0) {
            // check for name collisions
            if (globalCollides(packageName))
                throwSymbolAlreadyDefined(packageNamePos, packageName);

            // add it as a visible package
            file.visiblePackages[packageName] = pkg;
        } else {
            for (auto& tok : importedSymbolToks) {
                // check for name collisions
                if (globalCollides(tok.value))
                    throwSymbolAlreadyDefined(tok.position, tok.value);

                // perform a look up in the package we are importing from for
                // the symbol, specifying that the symbol must be public
                auto oimportedSymbol = pkg->globalTable.lookup(tok.value, tok.position, DefKind::Unknown, true);
                if (!oimportedSymbol)
                    throwError(tok.position, "symbol by name `{}` is not publically visible in package", tok.value);
            }
        }
    }
}