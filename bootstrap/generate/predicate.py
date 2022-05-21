from re import L
from typing import List, Optional
from dataclasses import dataclass

import llvm.ir as ir
import llvm.value as llvalue
import llvm.types as lltypes
from llvm.builder import IRBuilder
from syntax.ast import *
from syntax.token import Token
from depm import Symbol
from typecheck import *
import util
from .type_util import conv_type, is_nothing

@dataclass
class Predicate:
    ll_func: ir.Function
    func_params: List[Symbol]
    expr: ASTNode
    store_result: Optional[Symbol] = None

class PredicateGenerator:
    irb: IRBuilder
    var_block: ir.BasicBlock
    body: ir.FuncBody

    nothing_value: llvalue.Value = llvalue.Constant.Null(lltypes.Int1Type)

    def __init__(self):
        self.irb = IRBuilder()

    def generate(self, pred: Predicate):
        self.body = pred.ll_func.body

        if len(self.body) > 0:
            prev_block = self.body.last()
        else:
            prev_block = None

        self.var_block = self.body.append('vars')
        self.irb.move_to_start(self.var_block)

        for param in pred.func_params:
            if param.mutability == Symbol.Mutability.MUTABLE:
                param_var = self.irb.build_alloca(conv_type(param.type))
                self.irb.build_store(param.ll_value, param_var)
                param.ll_value = param_var
        
        bb = pred.ll_func.body.append()
        self.irb.build_br(bb)

        self.irb.move_to_start(bb)

        result = self.generate_expr(pred.expr)

        if self.var_block.instructions.first().is_terminator:
            self.body.remove(self.var_block)

            if prev_block:
                final_block = self.irb.block
                self.irb.move_to_end(prev_block)
                self.irb.build_br(bb)
                self.irb.move_to_end(final_block)
        elif prev_block:
            final_block = self.irb.block
            self.irb.move_to_end(prev_block)
            self.irb.build_br(self.var_block)
            self.irb.move_to_end(final_block)

        if result:
            if pred.store_result:
                self.irb.build_store(result, pred.store_result.ll_value)
            else:
                self.irb.build_ret(result)
        elif not pred.store_result:
            self.irb.build_ret()

    # ---------------------------------------------------------------------------- #

    def generate_block(self, stmts: List[ASTNode]):
        for stmt in stmts:
            match stmt:
                case VarDecl(var_lists, _):
                    for vlist in var_lists:
                        self.generate_var_list(vlist)
                case _:
                    self.generate_expr(stmt)

    def generate_var_list(self, var_list: VarList):
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
            else:
                ll_var = var_ll_init

            sym.ll_value = ll_var

    # ---------------------------------------------------------------------------- #

    def generate_expr(self, expr: ASTNode) -> Optional[llvalue.Value]:
        match expr:
            case Block(stmts):
                return self.generate_block(stmts)
            case TypeCast(src_expr, dest_type):
                return self.generate_type_cast(src_expr, dest_type)
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
                    if src_itype.size < dest_itype.size:
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

    def generate_binary_op_app(self, bop: BinaryOpApp) -> Optional[llvalue.Value]:
        lhs = self.generate_expr(bop.lhs)
        if not lhs:
            lhs = self.nothing_value

        # handle short-circuit evaluation
        match bop.overload.intrinsic_name:
            case 'land':
                start_block = self.irb.block
                true_block = self.body.append()
                exit_block = self.body.append()

                self.irb.build_cond_br(lhs, true_block, exit_block)

                self.irb.move_to_end(true_block)
                rhs = self.generate_expr(bop.rhs)
                self.irb.build_br(exit_block)

                self.irb.move_to_end(exit_block)
                and_result = self.irb.build_phi(lltypes.Int1Type)
                and_result.incoming.add(ir.PHIIncoming(llvalue.Constant.Bool(False), start_block))
                and_result.incoming.add(ir.PHIIncoming(rhs, true_block))

                return and_result
            case 'lor':
                start_block = self.irb.block
                false_block = self.body.append()
                exit_block = self.body.append()

                self.irb.build_cond_br(lhs, exit_block, false_block)

                self.irb.move_to_end(false_block)
                rhs = self.generate_expr(bop.rhs)
                self.irb.build_br(exit_block)

                self.irb.move_to_end(exit_block)
                or_result = self.irb.build_phi(lltypes.Int1Type)
                or_result.incoming.add(ir.PHIIncoming(llvalue.Constant.Bool(True), start_block))
                or_result.incoming.add(ir.PHIIncoming(rhs, false_block))
                
                return or_result
        rhs = self.generate_expr(bop.rhs)
        if not rhs:
            rhs = self.nothing_value

        match bop.overload.intrinsic_name:
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
                oper_call = self.irb.build_call(bop.overload.ll_value, lhs, rhs)

                if is_nothing(bop.rt_type):
                    return None

                return oper_call

    def generate_func_call(self, fc: FuncCall) -> Optional[llvalue.Value]:
        ll_args = []
        for arg in fc.args:
            ll_arg = self.generate_expr(arg)

            if ll_arg:
                ll_args.append(ll_arg)
            else:
                ll_args.append(self.nothing_value)

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
                return llvalue.Constant.Int(lltypes.Int32Type, get_rune_char_code(lit.value))
            case Token.Kind.NUMLIT:
                return llvalue.Constant.Int(conv_type(lit.type), int(lit.value.replace('_', '')))

    # ---------------------------------------------------------------------------- #

    def alloc_var(self, typ: Type) -> llvalue.Value:
        curr_block = self.irb.block
        self.irb.move_to_start(self.var_block)
        ll_var = self.irb.build_alloca(conv_type(typ, alloc_type=True))
        self.irb.move_to_end(curr_block)
        return ll_var

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


