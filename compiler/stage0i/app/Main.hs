module Main where

import qualified Syntax.Parser as Parser

main :: IO ()
main = do
    fpath <- getLine
    let fullPath = "../../tests/" ++ fpath
    result <- Parser.parseFile fullPath
    putStrLn $ case result of
        Left err -> show err
        Right ast -> show ast
