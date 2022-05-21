from collections import deque
from typing import Deque, Dict, Optional, Tuple, List
from dataclasses import dataclass, field

from report import TextSpan
from report.reporter import CompileError
from depm import *
from depm.source import SourceFile
from syntax.ast import *
from syntax.token import Token
from . import *
from .solver import Solver
import util

@dataclass
class Scope:
    func: Optional[FuncType]
    symbols: Dict[str, Symbol] = field(default_factory=dict)

class Walker:
    src_file: SourceFile
    solver: Solver
    scopes: Deque[Scope]

    def __init__(self, src_file: SourceFile):
        self.src_file = src_file
        self.solver = Solver(src_file)
        self.scopes = deque()

    def walk_definition(self, defin: ASTNode):
        try:
            match defin:
                case FuncDef():
                    self.walk_func_def(defin)

            self.solver.solve()
        finally:
            self.solver.reset()

    # ---------------------------------------------------------------------------- #

    def walk_func_def(self, fd: FuncDef):
        expect_body = self.validate_func_annotations(fd.symbol.name, fd.annots)
        
        if fd.body:
            if not expect_body:
                self.error('function must not have body', fd.body.span)

            self.push_scope(fd.type)

            for param in fd.params:
                self.define_local(param)

            self.walk_expr(fd.body)

            if not fd.type.rt_type == PrimitiveType.NOTHING:
                self.solver.assert_equiv(fd.type.rt_type, fd.body.type, fd.body.span)

            self.pop_scope()
        elif expect_body:
            self.error('function must have body', fd.span)

    INTRINSIC_FUNCS = set()

    def validate_func_annotations(self, func_name: str, annots: Annotations) -> bool:
        expect_body = True

        for aname, (aval, aspan) in annots.items():
            match aname:
                case 'extern':
                    if aval != '':
                        self.error('@extern does not take an argument', aspan)

                    expect_body = False
                case 'intrinsic':
                    if aval != '':
                        self.error(
                            '@intrinsic does not take an argument when applied to a function definition', 
                            aspan
                        )

                    if func_name not in self.INTRINSIC_FUNCS:
                        self.error(f'no intrinsic function named `{func_name}`', aspan)

                    expect_body = False
                case 'entry':
                    if aval != '':
                        self.error('@entry does not take an argument', aspan)

        return expect_body

    def walk_oper_def(self, od: OperDef):
        expects_body = self.validate_oper_annotations(od.annots)

        if od.body:
            if not expects_body:
                self.error('operator must not have a body')

            self.push_scope(od.overload.signature)

            for param in od.params:
                self.define_local(param)

            self.walk_expr(od.body)

            if not od.type.rt_type == PrimitiveType.NOTHING:
                self.solver.assert_equiv(od.type.rt_type, od.body.type, od.body.span)

            self.pop_scope()
        elif expects_body:
            self.error('operator must have a body')

    INTRINSIC_OPS = {
        'iadd', 'fadd', 'isub', 'fsub', 'imul', 'fmul', 'sdiv', 'udiv', 'fdiv',
        'smod', 'umod', 'fmod', 'lt', 'gt', 'lteq', 'gteq', 'land', 'lor', 'lnot',
        'ineg', 'fned', 'band', 'bor', 'bxor', 'eq', 'neq', 'shl', 'shr', 'compl'
    }

    def validate_oper_annotations(self, annots: Annotations):
        expects_body = True

        for aname, (aval, aspan) in annots.items():
            match aname:
                case 'intrinsic':
                    if aval == '':
                        self.error(
                            '@intrinsic requires an argument when applied to an operator definition', 
                            aspan
                        )
                    
                    if aval not in self.INTRINSIC_OPS:
                        self.error(f'no intrinsic operator named `{aval}`', aspan)

                    expects_body = False

        return expects_body


    # ---------------------------------------------------------------------------- #

    def walk_stmt(self, stmt: ASTNode):
        match stmt:
            case VarDecl():
                self.walk_var_decl(stmt)
            case _:
                self.walk_expr(stmt)

    def walk_var_decl(self, vd: VarDecl):
        for var_list in vd.var_lists:
            if var_list.initializer:
                self.walk_expr(var_list.initializer)

                assert len(var_list.symbols) == 1, 'tuple unpacking not implemented yet'

                for sym in var_list.symbols:
                    # TODO tuple unpacking
                    if sym.type:
                        self.solver.assert_equiv(
                            sym.type, 
                            var_list.initializer.type, 
                            var_list.initializer.span,
                        )
                    else:
                        sym.type = var_list.initializer.type

            for sym in var_list.symbols:
                self.define_local(sym)

    # ---------------------------------------------------------------------------- #

    def walk_expr(self, expr: ASTNode):
        match expr:
            case Block(stmts):
                self.push_scope()

                for stmt in stmts:
                    self.walk_stmt(stmt)

                self.pop_scope()
            case TypeCast(src, dest_type, span):
                self.walk_expr(src)
                self.solver.assert_cast(src.type, dest_type, span)
            case BinaryOpApp(op_tok, _, lhs, rhs):
                self.walk_expr(lhs)
                self.walk_expr(rhs)

                rt_type_var = self.solver.new_type_var(expr.span)
                oper_func_type = FuncType([lhs.type, rhs.type], rt_type_var)

                oper_var = self.solver.new_type_var(op_tok.span, op_tok.value)
                self.solver.add_operator_overloads(
                    oper_var, 
                    expr,
                    self.lookup_operator_overloads(op_tok, 2)
                )
                
                self.solver.assert_equiv(oper_var, oper_func_type, expr.span)
            case UnaryOpApp(op_tok, _, operand):
                self.walk_expr(operand)

                rt_type_var = self.solver.new_type_var(expr.span)
                oper_func_type = FuncType([operand.type], rt_type_var)

                oper_var = self.solver.new_type_var(op_tok.span, op_tok.value)
                self.solver.add_operator_overloads(
                    oper_var,
                    expr,
                    self.lookup_operator_overloads(op_tok, 1)
                )

                self.solver.assert_equiv(oper_var, oper_func_type, expr.span)
            case Indirect(elem, _):
                self.walk_expr(elem)
            case Dereference(ptr, span):
                self.walk_expr(ptr)
                elem_type_var = self.solver.new_type_var(span)
                self.solver.assert_equiv(ptr.type, PointerType(elem_type_var), span)
                expr.elem_type = elem_type_var
            case FuncCall(func, args, span):
                self.walk_expr(func)

                for arg in args:
                    self.walk_expr(arg)

                rt_type_var = self.solver.new_type_var(span)
                func_type = FuncType([arg.type for arg in args], rt_type_var)
                self.solver.assert_equiv(func.type, func_type, span)
                expr.rt_type = rt_type_var
            case Identifier(name, span):
                expr.symbol, expr.local = self.lookup_symbol(name, span)
            case Null():
                expr.type = self.solver.new_type_var(expr.span)
            case Literal():
                self.walk_literal(expr)
    
    INT_TYPES = [
        PrimitiveType.I64,
        PrimitiveType.I32,
        PrimitiveType.I16,
        PrimitiveType.I8,
        PrimitiveType.U64,
        PrimitiveType.U32,
        PrimitiveType.U16,
        PrimitiveType.U8,
    ]

    def walk_literal(self, lit: Literal):
        def prune_int_types_by_size(types: List[PrimitiveType], value: str, base: int) -> List[PrimitiveType]:
            bit_size = int(value, base=base).bit_length()

            return [typ for typ in types if typ.value >= bit_size]

        match lit.kind:
            case Token.Kind.FLOATLIT:
                float_var = self.solver.new_type_var(lit.span)
                self.solver.add_literal_overloads(float_var, [PrimitiveType.F64, PrimitiveType.F32])

                lit.type = float_var
            case Token.Kind.NUMLIT:
                num_var = self.solver.new_type_var(lit.span)

                num_types = prune_int_types_by_size(self.INT_TYPES, lit.value, 10)
                num_types += [PrimitiveType.F64, PrimitiveType.F32]

                self.solver.add_literal_overloads(num_var, num_types)

                lit.type = num_var
            case Token.Kind.INTLIT:
                trimmed_value, base, uns, lng = util.trim_int_lit(lit.value)

                int_types = self.INT_TYPES

                if uns:
                    int_types = [typ for typ in int_types if typ.name.startswith('U')]
                
                if lng:
                    int_types = [typ for typ in int_types if typ.name.endswith('64')]

                int_types = prune_int_types_by_size(int_types, trimmed_value, base)

                if len(int_types) == 0:
                    self.error('value too large to fit in an integral type', lit.span)
                elif len(int_types) == 1:
                    lit.type = int_types[0]
                else:
                    int_var = self.solver.new_type_var(lit.span)
                    self.solver.add_literal_overloads(int_var, int_types)

                    lit.type = int_var    

    # ---------------------------------------------------------------------------- #

    def lookup_operator_overloads(self, op_token: Token, arity: int) -> List[OperatorOverload]:
        if op := self.src_file.lookup_operator(op_token.kind, arity):
            return op.overloads

        self.error(f'no visible definitions for operator `{op_token.value}`', op_token.span)

    def lookup_symbol(self, name: str, span: TextSpan) -> Tuple[Symbol, bool]:
        for scope in self.scopes:
            if sym := scope.symbols.get(name):
                return sym, True
        
        if sym := self.src_file.lookup_symbol(name):
            return sym, False

        self.error(f'undefined symbol: `{name}`', span)

    def define_local(self, sym: Symbol):
        if sym.name in self.curr_scope.symbols:
            self.error(f'multiple symbols named `{sym.name}` defined in scope', sym.def_span)
        
        self.curr_scope.symbols[sym.name] = sym

    @property
    def curr_scope(self) -> Scope:
        return self.scopes[0]

    def push_scope(self, func: Optional[FuncType] = None):
        if func or len(self.scopes) == 0:
            self.scopes.appendleft(Scope(func))
        else:
            self.scopes.appendleft(Scope(self.scopes[0].func))

    def pop_scope(self):
        self.scopes.popleft()

    def error(self, msg: str, span: TextSpan):
        raise CompileError(msg, self.src_file, span)
