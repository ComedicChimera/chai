from typing import Dict, Optional, Tuple, List
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

@dataclass
class Scope:
    '''
    Represents a semantic scope in user code.

    Attributes
    ----------
    rt_type: Optional[Type]
        The return type within the given scope.
    symbols: Dict[str, Symbol]
        The map of symbols that are defined within this lexical scope.
    '''

    rt_type: Optional[Type]
    symbols: Dict[str, Symbol] = field(default_factory=dict)

class ControlAction(Flag):
    '''
    Used to represent the control flow actions that can terminate a control
    frame division.  This is a flag primarily to allow for efficient enumeration
    of the kinds of control flow that are valid within a control frame.  It also
    allows us to AND control actions together to determine the "net" control of
    an entire frame.
    '''

    NONE = 0
    LOOP = auto()
    FUNC = auto()

@dataclass
class ControlFrame:
    '''
    Represents the control flow structure of a function: each frame is the local
    control flow context of a given logical block structure.  Frames are
    comprised of divisions (divs) which contain component control flow actions
    of the given logical block structure: eg. each branch of an if tree is a
    single division within the control frame representing the if tree as a
    whole.  This system is used to check exhaustivity and perform block-type
    checking intuitively.

    Attributes
    ----------
    valid_actions: ControlAction
        A flag representing the valid control actions within a this control
        frame.  This flag can be comprised of multiple possible control actions:
        you can check if an action is valid by ANDing it with this field.
    divs: List[ControlAction]
        The list of frame divisions.
    div_action: ControlAction
        The action of the current division.  Note that this is actually a
        property with a special setter that updates the division action only if
        that division does not already have a meaningful action assigned to it:
        since a block can only be terminated by one kind of control flow.
    '''

    valid_actions: ControlAction
    divs: List[ControlAction] = field(default_factory=list)

    @property
    def div_action(self) -> ControlAction:
        return self.divs[-1]

    @div_action.setter
    def div_action(self, control: ControlAction):
        if not self.div_action:
            self.divs[-1] = control

    def new_div(self):
        '''Adds a new division to the frame.'''

        self.divs.append(ControlAction.NONE)

    def merge(self) -> ControlAction:
        '''Returns the net control action of the frame.'''
        
        return reduce(lambda a, b: a & b, self.divs)

    def is_valid(self, action: ControlAction) -> bool:
        '''
        Checks whether the given control action is valid within the frame.

        Params
        ------
        action: ControlAction
            The action to check.
        
        '''

        return bool(self.valid_actions & action)

# ---------------------------------------------------------------------------- #

