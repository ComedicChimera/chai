#include "patterns.hpp"

#include <unordered_map>

namespace chai {
    std::unordered_map<std::string, TokenKind> keywordPatterns = { 
        { "let", TokenKind::Let },

        { "if", TokenKind::If },
        { "elif", TokenKind::Elif },
        { "else", TokenKind::Else },
        { "while", TokenKind::While },
        { "do", TokenKind::Do },
        { "continue", TokenKind::Continue },
        { "break", TokenKind::Break },
        { "after", TokenKind::After },
        { "match", TokenKind::Match },
        { "case", TokenKind::Case },
        { "return", TokenKind::Return },
        { "end", TokenKind::End },

        { "def", TokenKind::Def },
        { "type", TokenKind::Type },
        { "oper", TokenKind::Oper },

        { "pub", TokenKind::Pub },
        { "vol", TokenKind::Vol },

        { "import", TokenKind::Import },
        { "from", TokenKind::From },

        { "as", TokenKind::As },

        { "i8", TokenKind::I8 },
        { "i16", TokenKind::I16 },
        { "i32", TokenKind::I32 },
        { "i64", TokenKind::I64 },
        { "u8", TokenKind::U8 },
        { "u16", TokenKind::U16 },
        { "u32", TokenKind::U32 },
        { "u64", TokenKind::U64 },
        { "f32", TokenKind::F32 },
        { "f64", TokenKind::F64 },
        { "rune", TokenKind::Rune },
        { "bool", TokenKind::Bool },
        { "nothing", TokenKind::Nothing },
    };

    std::optional<TokenKind> matchKeyword(const std::string& identifier) {
        if (keywordPatterns.contains(identifier))
            return keywordPatterns[identifier];

        return {};
    }

    std::unordered_map<std::string, TokenKind> symbolPatterns = {
        { "+", TokenKind::Plus },
        { "-", TokenKind::Minus },
        { "*", TokenKind::Star },
        { "/", TokenKind::Div },
        { "//", TokenKind::FloorDiv },
        { "%", TokenKind::Mod },
        { "**", TokenKind::Power },

        { "&", TokenKind::Amp },
        { "|", TokenKind::Pipe },
        { "^", TokenKind::Carret },
        { "<<", TokenKind::LShift },
        { ">>", TokenKind::RShift },
        { "~", TokenKind::Complement },

        { "&&", TokenKind::And },
        { "||", TokenKind::Or },
        { "!", TokenKind::Not },

        { "<", TokenKind::Lt },
        { ">", TokenKind::Gt },
        { "<=", TokenKind::LtEq },
        { ">=", TokenKind::GtEq },
        { "==", TokenKind::Eq },
        { "!=", TokenKind::NEq },

        { ".", TokenKind::Dot },
        { "..", TokenKind::RangeTo },
        { "...", TokenKind::Ellipsis },

        { "::", TokenKind::TypeProp },

        { "=", TokenKind::Assign },
        { "++", TokenKind::Increm },
        { "--", TokenKind::Decrem },

        { "(", TokenKind::LParen },
        { ")", TokenKind::RParen },
        { "[", TokenKind::LBracket },
        { "]", TokenKind::RBracket },
        { "{", TokenKind::LBrace },
        { "}", TokenKind::RBrace },
        { ",", TokenKind::Comma },
        { ":", TokenKind::Colon },
        { "->", TokenKind::Arrow },
        { ";", TokenKind::Semicolon },
    };

    std::optional<TokenKind> matchSymbol(const std::string& symbol) {
        if (symbolPatterns.contains(symbol))
            return symbolPatterns[symbol];

        return {};
    }
}
