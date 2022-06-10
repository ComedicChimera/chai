'''The module responsible for converting Chai predicates to LLVM.'''

__all__ = [
    'BodyPredicate',
    'PredicateGenerator'
]

import contextlib
from typing import List, Optional, Deque 
from dataclasses import dataclass
from collections import deque

import llvm.ir as ir
import llvm.value as llvalue
import llvm.types as lltypes
import util

from report import TextSpan
from llvm.builder import IRBuilder
from syntax.ast import *
from syntax.token import Token
from depm import Symbol
from typecheck import *

from .type_util import conv_type, is_nothing
from .debug_info import DebugInfoEmitter

@dataclass
class BodyPredicate:
    '''
    Represents an function or operator body that needs to be lowered.

    Attributes
    ----------
    ll_func: ir.Function
        The LLVM function this predicate is the body of.
    func_params: List[Symbol]
        The parameters to that function.
    expr: ASTNode
        The AST expression representing the body.
    '''

    ll_func: ir.Function
    func_params: List[Symbol]
    expr: ASTNode

@dataclass
class LLVMValueNode(ASTNode):
    '''
    An psuedo AST node that is used to "reinsert" LLVM values into the AST so
    that generation code can be easily reused for "desugaring" code.

    Attributes
    ----------
    ll_value: llvalue.Value
        The LLVM value stored by the node.
    '''

    ll_value: llvalue.Value

    _type: Type
    _span: TextSpan

    @property
    def type(self) -> Type:
        return self._type

    @property
    def span(self) -> TextSpan:
        return self._span

@dataclass
class LoopContext:
    '''
    Stores the basic blocks to jump to when loop control flow is encountered.

    Attributes
    ----------
    break_block: ir.BasicBlock
        The block to jump to when a break is encountered.
    continue_block: ir.BasicBlock
        The block to jump to when a continue is encountered.
    '''

    break_block: ir.BasicBlock
    continue_block: ir.BasicBlock

# ---------------------------------------------------------------------------- #

