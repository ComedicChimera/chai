from ctypes import c_uint, POINTER, c_char_p

from . import *
from .ir import *
from .value import Value
from .types import Type
from .metadata import DILocation

class IRBuilder(LLVMObject):
    def __init__(self):
        ctx = get_context()
        super().__init__(LLVMCreateBuilderInContext(ctx))
        ctx.take_ownership(self)

    def dispose(self):
        LLVMDisposeBuilder(self)

    @property
    def block(self) -> BasicBlock:
        return BasicBlock(LLVMGetInsertBlock(self))

    def move_before(self, instr: Instruction):
        LLVMPositionBuilderBefore(self, instr)

    def move_after(self, instr: Instruction):
        LLVMPositionBuilder(self, instr.parent, instr.ptr)

    def move_to_start(self, bb: BasicBlock):
        first = bb.instructions.first()

        if first:
            LLVMPositionBuilderBefore(self, first)
        else:
            LLVMPositionBuilder(self, bb, c_object_p())

    def move_to_end(self, bb: BasicBlock):
        LLVMPositionBuilderAtEnd(self, bb)

    @property
    def debug_location(self) -> Optional[DILocation]:
        if ptr := LLVMGetCurrentDebugLocation2(self):
            return DILocation(None, 0, 0, ptr=ptr)

        return None

    @debug_location.setter
    def debug_location(self, new_loc: Optional[DILocation]):
        LLVMSetCurrentDebugLocation2(self, new_loc.ptr if new_loc else None)

    # ---------------------------------------------------------------------------- #

    def build_ret(self, *values: Value) -> Terminator:
        match len(values):
            case 0:
                return Terminator(LLVMBuildRetVoid(self))
            case 1:
                return Terminator(LLVMBuildRet(self, values[0]))
            case n:
                value_arr = (c_object_p * n)(*(v.ptr for v in values))
                return Terminator(LLVMBuildAggregateRet(self, value_arr, n))

    def build_br(self, dest: BasicBlock) -> Terminator:
        return Terminator(LLVMBuildBr(self, dest))

    def build_cond_br(self, cond: Value, then_block: BasicBlock, else_block: BasicBlock) -> Terminator:
        return Terminator(LLVMBuildCondBr(self, cond, then_block, else_block))

    def build_switch(self, val: Value, default_block: BasicBlock, num_cases: int = 10) -> SwitchInstruction:
        return SwitchInstruction(LLVMBuildSwitch(self, val, default_block, num_cases))

    def build_unreachable(self) -> Terminator:
        return Terminator(LLVMBuildUnreachable(self))

    def build_add(self, lhs: Value, rhs: Value, nsw: bool = False, nuw: bool = False, name: str = "") -> Instruction:
        if nsw:
            return Instruction(LLVMBuildNSWAdd(self, lhs, rhs, name.encode()))
        elif nuw:
            return Instruction(LLVMBuildNUWAdd(self, lhs, rhs, name.encode()))
        else:
            return Instruction(LLVMBuildAdd(self, lhs, rhs, name.encode()))

    def build_sub(self, lhs: Value, rhs: Value, nsw: bool = False, nuw: bool = False, name: str = "") -> Instruction:
        if nsw:
            return Instruction(LLVMBuildNSWSub(self, lhs, rhs, name.encode()))
        elif nuw:
            return Instruction(LLVMBuildNUWSub(self, lhs, rhs, name.encode()))
        else:
            return Instruction(LLVMBuildSub(self, lhs, rhs, name.encode()))
    
    def build_mul(self, lhs: Value, rhs: Value, nsw: bool = False, nuw: bool = False, name: str = "") -> Instruction:
        if nsw:
            return Instruction(LLVMBuildNSWMul(self, lhs, rhs, name.encode()))
        elif nuw:
            return Instruction(LLVMBuildNUWMul(self, lhs, rhs, name.encode()))
        else:
            return Instruction(LLVMBuildMul(self, lhs, rhs, name.encode()))

    def build_udiv(self, lhs: Value, rhs: Value, exact: bool = False, name: str = "") -> Instruction:
        if exact:
            return Instruction(LLVMBuildExactUDiv(self, lhs, rhs, name.encode()))
        else:
            return Instruction(LLVMBuildUDiv(self, lhs, rhs, name.encode()))

    def build_sdiv(self, lhs: Value, rhs: Value, exact: bool = False, name: str = "") -> Instruction:
        if exact:
            return Instruction(LLVMBuildExactSDiv(self, lhs, rhs, name.encode()))
        else:
            return Instruction(LLVMBuildSDiv(self, lhs, rhs, name.encode()))  

    def build_urem(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildURem(self, lhs, rhs, name.encode()))  

    def build_srem(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildSRem(self, lhs, rhs, name.encode()))   

    def build_fadd(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFAdd(self, lhs, rhs, name.encode()))  

    def build_fsub(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFSub(self, lhs, rhs, name.encode())) 

    def build_fmul(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFMul(self, lhs, rhs, name.encode())) 

    def build_fdiv(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFDiv(self, lhs, rhs, name.encode())) 

    def build_frem(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFRem(self, lhs, rhs, name.encode()))

    def build_shl(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildShl(self, lhs, rhs, name.encode())) 

    def build_ashr(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildAShr(self, lhs, rhs, name.encode())) 

    def build_lshr(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildLShr(self, lhs, rhs, name.encode())) 

    def build_and(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildAnd(self, lhs, rhs, name.encode())) 

    def build_or(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildOr(self, lhs, rhs, name.encode())) 

    def build_xor(self, lhs: Value, rhs: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildXor(self, lhs, rhs, name.encode())) 

    def build_neg(self, v: Value, nsw: bool = False, nuw: bool = False, name: str = "") -> Instruction:
        if nsw:
            return Instruction(LLVMBuildNSWNeg(self, v, name.encode()))
        elif nuw:
            return Instruction(LLVMBuildNUWNeg(self, v, name.encode()))
        else:
            return Instruction(LLVMBuildNeg(self, v, name.encode()))

    def build_fneg(self, v: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFNeg(self, v, name.encode()))

    def build_not(self, v: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildNot(self, v, name.encode()))

    def build_alloca(self, typ: Type, name: str = "") -> AllocaInstruction:
        return AllocaInstruction(LLVMBuildAlloca(self, typ, name.encode()))

    def build_load(self, typ: Type, ptr: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildLoad2(self, typ, ptr, name.encode()))

    def build_store(self, v: Value, ptr: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildStore(self, v, ptr, name.encode()))

    def build_gep(self, typ: Type, ptr: Value, *indices: Value, name: str = "", in_bounds: bool = False) -> GEPInstruction:
        assert len(indices) > 0
        
        indices_arr_type = (c_object_p * len(indices))
        indices_arr = indices_arr_type(*(x.ptr for x in indices))

        if in_bounds:
            return GEPInstruction(LLVMBuildInBoundsGEP2(self, typ, ptr, indices_arr, len(indices_arr), name.encode()))
        else:
            return GEPInstruction(LLVMBuildGEP2(self, typ, ptr, indices_arr, len(indices_arr), name.encode()))

    def build_struct_gep(self, typ: Type, ptr: Value, ndx: int, name: str = "") -> GEPInstruction:
        return GEPInstruction(LLVMBuildStructGEP2(self, typ, ptr, ndx, name.encode()))

    def build_trunc(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildTrunc(self, src, dest_typ, name.encode()))

    def build_zext(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildZExt(self, src, dest_typ, name.encode()))

    def build_sext(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildSExt(self, src, dest_typ, name.encode()))

    def build_fp_trunc(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFPTrunc(self, src, dest_typ, name.encode()))

    def build_fp_to_ui(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFPToUI(self, src, dest_typ, name.encode()))

    def build_fp_to_si(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildFPToSI(self, src, dest_typ, name.encode()))

    def build_ui_to_fp(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildUIToFP(self, src, dest_typ, name.encode()))

    def build_si_to_fp(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildSIToFP(self, src, dest_typ, name.encode()))

    def build_ptr_to_int(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildPtrToInt(self, src, dest_typ, name.encode()))

    def build_int_to_ptr(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildIntToPtr(self, src, dest_typ, name.encode()))

    def build_bit_cast(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildBitCast(self, src, dest_typ, name.encode()))

    def build_addr_space_cast(self, src: Value, dest_typ: Type, name: str = "") -> Instruction:
        return Instruction(LLVMBuildAddrSpaceCast(self, src, dest_typ, name.encode()))

    def build_icmp(self, predicate: IntPredicate, lhs: Value, rhs: Value, name: str = "") -> ICmpInstruction:
        return ICmpInstruction(LLVMBuildICmp(self, predicate, lhs, rhs, name.encode()))

    def build_fcmp(self, predicate: RealPredicate, lhs: Value, rhs: Value, name: str = "") -> FCmpInstruction:
        return FCmpInstruction(LLVMBuildFCmp(self, predicate, lhs, rhs, name.encode()))

    def build_phi(self, typ: Type, name: str = "") -> PHINodeInstruction:
        return PHINodeInstruction(LLVMBuildPhi(self, typ, name.encode()))

    def build_call(self, func: Value, *args: Value, name: str = "") -> CallInstruction:
        if len(args) == 0:
            args_arr = (c_object_p * 0)()
        else:
            args_arr_type = c_object_p * len(args)
            args_arr = args_arr_type(*(x.ptr for x in args))

        return CallInstruction(LLVMBuildCall2(self, func.type, func, args_arr, len(args), name.encode()))

    def build_is_null(self, val: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildIsNull(self, val, name.encode()))

    def build_is_not_null(self, val: Value, name: str = "") -> Instruction:
        return Instruction(LLVMBuildIsNotNull(self, val, name.encode()))

    def build_insert_value(self, struct_val: Value, field_ndx: int, field_val: Value, name: str = "") -> InsertValueInstruction:
        return InsertValueInstruction(LLVMBuildInsertValue(struct_val, field_val, field_ndx, name.encode()))

    def build_extract_value(self, struct_val: Value, field_ndx: int, name: str = "") -> Instruction:
        return LLVMBuildExtractValue(struct_val, field_ndx, name.encode())

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMCreateBuilderInContext(ctx: Context) -> c_object_p:
    pass

