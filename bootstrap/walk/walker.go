package walk

import (
	"chaic/ast"
	"chaic/common"
	"chaic/depm"
	"chaic/report"
	"chaic/types"
)

// Walker is responsible for walking source files and performing semantic
// analysis on their definitions.
type Walker struct {
	// The Chai source file being walked.
	chFile *depm.ChaiFile

	// The list of created untyped nulls.
	untypedNulls []*types.UntypedNull

	// The list of created untypes numbers.
	untypedNumbers []*types.UntypedNumber

	// The stack of local scopes used to lookup symbols.
	localScopes []map[string]*common.Symbol

	// The return type of the enclosing function.  If this is `nil`, then there
	// is no enclosing function: ie. return statements are not valid.
	enclosingReturnType types.Type

	// The number loops until the outermost function block.
	loopDepth int
}

// WalkFile semantically analyzes the given source file.
func WalkFile(chFile *depm.ChaiFile) {
	w := &Walker{chFile: chFile}

	for _, def := range chFile.Definitions {
		w.walkDef(def)
	}
}

// walkDef walks a definition and catches any errors that occur.
func (w *Walker) walkDef(def ast.ASTNode) {
	// Catch any errors that occur while walking the definition.
	defer report.CatchErrors(w.chFile.AbsPath, w.chFile.ReprPath)

	// Ensure that the walker is reset.
	defer func() {
		w.untypedNulls = nil
		w.untypedNumbers = nil
		w.localScopes = nil
		w.enclosingReturnType = nil
		w.loopDepth = 0
	}()

	w.doWalkDef(def)

	// Finalize type inference.
	for _, null := range w.untypedNulls {
		if null.Value == nil {
			w.recError(null.Span, "unable to infer type for null")
		}
	}

	for _, num := range w.untypedNumbers {
		if num.Value == nil {
			num.Value = num.PossibleTypes[0]
		}
	}
}

// -----------------------------------------------------------------------------

// lookup looks up a symbol by name in all visible scopes.  If no symbol by
// the given name can be found, then an error is reported.
func (w *Walker) lookup(name string, span *report.TextSpan) *common.Symbol {
	// Traverse local scopes in reverse order to implement shadowing.
	for i := len(w.localScopes) - 1; i > -1; i-- {
		if sym, ok := w.localScopes[i][name]; ok {
			return sym
		}
	}

	// TODO: imports

	// Lookup in the global table.
	if sym, ok := w.chFile.Parent.SymbolTable[name]; ok {
		return sym
	}

	w.error(span, "undefined symbol: `%s`", name)
	return nil
}

// defineLocal defines a local symbol in the current local scope.  If the symbol
// is already defined, then an error is reported.
func (w *Walker) defineLocal(sym *common.Symbol) {
	currScope := w.localScopes[len(w.localScopes)-1]

	if _, ok := currScope[sym.Name]; ok {
		w.error(sym.DefSpan, "multiple symbols named `%s` defined in immediate local scope", sym.Name)
	}

	currScope[sym.Name] = sym
}

// pushScope pushes a new local scope onto the scope stack.
func (w *Walker) pushScope() {
	w.localScopes = append(w.localScopes, make(map[string]*common.Symbol))
}

// popScope removes the top local scope from the scope stack.
func (w *Walker) popScope() {
	w.localScopes = w.localScopes[:len(w.localScopes)-1]
}

// -----------------------------------------------------------------------------

// error reports an error on the given span that should abort walking of the
// current definition.
func (w *Walker) error(span *report.TextSpan, msg string, args ...interface{}) {
	panic(report.Raise(span, msg, args...))
}

// recError reports a recoverable error on the given span.
func (w *Walker) recError(span *report.TextSpan, msg string, args ...interface{}) {
	report.ReportCompileError(
		w.chFile.AbsPath,
		w.chFile.ReprPath,
		span,
		msg,
		args...,
	)
}

// warn reports a compile warning.
func (w *Walker) warn(span *report.TextSpan, msg string, args ...interface{}) {
	report.ReportCompileWarning(
		w.chFile.AbsPath,
		w.chFile.ReprPath,
		span,
		msg,
		args...,
	)
}
