from dataclasses import dataclass
from typing import List
from enum import Enum, auto

from chai.types import DataType
from chai.symbol import Symbol

# FuncDef represents a function definition
@dataclass
class FuncDef:
    sym: Symbol
    # TODO: body

# ASTDef is union of all AST definitions
ASTDef = FuncDef

# ---------------------------------------------------------------------------- #

# ValueCategory is an enumeration of the different kinds of value in Chai
class ValueCategory(Enum):
    # LValue has a well-defined place in memory: eg. a variable
    LValue = auto() 

    # RValue does not have a well-defined place in memory: eg. an integer
    # constant
    RValue = auto()

# ASTExpr is the base class for AST expressions
@dataclass
class ASTExpr:
    typ: DataType
    category: ValueCategory

# TODO: dataclass inheritance :(

# ASTIdent is an AST identifier
@dataclass
class ASTIdent(ASTExpr):
    sym: Symbol

    def __post_init__(self, sym: Symbol) -> None:
        super().__init__(sym.typ, ValueCategory.LValue)
        self.sym = sym

# ASTLit is an AST literal
@dataclass
class ASTLit(ASTExpr):
    value: str

    def __post_init__(self, typ: DataType, value: str) -> None:
        super().__init__(typ, ValueCategory.RValue)
        self.value = value