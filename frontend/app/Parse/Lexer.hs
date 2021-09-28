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

-- TODO: multiline strings


