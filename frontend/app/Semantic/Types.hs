module Semantic.Types where

data Type = Undetermined
    | Primitive PrimType

data PrimType = I8
    | I16
    | I32
    | I64
    | U8
    | U16
    | U32
    | U64
    | F32
    | F64
    | BOOL
    | RUNE