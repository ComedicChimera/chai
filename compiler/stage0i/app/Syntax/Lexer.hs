module Syntax.Lexer where

import Data.Functor

import Text.Parsec
import Text.Parsec.String (Parser)
import Text.Parsec.Language (emptyDef)
import qualified Text.Parsec.Token as Tok 

import Syntax.Internal

import Report.Message

lexer :: Tok.TokenParser ()
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


-- identifier reads in a new identifier
identifier :: Parser (String, TextPosition)
identifier = positionOf $ Tok.identifier lexer

-- keyword matches a keyword of the given value
keyword :: String -> Parser String
keyword s = lexeme $ string s

-- lexeme is Chai's "overload" of Parsec's lexeme to handle split-joins and
-- sensitive whitespace (ie. no unexpected newlines)
lexeme :: Parser a -> Parser a
lexeme p = do 
    x <- Tok.lexeme lexer p 
    splitJoin
    return x

-- sensitiveWhitespace is Chai's custom whitespace handling function for
-- dealing with whitespace inside of meaningful statements
sensitiveWhitespace :: Parser ()
sensitiveWhitespace = many (oneOf ['\r', ' ', '\t', '\f', '\v']) $> ()

splitJoin :: Parser ()
splitJoin = optional $ string "\\\n"

multilineStringLit :: Parser String
multilineStringLit = char '`' *> many (noneOf ['`']) <* char '`'
