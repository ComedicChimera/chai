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
import llvm.metadata as llmeta
import util

from report import TextSpan
from llvm.builder import IRBuilder
from syntax.ast import *
from syntax.token import Token
from depm import Symbol
from typecheck import *

from .type_util import conv_type, is_unit
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

@dataclass
class LHSLLData:
    '''
    Stores the data needed to generate an assignment to an LHS expression.

    Attributes
    ----------
    value_ptr: llvalue.Value
        The LLVM value pointer
    di_expr: llmeta.MDNode
        The DWARF DIExpression describing the `value_ptr` calculations.
    '''

    value_ptr: llvalue.Value
    di_expr: llmeta.MDNode

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
    unit_value: llvalue.Value

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
        self.unit_value = llvalue.Constant.Null(lltypes.Int1Type())
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

        # Generate all the body predicates.
        for body_pred in self.body_preds:
            self.generate_body_predicate(body_pred.ll_func, body_pred.func_params, body_pred.expr)

    # ---------------------------------------------------------------------------- #

    def generate_body_predicate(self, ll_func: ir.Function, func_params: List[Symbol], body_expr: ASTNode):
        '''
        Generates a body predicate.
        
        Params
        ------
        ll_func: ir.Function
            The IR function whose body is being generated.
        func_params: List[Symbol]
            The parameters to the function whose body is being generated.
        body_expr: ASTNode
            The predicate expression of the function body.
        '''

        # Set the predicate generator's body to the function body.
        self.body = ll_func.body

        # Begin the variable block: all variables are allocated at the start of
        # the function body in this block to avoid unnecessary allocations.
        self.var_block = self.body.append('vars')
        self.irb.move_to_start(self.var_block)

        # Emit a lexical scope to contain the body of the function.
        with self.die.emit_scope(body_expr.span):
            # Generate the parameter prelude.
            for i, param in enumerate(func_params):
                # Emit appropriate debug information for each parameter.
                self.die.emit_param_info(i, param)

                # Allocate a mutable variable for all mutable parameters.
                if param.mutability == Symbol.Mutability.MUTABLE:
                    param_var = self.irb.build_alloca(conv_type(param.type, True))
                    self.irb.build_store(param.ll_value, param_var)
                    param.ll_value = param_var

                    # TODO determine if we need to add variable debug
                    # declaration here
            
            # Add the entry block to the function body and jump to it from the
            # variable block.
            entry_b = ll_func.body.append('entry')
            self.irb.build_br(entry_b)

            # Position the builder at the start of the entry block so it can
            # start generating the actual body of the function.
            self.irb.move_to_start(entry_b)

            # Generate the function body.
            result = self.generate_expr(body_expr)

        # If the variable block contains only a terminator instruction, then
        # there are no variables declared so we can just remove the variable
        # block.
        if self.var_block.instructions.first().is_terminator:
            self.body.remove(self.var_block)
        
        # If the last block does not already have a terminator (eg. from a
        # return), then we add an appropriate function return terminator for it.
        if not self.irb.block.terminator:
            # Generate a return from the resultant value is it is not a unit
            # value.  If it is, then (effectively) nothing is returned from the
            # function so we can just generate a `ret void`.
            if result:
                self.irb.build_ret(result)
            else:
                self.irb.build_ret()

    # ---------------------------------------------------------------------------- #

    # NOTE Many of the generation functions return an optional instead of just
    # an LLVM value.  The reason is that any branch which values to the unit
    # value may not actually correspond to an LLVM value and should be pruned if
    # possible.

    def generate_if_tree(self, if_tree: IfTree) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR for an if/elif/else tree expression.

        Params
        ------
        if_tree: IfTree
            The if/elif/else tree to generate.

        Returns
        -------
        Optional[llvalue.Value]
            The yielded value of the block if it is not the unit value.
        '''

        # Prepare a list to store the incoming conditional blocks of the if
        # expression so we can build the PHI node at the end of the expressin as
        # necessary.
        incoming = []

        # Append the closing block of the if statement so we can jump to it from
        # the conditional blocks.
        if_end = self.body.append()

        # Whether the conditional branches ever actually reach the else block.
        never_end = True

        # Generate all the conditional branches of the block.
        for cond_branch in if_tree.cond_branches:
            # Emit a debug scope to contain the block.
            with self.die.emit_scope(cond_branch.body.span):
                # Generate the header variable as necessary.
                if cond_branch.header_var:
                    self.generate_var_decl(cond_branch.header_var)

                # Append the then and else blocks so we can reference them in
                # the conditional break of the if statement.
                if_then = self.body.append()
                if_else = self.body.append()

                # Generate the condition and use it to create the conditional branch.
                ll_cond = self.generate_expr(cond_branch.condition)
                self.irb.build_cond_br(ll_cond, if_then, if_else)

                # Generate the then block body.
                self.irb.move_to_end(if_then)
                ll_value = self.generate_expr(cond_branch.body)

            # If the then block isn't already terminated,
            if not self.irb.block.terminator:
                # Jump to the end of the if statement.
                self.irb.build_br(if_end)

                # Indicate the if block reaches the if_end.
                never_end = False

                # If there is a value, add it the list of incoming values to the
                # PHI node.
                if ll_value:
                    incoming.append(ir.PHIIncoming(ll_value, if_then))

            # Position the builder at the end of the else block for the start of
            # the next condition or else branch.
            self.irb.move_to_end(if_else)

        # Generate the else branch if it exists.
        if if_tree.else_branch:
            # Generate the body of the else branch within its own scope.
            with self.die.emit_scope(if_tree.else_branch.span):
                ll_value = self.generate_expr(if_tree.else_branch)
        
            # If the else branch is not already terminated, ...
            if not self.irb.block.terminator:
                # Jump to the ending block.
                self.irb.build_br(if_end)

                # Indicate the if block reaches the if_end.
                never_end = False

                # If there is a value, add it to the list of incoming values to
                # the if PHI node.
                if ll_value:
                    incoming.append(ir.PHIIncoming(ll_value, if_else))
        else:
            # Branch from the else block to the end of the if block.
            self.irb.build_br(if_end)

            # Since there is no else, the if expression can always reach the
            # terminator (if none of the conditions are true).
            never_end = False
 
        # If the if block never reaches the end block, then we can just remove
        # the end block and return nothing since all other branches exit early.
        if never_end:
            self.body.remove(if_end)
            return None

        # Otherwise, move the builder to the end block so it can continue
        # generating from the end of the if statement.
        self.irb.move_to_end(if_end)

        # If the if tree yields a value, then we build a PHI node of the
        # incoming values so handle the outbound values.
        if not is_unit(if_tree.type):
            # The debug location should always be None if we reach this point so
            # we don't need to clear it again.
            phi_node = self.irb.build_phi(conv_type(if_tree.type))
            phi_node.incoming.add(*incoming)

            return phi_node

    def generate_while_loop(self, while_loop: WhileLoop):
        '''
        Generates the LLVM IR for a while loop.

        Params
        ------
        while_loop: WhileLoop
            The while loop to generate.
        '''

        # Create a debug scope for the body of the loop.
        with self.die.emit_scope(while_loop.body.span):
            # Generate the loop header variable if it exists.
            if while_loop.header_var:
                self.generate_var_decl(while_loop.header_var)

            # Generate the loop header.
            loop_header = self.body.append()
            self.irb.build_br(loop_header)
            self.irb.move_to_end(loop_header)

            # Add the loop body and loop exit blocks so we can
            # jump to them from within the loop.
            loop_body = self.body.append()
            loop_exit = self.body.append()

            # Generate the conditional branch instruction to begin the loop.
            ll_cond = self.generate_expr(while_loop.condition)
            self.irb.build_cond_br(ll_cond, loop_body, loop_exit)

            # If there is a loop update statement (C-style loop), ...
            if while_loop.update_stmt:
                # Add the update block.
                loop_update = self.body.append()
                self.irb.move_to_start(loop_update)
                
                # Generate the statements contained therein.
                match stmt := while_loop.update_stmt:
                    case Assignment():
                        self.generate_assignment(stmt)
                    case IncDecStmt():
                        self.generate_incdec(stmt)
                    case _:
                        self.generate_expr(stmt)

                # Add a terminator if the update block does not self terminate
                # so that it can jump back to the loop header.
                if not self.irb.block.terminator:
                    self.irb.build_br(loop_header)

                # Set the loop continue block to the update block so that it is
                # jumped to at the end of each iteration and after any continues
                # which occur in the loop.
                loop_continue = loop_update
            else:
                # Otherwise, the loop continue is just the header.
                loop_continue = loop_header

            # Position the body at the start of the loop body.
            self.irb.move_to_start(loop_body)

            # Create the loop context and generate the while loop's body.
            self.push_loop_context(loop_exit, loop_continue)
            self.generate_expr(while_loop.body)
            self.pop_loop_context()

        # Add a terminator the loop block if necessary.
        if not self.irb.block.terminator:
            self.irb.build_br(loop_continue)

        # Position the builder at the end of the loop so it can continue
        # generating.
        self.irb.move_to_end(loop_exit)

    # ---------------------------------------------------------------------------- #

    def generate_block(self, stmts: List[ASTNode]) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR code for a block (sequence of statements).

        Params
        ------
        stmts: List[ASTNode]
            The list of statements in the block.

        Returns
        -------
        Optional[llvalue.Value]
            The yielded value of the block if it is not the unit value.
        '''

        # NOTE most branches of this function return nothing implying their
        # statements always yield the unit value (eg. let statements).

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
                            with self.debug_at_span(stmt.span):
                                self.irb.build_br(self.loop_context.break_block)

                            # Break statements always at least exit the current
                            # block so we return early since all code after them
                            # is always dead.
                            return
                        case Token.Kind.CONTINUE:
                            with self.debug_at_span(stmt.span):
                                self.irb.build_br(self.loop_context.continue_block)

                            # Continue statements always at least exit the current
                            # block so we return early since all code after them
                            # is always dead.
                            return
                case ReturnStmt(exprs):
                    with self.debug_at_span(stmt.span):
                        if len(exprs) == 1 and is_unit(exprs[0].type):
                            # Handle any `return ()`.
                            self.irb.build_ret()
                        else:
                            self.irb.build_ret(
                                *(self.coalesce_unit(self.generate_expr(expr)) 
                                for expr in exprs
                            ))

                    # Return statements always at least exit the current block
                    # so we return early since all code after them is always
                    # dead.
                    return
                case _:
                    # The value of the last statement in a block is always
                    # implicitly returned. For all other branches, it is the
                    # unit value; however, this branch may actually yield a
                    # value.
                    if (expr_value := self.generate_expr(stmt)) and i == len(stmts) - 1:
                        return expr_value

    def generate_var_decl(self, vd: VarDecl):
        '''
        Generates the LLVM IR for a variable declaration.

        Params
        ------
        vd: VarDecl
            The variable declaration to generate.
        '''

        for var_list in vd.var_lists:
            # If the variable list has an initializer, generate it.
            if var_list.initializer:
               
                var_ll_init = self.generate_expr(var_list.initializer)
            else:
                var_ll_init = None

            for sym in var_list.symbols:
                if var_ll_init:
                    # TODO handle pattern matching
                    ll_init = var_ll_init
                else:
                    # If there is no initializer, use `null` 
                    # TODO handle more complex null values
                    ll_init = self.get_null_value(sym.type)
               
                if sym.mutability == Symbol.Mutability.MUTABLE:
                    # We only need to actually allocate a variable if the symbol
                    # is mutable.
                    with self.debug_at_span(sym.def_span):
                        ll_var = self.alloc_var(sym.type)

                        # ll_init is never None so it is always safe to store it.
                        self.irb.build_store(ll_init, ll_var)

                    # Emit the debug information for the variable.
                    self.die.emit_local_var_decl(sym, ll_var, self.var_block.instructions.last())
                else:
                    # Otherwise, we can just point the symbol the yielded value
                    # of the initializer.
                    ll_var = var_ll_init

                    # Emit the debug information for the variable.
                    self.die.emit_local_var_decl(sym, ll_var, self.irb.block)

                sym.ll_value = ll_var        

    def generate_assignment(self, assign: Assignment):
        '''
        Generates the LLVM IR for an assignment statement.

        Params
        ------
        assign: Assignment
            The assignment statement to generate.
        '''

        # TODO generate DWARF debugging info for assignment

        # Generate all the LHS value pointers for the assignment.
        lhss = [self.generate_lhs_expr(lhs_expr) for lhs_expr in assign.lhs_exprs]

        with self.debug_at_span(assign.span):
            if len(assign.lhs_exprs) == len(assign.rhs_exprs):       
                if len(assign.compound_ops) == 0:
                    # If there is no compound assignment operation, then the RHS
                    # LLVM IR is just the yielded value of the RHS expressions.
                    rhss = [
                        # Coalesce unit values to a value so we can assign
                        # safely.

                        # NOTE should we just prune the assignment entirely...
                        self.coalesce_unit(self.generate_expr(rhs_expr))
                        for rhs_expr in assign.rhs_exprs
                    ]
                else:
                    # Otherwise, we generate a binary operator application
                    # between the LHS value and the RHS value and then store
                    # that into the LHS.
                    rhss = [
                        self.generate_binary_op_app(BinaryOpApp(
                            op, 
                            # Since we have already generated most of the LHS,
                            # all we need to do it is load from the LHS pointer
                            # to obtain the LHS value.  This avoids
                            # recalculating the LHS multiple times.
                            LLVMValueNode(
                                self.irb.build_load(lltypes.PointerType.from_type(lhs.type).elem_type, lhs), 
                                lhs_expr.type, 
                                lhs_expr.span
                            ), 
                            rhs_expr, lhs_expr.type
                        ))
                        for lhs, lhs_expr, rhs_expr, op in
                        zip(lhss, assign.lhs_exprs, assign.rhs_exprs, assign.compound_ops)
                    ]
                
                # Store the RHS into the LHS pointer.
                for lhs, rhs in zip(lhss, rhss):
                    self.irb.build_store(rhs, lhs)
            else:
                # TODO pattern matching
                raise NotImplementedError()

    def generate_incdec(self, incdec: IncDecStmt):
        '''
        Generates the LLVM IR for an increment/decrement statement.

        Params
        ------
        incdec: IncDecStmt
            The increment/decrement statement to generate.
        '''

        with self.debug_at_span(incdec.span):
            # Generate the LHS value pointer.
            lhs_ptr = self.generate_lhs_expr(incdec.lhs_operand)
            lhs_ll_type = conv_type(incdec.lhs_operand.type)

            # Extract the value of the LHS pointer so it can be used in the
            # arithmetic operation.
            lhs_value = self.irb.build_load(lhs_ll_type, lhs_ptr)

            # Create the constant one needed for the increment/decrement.
            one_value = llvalue.Constant.Int(lltypes.IntegerType.from_type(lhs_ll_type), 1)

            # Generate the binary operator application, `lhs + 1`, reusing the
            # calculated LHS and the generated one value.
            rhs_value = self.generate_binary_op_app(BinaryOpApp(
                incdec.op, 
                LLVMValueNode(lhs_value, incdec.lhs_operand.type, incdec.span),
                LLVMValueNode(one_value, incdec.lhs_operand.type, incdec.span)
            ))

            # Store the result of the sum back into the LHS value pointer.
            self.irb.build_store(rhs_value, lhs_ptr)

    def generate_lhs_expr(self, lhs_expr: ASTNode) -> LHSLLData:
        '''
        Generate the LLVM IR for accessing the value pointer for a mutable
        value.  This is used to assign to mutable values which are actually just
        compiled as pointers.

        For example, if you wanted to compile `x = 10`, you couldn't just use
        the the regular expression generation code to generate the `x` on the
        LHS because that would give you the value of `x`.  Instead, you need to
        access the pointer to the variable `x` which is a different operation.

        Params
        ------
        lhs_expr: ASTNode
            The LHS expression to generate.
        '''

        # NOTE I don't know that we actually need to use llvm.dbg.value... I am
        # not exactly sure when it should be used.

        match lhs_expr:
            case Identifier(symbol=sym):
                # If an identifier occurs as an LHS expression, it is always
                # mutable meaning its `ll_value` is its value pointer so we can
                # just return that without loading it.
                return sym.ll_value
            case Dereference(ptr):
                # Dereferencing on the LHS actually compiles as a NO-OP since
                # all the notation `*x = ...` means is assign to the value
                # pointed to by `x`, but since all we need to assign is the
                # pointer, namely `x`, we can just return `x`.
                return self.generate_expr(ptr)
            case _:
                raise NotImplementedError()

    # ---------------------------------------------------------------------------- #

    def generate_expr(self, expr: ASTNode) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR for any expression node.

        Params
        ------
        expr: ASTNode
            The expression node to generate.
        '''

        match expr:
            case IfTree():
                return self.generate_if_tree(expr)
            case WhileLoop():
                return self.generate_while_loop(expr)
            case Block(stmts):
                return self.generate_block(stmts)
            case TypeCast():
                return self.generate_type_cast(expr)
            case BinaryOpApp():
                return self.generate_binary_op_app(expr)
            case UnaryOpApp():
                return self.generate_unary_op_app(expr)
            case FuncCall():
                return self.generate_func_call(expr)
            case Indirect(elem):
                with self.debug_at_span(expr.span):
                    # TEMPORARY STACK ALLOCATION CODE
                    ptr = self.alloc_var(elem.type)

                    # TODO handle unit pointers
                    ll_elem = self.generate_expr(elem)
                    self.irb.build_store(ll_elem, ptr)

                    return ptr
            case Dereference(ptr):
                # TODO add any necessary dereference safely checks
                with self.debug_at_span(expr.span):
                    return self.irb.build_load(ptr)
            case Identifier(symbol=sym):
                if is_unit(sym):
                    # Symbols which are unit typed can just be replaced with the
                    # unit value so we just return `None` indicating that this
                    # identifier usage can be pruned.
                    return None
                elif sym.mutability == Symbol.Mutability.MUTABLE:
                    # Mutable symbols store a value pointer not just the value
                    # itself: in order to use the value, we have to load it from
                    # the pointer.
                    with self.debug_at_span(expr.span):
                        return self.irb.build_load(conv_type(sym.type), sym.ll_value)
                else:
                    return sym.ll_value
            case Literal():
                return self.generate_literal(expr)
            case Null(type=typ):
                return self.get_null_value(typ)
            case LLVMValueNode(ll_value=value):
                return value
            case _:
                raise NotImplementedError()

    def generate_type_cast(self, tc: TypeCast) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR for a type cast.

        Params
        ------
        tc: TypeCast
            The type cast to generate.
        '''

        if is_unit(tc.dest_type):
            # If we are casting to a unit type, then we don't actually need to
            # do anything and can just return nothing.
            return None
        elif tc.src_expr.type == tc.dest_type:
            # If the two types are equivalent, then we don't need to generate a
            # cast (just type semantics).
            return tc.src_expr

        # Get the inner types of the two types we are casting so we don't need
        # to add additional checks for type variables, aliases, etc.
        src_itype = tc.src_expr.type.inner_type()
        dest_itype = tc.dest_type.inner_type()

        # The generate the LLVM IR value to cast.  We know that this is never
        # `unit` since `unit` can only be cast to `unit` which we already
        # checked.
        src_ll_val = self.generate_expr(tc.src_expr)

        # Get the destination LLVM type so we can reuse it :)
        dest_ll_type = conv_type(dest_itype)

        with self.debug_at_span(tc.span):
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
        '''
        Generates the LLVM IR for a binary operator application.

        Params
        ------
        bop_app: BinaryOpApp
            The binary operator application to generate.
        '''

        # Generate the LHS operand now, coalescing any unit values we encounter
        # so they can be used safely.  Note that we do NOT want to generate the
        # code for the RHS operand just yet since short-circuit operators may
        # not want it to always run.
        lhs = self.coalesce_unit(self.generate_expr(bop_app.lhs))

        # Handle short-circuit evaluation operators.
        match bop_app.op.overload.intrinsic_name:
            case 'land':
                # The IR generated here is equivalent to the following:
                #
                # if a => b else => False end

                start_block = self.irb.block
                true_block = self.body.append()
                exit_block = self.body.append()

                self.irb.build_cond_br(lhs, true_block, exit_block)

                self.irb.move_to_end(true_block)
                rhs = self.generate_expr(bop_app.rhs)
                self.irb.build_br(exit_block)

                self.irb.move_to_end(exit_block)

                with self.debug_at_span(bop_app.span):
                    and_result = self.irb.build_phi(lltypes.Int1Type())
                    and_result.incoming.add(ir.PHIIncoming(llvalue.Constant.Bool(False), start_block))
                    and_result.incoming.add(ir.PHIIncoming(rhs, true_block))

                    return and_result
            case 'lor':
                # The IR generated here is equivalent to the following:
                #
                # if a => true else => b end

                start_block = self.irb.block
                false_block = self.body.append()
                exit_block = self.body.append()

                self.irb.build_cond_br(lhs, exit_block, false_block)

                self.irb.move_to_end(false_block)
                rhs = self.generate_expr(bop_app.rhs)
                self.irb.build_br(exit_block)

                self.irb.move_to_end(exit_block)

                with self.debug_at_span(bop_app.span):
                    or_result = self.irb.build_phi(lltypes.Int1Type())
                    or_result.incoming.add(ir.PHIIncoming(llvalue.Constant.Bool(True), start_block))
                    or_result.incoming.add(ir.PHIIncoming(rhs, false_block))
                    
                    return or_result

        # Now we can freely generate the RHS operand since no other operators
        # can short-circuit.
        rhs = self.coalesce_unit(self.generate_expr(bop_app.rhs))

        
        with self.debug_at_span(bop_app.span):
            # Handle intrinsic operators: this match expression is basically
            # just a table mapping names to LLVM instructions -- not any really
            # exciting logic here.
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
                    # Not an intrinsic operator => generate a function call.
                    oper_call = self.irb.build_call(bop_app.op.overload.ll_value, lhs, rhs)

                    # If the function returns unit, we return `None`: no actual
                    # LLVM value returned.
                    if is_unit(bop_app.rt_type):
                        return None

                    return oper_call

    def generate_unary_op_app(self, uop_app: UnaryOpApp) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR for a unary operator application.

        Params
        ------
        uop_app: UnaryOpApp
            The unary operator application to generate.
        '''

        # Generate the operand to unary operator, coalesing any unit types so
        # they can be used safely.
        operand = self.coalesce_unit(self.generate_expr(uop_app.operand))

        with self.debug_at_span(uop_app.span):
            # Handle all the intrinsic unary operators.
            match uop_app.op.overload.intrinsic_name:
                case 'ineg':
                    return self.irb.build_neg(operand)
                case 'fneg':
                    return self.irb.build_fneg(operand)
                case 'compl' | 'not':
                    return self.irb.build_not(operand)
                case _:
                    # Not an intrinsic operator => build a function call.
                    oper_call = self.irb.build_call(uop_app.op.overload.ll_value, operand)
                    
                    # If the function returns unit, then we return `None`: no
                    # actual LLVM value returned.
                    if is_unit(uop_app.rt_type):
                        return None

                    return oper_call

    def generate_func_call(self, fc: FuncCall) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR for a function call.

        Params
        ------
        fc: FuncCall
            The function call to generate.
        '''

        # Generate the argument values to the function.  Since we don't prune
        # unit arguments functions, we need to generate unit values to pass to
        # any functions which accept unit arguments.
        ll_args = [self.coalesce_unit(self.generate_expr(arg)) for arg in fc.args]

        with self.debug_at_span(fc.span):
            # Build the function call itself.
            ll_call = self.irb.build_call(
                self.generate_expr(fc.func),
                *ll_args
            )

        # If the function returns unit, then we return `None`: no actual LLVM
        # value returned.
        if is_unit(fc.rt_type):
            return None

        return ll_call

    def generate_literal(self, lit: Literal) -> Optional[llvalue.Value]:
        '''
        Generates the LLVM IR for a literal.

        Params
        ------
        lit: Literal
            The literal to generate.
        '''

        match lit.kind:
            case Token.Kind.BOOLLIT:
                return llvalue.Constant.Bool(lit.value == 'true')
            case Token.Kind.INTLIT:
                # We need to trim the prefixes and suffixes off of any integer
                # literals and use them to determine the base before we convert
                # it to any an integer value.
                trimmed_value, base, _, _ = util.trim_int_lit(lit.value)

                return llvalue.Constant.Int(conv_type(lit.type), int(trimmed_value, base))   
            case Token.Kind.FLOATLIT:
                return llvalue.Constant.Real(conv_type(lit.type), float(lit.value.replace('_', '')))
            case Token.Kind.RUNELIT:
                # Rune literals are just integer values: we convert them to
                # their Unicode code point value.
                return llvalue.Constant.Int(lltypes.Int32Type(), get_rune_char_code(lit.value))
            case Token.Kind.NUMLIT:
                return llvalue.Constant.Int(conv_type(lit.type), int(lit.value.replace('_', '')))

    # ---------------------------------------------------------------------------- #

    def alloc_var(self, typ: Type) -> llvalue.Value:
        '''
        Allocates a new variable in the variable block for the enclosing function.

        Params
        ------
        typ: Type
            The Chai type of the variable to allocate.

        Returns
        -------
        llvalue.Value
            The pointer to the allocated variable value.
        '''

        # Save the current block so we can move back to it.
        curr_block = self.irb.block

        # Move to before the last instruction of the variable block: namely, the
        # `br` terminator at the end of the block.
        self.irb.move_before(self.var_block.instructions.last())

        # Build the actually `alloca` instruction for the variable.
        ll_var = self.irb.build_alloca(conv_type(typ, alloc_type=True))

        # Move back to the current block we were generating in.
        self.irb.move_to_end(curr_block)

        # Return the allocated variable value pointer.
        return ll_var

    def coalesce_unit(self, value: Optional[llvalue.Value]) -> llvalue.Value:
        '''
        If the given optional LLVM value is `None`, it returns the unit value.
        Otherwise, it returns the given value  This is used to get rid of
        unwanting `None` (when a unit value is actually needed).

        Params
        ------
        value: Optional[llvalue.Value]
            The LLVM value to coalesce.
        '''

        if value:
            return value

        return self.unit_value

    def get_null_value(self, typ: Type) -> llvalue.Value:
        '''
        Returns the null value for a type.

        Params
        ------
        typ: Type
            The type whose null value to return.  This type is assumed to be
            nullable.
        '''

        # TODO handle more complex null values

        return llvalue.Constant.Null(conv_type(typ))

    # ---------------------------------------------------------------------------- #

    @property
    def loop_context(self) -> LoopContext:
        '''
        Returns the top loop context on the loop context stack.
        '''

        return self.loop_contexts[0]

    def push_loop_context(self, bb: ir.BasicBlock, cb: ir.BasicBlock):
        '''
        Pushes a new loop context onto the loop context stack.

        Params
        ------
        bb: ir.BasicBlock
            The block to jump to when a break is encountered.
        cb: ir.BasicBlock
            The block to jump to when a continue is encountered.
        '''

        self.loop_contexts.appendleft(LoopContext(bb, cb))

    def pop_loop_context(self):
        '''
        Pops the top loop context off the loop context stack.
        '''

        self.loop_contexts.popleft()

    # ---------------------------------------------------------------------------- #

    @contextlib.contextmanager
    def debug_at_span(self, span: TextSpan):
        '''
        Sets the debug location within the context controlled by the returned
        context manager: the debug location is set back to whatever value it was
        before the context was entered once the context is exited.

        Params
        ------
        span: TextSpan
            The span to use to calculate the debug location.
        '''

        prev_di_loc = self.irb.debug_location
        self.irb.debug_location = self.die.as_di_location(span)

        yield

        self.irb.debug_location = prev_di_loc


def get_rune_char_code(rune_val: str) -> int:
    '''
    Returns the Unicode character code for a rune literal.  This handles both
    regular runes and escape codes.

    Params
    ------
    rune_val: str
        The escape code, including the leading slash
    '''

    if rune_val[0] == '\\':
        # Chop off the leading `\`: it is not meaningful anymore.
        rune_val = rune_val[1:]

        # Convert the rune using the decoding "table" below.
        match rune_val:
            case 'n':
                return 10
            case '\\':
                return 92
            case 'v':
                return 11
            case 'r':
                return 13
            case '0':
                return 0
            case 'a':
                return 7
            case 'b':
                return 8
            case 'f':
                return 12
            case '\'':
                return 39
            case '\"':
                return 34
            case 't':
                return 9
            case _:
                # Not single character escape code: the value of the rune is
                # just an hexadecimal integer so we can just convert it without
                # checking the prefix :)  Note that `[1:]` is not redundant here:
                # it is chopping off the prefix code not just the `\`.
                return int(rune_val[1:], base=16) 

    # Otherwise, just use the built-in `ord(c)` to convert it.   
    return ord(rune_val)