@llvm_api
def LLVMPositionBuilder(b: IRBuilder, bb: BasicBlock, instr: c_object_p):
    pass

@llvm_api
def LLVMPositionBuilderBefore(b: IRBuilder, instr: Instruction):
    pass

@llvm_api
def LLVMPositionBuilderAtEnd(b: IRBuilder, bb: BasicBlock):
    pass

@llvm_api
def LLVMGetInsertBlock(b: IRBuilder) -> c_object_p:
    pass

@llvm_api
def LLVMDisposeBuilder(b: IRBuilder):
    pass

# ---------------------------------------------------------------------------- #

# NOTE Most of these functions were generated automatically using a scuffed
# script I wrote so I wouldn't lose my mind so if the names look weird that is
# why.

@llvm_api
def LLVMBuildRetVoid(b: IRBuilder) -> c_object_p:
    pass

@llvm_api
def LLVMBuildRet(b: IRBuilder, v: Value) -> c_object_p:
    pass

@llvm_api
def LLVMBuildAggregateRet(b: IRBuilder, values: POINTER(c_object_p), n: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMBuildBr(p0: IRBuilder, dest: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMBuildCondBr(p0: IRBuilder, _if: Value, then: BasicBlock, _else: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMBuildSwitch(p0: IRBuilder, v: Value, _else: BasicBlock, num_cases: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMBuildUnreachable(p0: IRBuilder) -> c_object_p:
    pass

@llvm_api
def LLVMBuildAdd(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNSWAdd(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNUWAdd(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFAdd(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildSub(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNSWSub(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNUWSub(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFSub(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildMul(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNSWMul(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNUWMul(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFMul(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildUDiv(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildExactUDiv(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildSDiv(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildExactSDiv(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFDiv(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildURem(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildSRem(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFRem(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildShl(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildLShr(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildAShr(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildAnd(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildOr(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildXor(p0: IRBuilder, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNeg(p0: IRBuilder, v: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNSWNeg(b: IRBuilder, v: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNUWNeg(b: IRBuilder, v: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFNeg(p0: IRBuilder, v: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildNot(p0: IRBuilder, v: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildAlloca(p0: IRBuilder, ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildLoad2(p0: IRBuilder, ty: Type, pointer_val: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildStore(p0: IRBuilder, val: Value, ptr: Value) -> c_object_p:
    pass

@llvm_api
def LLVMBuildGEP2(b: IRBuilder, ty: Type, pointer: Value, indices: POINTER(c_object_p), num_indices: c_uint, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildInBoundsGEP2(b: IRBuilder, ty: Type, pointer: Value, indices: POINTER(c_object_p), num_indices: c_uint, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildStructGEP2(b: IRBuilder, ty: Type, pointer: Value, idx: c_uint, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildGlobalString(b: IRBuilder, str: c_char_p, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildGlobalStringPtr(b: IRBuilder, str: c_char_p, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildTrunc(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildZExt(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildSExt(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFPToUI(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFPToSI(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildUIToFP(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildSIToFP(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFPTrunc(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFPExt(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildPtrToInt(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildIntToPtr(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildBitCast(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildAddrSpaceCast(p0: IRBuilder, val: Value, dest_ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildICmp(p0: IRBuilder, op: IntPredicate, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildFCmp(p0: IRBuilder, op: RealPredicate, l_h_s: Value, r_h_s: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildPhi(p0: IRBuilder, ty: Type, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildCall2(p0: IRBuilder, p1: Type, fn: Value, args: POINTER(c_object_p), num_args: c_uint, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildIsNull(p0: IRBuilder, val: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildIsNotNull(p0: IRBuilder, val: Value, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMGetCurrentDebugLocation2(builder: IRBuilder) -> c_object_p:
    pass

@llvm_api
def LLVMSetCurrentDebugLocation2(builder: IRBuilder, loc: c_object_p):
    pass

@llvm_api
def LLVMBuildInsertValue(builder: IRBuilder, agg_val: Value, elt_val: Value, ndx: c_uint, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMBuildExtractValue(builder: IRBuilder, agg_val: Value, ndx: c_uint, name: c_char_p) -> c_object_p:
    pass