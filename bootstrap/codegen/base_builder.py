from typing import List

from llvm.builder import IRBuilder

from . import *

class BaseLLBuilder:
    '''
    Acts an API adapter between the Chai compiler and the LLVM bindings. It is
    directed by the Lowerer and handles the actual generation of LLVM IR.

    Methods
    -------
    finalize() -> llmod.Module
        Finalizes generation and returns the resultant LLVM module.
    '''

    # The package being used to generate LLVM IR.
    pkg: Package

    # The LLVM module being built.
    mod: llmod.Module

    # The LLVM IR builder.
    irb: IRBuilder

    def __init__(self, pkg: Package, debug: bool) -> 'BaseLLBuilder':
        '''
        Params
        ------
        pkg: Package
            The package being used to generate LLVM IR.
        '''
        
        self.pkg = pkg
        self.mod = llmod.Module(pkg.name)
        
        self.irb = IRBuilder()

    def finalize(self) -> llmod.Module:
        if self.dib:
            self.dib.finalize()

        return self.mod

    # ---------------------------------------------------------------------------- #

    def begin_file(self, src_file: SourceFile):
        '''Builds the LLVM prelude for a new file.'''

        # TODO import generation

    def build_func_decl(
        self, 
        name: str, 
        params: List[Symbol], 
        rt_type: Type,
        external: bool = False,
        call_conv = ''
    ) -> ir.Function:
        '''
        Builds an LLVM function declaration.

        Params
        ------
        name: str
            The mangled name of the function.
        params: List[Symbol],
            The list of parameters to the function.
        rt_type: Type
            The return type of the function.
        external: bool
            Whether the function is defined externally.
        call_conv: str
            The name of the calling convention of the function.  If the value is
            an empty string, then the default Chai calling convention is used.
        '''

        ll_func_type = lltypes.FunctionType(
            [self.convert_type(x.type) for x in params], 
            self.convert_type(rt_type)
        )
        ll_func = self.mod.add_function(name, ll_func_type) 

        if external:
            ll_func.linkage = llvalue.Linkage.EXTERNAL
        else:
            ll_func.linkage = llvalue.Linkage.INTERNAL

        match call_conv:
            case 'win64':
                ll_func.call_conv = ir.CallConv.WIN64

        return ll_func

    # ---------------------------------------------------------------------------- #

    def convert_type(self, typ: Type) -> lltypes.Type:
        match typ:
            case PrimitiveType():
                match typ:
                    case PrimitiveType.BOOL | PrimitiveType.NOTHING:
                        return lltypes.Int1Type()
                    case PrimitiveType.U8 | PrimitiveType.I8:
                        return lltypes.Int8Type()
                    case PrimitiveType.I16 | PrimitiveType.U16:
                        return lltypes.Int16Type()
                    case PrimitiveType.I32 | PrimitiveType.U32:
                        return lltypes.Int32Type()
                    case PrimitiveType.I64 | PrimitiveType.U64:
                        return lltypes.Int64Type()
                    case PrimitiveType.F32:
                        return lltypes.FloatType()
                    case PrimitiveType.F64:
                        return lltypes.DoubleType()
            case PointerType(elem_type):
                return lltypes.PointerType(self.convert_type(elem_type))
            case _:
                raise NotImplementedError()
        
            

