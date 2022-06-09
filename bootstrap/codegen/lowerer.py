from typing import List
from dataclasses import dataclass

from syntax.ast import *

from . import *
from .base_builder import BaseLLBuilder
from .dbg_builder import DebugLLBuilder

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

# ---------------------------------------------------------------------------- #

class Lowerer:
    '''
    Responsible for lowering a Chai package into an LLVM module.

    Methods
    -------
    lower() -> llmod.Module
    '''

    # The package being lowered to LLVM IR.
    pkg: Package

    # The prefix to add before every symbol to prevent linking collisions.
    pkg_prefix: str

    # The Lowerer's LLVM builder.
    b: BaseLLBuilder

    # The list of body predicates to lower.
    body_predicates: List[BodyPredicate]

    def __init__(self, pkg: Package, debug: bool = True):
        '''
        Params
        ------
        pkg: Package
            The package to convert.
        '''

        self.pkg = pkg
        self.pkg_prefix = f'p{pkg.id}.'

        self.b = DebugLLBuilder(pkg, debug) if debug else BaseLLBuilder(pkg, debug)
        self.body_predicates = []

    def lower(self) -> llmod.Module:
        '''
        Lowers the package to an LLVM module.

        Returns
        -------
        llmod.Module
            The converted module.
        '''
        
        # Lower all the definitions: more precisely, generate definition stubs
        # for each definition so that they can be easily referenced when we are
        # generating their bodies.
        for src_file in self.pkg.files:
            self.b.begin_file(src_file)

            for defin in src_file.definitions:
                self.lower_definition(defin)

        # Lower all the body predicates.
        for body_pred in self.body_predicates:
            self.lower_body_predicate(body_pred.ll_func, body_pred.func_params, body_pred.expr)

        return self.b.module

    # ---------------------------------------------------------------------------- #

    def lower_definition(self, defin: ASTNode):
        '''
        Lowers a definition in a source file.

        Params
        ------
        defin: ASTNode
            The AST node of the definition to lower.
        '''

        match defin:
            case FuncDef():
                self.lower_func_def(defin)
            case OperDef():
                pass

    def lower_func_def(self, fd: FuncDef):
        '''
        Lowers a function definition.

        Params
        ------
        fd: FuncDef
            The function definition to lower.
        '''

        # Definitions of intrinsic functions aren't lowered, but rather replaced
        # with LLVM code snippets at their callsites.
        if 'intrinsic' in fd.annots:
            return

        # Check for any annotations which would cause the name not to be
        # mangled: eg. `@extern`.
        if any(x in fd.annots for x in ['entry', 'extern']):
            mangled_name = fd.symbol.name
        else:
            mangled_name = self.pkg_prefix + fd.symbol.name

        # Create the LLVM function declaration.
        ll_func = self.b.build_func_decl(
            mangled_name,
            fd.params,
            fd.type.rt_type,
            'extern' in fd.annots,
            fd.annots.get('callconv', '')
        )

        # Update the function's symbol's LLVM value to point to the newly
        # created LLVM function.
        fd.symbol.ll_value = ll_func

        # Add the body as a predicate to lower later if it exists.
        if fd.body:
            self.body_predicates.append(BodyPredicate(ll_func, fd.params, fd.body))

    def lower_oper_def(self, od: OperDef):
        '''
        Lowers an operator definition.

        Params
        ------
        od: OperDef
            The operator definition to lower.
        '''

        # Definitions of intrinsic operators aren't lowered, but rather replaced
        # with LLVM code snippets at their applications.
        if 'intrinsic' in od.annots:
            return

        # Create the LLVM function corresponding to the operator.
        ll_func = self.b.build_func_decl(
            self.pkg_prefix + 'oper.overload.' + od.overload.id,
            od.params,
            od.type.rt_type,
        )

        # Update the operator overload's LLVM value to point to the newly
        # created LLVM function.
        od.overload.ll_value = ll_func

        # Add the body as a predicate to lower later if it exists.
        if od.body:
            self.body_predicates.append(BodyPredicate(ll_func, od.params, od.body))

    # ---------------------------------------------------------------------------- #

    def lower_body_predicate(self, ll_func: ir.Function, params: List[Symbol], body_expr: ASTNode):
        '''
        Lowers a body predicate.

        Params
        ------
        ll_func: ir.Function
            The LLVM function whose body is being lowered.
        params: List[Symbol]
            The parameters to the function.
        body_expr: ASTNode
            The body expression AST node.
        '''

        # Update the parameter names appropriately and set the parameter symbols
        # to point the values of their LLVM parameters.
        for param, ll_param in zip(params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param