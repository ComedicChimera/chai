module Syntax.Lexer where

import Data.Functor
import Data.Maybe

import Text.Parsec
import Text.Parsec.String (Parser)
import Text.Parsec.Language (emptyDef)
import qualified Text.Parsec.Token as Tok 

import Syntax.Internal
import qualified Syntax.AST as AST

import Semantic.Types

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

-- isNext returns whether or not a given token string occurred next.  The token
-- is read if so.  It is useful for cases where the token value is insignificant
-- but its presence is determinant.
isNext :: String -> Parser Bool
isNext tokValue = do
    maybe <- optionMaybe $ string tokValue
    return $ isJust maybe
       
-- identifier reads in a new identifier
identifier :: Parser AST.Identifier
identifier = do 
    (name, pos) <- positionOf $ Tok.identifier lexer
    sensitiveWhitespace
    return $ AST.Identifier { AST.idName = name, AST.idPos = pos, AST.idType = Undetermined}

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

-- el runs a parser and accepts any kind of whitespace after it (including
-- newlines).  It is basically a wrapper around `Tok.lexeme`
el :: Parser a -> Parser a
el = Tok.lexeme lexer

-- splitJoin parses a split join parser
splitJoin :: Parser ()
splitJoin = optional $ string "\\\n"

-- multilineStringLit parses a multiline/raw string literal
multilineStringLit :: Parser String
multilineStringLit = char '`' *> many (noneOf ['`']) <* char '`'
