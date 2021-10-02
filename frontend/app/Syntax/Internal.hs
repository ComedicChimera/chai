{-# LANGUAGE FlexibleInstances, MultiParamTypeClasses #-}

module Syntax.Internal where

import System.IO
import Data.Functor.Identity

import Text.Parsec
import Text.Parsec.Token


-- -- Create an instance of the Handle type for our parser so it can read chars
-- -- from an input stream as opposed to having to read in the whole text file and
-- -- then parse it.  The `parseFile` function provided by Parser ensures that the
-- -- file and stream is handled properly without ever having to store a huge file
-- -- string in memory
-- instance Stream Handle IO Char where
--     uncons h = do
--         isEOF <- hIsEOF h
--         if isEOF then
--             return Nothing
--         else do
--             c <- hGetChar h
--             return $ Just (c, h)

-- ChaiLexer is like Parsec's TokenParser but customized a bit for Chai's needs
type ChaiLexer = GenTokenParser String () Identity 

-- ChaiParser is like Parsec's Parser customized for Chai's parser's default
-- behavior (eg. managing symbols during parsing)
type ChaiParser = Parsec String ()
