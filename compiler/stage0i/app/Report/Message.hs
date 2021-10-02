module Report.Message where

-- TextPosition is a phase-agnostic way of representing the position of a given
-- error or warning within user source text
data TextPosition = TextPosition {
    textPosStartLine :: Int,
    textPosStartCol :: Int,
    textPosEndLine :: Int,
    textPosEndCol :: Int
}

-- CompileMessage represents a message generating during compilation about some
-- part of user source
data CompileMessage = CompileMessage {
    -- msgContent is the actual message to be displayed
    msgContent :: String,

    -- msgSrcFilePath is the absolute path to the source file in question
    msgSrcFilePath :: String,

    -- msgRelPath is a function for converting the absolute path into
    -- a more meaningful relative path.  Generally, this arises from a 
    -- partial application of method function of a module
    msgRelPath :: String -> String,

    -- msgPos is the text position of the message within the given source file
    msgPos :: TextPosition
}