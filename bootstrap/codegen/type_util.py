from typecheck import *
import llvm.types as lltypes

def conv_type(typ: Type, alloc_type: bool = False, rt_type: bool = False) -> lltypes.Type:
    match inner_typ := typ.inner_type():
        case PrimitiveType():
            if rt_type and is_unit(inner_typ):
                return lltypes.VoidType()

            return conv_prim_type(inner_typ)
        case PointerType(elem_type):
            return lltypes.PointerType(conv_type(elem_type))
        case _:
            raise NotImplementedError()

def conv_prim_type(prim_typ: PrimitiveType) -> lltypes.Type:
    match prim_typ:
        case PrimitiveType.BOOL | PrimitiveType.UNIT:
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

def is_unit(typ: Type) -> bool:
    return typ == PrimitiveType.UNIT