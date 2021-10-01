module Report.Message where

data TextPosition = TextPosition {
    textPosStartLine :: Int,
    textPosStartCol :: Int,
    textPosEndLine :: Int,
    textPosEndCol :: Int
}