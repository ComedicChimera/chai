from ctypes import c_char_p, c_uint, POINTER, byref, c_uint64
from dataclasses import dataclass
from typing import Iterator, Optional, Dict
from enum import Enum, auto
from abc import ABC, abstractmethod

from . import *
from .types import Type
from .value import UserValue, Value, GlobalValue
from .metadata import DISubprogram

class Attribute(LLVMObject):
    class Kind(LLVMEnum):
        ALWAYS_INLINE = 1
        ARG_MEM_ONLY = 2
        BUILTIN = 3
        COLD = 4
        CONVERGENT = 5
        DISABLE_SANITIZER_INSTRUMENTATION = 6
        HOT = 7
        IMM_ARG = 8
        IN_REG = 9
        INACCESSIBLE_MEM_ONLY = 10
        INACCESSIBLE_MEM_OR_ARG_MEM_ONLY = 11
        INLINE_HINT = 12
        JUMP_TABLE = 13
        MIN_SIZE = 14
        MUST_PROGRESS = 15
        NAKED = 16
        NEST = 17
        NO_ALIAS = 18
        NO_BUILTIN = 19
        NO_CALLBACK = 20
        NO_CAPTURE = 21
        NO_CF_CHECK = 22
        NO_DUPLICATE = 23
        NO_FREE = 24
        NO_IMPLICIT_FLOAT = 25
        NO_INLINE = 26
        NO_MERGE = 27
        NO_PROFILE = 28
        NO_RECURSE = 29
        NO_RED_ZONE = 30
        NO_RETURN = 31
        NO_SANITIZE_COVERAGE = 32
        NO_SYNC = 33
        NO_UNDEF = 34
        NO_UNWIND = 35
        NON_LAZY_BIND = 36
        NON_NULL = 37
        NULL_POINTER_IS_VALID = 38
        OPT_FOR_FUZZING = 39
        OPTIMIZE_FOR_SIZE = 40
        OPTIMIZE_NONE = 41
        READ_NONE = 42
        READ_ONLY = 43
        RETURNED = 44
        RETURNS_TWICE = 45
        S_EXT = 46
        SAFE_STACK = 47
        SANITIZE_ADDRESS = 48
        SANITIZE_HW_ADDRESS = 49
        SANITIZE_MEM_TAG = 50
        SANITIZE_MEMORY = 51
        SANITIZE_THREAD = 52
        SHADOW_CALL_STACK = 53
        SPECULATABLE = 54
        SPECULATIVE_LOAD_HARDENING = 55
        STACK_PROTECT = 56
        STACK_PROTECT_REQ = 57
        STACK_PROTECT_STRONG = 58
        STRICT_FP = 59
        SWIFT_ASYNC = 60
        SWIFT_ERROR = 61
        SWIFT_SELF = 62
        UW_TABLE = 63
        WILL_RETURN = 64
        WRITE_ONLY = 65
        Z_EXT = 66
        LAST_ENUM_ATTR = 66
        FIRST_TYPE_ATTR = 67
        BY_REF = 67
        BY_VAL = 68
        ELEMENT_TYPE = 69
        IN_ALLOCA = 70
        PREALLOCATED = 71
        STRUCT_RET = 72
        LAST_TYPE_ATTR = 72
        FIRST_INT_ATTR = 73
        ALIGNMENT = 73
        ALLOC_SIZE = 74
        DEREFERENCEABLE = 75
        DEREFERENCEABLE_OR_NULL = 76
        STACK_ALIGNMENT = 77
        V_SCALE_RANGE = 78

    class Variant(Enum):
        ENUM = auto()
        STRING = auto()
        TYPE = auto()

    _variant: Variant

    def __init__(self, **kwargs):
        if ptr := kwargs.get('ptr'):
            assert isinstance(ptr, c_object_p)

            super().__init__(ptr)

            if LLVMIsEnumAttribute(self) == 1:
                self._variant = Attribute.Variant.ENUM
            elif LLVMIsStringAttribute(self) == 1:
                self._variant = Attribute.Variant.STRING
            else:
                self._variant = Attribute.Variant.TYPE
        else:
            assert (kind := kwargs.get('kind'))

            ctx = kwargs.get('context', Context.global_ctx())

            if isinstance(kind, str):
                assert (value := kwargs.get('kind'))

                value_bytes = value.encode()
                super().__init__(LLVMCreateStringAttribute(ctx, kind, value_bytes, len(value_bytes)))
                self._variant = Attribute.Variant.STRING
            else:
                assert isinstance(kind, Attribute.Kind)

                if value := kwargs.get('value'):
                    if isinstance(value, Type):
                        super().__init__(LLVMCreateTypeAttribute(ctx, kind, value))
                        self._variant = Attribute.Variant.TYPE
                        return

                    super().__init__(LLVMCreateEnumAttribute(ctx, kind, value))
                else:
                    super().__init__(LLVMCreateEnumAttribute(ctx, kind, 0))

                self._variant = Attribute.Variant.ENUM

    @property
    def kind(self) -> Kind | str:
        match self._variant:
            case Attribute.Variant.ENUM | Attribute.Variant.TYPE:
                return Attribute.Kind(LLVMGetEnumAttributeKind(self))
            case Attribute.Variant.STRING:
                length = c_uint()
                return str(LLVMGetStringAttributeKind(self, byref(length)), encoding='utf-8')

    @property
    def value(self) -> int | str | Type:
        match self._variant:
            case Attribute.Variant.ENUM:
                return LLVMGetEnumAttributeValue(self)
            case Attribute.Variant.TYPE:
                ptr = LLVMGetTypeAttributeValue(self)
                return Type(ptr)
            case Attribute.Variant.STRING:
                length = c_uint()
                return str(LLVMGetStringAttributeValue(self, byref(length)), encoding='utf-8')

    @property
    def variant(self) -> Variant:
        return self._variant 

