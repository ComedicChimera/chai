from collections import deque
from typing import Deque, Dict, Optional, Tuple
from dataclasses import dataclass, field

from report import CompileError, TextSpan
from depm import Symbol
from depm.source import SourceFile
from syntax.ast import *
from . import FuncType, PointerType
from .solver import Solver, TypeVariable

@dataclass
class Scope:
    func: Optional[FuncType]
    symbols: Dict[str, Symbol] = field(default_factory=dict)

class Walker:
    srcfile: SourceFile
    solver: Solver
    scopes: Deque[Scope]

    def __init__(self, srcfile: SourceFile):
        self.srcfile = srcfile
        self.solver = Solver(srcfile)
        self.scopes = deque()

    def walk_file(self):
        for defin in self.srcfile.definitions:
            self.walk_definition(defin)

            self.solver.solve()

    # ---------------------------------------------------------------------------- #

    def walk_definition(self, defin: ASTNode):
        match defin:
            case FuncDef():
                self.walk_func_def(defin)

    def walk_func_def(self, fd: FuncDef):
        # TODO validate annotations
        
        if fd.body:
            self.push_scope(fd.func_id.type)

            for param in fd.func_params:
                self.define_local(Symbol(
                    param.name,
                    self.srcfile.parent.id,
                    param.type,
                    Symbol.Kind.VALUE,
                    Symbol.Mutability.NEVER_MUTATED,
                    None,  # never used
                ))

            self.walk_expr(fd.body)

            self.solver.assert_equiv(fd.func_id.type.rt_type, fd.body.type, fd.body.span)

            for param in fd.func_params:
                if self.curr_scope.symbols[param.name].mutability == Symbol.Mutability.MUTABLE:
                    param.mutated = True

            self.pop_scope()

    # ---------------------------------------------------------------------------- #

    def walk_stmt(self, stmt: ASTNode):
        match stmt:
            case VarDecl():
                self.walk_var_decl(stmt)

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
                self.solver.assert_cast(src.type, dest_type, span)
            case Dereference(ptr, span):
                elem_type_var = self.solver.new_type_var(span)
                self.solver.assert_equiv(ptr.type, PointerType(elem_type_var), span)
                expr.elem_type = elem_type_var
            case FuncCall(func, args, span):
                rt_type_var = self.solver.new_type_var(span)
                func_type = FuncType([arg.type for arg in args], rt_type_var)
                self.solver.assert_equiv(func.type, func_type, span)
                expr.rt_type = rt_type_var
            case Identifier(name, span):
                expr.symbol, expr.local = self.lookup(name, span)
            case Null():
                expr.type = self.solver.new_type_var(expr.span)
            case Literal():
                self.walk_literal(expr)
                
    def walk_literal(self, lit: Literal):
        pass

    # ---------------------------------------------------------------------------- #

    def lookup(self, name: str, span: TextSpan) -> Tuple[Symbol, bool]:
        for scope in self.scopes:
            if sym := scope.symbols.get(name):
                return sym, True
        
        if sym := self.srcfile.parent.symbol_table.get(name):
            return sym, False

        self.error(f'undefined symbol: `{name}`')

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
        raise CompileError(msg, self.srcfile.rel_path, span)
