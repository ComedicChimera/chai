module Syntax.AST where

import Report.Message (TextPosition)
import Semantic.Types (Type)

data Identifier = Identifier {
    idName :: String,
    idPos :: TextPosition,
    idType :: Type
}

data Definition = Func { funcName :: Identifier, funcArgs :: [FuncArg], funcReturnType :: Type}

data FuncArg = FuncArg { argName :: Identifier, argType :: Type, argFlags :: [FuncArgFlag]}
data FuncArgFlag = IsOptional | IsIndefinite | IsByRef