class AttributeSet(ABC):
    KeyType = Attribute.Kind | str 

    _ndx: int
    _attr_dict: Dict[KeyType, Attribute]

    def __init__(self, ndx: int):
        self._ndx = ndx

        attr_arr_len = self._get_attr_count()
        attr_arr = (c_object_p * attr_arr_len)()
        self._get_attrs(attr_arr)

        self._attr_dict = {attr.kind: attr for attr in map(lambda p: Attribute(ptr=p), attr_arr)} 

    def __len__(self) -> int:
        return len(self._attr_dict)

    def __getitem__(self, key: KeyType) -> Attribute:
        return self._attr_dict(key)

    def __iter__(self) -> Iterator[Attribute]:
        return self._attr_dict.values()

    def add(self, attr: Attribute):
        self._attr_dict[attr.kind] = attr
        self._add_attr(attr)

    def remove(self, key: KeyType):
        del self._attr_dict[key]

        if isinstance(key, str):
            key_bytes = key.encode()
            self._remove_str_attr(key_bytes)
        else:
            self._remove_enum_attr(key)
            
    @abstractmethod
    def _get_attr_count(self) -> int:
        pass

    @abstractmethod
    def _get_attrs(self, attr_arr: POINTER(c_object_p)):
        pass

    @abstractmethod
    def _add_attr(self, attr: Attribute):
        pass

    @abstractmethod
    def _remove_str_attr(self, key_bytes: bytes):
        pass

    @abstractmethod
    def _remove_enum_attr(self, key: Attribute.Kind):
        pass

# ---------------------------------------------------------------------------- #