class Walker:
    '''
    Responsible for performing semantic analysis by walking Chai's AST.

    Methods
    -------
    walk_definition(defin: ASTNode)
    '''

    # The source file of the AST nodes being walked.
    src_file: SourceFile

    # THe global type solver used for type checking.
    solver: Solver

    # The semantic scope stack: this should not be accessed directly within
    # walking methods, but rather using the helper methods defined at the bottom
    # of Walker.
    scopes: List[Scope]

    # The control frame stack: this should not be accessed directly within
    # walking methods, but rather using the helper methods defined at the bottom
    # of Walker.
    frames: List[ControlFrame]

    def __init__(self, src_file: SourceFile):
        '''
        Params
        ------
        src_file: SourceFile
            The source file of the AST nodes being walked.
        '''

        self.src_file = src_file
        self.solver = Solver(src_file)
        self.scopes = []
        self.frames = []

    def walk_definition(self, defin: ASTNode):
        '''
        Walks a definition AST node.

        Params
        ------
        defin: ASTNode
            The definition AST node to walk.
        '''

        match defin:
            case FuncDef():
                self.walk_func_def(defin)
            case OperDef():
                self.walk_oper_def(defin)
            case _:
                raise NotImplementedError()

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
                case 'abientry':
                    if aval != '':
                        self.error('@abientry does not take an argument', aspan)

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
        'smod', 'umod', 'fmod', 'slt', 'ult', 'flt', 'sgt', 'ugt', 'fgt', 'slteq', 
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
            self.push_scope(fd.type.rt_type)
            self.push_frame(ControlAction.FUNC)

            for param in fd.params:
                self.define_local(param)

            if fd.type.rt_type == PrimitiveType.UNIT:
                self.walk_expr(fd.body, False, False)
            else:
                self.walk_expr(fd.body, True, False)

                if self.div_action != ControlAction.FUNC:
                    self.solver.assert_equiv(fd.type.rt_type, fd.body.type, fd.body.span)

            self.pop_frame()
            self.pop_scope()

            self.solver.solve()
        finally:
            self.solver.reset()

    # ---------------------------------------------------------------------------- #

    def walk_if_tree(self, if_tree: IfTree, expects_value: bool, must_yield: bool):
        if_rt_type = None

        self.push_frame(initial_div=False)  

        for cond_branch in if_tree.cond_branches:
            self.push_scope()
            self.curr_frame.new_div()

            if cond_branch.header_var:
                self.walk_var_decl(cond_branch.header_var)

            self.walk_expr(cond_branch.condition, True)
            self.solver.assert_equiv(cond_branch.condition.type, PrimitiveType.BOOL, cond_branch.condition.span)
                 
            self.walk_expr(cond_branch.body, expects_value, False)

            if expects_value and not self.div_action:
                if if_rt_type:
                    self.solver.assert_equiv(cond_branch.body.type, if_rt_type, cond_branch.body.span)
                else:
                    if_rt_type = cond_branch.body.type

            self.pop_scope()

        if if_tree.else_branch:
            self.push_scope()
            self.curr_frame.new_div()

            self.walk_expr(if_tree.else_branch, expects_value, False)

            if expects_value and not self.div_action:
                self.solver.assert_equiv(if_tree.else_branch.type, if_rt_type, if_tree.else_branch.span)

            self.pop_scope()
        elif expects_value:
            self.error('if tree is not exhaustive: include an else', if_tree.span)

        if if_rt_type:
            if_tree.rt_type = if_rt_type
        elif expects_value and must_yield:
            self.error('unconditional jump when value is expected', if_tree.span)

        self.pop_frame()
            
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
        
        self.push_frame(ControlAction.LOOP)
        self.walk_expr(while_loop.body, False)
        self.pop_frame()

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
                        if self.curr_frame.is_valid(ControlAction.LOOP):
                            self.div_action = ControlAction.LOOP
                        else:
                            self.error(f'{keyword.value} cannot be used outside a loop', keyword.span)
            case ReturnStmt(exprs):
                if expects_value and must_yield:
                    self.error(f'unconditional jump when value is expected', stmt.span)

                for expr in exprs:
                    self.walk_expr(expr, True)

                if self.curr_frame.is_valid(ControlAction.FUNC):
                    self.div_action = ControlAction.FUNC
                else:
                    self.error(f'return cannot be used outside a function', stmt.span)

                match len(exprs):
                    case 0:
                        self.solver.assert_equiv(self.curr_scope.rt_type, PrimitiveType.UNIT, stmt.span)
                    case 1:
                        self.solver.assert_equiv(self.curr_scope.rt_type, exprs[0].type, stmt.span)
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

        if not self.try_mark_mutable(incdec.lhs_operand):
            self.error('cannot mutate an immutable value', incdec.lhs_operand)

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
        return self.scopes[-1]

    def push_scope(self, rt_type: Optional[Type] = None):
        if rt_type or len(self.scopes) == 0:
            self.scopes.append(Scope(rt_type))
        else:
            self.scopes.append(Scope(self.curr_scope.rt_type))

    def pop_scope(self):
        self.scopes.pop()

    # ---------------------------------------------------------------------------- #

    @property
    def curr_frame(self) -> ControlFrame:
        return self.frames[-1]

    def push_frame(self, valid_control: ControlAction = ControlAction.NONE, initial_div: bool = True):
        if len(self.frames) == 0:
            self.frames.append(ControlFrame(valid_control))
        else:
            self.frames.append(ControlFrame(self.curr_frame.valid_actions | valid_control))

        if initial_div:
            self.curr_frame.new_div()

    def pop_frame(self):
        frame = self.frames.pop()

        if len(self.frames) > 0:
            self.div_action = frame.merge()

    @property
    def div_action(self) -> ControlAction:
        return self.curr_frame.div_action

    @div_action.setter
    def div_action(self, action: ControlAction):
        self.curr_frame.div_action = action

    # ---------------------------------------------------------------------------- #

    def error(self, msg: str, span: TextSpan):
        raise CompileError(msg, self.src_file, span)
