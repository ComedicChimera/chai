'''The package responsible for performing LLVM code generation.'''

from depm import *
from depm.source import Package, SourceFile
from typecheck import *
import llvm.module as llmod
import llvm.value as llvalue
import llvm.types as lltypes
import llvm.ir as ir