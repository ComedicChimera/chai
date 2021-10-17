from dataclasses import dataclass
from enum import Enum, auto
from typing import Dict, List, MutableSet, Optional, Callable

from . import ChaiCompileError, ChaiFail, TextPosition
from .types import DataType
from .report import report

class DefKind(Enum):
    ValueDef = auto()

    @staticmethod
    def repr(kind) -> str:
        if kind == DefKind.ValueDef:
            return 'value'

        # TODO: rest

class Mutability(Enum):
    Mutable = auto()
    Immutable = auto()
    NeverMutated = auto()

# NOTE: All fields marked `Optional` in symbol are field that are bound late --
# they may not be immediately determined when the symbol is initialized.
@dataclass
class Symbol:
    name: str
    typ: Optional[DataType]
    def_pos: Optional[TextPosition]
    parent_pkg_id: int
    public: bool
    def_kind: DefKind
    mutability: Mutability

# SymbolReference is a usage of a symbol in user source code that is has yet to
# be resolved. 
@dataclass
class SymbolReference:
    pkg_id: int
    rel_path: str
    pos: TextPosition
    asserts_mutable: bool


# SymbolTable is the global symbol table for a Chai package.  It is not
# responsible for providing an index of defined symbols but also for resolving
# and handling undefined symbols as they are used.  Chai attempts to fully parse
# and symbolically analyze in one pass -- the table keeps track of the symbolic
# state of the program during this process.  Once package initialization is
# finished, this symbol table can be used to report all the undefined symbols.
class SymbolTable:
    # lookup_table is a dictionary of all visible symbols OR all global symbols
    # whose definitions have been requested (so that all dependencies can share
    # the same symbol -- when it is resolved, it will be resolved for all
    # dependencies.
    lookup_table: Dict[str, Symbol] = {}

    # unresolved is a table of all references to unresolved symbols.  As symbols
    # are resolved, entries are removed from this table.  All entries remaining
    # in this table after package initialization are reported unresolved.
    unresolved: Dict[str, List[SymbolReference]] = {}

    # resolution_failed is a set of symbol names for which resolution failed.
    # This prevents symbols from being defined multiple times if their initial
    # resolution failed.
    resolution_failed: MutableSet[str] = set()

    # pkg_id is the ID of the package that this symbol table belongs to
    pkg_id: int

    def __init__(self, pkg_id: int) -> None:
        self.pkg_id = pkg_id

    # define defines a new global symbol and resolves all those symbols
    # depending on it. This function returns a symbol which is the Symbol
    # instance that should be used for this symbol henceforth -- it may not be
    # the symbol passed in.  It may throw an error if the symbol is already
    # concretely defined (ie. not just depended on).  It returns None if the
    # definition fails.  It also requires the relative path to the file that
    # defines the symbol for purposes of error reporting.
    def define(self, sym: Symbol, rel_path: str) -> Symbol:
        # if resolution has already been attempted and failed for this symbol
        # then trying to define it again is equivalent to a multi-definition
        if sym.name in self.resolution_failed:
            raise ChaiCompileError(rel_path, sym.def_pos, f'symbol defined multiple times: `{sym.name}`')
        # if the symbol is unresolved, then its definition is ok and this
        # definition completes the previously defined symbol and removes
        # unresolved entries corresponding to the symbol.
        elif sym.name in self.unresolved:
            return self._resolve(sym)
        # symbol is already concretely defined => error
        elif sym.name in self.lookup_table:
            raise ChaiCompileError(rel_path, sym.def_pos, f'symbol defined multiple times: `{sym.name}`')
        # otherwise, just stick the completed symbol in the look up table
        else:
            self.lookup_table[sym.name] = sym
            return sym

    # lookup attempts to retrieve a globally declared symbol matching the given
    # characteristics. It may throw an error if it finds a partial match (ie. a
    # symbol that matches in name but not other properties).  However, if it
    # does not find a symbol matching in name, then it will NOT throw an error,
    # but rather record the symbol reference to see if a matching symbol is
    # defined later -- a declared-by-usage symbol is returned in this case.
    def lookup(self, pkg_id: int, rel_path: str, pos: TextPosition, name: str, def_kind: DefKind, mutability: Mutability) -> Symbol:
        if name in self.lookup_table:
            decl_sym = self.lookup_table[name]

            # check for partial matches (ie. non-matching def kinds, etc.)
            if decl_sym.def_kind != def_kind:
                raise ChaiCompileError(rel_path, pos, f'cannot use {decl_sym.def_kind} as {def_kind}')
            elif decl_sym.mutability == Mutability.Immutable and mutability == Mutability.Mutable:
                raise ChaiCompileError(rel_path, pos, 'cannot mutate an immutable value')

            # publicity is unique: if the symbol is unresolved, then if the
            # symbol must be public, then we simply mark it as public.  If it is
            # already defined, however, then we error if it lookup is externally
            # and the defined symbol is not.
            if not decl_sym.public and self.pkg_id != pkg_id:
                if name in self.unresolved:
                    # update the publicity of the symbol
                    decl_sym.public = True
                else:
                    # error
                    raise ChaiCompileError(rel_path, pos, f'symbol `{name}` is not publically visible')
           
            # update mutability => this happens in all cases
            if decl_sym.mutability == Mutability.NeverMutated and mutability != Mutability.NeverMutated:
                decl_sym.mutability = mutability

            # add the symbol reference to list of unresolved as necessary
            if name in self.unresolved:
                self.unresolved[name].append(SymbolReference(pkg_id, rel_path, pos, mutability == Mutability.Mutable))

            # return the matched symbol
            return decl_sym
        else:
            self.lookup_table[name] = Symbol(name, None, None, self.pkg_id, self.pkg_id != pkg_id, def_kind, mutability)
            self.unresolved[name].append(SymbolReference(pkg_id, rel_path, pos, mutability == Mutability.Mutable))
            return self.lookup_table[name]

    # report_unresolved reports all remaining unresolved symbols as unresolved
    # => package initialization has finished; no more symbols can resolve
    def report_unresolved(self) -> None:
        for name in self.unresolved:
            try:
                self._error_unresolved(name, f'undefined symbol: `{name}`')
            except ChaiFail:
                pass

    # ---------------------------------------------------------------------------- #

    # _resolve attempts to resolve a declared-by-usage symbol with its actual
    # definition. This function returns the declared symbol if resolution
    # succeeds and None if it fails.
    def _resolve(self, sym: Symbol) -> Symbol:
        # we want to first check that the properties of the declared symbol
        # match up with those of the defined symbol.  If not, then we need
        # to error on all the unresolved usages appropriately.
        decl_sym = self.lookup_table[sym.name]

        # if the declared symbol is marked as public, then it was required from
        # another package.  Thus, if the defined symbol is not public then we
        # want to error on all unresolved that are from another package and
        # fail.
        if decl_sym.public and not sym.public:
            self._error_unresolved(sym.name, f'symbol `{sym.name}` is not publically visible', lambda s: s.pkg_id != self.pkg_id)

        if decl_sym.def_kind != sym.def_kind:
            self._error_unresolved(sym.name, f'cannot use {sym.def_kind} as {decl_sym.def_kind}')
        
        # mutability is only an issue if the declared symbol is marked as
        # mutable and the actual symbol is immutable.
        if decl_sym.mutability == Mutability.Mutable and sym.mutability == Mutability.Immutable:
            self._error_unresolved(sym.name, f'cannot mutate an immutable value', lambda s: s.asserts_mutable)

        # we want to copy over all the symbol's properties to the previously
        # declared symbol to resolve all dependencies of it.  This
        # previously declared symbol will be returned as the defined symbol.
        decl_sym.name = sym.name
        decl_sym.def_kind = sym.def_kind
        decl_sym.def_pos = sym.def_pos
        decl_sym.typ = sym.typ
        decl_sym.public = sym.public

        # we only want to update the decl_sym's mutablility if the defined
        # symbol is not marked `NeverMutated`.
        if sym.mutability != Mutability.NeverMutated:
            decl_sym.mutability = sym.mutability

        # remove the unnecessary unresolved entries
        self.unresolved.pop(sym.name)

        # return the declared symbol (now concretely defined) as the actual
        # symbol -- so that way all symbol references are consistent
        return decl_sym

    # _error_unresolved throws an error for each unresolved reference to a
    # specify symbol.  The error message is passed in -- this can be any error.
    # A filtering condition can be passed in to only error on certain symbols.
    def _error_unresolved(self, name: str, msg: str, filter_cond: Optional[Callable[[SymbolReference], bool]] = None) -> None:
        if filter_cond:
            for ref in self.unresolved[name]:
                if filter_cond(ref):
                    report.report_compile_error(ChaiCompileError(ref.rel_path, ref.pos, msg))

            # filter out the symbols we have already errored upon
            self.unresolved[name] = list(filter(lambda x: not filter_cond(x), self.unresolved[name]))
        else:
            for ref in self.unresolved[name]:
                report.report_compile_error(ChaiCompileError(ref.rel_path, ref.pos, msg))

            # remove these unresolved so we don't error on them more than we have to
            self.unresolved.pop(name)
            
        # mark the symbol as having failed resolution
        self.resolution_failed.add(name)

        # stop file parsing and analysis
        raise ChaiFail()
