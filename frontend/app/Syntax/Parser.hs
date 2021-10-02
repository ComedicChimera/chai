module Syntax.Parser (parseFile) where

import System.IO

import Text.Parsec
import qualified Text.Parsec.Token as Tok

import Report.Message
import Syntax.Internal
import qualified Syntax.AST as AST
import qualified Syntax.Lexer as Lex

-- parseFile runs Chai's parser on a file at the provided file path
parseFile :: String -> IO (Either CompileMessage [AST.Definition])
parseFile fpath = do 
    sourceText <- readFile fpath
    let result = runParser file () fpath sourceText
    return $ case result of
        Right defs -> Right defs
        Left err -> Left CompileMessage { 
            -- TODO: figure out how to get the message content from the parse error
            msgContent = "",
            msgPos = let pos = errorPos err in TextPosition { 
                textPosStartLine = sourceLine pos,
                textPosStartCol = sourceColumn pos,
                textPosEndLine = sourceLine pos,
                textPosEndCol = sourceColumn pos + 1
            },
            msgSrcFilePath = fpath,
            msgRelPath = id
        }

-- file is start symbol for the Chai grammar
file :: ChaiParser [AST.Definition]
file = Tok.whiteSpace Lex.lexer *> many topDef

-- topDef is a top level definition
topDef :: ChaiParser AST.Definition
topDef = funcDef

-- funcDef is a function definition
funcDef :: ChaiParser AST.Definition
funcDef = do
    Lex.keyword "def"
    -- TODO: figure out position computation
    return AST.Func {}

    