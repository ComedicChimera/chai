module Chai.Lexer where

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

-- Handling some Chai specific lexeme semantics
lexeme :: Parser a -> Parser a
lexeme p = do 
    x <- Tok.lexeme lexer p 
    splitJoin
    return x

-- Split-join handling (optionally reads in a split-join)
splitJoin :: Parser ()
splitJoin = optional $ string "\\\n"

-- Keyword handling
keyword :: String -> Parser String
keyword k = Tok.lexeme lexer $ string k

-- Multiline/raw string literals
multilineStringLit :: Parser String
multilineStringLit = char '`' *> many (noneOf ['`']) <* char '`'