class CallConv(LLVMEnum):
    C = 0
    FAST = 8
    COLD = 9
    GHC = 10
    HIPE = 11
    WEBKIT_JS = 12
    ANY_REG = 13
    PRESERVE_MOST = 14
    PRESERVEALL = 15
    SWIFT = 16
    CXX_FAST_TLS = 17
    TAIL = 18,
    CFGUARD_CHECK = 19
    SWIFTTAIL = 20
    FIRSTTARGETCC = 64
    X86_STDCALL = 64
    X86_FASTCALL = 65
    ARM_APCS = 66
    ARM_AAPCS = 67
    ARM_AAPCS_VFP = 68,
    MSP430_INTR = 69
    X86_THISCALL = 70
    PTX_KERNEL = 71
    PTX_DEVICE = 72
    SPIR_FUNC = 75
    SPIR_KERNEL = 76
    INTEL_OCL_BI = 77
    X86_64_SYSV = 78
    WIN64 = 79
    X86_VECTORCALL = 80
    HHVM = 81
    HHVM_C = 82
    X86_INTR = 83
    AVR_INTR = 84
    AVR_SIGNAL = 85
    AVR_BUILTIN = 86
    AMDGPU_VS = 87
    AMDGPU_GS = 88
    AMDGPU_PS = 89
    AMDGPU_CS = 90,
    AMDGPU_KERNEL = 91
    X86_REGCALL = 92
    AMDGPU_HS = 93
    MSP430_BUILTIN = 94
    AMDGPU_LS = 95
    AMDGPU_ES = 96
    AARCH64_VECTORCALL = 97
    AARCH64_SVE_VECTORCALL = 98
    WASM_EMSCRIPTEN_INVOKE = 99
    AMDGPU_GFX = 100
    M68K_INTR = 101
    MAXID = 1023

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

    @property
    def parent(self) -> 'BasicBlock':
        return BasicBlock(LLVMGetInstructionParent(self))

class IntPredicate(LLVMEnum):
    EQ = 32
    NE = auto()
    UGT = auto()
    UGE = auto()
    ULT = auto()
    ULE = auto()
    SGT = auto()
    SGE = auto()
    SLT = auto()
    SLE = auto()

