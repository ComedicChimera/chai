from collections import deque
from typing import Deque, Dict, Optional, Tuple, List
from dataclasses import dataclass, field
from enum import Flag, auto
from functools import reduce

from report import TextSpan
from report.reporter import CompileError
from depm import *
from depm.source import SourceFile
from syntax.ast import *
from syntax.token import Token
from . import *
from .solver import Solver
import util

class Control(Flag):
    NONE = 0
    LOOP = auto()
    FUNC = auto()

@dataclass
class Scope:
    func: Optional[FuncType]
    symbols: Dict[str, Symbol] = field(default_factory=dict)

@dataclass
class ControlContext:
    valid_control: Control
    context_divs: Deque[Control] = field(default_factory=deque)

    @property
    def curr_control(self) -> Control:
        return self.context_divs[0]

    @curr_control.setter
    def curr_control(self, control: Control):
        if not self.curr_control:
            self.context_divs[0] = control

    def new_div(self):
        self.context_divs.appendleft(Control.NONE)

    def merge(self) -> Control:
        return reduce(lambda a, b: a & b, self.context_divs)

class Walker:
    src_file: SourceFile
    solver: Solver
    scopes: Deque[Scope]
    control_contexts: Deque[ControlContext]

    def __init__(self, src_file: SourceFile):
        self.src_file = src_file
        self.solver = Solver(src_file)
        self.scopes = deque()
        self.control_contexts = deque()

    def walk_definition(self, defin: ASTNode):
        match defin:
            case FuncDef():
                self.walk_func_def(defin)

    # ---------------------------------------------------------------------------- #

    def walk_func_def(self, fd: FuncDef):
        expect_body = self.validate_func_annotations(fd.symbol.name, fd.annots)
        
        if fd.body:
            if not expect_body:
                self.error('function must not have body', fd.body.span)

            self.walk_func_body(fd)
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
        expects_body = self.validate_oper_annotations(od)

        if od.body:
            if not expects_body:
                self.error('operator must not have a body')

            self.walk_func_body(od)
        elif expects_body:
            self.error('operator must have a body')

    INTRINSIC_OPS = {
        'iadd', 'fadd', 'isub', 'fsub', 'imul', 'fmul', 'sdiv', 'udiv', 'fdiv',
        'smod', 'umod', 'fmod', 'slt', 'ult', 'flt', 'sgt' 'ugt', 'fgt', 'slteq', 
        'ulteq', 'flteq', 'sgteq', 'ugteq', 'fgteq', 'ieq', 'feq', 'ineq', 'fneq',
        'land', 'lor', 'lnot', 'ineg', 'fneg', 'band', 'bor', 'bxor', 'shl', 'ashr', 
        'lshr', 'compl'
    }

    def validate_oper_annotations(self, od: OperDef):
        expects_body = True

        for aname, (aval, aspan) in od.annots.items():
            match aname:
                case 'intrinsic':
                    if aval == '':
                        self.error(
                            '@intrinsic requires an argument when applied to an operator definition', 
                            aspan
                        )
                    
                    if aval not in self.INTRINSIC_OPS:
                        self.error(f'no intrinsic operator named `{aval}`', aspan)

                    od.overload.intrinsic_name = aval
                    expects_body = False

        return expects_body

    def walk_func_body(self, fd: FuncDef | OperDef):
        try:
            self.push_scope(fd.type)
            self.push_context(Control.FUNC)

            for param in fd.params:
                self.define_local(param)

            if fd.type.rt_type == PrimitiveType.NOTHING:
                self.walk_expr(fd.body, False, False)
            else:
                self.walk_expr(fd.body, True, False)

                if self.curr_control != Control.FUNC:
                    self.solver.assert_equiv(fd.type.rt_type, fd.body.type, fd.body.span)

            self.pop_context()
            self.pop_scope()

            self.solver.solve()
        finally:
            self.solver.reset()

    # ---------------------------------------------------------------------------- #

    def walk_if_tree(self, if_tree: IfTree, expects_value: bool, must_yield: bool):
        if_rt_type = None

        self.push_context(initial_div=False)  

        for cond_branch in if_tree.cond_branches:
            self.push_scope()
            self.add_context_div()

            if cond_branch.header_var:
                self.walk_var_decl(cond_branch.header_var)

            self.walk_expr(cond_branch.condition, True)
            self.solver.assert_equiv(cond_branch.condition.type, PrimitiveType.BOOL, cond_branch.condition.span)
                 
            self.walk_expr(cond_branch.body, expects_value, False)

            if expects_value and not self.curr_control:
                if if_rt_type:
                    self.solver.assert_equiv(cond_branch.body.type, if_rt_type, cond_branch.body.span)
                else:
                    if_rt_type = cond_branch.body.type

            self.pop_scope()

        if if_tree.else_branch:
            self.push_scope()
            self.add_context_div()

            self.walk_expr(if_tree.else_branch, expects_value, False)

            if expects_value and not self.curr_control:
                self.solver.assert_equiv(if_tree.else_branch.type, if_rt_type, if_tree.else_branch.span)

            self.pop_scope()
        elif expects_value:
            self.error('if tree is not exhaustive: include an else', if_tree.span)

        if if_rt_type:
            if_tree.rt_type = if_rt_type
        elif expects_value and must_yield:
            self.error('unconditional jump when value is expected', if_tree.span)

        self.pop_context()
            
    def walk_while_loop(self, while_loop: WhileLoop, expects_value: bool):
        if expects_value:
            raise NotImplementedError()

        self.push_scope()

        if while_loop.header_var:
            self.walk_var_decl(while_loop.header_var)

        self.walk_expr(while_loop.condition, True)
        self.solver.assert_equiv(while_loop.condition.type, PrimitiveType.BOOL, while_loop.condition.span)

        if while_loop.update_stmt:
            self.walk_stmt(while_loop.update_stmt, False, False)
        
        self.push_context(Control.LOOP)
        self.walk_expr(while_loop.body, False)
        self.pop_context()

        self.pop_scope()

    # ---------------------------------------------------------------------------- #

    def walk_stmt(self, stmt: ASTNode, expects_value: bool, must_yield: bool):
        match stmt:
            case VarDecl():
                self.walk_var_decl(stmt)
            case Assignment():
                self.walk_assignment(stmt)
            case IncDecStmt():
                self.walk_inc_dec_stmt(stmt)
            case KeywordStmt(keyword):
                if expects_value and must_yield:
                    self.error(f'unconditional jump when value is expected', keyword.span)

                match keyword.kind:
                    case Token.Kind.BREAK | Token.Kind.CONTINUE:
                        if Control.LOOP & self.valid_control:
                            self.curr_control = Control.LOOP
                        else:
                            self.error(f'{keyword.value} cannot be used outside a loop', keyword.span)
            case ReturnStmt(exprs):
                if expects_value and must_yield:
                    self.error(f'unconditional jump when value is expected', stmt.span)

                for expr in exprs:
                    self.walk_expr(expr, True)

                if Control.FUNC & self.valid_control:
                    self.curr_control = Control.FUNC
                else:
                    self.error(f'return cannot be used outside a function', stmt.span)

                match len(exprs):
                    case 0:
                        self.solver.assert_equiv(self.curr_scope.func.rt_type, PrimitiveType.NOTHING, stmt.span)
                    case 1:
                        self.solver.assert_equiv(self.curr_scope.func.rt_type, exprs[0].type, stmt.span)
                    case _:
                        raise NotImplementedError()
            case _:
                self.walk_expr(stmt, expects_value, must_yield)

    def walk_var_decl(self, vd: VarDecl):
        for var_list in vd.var_lists:
            if var_list.initializer:
                self.walk_expr(var_list.initializer, True)

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

    def walk_assignment(self, assign: Assignment):
        if len(assign.lhs_exprs) == len(assign.rhs_exprs):
            if len(assign.compound_ops):
                cpd_op_overloads = self.lookup_operator_overloads(assign.compound_ops[0].token, 2)
            else:
                cpd_op_overloads = None

            for i, (lhs, rhs) in enumerate(zip(assign.lhs_exprs, assign.rhs_exprs)):
                self.walk_expr(lhs, True)
                self.walk_expr(rhs, True)

                if lhs.category != ValueCategory.LVALUE:
                    self.error('cannot mutate an R-value', lhs.span)

                if not self.try_mark_mutable(lhs):
                    self.error('cannot mutate an immutable value', lhs.span)

                if cpd_op_overloads:
                    rhs_rt_type = self.check_oper_app(
                        assign.compound_ops[i], 
                        cpd_op_overloads, 
                        rhs.span, lhs, rhs
                    )
                    self.solver.assert_equiv(lhs.type, rhs_rt_type, rhs.span)
                else:
                    self.solver.assert_equiv(lhs.type, rhs.type, rhs.span)                    
        else:
            # TODO pattern matching
            raise NotImplementedError()

    def walk_inc_dec_stmt(self, incdec: IncDecStmt):
        self.walk_expr(incdec.lhs_operand, True)

        int_type_var = self.solver.new_type_var(incdec.op.token.span)
        self.solver.add_literal_overloads(int_type_var, self.INT_TYPES)

        overloads = self.lookup_operator_overloads(incdec.op.token, 2)

        rhs_rt_type = self.check_oper_app(
            incdec.op, 
            overloads, 
            incdec.span,
            incdec.lhs_operand,
            Literal(Token.Kind.INTLIT, '', incdec.op.token.span, int_type_var)
        )

        self.solver.assert_equiv(incdec.lhs_operand.type, rhs_rt_type, incdec.span)

    def try_mark_mutable(self, lhs_expr: ASTNode) -> bool:
        match lhs_expr:
            case Identifier(symbol=sym):
                if sym.mutability == Symbol.Mutability.IMMUTABLE:
                    return False
                else:
                    sym.mutability = Symbol.Mutability.MUTABLE
                    return True
            case Dereference(ptr):
                return not ptr.type.inner_type().const

        return False

    # ---------------------------------------------------------------------------- #

    def walk_expr(self, expr: ASTNode, expects_value: bool, must_yield: bool = True):
        match expr:
            case IfTree():
                self.walk_if_tree(expr, expects_value, must_yield)
            case WhileLoop():
                self.walk_while_loop(expr, expects_value)
            case Block(stmts):
                self.push_scope()

                for i, stmt in enumerate(stmts):
                    self.walk_stmt(stmt, expects_value and i == len(stmts) - 1, must_yield) 

                self.pop_scope()
            case TypeCast(src, dest_type, span):
                self.walk_expr(src, True)

                self.solver.assert_cast(src.type, dest_type, span)
            case BinaryOpApp(op, lhs, rhs):
                self.walk_expr(lhs, True)
                self.walk_expr(rhs, True)

                overloads = self.lookup_operator_overloads(op.token, 2)

                expr.rt_type = self.check_oper_app(op, overloads, expr.span, lhs, rhs)
            case UnaryOpApp(op, operand):
                self.walk_expr(operand, True)

                overloads = self.lookup_operator_overloads(op.token, 1)

                expr.rt_type = self.check_oper_app(op, overloads, expr.span, operand)
            case Indirect(elem, const):
                self.walk_expr(elem, True)

                expr.ptr_type = PointerType(elem.type, const or not self.try_mark_mutable(elem))
            case Dereference(ptr, span):
                self.walk_expr(ptr, True)

                elem_type_var = self.solver.new_type_var(span)
                self.solver.assert_equiv(ptr.type, PointerType(elem_type_var), span)         
                expr.elem_type = elem_type_var
            case FuncCall(func, args, span):
                self.walk_expr(func, True)

                for arg in args:
                    self.walk_expr(arg, True)

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

    def check_oper_app(
        self, 
        op: AppliedOperator, 
        overloads: List[OperatorOverload], 
        span: TextSpan,
        *operands: ASTNode, 
    ) -> Type:

        rt_type_var = self.solver.new_type_var(span)
        oper_func_type = FuncType([operand.type for operand in operands], rt_type_var)

        oper_var = self.solver.new_type_var(op.token.span, f'({op.token.value})')
        self.solver.add_operator_overloads(
            oper_var,
            op,
            overloads
        )

        self.solver.assert_equiv(oper_var, oper_func_type, span)
        return rt_type_var
    
    # ---------------------------------------------------------------------------- #

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
            self.scopes.appendleft(Scope(self.curr_scope.func))

    def pop_scope(self):
        self.scopes.popleft()

    # ---------------------------------------------------------------------------- #

    @property
    def curr_control(self) -> Control:
        return self.control_contexts[0].curr_control

    @curr_control.setter
    def curr_control(self, control: Control):
        self.control_contexts[0].curr_control = control

    @property
    def valid_control(self) -> Control:
        return self.control_contexts[0].valid_control

    def push_context(self, valid_control: Control = Control.NONE, initial_div: bool = True):
        if len(self.control_contexts) == 0:
            self.control_contexts.appendleft(ControlContext(valid_control))
        else:
            self.control_contexts.appendleft(ControlContext(self.valid_control | valid_control))

        if initial_div:
            self.control_contexts[0].new_div()

    def add_context_div(self):
        self.control_contexts[0].new_div()

    def pop_context(self):
        ctx = self.control_contexts.popleft()

        if len(self.control_contexts) > 0:
            self.curr_control = ctx.merge()

    # ---------------------------------------------------------------------------- #

    def error(self, msg: str, span: TextSpan):
        raise CompileError(msg, self.src_file, span)
