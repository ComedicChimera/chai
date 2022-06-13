'''Provides Chai's implementation of lexical tokens.'''

from dataclasses import dataclass
from enum import Enum, auto

from report import TextSpan

@dataclass
class Token:
    '''
    Represents a single lexical token.

    Attributes
    ----------
    kind: Token.Kind
        The lexical kind of the token.
    value: str
        The string value stored by the token.
    span: TextSpan
        The positional span over which the token extends.
    '''

    class Kind(Enum):
        '''Enumerates the different kinds of tokens.'''

        NEWLINE = auto()

        PACKAGE = auto()
        DEF = auto()
        OPER = auto()

        LET = auto()
        CONST = auto()
        
        IF = auto()
        ELIF = auto()
        ELSE = auto()
        WHILE = auto()
        END = auto()

        BREAK = auto()
        CONTINUE = auto()
        RETURN = auto()

        AS = auto()

        BOOL = auto()
        I8 = auto()
        U8 = auto()
        I16 = auto()
        U16 = auto()
        I32 = auto()
        U32 = auto()
        I64 = auto()
        U64 = auto()
        F32 = auto()
        F64 = auto()
        UNIT = auto()

        PLUS = auto()
        MINUS = auto()
        STAR = auto()
        DIV = auto()
        MOD = auto()
        POWER = auto()

        LT = auto()
        LTEQ = auto()
        GT = auto()
        GTEQ = auto()
        EQ = auto()
        NEQ = auto()

        LSHIFT = auto()
        RSHIFT = auto()
        COMPL = auto()

        AMP = auto()
        PIPE = auto()
        CARRET = auto()

        AND = auto()
        OR = auto()
        NOT = auto()

        ASSIGN = auto()
        INCREMENT = auto()
        DECREMENT = auto()

        LPAREN = auto()
        RPAREN = auto()
        LBRACE = auto()
        RBRACE = auto()
        LBRACKET = auto()
        RBRACKET = auto()
        COMMA = auto()
        COLON = auto()
        SEMICOLON = auto()
        DOT = auto()
        ARROW = auto()
        ATSIGN = auto()

        STRINGLIT = auto()
        RUNELIT = auto()
        NUMLIT = auto()
        FLOATLIT = auto()
        INTLIT = auto()
        BOOLLIT = auto()
        IDENTIFIER = auto()
        NULL = auto()

        EOF = auto()

    kind: Kind
    value: str
    span: TextSpan
