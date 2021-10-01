module Syntax.Parser (parseFile) where

import qualified Syntax.AST as AST

parseFile :: String -> IO [AST.Definition]
parseFile fpath = return []