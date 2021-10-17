from dataclasses import dataclass
from typing import List

from chai.symbol import Symbol

# FuncDef represents a function definition
@dataclass
class FuncDef:
    sym: Symbol
    # TODO: body


# ASTDef is union of all AST definitions
ASTDef = FuncDef