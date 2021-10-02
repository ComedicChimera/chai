module Syntax.Parser (parseFile) where

import Data.Functor

import Text.Parsec
import Text.Parsec.String
import qualified Text.Parsec.Token as Tok

import Syntax.Internal
import qualified Syntax.AST as AST
import qualified Syntax.Lexer as Lex

import Report.Message
import Semantic.Types

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
file :: Parser [AST.Definition]
file = Tok.whiteSpace Lex.lexer *> many topDef

-- topDef is a top level definition
topDef :: Parser AST.Definition
topDef = funcDef

-- funcDef is a function definition
funcDef :: Parser AST.Definition
funcDef = do
    Lex.keyword "def"
    (name, pos) <- Lex.identifier
    args <- between (Lex.lexeme $ char '(') (Lex.lexeme $ char ')') funcArg
    return AST.Func {}

-- funcArg is a function argument
funcArg :: Parser AST.FuncArg
funcArg = pure AST.FuncArg {}

-- typeLabel represents a Chai type label
typeLabel :: Parser Type
typeLabel = pure Undetermined

-- primTypeLabel represents a primitive type label
primTypeLabel :: Parser PrimType
primTypeLabel = try (Lex.keyword "u8" $> U8)
    <||> (Lex.keyword "u16" $> U16)
    <||> (Lex.keyword "u32" $> U32)
    <||> (Lex.keyword "u64" $> U64)
    <||> (Lex.keyword "i8" $> I8)
    <||> (Lex.keyword "i16" $> I16)
    <||> (Lex.keyword "i32" $> I32)
    <||> (Lex.keyword "i64" $> I64)
    <||> (Lex.keyword "f32" $> F32)
    <||> (Lex.keyword "f64" $> F64)
    <||> (Lex.keyword "nothing" $> NothingType)
    <||> (Lex.keyword "bool" $> Bool)
    <|> (Lex.keyword "rune" $> Rune)
    