class PredicateGenerator:
    '''
    Responsible for converting predicates (ie. executable expressions such as
    function bodies and initializers) into LLVM IR.

    Methods
    -------
    add_predicate(pred: BodyPredicate)
    generate()
    '''

    # The LLVM IR builder.
    irb: IRBuilder

    # The LLVM debug info emitter: may be None if no info is to be emitted.
    die: Optional[DebugInfoEmitter]

    # The block at the start of the predicate used to store the actual variable
    # allocations: the compiler automatically factors all variable allocations
    # to a special block at the start of the predicate to ensure memory is used
    # efficiently.  NOTE There may be an LLVM optimization pass that does this
    # automatically, but this is fairly easy to implement and possibly saves
    # the optimizer some time.
    var_block: ir.BasicBlock

    # The function body to append basic blocks to.
    body: ir.FuncBody

    # The stack of look contexts.  This should be accessed through the helper
    # methods defined at the bottom of the PredicateGenerator.
    loop_contexts: Deque[LoopContext]

    # A reusable LLVM value used whenever `()` must be treated as an actual
    # value: ie. cannot be pruned from source code.
    nothing_value: llvalue.Value

    # The list of body predicates to generate.
    body_preds: List[BodyPredicate]

    def __init__(self, die: Optional[DebugInfoEmitter]):
        '''
        Params
        ------
        die: Optional[DebugInfoEmitter]
            The optional debug info emitter: `None` if no debug info is to be
            emitted.
        '''

        self.irb = IRBuilder()
        self.die = die

        self.loop_contexts = deque()
        self.nothing_value = llvalue.Constant.Null(lltypes.Int1Type())
        self.body_preds = []

    def add_predicate(self, pred: BodyPredicate):
        '''
        Adds a predicate to the list of predicates to generate.

        Params
        ------
        pred: BodyPredicate
            The predicate to generate.
        '''

        # Predicate is a body predicate...
        if isinstance(pred, BodyPredicate):
            self.body_preds.append(pred)

    def generate(self):
        '''Generates all the added predicates.'''

        for body_pred in self.body_preds:
            self.generate_body_predicate(self, body_pred.ll_func, body_pred.func_params, body_pred.expr)

    # ---------------------------------------------------------------------------- #

    def generate_body_predicate(self, ll_func: ir.Function, func_params: List[Symbol], body_expr: ASTNode):
        self.body = ll_func.body

        self.var_block = self.body.append('vars')
        self.irb.move_to_start(self.var_block)

        for param in func_params:
            if param.mutability == Symbol.Mutability.MUTABLE:
                param_var = self.irb.build_alloca(conv_type(param.type, True))
                self.irb.build_store(param.ll_value, param_var)
                param.ll_value = param_var
        
        bb = ll_func.body.append()
        self.irb.build_br(bb)

        self.irb.move_to_start(bb)

        with self.maybe_debug_scope(body_expr.span):
            result = self.generate_expr(body_expr)

        if self.var_block.instructions.first().is_terminator:
            self.body.remove(self.var_block)
        
        if not self.irb.block.terminator:
            if result:
                self.irb.build_ret(result)
            else:
                self.irb.build_ret()

    # ---------------------------------------------------------------------------- #

    def generate_if_tree(self, if_tree: IfTree) -> Optional[llvalue.Value]:
        incoming = []

        if_end = self.body.append()
        never_end = True

        for cond_branch in if_tree.cond_branches:
            if cond_branch.header_var:
                self.generate_var_decl(cond_branch.header_var)

            if_then = self.body.append()
            if_else = self.body.append()

            ll_cond = self.generate_expr(cond_branch.condition)
            self.irb.build_cond_br(ll_cond, if_then, if_else)

            self.irb.move_to_end(if_then)
            ll_value = self.generate_expr(cond_branch.body)

            if not self.irb.block.terminator:
                never_end = False
                self.irb.build_br(if_end)

                if ll_value:
                    incoming.append(ir.PHIIncoming(ll_value, if_then))

            self.irb.move_to_end(if_else)

        if if_tree.else_branch:
            ll_value = self.generate_expr(if_tree.else_branch)

            if ll_value:
                incoming.append(ir.PHIIncoming(ll_value, if_else))

            never_end = bool(self.irb.block.terminator)
        else:
            never_end = False
        
        if not self.irb.block.terminator:
            self.irb.build_br(if_end)
 
        if never_end:
            self.body.remove(if_end)
            return None

        self.irb.move_to_end(if_end)

        if is_nothing(if_tree.type):
            return None

        if len(incoming) > 0:
            phi_node = self.irb.build_phi(conv_type(if_tree.type))
            phi_node.incoming.add(*incoming)

            return phi_node

    def generate_while_loop(self, while_loop: WhileLoop):
        if while_loop.header_var:
            self.generate_var_decl(while_loop.header_var)

        loop_header = self.body.append()
        self.irb.build_br(loop_header)
        self.irb.move_to_end(loop_header)

        loop_body = self.body.append()
        loop_exit = self.body.append()

        ll_cond = self.generate_expr(while_loop.condition)
        self.irb.build_cond_br(ll_cond, loop_body, loop_exit)

        if while_loop.update_stmt:
            loop_update = self.body.append()
            self.irb.move_to_start(loop_update)
            
            match stmt := while_loop.update_stmt:
                case Assignment():
                    self.generate_assignment(stmt)
                case IncDecStmt():
                    self.generate_incdec(stmt)
                case _:
                    self.generate_expr(stmt)

            if not self.irb.block.terminator:
                self.irb.build_br(loop_header)

            loop_continue = loop_update
        else:
            loop_continue = loop_header

        self.irb.move_to_end(loop_body)

        self.push_loop_context(loop_exit, loop_continue)
        self.generate_expr(while_loop.body)
        self.pop_loop_context()

        if not self.irb.block.terminator:
            self.irb.build_br(loop_continue)

        self.irb.move_to_end(loop_exit)

    # ---------------------------------------------------------------------------- #

    def generate_block(self, stmts: List[ASTNode]) -> Optional[llvalue.Value]:
        for i, stmt in enumerate(stmts):
            match stmt:
                case VarDecl():
                    self.generate_var_decl(stmt)
                case Assignment():
                    self.generate_assignment(stmt)
                case IncDecStmt():
                    self.generate_incdec(stmt)
                case KeywordStmt(keyword):
                    match keyword.kind:
                        case Token.Kind.BREAK:
                            self.irb.build_br(self.loop_context.break_block)
                            return
                        case Token.Kind.CONTINUE:
                            self.irb.build_br(self.loop_context.continue_block)
                            return
                case ReturnStmt(exprs):
                    self.irb.build_ret(*(self.generate_expr(expr) for expr in exprs))
                    return
                case _:
                    if (expr_value := self.generate_expr(stmt)) and i == len(stmts) - 1:
                        return expr_value

    def generate_var_decl(self, vd: VarDecl):
        for var_list in vd.var_lists:
            if var_list.initializer:
                var_ll_init = self.generate_expr(var_list.initializer)

                if not var_ll_init:
                    var_ll_init = self.nothing_value
            else:
                var_ll_init = None

            for sym in var_list.symbols:
                if not var_ll_init:
                    ll_init = llvalue.Constant.Null(conv_type(sym.type))
                else:
                    # TODO handle pattern matching
                    ll_init = var_ll_init

                if sym.mutability == Symbol.Mutability.MUTABLE:
                    ll_var = self.alloc_var(sym.type)

                    if ll_init:
                        self.irb.build_store(ll_init, ll_var)
                else:
                    ll_var = var_ll_init

                sym.ll_value = ll_var        

    def generate_assignment(self, assign: Assignment):
        lhss = [self.generate_lhs_expr(lhs_expr) for lhs_expr in assign.lhs_exprs]

        if len(assign.lhs_exprs) == len(assign.rhs_exprs):
            if len(assign.compound_ops) == 0:
                rhss = [
                    self.coalesce_nothing(self.generate_expr(rhs_expr))
                    for rhs_expr in assign.rhs_exprs
                ]
            else:
                rhss = [
                    self.generate_binary_op_app(BinaryOpApp(
                        op, 
                        LLVMValueNode(self.irb.build_load(lltypes.PointerType.from_type(lhs.type).elem_type, lhs), lhs_expr.type, lhs_expr.span), 
                        rhs_expr, lhs_expr.type
                    ))
                    for lhs, lhs_expr, rhs_expr, op in
                    zip(lhss, assign.lhs_exprs, assign.rhs_exprs, assign.compound_ops)
                ]
            
            for lhs, rhs in zip(lhss, rhss):
                self.irb.build_store(rhs, lhs)
        else:
            # TODO pattern matching
            raise NotImplementedError()

    def generate_incdec(self, incdec: IncDecStmt):
        lhs_ptr = self.generate_lhs_expr(incdec.lhs_operand)
        lhs_ll_type = conv_type(incdec.lhs_operand.type)
        lhs_value = self.irb.build_load(lhs_ll_type, lhs_ptr)
        one_value = llvalue.Constant.Int(lltypes.IntegerType.from_type(lhs_ll_type), 1)

        rhs_value = self.generate_binary_op_app(BinaryOpApp(
            incdec.op, 
            LLVMValueNode(lhs_value, incdec.lhs_operand.type, incdec.span),
            LLVMValueNode(one_value, incdec.lhs_operand.type, incdec.span)
        ))

        self.irb.build_store(rhs_value, lhs_ptr)

    def generate_lhs_expr(self, lhs_expr: ASTNode) -> llvalue.Value:
        match lhs_expr:
            case Identifier(symbol=sym):
                return sym.ll_value
            case Dereference(ptr):
                return self.generate_expr(ptr)
            case _:
                raise NotImplementedError()

    # ---------------------------------------------------------------------------- #

    def generate_expr(self, expr: ASTNode) -> Optional[llvalue.Value]:
        match expr:
            case IfTree():
                return self.generate_if_tree(expr)
            case WhileLoop():
                return self.generate_while_loop(expr)
            case Block(stmts):
                return self.generate_block(stmts)
            case TypeCast(src_expr, dest_type):
                return self.generate_type_cast(src_expr, dest_type)
            case BinaryOpApp():
                return self.generate_binary_op_app(expr)
            case UnaryOpApp():
                return self.generate_unary_op_app(expr)
            case FuncCall():
                return self.generate_func_call(expr)
            case Indirect(elem):
                # TEMPORARY STACK ALLOCATION CODE
                ptr = self.alloc_var(elem.type)

                # TODO handle nothing pointers
                ll_elem = self.generate_expr(elem)
                self.irb.build_store(ll_elem, ptr)

                return ptr
            case Identifier(symbol=sym):
                if is_nothing(sym):
                    return None
                elif sym.mutability == Symbol.Mutability.MUTABLE:
                    return self.irb.build_load(conv_type(sym.type), sym.ll_value)
                else:
                    return sym.ll_value
            case Literal():
                return self.generate_literal(expr)
            case Null(type=typ):
                return llvalue.Constant.Null(conv_type(typ))
            case LLVMValueNode(ll_value=value):
                return value
            case _:
                raise NotImplementedError()

    def generate_type_cast(self, src_expr: ASTNode, dest_type: Type) -> Optional[llvalue.Value]:
        if is_nothing(dest_type):
            return None
        elif src_expr.type == dest_type:
            return src_expr

        src_itype = src_expr.type.inner_type()
        dest_itype = dest_type.inner_type()

        src_ll_val = self.generate_expr(src_expr)
        dest_ll_type = conv_type(dest_itype)

        match (src_itype, dest_itype):
            case (PrimitiveType(), PrimitiveType()):
                # int to int
                if dest_itype.is_integral and src_itype.is_integral:
                    # small int to large int
                    if src_itype.usable_width < dest_itype.usable_width:
                        # small signed int to large signed int 
                        if src_itype.is_signed and dest_itype.is_signed:
                            return self.irb.build_sext(src_ll_val, dest_ll_type)
                        # small unsigned int to large signed/unsigned int
                        # or small signed int to large unsigned int
                        else:
                            return self.irb.build_zext(src_ll_val, dest_ll_type)
                    # large int to small int
                    else:
                        return self.irb.build_trunc(src_ll_val, dest_ll_type)

                # TODO other primitive casts
                raise NotImplementedError()
            case (PointerType(), PointerType()):
                return self.irb.build_bit_cast(src_ll_val, conv_type(dest_itype))
            case _:
                raise NotImplementedError()

    def generate_binary_op_app(self, bop_app: BinaryOpApp) -> Optional[llvalue.Value]:
        lhs = self.coalesce_nothing(self.generate_expr(bop_app.lhs))

        # handle short-circuit evaluation
        match bop_app.op.overload.intrinsic_name:
            case 'land':
                start_block = self.irb.block
                true_block = self.body.append()
                exit_block = self.body.append()

                self.irb.build_cond_br(lhs, true_block, exit_block)

                self.irb.move_to_end(true_block)
                rhs = self.generate_expr(bop_app.rhs)
                self.irb.build_br(exit_block)

                self.irb.move_to_end(exit_block)
                and_result = self.irb.build_phi(lltypes.Int1Type())
                and_result.incoming.add(ir.PHIIncoming(llvalue.Constant.Bool(False), start_block))
                and_result.incoming.add(ir.PHIIncoming(rhs, true_block))

                return and_result
            case 'lor':
                start_block = self.irb.block
                false_block = self.body.append()
                exit_block = self.body.append()

                self.irb.build_cond_br(lhs, exit_block, false_block)

                self.irb.move_to_end(false_block)
                rhs = self.generate_expr(bop_app.rhs)
                self.irb.build_br(exit_block)

                self.irb.move_to_end(exit_block)
                or_result = self.irb.build_phi(lltypes.Int1Type())
                or_result.incoming.add(ir.PHIIncoming(llvalue.Constant.Bool(True), start_block))
                or_result.incoming.add(ir.PHIIncoming(rhs, false_block))
                
                return or_result

        rhs = self.coalesce_nothing(self.generate_expr(bop_app.rhs))

        match bop_app.op.overload.intrinsic_name:
            case 'iadd':
                return self.irb.build_add(lhs, rhs)
            case 'isub':
                return self.irb.build_sub(lhs, rhs)
            case 'imul':
                return self.irb.build_mul(lhs, rhs)
            case 'sdiv':
                return self.irb.build_sdiv(lhs, rhs)
            case 'udiv':
                return self.irb.build_udiv(lhs, rhs)
            case 'smod':
                return self.irb.build_srem(lhs, rhs)
            case 'umod':
                return self.irb.build_urem(lhs, rhs)
            case 'fadd':
                return self.irb.build_fadd(lhs, rhs)
            case 'fsub':
                return self.irb.build_fsub(lhs, rhs)
            case 'fmul':
                return self.irb.build_fmul(lhs, rhs)
            case 'fdiv':
                return self.irb.build_fdiv(lhs, rhs)
            case 'fmod':
                return self.irb.build_frem(lhs, rhs)
            case 'shl':
                return self.irb.build_shl(lhs, rhs)
            case 'ashr':
                return self.irb.build_ashr(lhs, rhs)
            case 'lshr':
                return self.irb.build_lshr(lhs, rhs)
            case 'band':
                return self.irb.build_and(lhs, rhs)
            case 'bor':
                return self.irb.build_or(lhs, rhs)
            case 'bxor':
                return self.irb.build_xor(lhs, rhs)
            case 'slt':
                return self.irb.build_icmp(ir.IntPredicate.SLT, lhs, rhs)
            case 'ult':
                return self.irb.build_icmp(ir.IntPredicate.ULT, lhs, rhs)
            case 'flt':
                return self.irb.build_fcmp(ir.RealPredicate.ULT, lhs, rhs)
            case 'sgt':
                return self.irb.build_icmp(ir.IntPredicate.SGT, lhs, rhs)
            case 'ugt':
                return self.irb.build_icmp(ir.IntPredicate.UGT, lhs, rhs)
            case 'fgt':
                return self.irb.build_fcmp(ir.RealPredicate.UGT, lhs, rhs)
            case 'slteq':
                return self.irb.build_icmp(ir.IntPredicate.SLE, lhs, rhs)
            case 'ulteq':
                return self.irb.build_icmp(ir.IntPredicate.ULE, lhs, rhs)
            case 'flteq':
                return self.irb.build_fcmp(ir.RealPredicate.ULE, lhs, rhs)
            case 'sgteq':
                return self.irb.build_icmp(ir.IntPredicate.SGE, lhs, rhs)
            case 'ugteq':
                return self.irb.build_icmp(ir.IntPredicate.UGE, lhs, rhs)
            case 'fgteq':
                return self.irb.build_fcmp(ir.RealPredicate.UGE, lhs, rhs)
            case 'ieq':
                return self.irb.build_icmp(ir.IntPredicate.EQ, lhs, rhs)
            case 'feq':
                return self.irb.build_fcmp(ir.RealPredicate.UEQ, lhs, rhs)
            case 'ineq':
                return self.irb.build_icmp(ir.IntPredicate.NE, lhs, rhs)
            case 'fneq':
                return self.irb.build_fcmp(ir.RealPredicate.UNE, lhs, rhs)
            case '':
                # Not an intrinsic operator.
                oper_call = self.irb.build_call(bop_app.op.overload.ll_value, lhs, rhs)

                if is_nothing(bop_app.rt_type):
                    return None

                return oper_call

    def generate_unary_op_app(self, uop_app: UnaryOpApp) -> Optional[llvalue.Value]:
        operand = self.coalesce_nothing(self.generate_expr(uop_app.operand))

        match uop_app.op.overload.intrinsic_name:
            case 'ineg':
                return self.irb.build_neg(operand)
            case 'fneg':
                return self.irb.build_fneg(operand)
            case 'compl' | 'not':
                return self.irb.build_not(operand)
            case _:
                # Not an intrinsic operator
                oper_call = self.irb.build_call(uop_app.op.overload.ll_value, operand)
                
                if is_nothing(uop_app.rt_type):
                    return None

                return oper_call

    def generate_func_call(self, fc: FuncCall) -> Optional[llvalue.Value]:
        ll_args = [self.coalesce_nothing(self.generate_expr(arg)) for arg in fc.args]

        ll_call = self.irb.build_call(
            self.generate_expr(fc.func),
            *ll_args
        )

        if is_nothing(fc.rt_type):
            return None

        return ll_call

    def generate_literal(self, lit: Literal) -> Optional[llvalue.Value]:
        match lit.kind:
            case Token.Kind.BOOLLIT:
                return llvalue.Constant.Bool(lit.value == 'true')
            case Token.Kind.INTLIT:
                trimmed_value, base, _, _ = util.trim_int_lit(lit.value)

                return llvalue.Constant.Int(conv_type(lit.type), int(trimmed_value, base))   
            case Token.Kind.FLOATLIT:
                return llvalue.Constant.Real(conv_type(lit.type), float(lit.value.replace('_', '')))
            case Token.Kind.RUNELIT:
                return llvalue.Constant.Int(lltypes.Int32Type(), get_rune_char_code(lit.value))
            case Token.Kind.NUMLIT:
                return llvalue.Constant.Int(conv_type(lit.type), int(lit.value.replace('_', '')))

    # ---------------------------------------------------------------------------- #

    def alloc_var(self, typ: Type) -> llvalue.Value:
        curr_block = self.irb.block
        self.irb.move_to_start(self.var_block)
        ll_var = self.irb.build_alloca(conv_type(typ, alloc_type=True))
        self.irb.move_to_end(curr_block)
        return ll_var

    def coalesce_nothing(self, value: Optional[llvalue.Value]) -> llvalue.Value:
        if value:
            return value

        return self.nothing_value

    # ---------------------------------------------------------------------------- #

    @property
    def loop_context(self) -> LoopContext:
        return self.loop_contexts[0]

    def push_loop_context(self, bb: ir.BasicBlock, cb: ir.BasicBlock):
        self.loop_contexts.appendleft(LoopContext(bb, cb))

    def pop_loop_context(self):
        self.loop_contexts.popleft()

    # ---------------------------------------------------------------------------- #

    def maybe_emit_location(self, span: TextSpan):
        if self.die:
            self.irb.debug_location = self.die.as_di_location(span)

    @contextlib.contextmanager
    def maybe_debug_scope(self, span: TextSpan):
        if self.die:
            self.die.push_scope(span)
            yield
            self.die.pop_scope()
        else:
            yield

def get_rune_char_code(rune_val: str) -> int:
    match rune_val:
        case '\\n':
            return 10
        case '\\\\':
            return 92
        case '\\v':
            return 11
        case '\\r':
            return 13
        case '\\0':
            return 0
        case '\\a':
            return 7
        case '\\b':
            return 8
        case '\\f':
            return 12
        case '\\\'':
            return 39
        case '\\\"':
            return 34
        case '\\t':
            return 9
        case _:
            if rune_val.startswith(('\\x', '\\u', '\\U')):
                return int(rune_val[2:], base=16) 
            else:
                return ord(rune_val)


