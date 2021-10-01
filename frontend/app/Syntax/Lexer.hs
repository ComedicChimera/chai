module Syntax.Lexer where

import Control.Monad

import Text.Parsec
import Text.Parsec.String (Parser)
import Text.Parsec.Language (emptyDef)
import qualified Text.Parsec.Token as Tok 

lexer :: Tok.TokenParser ()
lexer = Tok.makeTokenParser style
    where
        ops = ["+", "*", "-", "/", "//"] -- TODO: rest
        keywords = ["def", "import"] -- TODO: rest
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


-- lexemeEOL works exactly like Chai's `lexeme` but it does not skip newlines.
-- It is useful for language elements after which a newline might be significant
lexemeEOL :: Parser a -> Parser a
lexemeEOL p = do
    x <- p
    sensitiveWhiteSpace
    splitJoin
    return x

-- lexeme is Chai's "overload" of Parsec's lexeme to handle split-joins
lexeme :: Parser a -> Parser a
lexeme p = do 
    x <- Tok.lexeme lexer p 
    splitJoin
    return x

-- sensitiveWhiteSpace skips all standard whitespace except for newlines
sensitiveWhiteSpace :: Parser ()
sensitiveWhiteSpace = void $ many $ oneOf [' ', '\t', '\v', '\f', '\r']

splitJoin :: Parser ()
splitJoin = optional $ string "\\\n"

multilineStringLit :: Parser String
multilineStringLit = char '`' *> many (noneOf ['`']) <* char '`'