class ICmpInstruction(Instruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def predicate(self) -> IntPredicate:
        return IntPredicate(LLVMGetICmpPredicate(self))

    @staticmethod
    def from_instr(instr: Instruction) -> 'ICmpInstruction':
        assert instr.opcode == OpCode.I_CMP

        return ICmpInstruction(instr.ptr)

class RealPredicate(LLVMEnum):
    OEQ = 1
    OGT = auto()
    OGE = auto()           
    OLT = auto()    
    OLE = auto()            
    ONE = auto()            
    ORD = auto()
    UNO = auto()            
    UEQ = auto()
    UGT = auto()        
    UGE = auto()            
    ULT = auto()     
    ULE = auto()            
    UNE = auto()     

class FCmpInstruction(Instruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def predicate(self) -> RealPredicate:
        return RealPredicate(LLVMGetFCmpPredicate(self))

    @staticmethod
    def from_instr(instr: Instruction) -> 'FCmpInstruction':
        return FCmpInstruction(instr.ptr)

class CallSiteAttributeSet(AttributeSet):
    _call_instr: 'CallInstruction'

    def __init__(self, call_instr: 'CallInstruction', ndx: int):
        self._call_instr = call_instr
        super().__init__(ndx)

    def _get_attr_count(self) -> int:
        return LLVMGetCallSiteAttributeCount(self._call_instr, self._ndx)

    def _get_attrs(self, attr_arr: POINTER(c_object_p)):
        LLVMGetCallSiteAttributes(self._call_instr, self._ndx, attr_arr)

    def _add_attr(self, attr: Attribute):
        LLVMAddCallSiteAttribute(self._call_instr, self._ndx, attr)

    def _remove_enum_attr(self, key: Attribute.Kind):
        LLVMRemoveCallSiteEnumAttribute(self._call_instr, self._ndx, key)

    def _remove_str_attr(self, key_bytes: bytes):
        LLVMRemoveCallSiteStringAttribute(self._call_instr, self._ndx, key_bytes, len(key_bytes))

class CallInstruction(Instruction):
    _func_attr_set: CallSiteAttributeSet
    _rt_attr_set: CallSiteAttributeSet
    _param_attr_sets: Dict[int, CallSiteAttributeSet]

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)
        self._func_attr_set = CallSiteAttributeSet(self, 0)
        self._rt_attr_set = CallSiteAttributeSet(self, -1)

    @property
    def func(self) -> 'Function':
        return Function(LLVMGetCalledValue(self))

    @property
    def call_conv(self) -> CallConv:
        return CallConv(LLVMGetInstructionCallConv(self))

    @call_conv.setter
    def call_conv(self, cc: CallConv):
        LLVMSetInstructionCallConv(self, cc)

    @property
    def is_tail_call(self) -> bool:
        return LLVMIsTailCall(self) == 1

    @is_tail_call.setter
    def is_tail_call(self, new_tail_call: bool):
        LLVMSetTailCall(self, int(new_tail_call))

    @property
    def func_attrs(self) -> CallSiteAttributeSet:
        return self._func_attr_set

    @property
    def rt_attrs(self) -> CallSiteAttributeSet:
        return self._rt_attr_set

    def param_attrs(self, ndx: int) -> CallSiteAttributeSet:
        if param_attr_set := self._param_attr_sets.get(ndx):
            return param_attr_set

        param_attr_set = CallSiteAttributeSet(self, ndx)
        self._param_attr_sets = param_attr_set
        return param_attr_set

class Terminator(Instruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    class _Successors:
        _term: 'Terminator'

        def __init__(self, term: 'Terminator'):
            self._term = term

        def __len__(self) -> int:
            return LLVMGetNumSuccessors(self._term)

        def __getitem__(self, ndx: int) -> 'BasicBlock':
            if 0 <= ndx < len(self):
                return BasicBlock(LLVMGetSuccessor(self._term, ndx))

            raise IndexError()

        def __setitem__(self, ndx: int, bb: 'BasicBlock'):
            if 0 <= ndx < len(self):
                LLVMSetSuccessor(self._term, ndx, bb)

        def __iter__(self) -> Iterator['BasicBlock']:
            for i in range(len(self)):
                yield BasicBlock(LLVMGetSuccessor(self._term, i))

        def __reversed__(self) -> Iterator['BasicBlock']:
            for i in range(len(self)-1, -1, -1):
                yield BasicBlock(LLVMGetSuccessor(self._term, i))

    @property
    def successors(self) -> _Successors:
        return Terminator._Successors(self)

    @property
    def is_conditional(self) -> bool:
        return LLVMIsConditional(self) == 1

    @staticmethod
    def from_instr(instr: Instruction) -> 'Terminator':
        assert instr.is_terminator

        return Terminator(instr.ptr)

class ConditionalInstruction(Terminator):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def condition(self) -> Value:
        return Value(LLVMGetCondition(self))

    @condition.setter
    def condition(self, new_cond: Value):
        LLVMSetCondition(self, new_cond)

    @staticmethod
    def from_term(term: Terminator) -> 'ConditionalInstruction':
        assert term.is_conditional

        return ConditionalInstruction(term.ptr)

class SwitchInstruction(ConditionalInstruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    def add_case(self, case_val: Value, case_block: 'BasicBlock'):
        LLVMAddCase(self, case_val, case_block)

    @property
    def default(self) -> 'BasicBlock':
        return BasicBlock(LLVMGetSwitchDefaultDest(self))

    @staticmethod
    def from_cond_instr(cond_instr: ConditionalInstruction) -> 'SwitchInstruction':
        assert cond_instr.opcode == OpCode.SWITCH

        return SwitchInstruction(cond_instr.ptr)

class AllocaInstruction(Instruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def allocated_type(self) -> Type:
        return Type(LLVMGetAllocatedType(self))
    
    @staticmethod
    def from_instr(instr: Instruction) -> 'AllocaInstruction':
        assert instr.opcode == OpCode.ALLOCA

        return AllocaInstruction(instr.ptr)

class GEPInstruction(Instruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    @property
    def in_bounds(self) -> bool:
        return LLVMIsInBounds(self) == 1

    @in_bounds.setter
    def in_bounds(self, in_bounds: bool):
        LLVMSetIsInBounds(self, int(in_bounds))

    @staticmethod
    def from_instr(instr: Instruction) -> 'GEPInstruction':
        assert instr.opcode == OpCode.GET_ELEMENT_PTR

        return GEPInstruction(instr.ptr)

@dataclass
class PHIIncoming:
    value: Value
    block: 'BasicBlock'

class PHINodeInstruction(Instruction):
    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

    class _IncomingList:
        _pni: 'PHINodeInstruction'

        def __init__(self, pni: 'PHINodeInstruction'):
            self._pni = pni

        def __len__(self) -> int:
            return LLVMCountIncoming(self._pni)

        def __getitem__(self, ndx: int) -> PHIIncoming:
            if 0 <= ndx < len(self):
                return self._get_at(ndx)

            raise IndexError()

        def __iter__(self) -> Iterator[PHIIncoming]:
            for i in range(len(self)):
                yield self._get_at(i)

        def __reversed__(self) -> Iterator[PHIIncoming]:
            for i in range(len(self)-1, -1, -1):
                yield self._get_at(i)

        def add(self, *incoming: PHIIncoming):
            arr_type = c_object_p * len(incoming)
            
            values_arr = arr_type(*(x.value.ptr for x in incoming))
            blocks_arr = arr_type(*(x.block.ptr for x in incoming))

            LLVMAddIncoming(self._pni, values_arr, blocks_arr, len(incoming))

        def _get_at(self, ndx: int) -> PHIIncoming:
            val = LLVMGetIncomingValue(self._pni, ndx)
            block = LLVMGetIncomingBlock(self._pni, ndx)

            return PHIIncoming(val, block)

    @property
    def incoming(self) -> _IncomingList:
        return PHINodeInstruction._IncomingList(self)

# ---------------------------------------------------------------------------- # 

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

class FuncAttributeSet(AttributeSet):
    _func: 'Function'

    def __init__(self, func: 'Function', ndx: int):
        self._func = func
        super().__init__(ndx)

    def _get_attr_count(self) -> int:
        return LLVMGetAttributeCountAtIndex(self._func, self._ndx)

    def _get_attrs(self, attr_arr: POINTER(c_object_p)):
        LLVMGetAttributesAtIndex(self._func, self._ndx, attr_arr)

    def _add_attr(self, attr: Attribute):
        LLVMAddAttributeAtIndex(self._func, self._ndx, attr)

    def _remove_enum_attr(self, key: Attribute.Kind):
        LLVMRemoveEnumAttributeAtIndex(self._func, self._ndx, key)

    def _remove_str_attr(self, key_bytes: bytes):
        LLVMRemoveStringAttributeAtIndex(self._func, self._ndx, key_bytes, len(key_bytes))

class FuncParam(Value):
    _attr_set: FuncAttributeSet

    def __init__(self, ptr: c_object_p, func: 'Function', attr_ndx: int):
        super().__init__(ptr)

        self._attr_set = FuncAttributeSet(func, attr_ndx)

    @property
    def attr_set(self):
        return self._attr_set

class FuncBody:
    _func: 'Function'
    _name_counter: int

    def __init__(self, func: 'Function'):
        self._func = func
        self._name_counter = 1

    def __iter__(self) -> Iterator[BasicBlock]:
        bb_ptr = LLVMGetFirstBasicBlock(self._func)

        while bb_ptr:
            yield BasicBlock(bb_ptr)

            bb_ptr = LLVMGetNextBasicBlock(bb_ptr)

    def __reversed__(self) -> Iterator[BasicBlock]:
        bb_ptr = LLVMGetLastBasicBlock(self._func)

        while bb_ptr:
            yield BasicBlock(bb_ptr)

            bb_ptr = LLVMGetPreviousBasicBlock(bb_ptr)

    def __len__(self) -> int:
        return LLVMCountBasicBlocks(self._func)

    def first(self) -> Optional[BasicBlock]:
        bb_ptr = LLVMGetFirstBasicBlock(self._func)

        if bb_ptr:
            return BasicBlock(bb_ptr)

        return None

    def last(self) -> Optional[BasicBlock]:
        bb_ptr = LLVMGetLastBasicBlock(self._func)

        if bb_ptr:
            return BasicBlock(bb_ptr)

        return None

    def entry(self) -> Optional[BasicBlock]:
        bb_ptr = LLVMGetEntryBasicBlock(self._func)

        if bb_ptr:
            return BasicBlock(bb_ptr)

        return None

    def append(self, name: str = "") -> BasicBlock:
        if not name:
            name = f'bb{self._name_counter}'
            self._name_counter += 1

        return BasicBlock(LLVMAppendBasicBlockInContext(
            get_context(),
            self._func,
            name.encode(),
        ))

    def insert(self, before: BasicBlock, name: str = "") -> BasicBlock:
        if not name:
            name = f'bb{len(self._name_counter)}'
            self._name_counter += 1

        return BasicBlock(LLVMInsertBasicBlockInContext(
            get_context(),
            before,
            name.encode()
        ))

    def remove(self, bb: BasicBlock):
        LLVMDeleteBasicBlock(bb)

class Function(GlobalValue):
    _attr_set: FuncAttributeSet
    _rt_attr_set: FuncAttributeSet

    class FuncParamsView:
        _func: 'Function'

        def __init__(self, func: 'Function'):
            self._func = func

        def __len__(self) -> int:
            return LLVMCountParams(self._func)

        def __getitem__(self, ndx: int) -> FuncParam:
            if 0 <= ndx < len(self):
                return FuncParam(LLVMGetParam(self._func, ndx), self._func, ndx + 1)
            else:
                raise IndexError()

        def __iter__(self) -> Iterator[FuncParam]:
            for i in range(len(self)):
                yield FuncParam(LLVMGetParam(self._func, i), self._func, i + 1)

    def __init__(self, ptr: c_object_p):
        super().__init__(ptr)

        self._attr_set = FuncAttributeSet(self, 0)
        self._rt_attr_set = FuncAttributeSet(self, -1)

    @property
    def call_conv(self) -> CallConv:
        return CallConv(LLVMGetFunctionCallConv(self))

    @call_conv.setter
    def call_conv(self, cc: CallConv):
        LLVMSetFunctionCallConv(self, cc)

    @property
    def params(self) -> FuncParamsView:
        return Function.FuncParamsView(self)

    @property
    def attr_set(self) -> FuncAttributeSet:
        return self._attr_set

    @property
    def rt_attr_set(self) -> FuncAttributeSet:
        return self._rt_attr_set

    @property
    def body(self) -> FuncBody:
        return FuncBody(self)

    @property
    def di_sub_program(self) -> DISubprogram:
        return DISubprogram(LLVMGetSubprogram(self))

    @di_sub_program.setter
    def di_sub_program(self, dsp: DISubprogram):
        LLVMSetSubprogram(self, dsp)

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

@llvm_api
def LLVMGetInstructionParent(instr: Instruction) -> c_object_p:
    pass

@llvm_api
def LLVMIsEnumAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateEnumAttribute(ctx: Context, kind: Attribute.Kind, val: c_uint64) -> c_object_p:
    pass

@llvm_api
def LLVMGetEnumAttributeKind(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMGetEnumAttributeValue(attr: Attribute) -> c_uint64:
    pass

@llvm_api
def LLVMIsStringAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateStringAttribute(ctx: Context, kind: c_char_p, klen: c_uint, val: c_char_p, vlen: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetStringAttributeKind(attr: Attribute, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMGetStringAttributeValue(attr: Attribute, length: POINTER(c_uint)) -> c_char_p:
    pass

@llvm_api
def LLVMIsTypeAttribute(attr: Attribute) -> c_enum:
    pass

@llvm_api
def LLVMCreateTypeAttribute(ctx: Context, kind: Attribute.Kind, val: Type) -> c_object_p:
    pass

@llvm_api
def LLVMGetTypeAttributeValue(attr: Attribute) -> c_object_p:
    pass

@llvm_api
def LLVMGetFunctionCallConv(func: Function) -> c_uint:
    pass

@llvm_api
def LLVMSetFunctionCallConv(func: Function, cc: CallConv):
    pass

@llvm_api
def LLVMCountParams(func: Function) -> c_uint:
    pass

@llvm_api
def LLVMGetParam(func: Function, ndx: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetAttributeCountAtIndex(func: Function, ndx: c_uint) -> c_uint:
    pass

@llvm_api
def LLVMGetAttributesAtIndex(func: Function, ndx: c_uint, attrs: POINTER(c_object_p)):
    pass

@llvm_api
def LLVMAddAttributeAtIndex(func: Function, ndx: c_uint, attr: Attribute):
    pass

@llvm_api
def LLVMRemoveEnumAttributeAtIndex(func: Function, ndx: c_uint, kind: Attribute.Kind):
    pass

@llvm_api
def LLVMRemoveStringAttributeAtIndex(func: Function, ndx: c_uint, kind: c_char_p, length: c_uint):
    pass

@llvm_api
def LLVMGetFirstBasicBlock(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMGetNextBasicBlock(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetLastBasicBlock(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMGetPreviousBasicBlock(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetEntryBasicBlock(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMCountBasicBlocks(func: Function) -> c_uint:
    pass

@llvm_api
def LLVMAppendBasicBlockInContext(ctx: Context, func: Function, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMInsertBasicBlockInContext(ctx: Context, bb: BasicBlock, name: c_char_p) -> c_object_p:
    pass

@llvm_api
def LLVMDeleteBasicBlock(bb: BasicBlock) -> c_object_p:
    pass

@llvm_api
def LLVMGetInstructionCallConv(ci: CallInstruction) -> c_enum:
    pass

@llvm_api
def LLVMSetInstructionCallConv(ci: CallInstruction, cc: CallConv):
    pass

@llvm_api
def LLVMIsTailCall(ci: CallInstruction) -> c_enum:
    pass

@llvm_api
def LLVMSetTailCall(ci: CallInstruction, is_tail_call: c_enum):
    pass

@llvm_api
def LLVMGetCalledValue(ci: CallInstruction) -> c_object_p:
    pass

@llvm_api
def LLVMGetCallSiteAttributeCount(ci: CallInstruction, ndx: c_uint) -> c_uint:
    pass

@llvm_api
def LLVMGetCallSiteAttributes(ci: CallInstruction, ndx: c_uint, arr: POINTER(c_object_p)):
    pass

@llvm_api
def LLVMAddCallSiteAttribute(ci: CallInstruction, ndx: c_uint, attr: Attribute):
    pass

@llvm_api
def LLVMRemoveCallSiteEnumAttribute(ci: CallInstruction, ndx: c_uint, key: Attribute.Kind):
    pass

@llvm_api
def LLVMRemoveCallSiteStringAttribute(ci: CallInstruction, ndx: c_uint, key: c_char_p, key_len: c_uint):
    pass

@llvm_api
def LLVMGetNumSuccessors(term: Terminator) -> c_uint:
    pass

@llvm_api
def LLVMGetSuccessor(term: Terminator, ndx: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMSetSuccessor(term: Terminator, ndx: c_uint, bb: BasicBlock):
    pass

@llvm_api
def LLVMIsConditional(term: Terminator) -> c_enum:
    pass 

@llvm_api 
def LLVMGetCondition(cb: ConditionalInstruction) -> c_object_p:
    pass

@llvm_api
def LLVMSetCondition(cb: ConditionalInstruction, cond: Value):
    pass

@llvm_api
def LLVMAddCase(switch: SwitchInstruction, on_val: Value, dest: BasicBlock):
    pass

@llvm_api
def LLVMGetSwitchDefaultDest(switch: SwitchInstruction) -> c_object_p:
    pass

@llvm_api
def LLVMGetAllocatedType(alloca: AllocaInstruction) -> c_object_p:
    pass

@llvm_api
def LLVMIsInBounds(gep: GEPInstruction) -> c_enum:
    pass

@llvm_api
def LLVMSetIsInBounds(gep: GEPInstruction, is_in_bounds: c_enum):
    pass

@llvm_api
def LLVMAddIncoming(pni: PHINodeInstruction, values: POINTER(c_object_p), blocks: POINTER(c_object_p), count: c_uint):
    pass

@llvm_api
def LLVMCountIncoming(pni: PHINodeInstruction) -> c_uint:
    pass

@llvm_api
def LLVMGetIncomingValue(pni: PHINodeInstruction, ndx: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetIncomingBlock(pni: PHINodeInstruction, ndx: c_uint) -> c_object_p:
    pass

@llvm_api
def LLVMGetICmpPredicate(icmp: ICmpInstruction) -> c_enum:
    pass

@llvm_api
def LLVMGetFCmpPredicate(fcmp: FCmpInstruction) -> c_enum:
    pass

@llvm_api
def LLVMGetSubprogram(func: Function) -> c_object_p:
    pass

@llvm_api
def LLVMSetSubprogram(func: Function, dsp: DISubprogram):
    pass 