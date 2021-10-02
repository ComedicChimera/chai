{-# LANGUAGE FlexibleInstances, MultiParamTypeClasses #-}

module Syntax.Internal where

import Data.Functor.Identity

import Text.Parsec
import Text.Parsec.String

import Report.Message

-- -- ChaiLexer is like Parsec's TokenParser but customized a bit for Chai's needs
-- type ChaiLexer = GenTokenParser String () Identity 

-- -- ChaiParser is like Parsec's Parser customized for Chai's parser's default
-- -- behavior (eg. managing symbols during parsing)
-- type ChaiParser = Parsec String ()

-- positionOf captures a text position around a given token
positionOf :: Parser a -> Parser (a, TextPosition)
positionOf p = do
    startPos <- getPosition
    x <- p
    endPos <- getPosition
    return (x, TextPosition {
        textPosStartLine = sourceLine startPos,
        textPosStartCol = sourceColumn startPos,
        textPosEndLine = sourceLine endPos,
        textPosEndCol = sourceColumn endPos
    })

-- <||> acts a special combinator that is equivalent to a <|> try b
(<||>) :: Parser a -> Parser a -> Parser a
(<||>) p q = p <|> try q
