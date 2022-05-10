from dataclasses import dataclass
from enum import Enum, auto

from report import Position

@dataclass
class Token:
    class Kind(Enum):
        Newline = auto()

        Def = auto()
        Let = auto()
        If = auto()

        Plus = auto()
        Increment = auto()

        StringLit = auto()
        RuneLit = auto()
        NumLit = auto()
        FloatLit = auto()
        IntLit = auto()
        Identifier = auto()

        EndOfFile = auto()

    kind: Kind
    value: str
    pos: Position
