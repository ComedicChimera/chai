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

    nothing_value: lltypes.Type = llvalue.Constant.Null(lltypes.Int1Type)

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

        if len(self.var_block.instructions) == 1:
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
                curr_block = self.irb.block
                self.irb.move_to_end(self.var_block)

                ll_var = self.irb.build_alloca(conv_type(sym.type, True))
                self.irb.build_store(ll_init, ll_var)

                self.irb.move_to_end(curr_block)
            else:
                ll_var = var_ll_init

            sym.ll_value = ll_var

    # ---------------------------------------------------------------------------- #

    def generate_expr(self, expr: ASTNode) -> Optional[llvalue.Value]:
        match expr:
            case Block(stmts):
                return self.generate_block(stmts)
            case FuncCall():
                return self.generate_func_call(expr)
            case Indirect(ptr):
                return self.irb.build_load(conv_type(expr.type.elem_type), self.generate_expr(ptr))
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


    def generate_func_call(self, fc: FuncCall) -> Optional[llvalue.Value]:
        ll_args = []
        for arg in fc.args:
            ll_arg = self.generate_expr(arg)

            if ll_arg:
                ll_args.append(ll_arg)
            else:
                ll_args.append(self.nothing_value)

        ll_call = self.irb.build_call(
            conv_type(fc.rt_type, rt_type=True),
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


