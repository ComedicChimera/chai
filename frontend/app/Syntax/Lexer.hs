module Syntax.Lexer where

import Control.Monad

import Text.Parsec
import Text.Parsec.String (Parser)
import Text.Parsec.Language (emptyDef)
import qualified Text.Parsec.Token as Tok 

import Syntax.Internal

lexer :: ChaiLexer
lexer = Tok.makeTokenParser style
    where
        ops = ["+", "*", "-", "/", "//"] -- TODO: rest
        keywords = ["def", "end"] -- TODO: rest
        style = emptyDef {
            Tok.commentLine = "#",
            Tok.commentStart = "#!",
            Tok.commentEnd = "!#",
            Tok.nestedComments = True,
            Tok.identStart = letter <|> char '_',
            Tok.identLetter = alphaNum <|> char '_',
            Tok.reservedNames = keywords,
            Tok.reservedOpNames = ops,
            Tok.caseSensitive = True
        }


-- keyword matches a keyword of the given value
keyword :: String -> ChaiParser String
keyword s = lexeme $ string s

-- keywordEOL matches a keyword at the end of a line
keywordEOL :: String -> ChaiParser String
keywordEOL s = lexemeEOL $ string s

-- lexemeEOL works exactly like Chai's `lexeme` but it does not skip newlines.
-- It is useful for language elements after which a newline might be significant
lexemeEOL :: ChaiParser a -> ChaiParser a
lexemeEOL p = do
    x <- p
    sensitiveWhiteSpace
    splitJoin
    return x

-- lexeme is Chai's "overload" of Parsec's lexeme to handle split-joins
lexeme :: ChaiParser a -> ChaiParser a
lexeme p = do 
    x <- Tok.lexeme lexer p 
    splitJoin
    return x

-- sensitiveWhiteSpace skips all standard whitespace except for newlines
sensitiveWhiteSpace :: ChaiParser ()
sensitiveWhiteSpace = void $ many $ oneOf [' ', '\t', '\v', '\f', '\r']

splitJoin :: ChaiParser ()
splitJoin = optional $ string "\\\n"

multilineStringLit :: ChaiParser String
multilineStringLit = char '`' *> many (noneOf ['`']) <* char '`'
