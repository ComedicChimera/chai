#include "parser.hpp"

#include <format>

#include "report/message.hpp"

namespace chai {
    template<typename ...T>
    void Parser::throwError(const TextPosition& pos, const std::string& msg, T... args) {
        throw CompileMessage(std::format(msg, args), pos, file.parent->parent->getErrorPath(file.filePath));
    }

    void Parser::throwSymbolAlreadyDefined(const TextPosition& pos, const std::string& name) {
        throwError(pos, "symbol already defined in scope: `{}`", name);
    }
}