from ctypes import c_char_p
from typing import Iterator, Optional

from . import *
from .value import UserValue, Value

class OpCode(LLVMEnum):
   RET              = 1
   BR               = 2
   SWITCH           = 3
   INDIRECT_BR      = 4
   INVOKE           = 5
   UNREACHABLE      = 7
   CALL_BR          = 67
   F_NEG            = 66
   ADD              = 8
   F_ADD            = 9
   SUB              = 10
   F_SUB            = 11
   MUL              = 12
   F_MUL            = 13
   U_DIV            = 14
   S_DIV            = 15
   F_DIV            = 16
   U_REM            = 17
   S_REM            = 18
   F_REM            = 19
   SHL              = 20
   L_SHR            = 21
   A_SHR            = 22
   AND              = 23
   OR               = 24
   XOR              = 25
   ALLOCA           = 26
   LOAD             = 27
   STORE            = 28
   GET_ELEMENT_PTR  = 29
   TRUNC            = 30
   Z_EXT            = 31
   S_EXT            = 32
   FP_TO_UI         = 33
   FP_TO_SI         = 34
   UI_TO_FP         = 35
   SI_TO_FP         = 36
   FP_TRUNC         = 37
   FP_EXT           = 38
   PTR_TO_INT       = 39
   INT_TO_PTR       = 40
   BIT_CAST         = 41
   ADDR_SPACE_CAST  = 60
   I_CMP            = 42
   F_CMP            = 43
   PHI              = 44
   CALL             = 45
   SELECT           = 46
   USER_OP1         = 47
   USER_OP2         = 48
   VA_ARG           = 49
   EXTRACT_ELEMENT  = 50
   INSERT_ELEMENT   = 51
   SHUFFLE_VECTOR   = 52
   EXTRACT_VALUE    = 53
   INSERT_VALUE     = 54
   FREEZE           = 68
   FENCE            = 55
   ATOMIC_CMP_XCHG  = 56
   ATOMIC_RMW       = 57
   RESUME           = 58
   LANDING_PAD      = 59
   CLEANUP_RET      = 61
   CATCH_RET        = 62
   CATCH_PAD        = 63
   CLEANUP_PAD      = 64
   CATCH_SWITCH     = 65

class Instruction(UserValue):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def is_terminator(self) -> bool:
        return bool(LLVMIsATerminatorInst(self))

    @property
    def opcode(self) -> OpCode:
        return OpCode(LLVMGetInstructionOpcode(self))

class BasicBlock(Value):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @staticmethod
    def from_value(v: Value) -> 'BasicBlock':
        assert LLVMValueIsBasicBlock(v) == 1

        return BasicBlock(LLVMValueAsBasicBlock(v))

    @property
    def terminator(self) -> Optional[Instruction]:
        term_ptr = LLVMGetBasicBlockTerminator(self)

        if term_ptr:
            return Instruction(term_ptr)
        
        return None

    class _Instructions:
        _bb: 'BasicBlock'

        def __init__(self, bb: 'BasicBlock'):
            self._bb = bb

        def __iter__(self) -> Iterator[Instruction]:
            instr_ptr = LLVMGetFirstInstruction(self._bb)

            while instr_ptr:
                yield Instruction(instr_ptr)

                instr_ptr = LLVMGetNextInstruction(instr_ptr)

        def __reversed__(self) -> Iterator[Instruction]:
            instr_ptr = LLVMGetLastInstruction(self._bb)

            while instr_ptr:
                yield Instruction(instr_ptr)

                instr_ptr = LLVMGetPreviousInstruction(instr_ptr)

        def first(self) -> Optional[Instruction]:
            instr_ptr = LLVMGetFirstInstruction(self._bb)

            if instr_ptr:
                return Instruction(instr_ptr)

            return None

        def last(self) -> Optional[Instruction]:
            instr_ptr = LLVMGetLastInstruction(self._bb)

            if instr_ptr:
                return Instruction(instr_ptr)

            return None
        
    @property
    def instructions(self) -> _Instructions:
        return BasicBlock._Instructions(self)

# ---------------------------------------------------------------------------- #

@llvm_api
def LLVMValueIsBasicBlock(v: Value) -> c_enum:
    pass

@llvm_api
def LLVMValueAsBasicBlock(v: Value) -> c_object_p:
    pass

@llvm_api
def LLVMGetBasicBlockTerminator(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetFirstInstruction(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetNextInstruction(bb: Instruction) -> c_object_p:
    pass

@llvm_api
def LLVMGetLastInstruction(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetPreviousInstruction(instr: Instruction) -> c_object_p:
    pass

@llvm_api
def LLVMIsATerminatorInst(instr: Instruction) -> c_enum:
    pass

@llvm_api
def LLVMGetInstructionOpcode(instr: Instruction) -> c_enum:
    pass