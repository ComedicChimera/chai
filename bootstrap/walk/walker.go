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

	// The stack of local scopes used to lookup symbols.
	localScopes []map[string]*common.Symbol

	// The return type of the enclosing function.  If this is `nil`, then there
	// is no enclosing function: ie. return statements are not valid.
	enclosingReturnType types.Type

	// The list of untyped nulls that occur in user source code.
	nullValues []*types.UntypedNull

	// The number loops until the outermost function block.
	loopDepth int
}

// WalkFile semantically analyzes the given source file.
func WalkFile(chFile *depm.ChaiFile) {
	w := &Walker{chFile: chFile}

	for _, def := range chFile.Definitions {
		w.walkDef(def)

		// Clear the list of nulls between walks.
		w.nullValues = nil
	}
}

// walkDef walks a definition and catches any errors that occur.
func (w *Walker) walkDef(def ast.ASTNode) {
	// Catch any errors that occur while walking the definition.
	defer report.CatchErrors(w.chFile.AbsPath, w.chFile.ReprPath)

	w.doWalkDef(def)

	// Ensure that all null values are inferred.
	for _, nullValue := range w.nullValues {
		if nullValue.InferredType == nil {
			w.recError(nullValue.Span, "cannot infer type of null")
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

	// TODO: imports/scripts?

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
