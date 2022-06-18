'''
The main backend of the Chai compiler: responsible for converting Chai into LLVM.
'''

__all__ = ['Generator']

import llvm.value as llvalue
import llvm.types as lltypes
from llvm.module import Module as LLModule
from depm.source import Package
from syntax.ast import *

from .type_util import conv_type, is_unit
from .predicate import *
from .debug_info import DebugInfoEmitter

class Generator:
    '''
    Responsible for converting a Chai package into an LLVM module.

    Methods
    -------
    generate() -> LLModule
    '''

    # The package being converted to LLVM.
    pkg: Package

    # The global package prefix used for name mangling: by adding this prefix,
    # we prevent link collisions between identically named functions defined in
    # different packages.
    pkg_prefix: str

    # The LLVM module being generated from the package.
    mod: LLModule

    # The debug info emitter.
    die: DebugInfoEmitter

    # The predicate generator: used to convert expressions into LLVM IR.
    pred_gen: PredicateGenerator

    def __init__(self, pkg: Package):
        '''
        Params
        ------
        pkg: Package
            The package to generate.
        debug: bool
            Whether to emit debug information.
        '''

        self.pkg = pkg
        self.pkg_prefix = f'p{pkg.id}.'

        self.mod = LLModule(pkg.name)

        self.die = DebugInfoEmitter(pkg, self.mod)

        self.pred_gen = PredicateGenerator(self.die)

    def generate(self) -> LLModule:
        '''
        Generates the package and returns the produced LLVM module.
        '''

        # Start by generating all the global declarations.
        for src_file in self.pkg.files:
            self.die.emit_file_header(src_file)

            for defin in src_file.definitions:
                match defin:
                    case FuncDef():
                        self.generate_func_def(defin)
                    case OperDef():
                        self.generate_oper_def(defin)

        # Then, generate all the predicates of those global declarations: since
        # predicate bodies can refer to any global declarations, we have to
        # generate them last so that all the declarations are visible.
        self.pred_gen.generate()

        # Finalize debug info.
        self.die.finalize()

        # Return the completed module.
        return self.mod

    # ---------------------------------------------------------------------------- #

    def generate_func_def(self, fd: FuncDef):
        '''
        Generate a function definition.

        Params
        ------
        fd: FuncDef
            The function definition to generate.
        '''

        # TODO implementation of intrinsic functions
        if 'intrinsic' in fd.annots:
            return

        # Whether to apply standard name mangling.
        mangle = True

        # Whether this function definition should be marked external.
        public = False

        for annot in fd.annots:
            match annot:
                case 'extern' | 'abientry':
                    # Both @extern and @abientry stop name mangling and make
                    # the symbol public.  @extern because the name of the
                    # function has to match the name of the external symbol.
                    mangle = False
                    public = True

        # Determine the LLVM name based on whether or not it should be mangled
        # from the Chai name.
        if mangle:
            ll_name = self.pkg_prefix + fd.symbol.name
        else:
            ll_name = fd.symbol.name

        # Create the LLVM function type.
        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in fd.params],
            conv_type(fd.type.inner_type().rt_type, rt_type=True)
        )

        # Create the LLVM function.
        ll_func = self.mod.add_function(ll_name, ll_func_type)

        # Mark it as external if necessary.
        ll_func.linkage = llvalue.Linkage.EXTERNAL if public else llvalue.Linkage.INTERNAL

        # Assign all the function parameters their corresponding LLVM values and
        # update the LLVM parameters with their appropriate names.
        for param, ll_param in zip(fd.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        # Assign the function its corresponding LLVM function value.
        fd.symbol.ll_value = ll_func

        # Add the function body as a predicate to generate if it exists.
        if fd.body:
            self.pred_gen.add_predicate(BodyPredicate(ll_func, fd.params, fd.body, not is_unit(fd.type.rt_type)))

        # Emit function debug info if it is not external/bodiless.
        if fd.body:
            ll_func.di_sub_program = self.die.emit_function_info(fd, ll_name)

    def generate_oper_def(self, od: OperDef):
        '''
        Generate an operator definition.

        Params
        ------
        od: OperDef
            The operator definition to generate.
        '''

        # Intrinsic operators are not generated: their applications are replaced
        # with LLVM code snippets.
        if 'intrinsic' in od.annots:
            return

        # All operators have the LLVM name `oper.overload.[overload_id]`.
        ll_name = f'{self.pkg_prefix}.oper.overload.{od.overload.id}'

        # Generate the function type for the operator definition.
        ll_func_type = lltypes.FunctionType(
            [conv_type(x.type) for x in od.params],
            conv_type(od.type.inner_type().rt_type, rt_type=True)
        )

        # Generate the LLVM function for the operator definition.
        ll_func = self.mod.add_function(ll_name, ll_func_type)
        
        # TODO add support for exported operators.
        ll_func.linkage = llvalue.Linkage.INTERNAL

        # Assign all the operator parameters their corresponding LLVM values and
        # update the LLVM parameters with their appropriate names.
        for param, ll_param in zip(od.params, ll_func.params):
            ll_param.name = param.name
            param.ll_value = ll_param

        # Assign the overload its corresponding LLVM function value.
        od.overload.ll_value = ll_func

        # Add the overload body as a predicate to generate if it exists.
        if od.body:
            self.pred_gen.add_predicate(BodyPredicate(ll_func, od.params, od.body, not is_unit(od.overload.signature.rt_type)))

        # Emit operator debug info.
        ll_func.di_sub_program = self.die.emit_oper_info(od, ll_name)

    def generate_record_type_def(self, rtd: RecordTypeDef):
        '''
        Generates a record type definition.

        Params
        ------
        rtd: RecordTypeDef
            The record type definition.
        '''

        # Get the associated record type.
        rec_type = rtd.type.inner_type() 
        
        # Determine the order in which to generate fields.
        if 'keep_order' in rtd.annots:
            fields = rec_type.all_fields
        else:
            # Unless the user has indicated otherwise, we want to reorder fields
            # to minimize alignment padding. 

            # First, we sort the fields in descending order by size.
            sorted_fields = sorted(
                rec_type.all_fields, 
                key=lambda x: x[1].size, 
                reverse=True
            )

            # Then, until we have added all fields to the struct...
            fields = []
            current_offset = 0
            while len(sorted_fields) > 0:
                # The maximum offset % align which is corresponds to minimum
                # padding required to insert a field.
                max_mod = 0

                # The field index in the sorted fields list with the minimum
                # padding needed for it be inserted aligned.
                min_padding_i = 0

                # We consider each field which hasn't yet be inserted in order
                # from largest to smallest size (this is generally more
                # efficient).
                for i in range(len(sorted_fields)):
                    field = sorted_fields[i]

                    mod = current_offset % field.type.align
 
                    if mod == 0:
                        # If the field can be inserted with no padding, we add
                        # it to the struct fields, remove it from the list of
                        # remaining fields, and stop searching.
                        fields.append(field)
                        sorted_fields.pop(i)
                        break
                    elif max_mod < mod:
                        # Otherwise, if the padding needed to insert this field
                        # is less than the minimum padding needed to insert any
                        # other field we have considered so far, we mark it as
                        # the field to insert if we can't find any zero padding
                        # fields and record its minimum padding value
                        # (represented by its offset % align).
                        min_padding_i = i
                        max_mod = mod
                else:
                    # If no zero padding field can be found, we insert the field
                    # with minimum alignment padding.
                    fields.append(sorted_fields.pop(min_padding_i))

        # We then create an LLVM struct based on the record fields.
        ll_struct = lltypes.StructType(
            *(conv_type(field.type) for field in fields), 
            name=self.pkg_prefix + rtd.sym.name,
            packed="packed" in rtd.annots
        )

        # Update the record types `ll_type` appropriately.
        rec_type.ll_type = ll_struct

        # Add a record data info entry.
        self.pred_gen.add_record_info(rec_type, RecordInfo(
            {field.name: i for i, field in enumerate(fields)},
            rtd.field_inits
        ))

        # Update the record type's alignment: the first field is always the
        # largest field in the struct; therefore, it is the field determining
        # the alignment.
        rec_type.align = fields[0].type.align

                    
                



