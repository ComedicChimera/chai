package walk

import (
	"chai/ast"
	"chai/depm"
	"chai/report"
	"chai/typing"
	"fmt"
)

// Walker is the structure responsible for performing the majority of semantic
// analysis of Chai code.  It walks the AST, types all nodes on the AST, looks
// up symbols, checks mutability, and performs all other necessary checks.  One
// walker should be created per source file.
type Walker struct {
	// uni is the shared universe for the project.
	uni *depm.Universe

	// chFile is the Chai file currently being walked.
	chFile *depm.ChaiFile

	// scopes the stack of scopes that are pushed and popped during semantic
	// analysis.  The properties of the parent scope are propagated down to sub
	// scopes; however, symbols should still be looked up in scope-reverse order
	// (ie. top of the stack to bottom) to support shadowing.
	scopes []*Scope

	// solver is the type solver used for the file this walker is analyzing.
	solver *typing.Solver

	// deps is a map of the dependencies of the definition currently being
	// walked.
	deps map[string]struct{}
}

// Scope represents a local scope inside a function or expression.  This stores
// both variables and context flags such as the enclosing function.
type Scope struct {
	// Vars are the scope-local variables: ie. only the variables that are
	// specifically defined in this scope -- they will be visible in all scopes
	// below this one.
	Vars map[string]*depm.Symbol

	// Func is the enclosing function type of the scope.
	Func *typing.FuncType

	// LocalArgs is the list of argument symbols corresponding to Func if it
	// exists.  These symbols are considered to be in pseudoscope above the
	// local symbols of the function.
	LocalArgs []*depm.Symbol

	// IsFuncTopScope indicates whether or not this scope is the top most scope
	// of its enclosing function.  It is used to denote when function arguments
	// should be looked up.
	IsFuncTopScope bool

	// LocalMuts is a map of local mutabilities: pointers to fields on the AST
	// to be updated with their symbol's mutability.  The key is the name of the
	// symbol to fetch the mutability from.
	// NOTE: This entire construct exists because I can't actually store the
	// symbols themselves on the AST because of Go's weird import rules so I
	// have to update them late.
	LocalMuts map[string]*int
}

// -----------------------------------------------------------------------------

// NewWalker creates a new walker for the given source file.
func NewWalker(uni *depm.Universe, chFile *depm.ChaiFile) *Walker {
	return &Walker{
		uni:    uni,
		chFile: chFile,
		solver: typing.NewSolver(chFile.Context, chFile.Parent.ID),
	}
}

// -----------------------------------------------------------------------------

// reportError reports a compile error in the walker.  It supports formatted
// error messages by default.
func (w *Walker) reportError(pos *report.TextPosition, msg string, args ...interface{}) {
	report.ReportCompileError(
		w.chFile.Context,
		pos,
		fmt.Sprintf(msg, args...),
	)
}

// -----------------------------------------------------------------------------

// defineLocal defines a new local variable.  It returns if the definition was
// successful (ie. did not conflict with any other local symbols). It throws an
// error as necessary.
func (w *Walker) defineLocal(sym *depm.Symbol) bool {
	if _, ok := w.topScope().Vars[sym.Name]; ok {
		w.reportError(sym.DefPosition, "multiple symbols defined in scope with name `%s`", sym.Name)
		return false
	} else {
		w.topScope().Vars[sym.Name] = sym
		return true
	}
}

// lookup looks up a symbol including local scopes and function arguments.  It
// throws an error if the symbol is not defined.
func (w *Walker) lookup(name string, pos *report.TextPosition) (*depm.Symbol, bool) {
	// lookup scopes in reverse order
	for i := len(w.scopes) - 1; i >= 0; i-- {
		scope := w.scopes[i]

		if sym, ok := scope.Vars[name]; ok {
			return sym, true
		}

		// check for function arguments
		if scope.IsFuncTopScope {
			for _, argSym := range scope.LocalArgs {
				if argSym.Name == name {
					return argSym, true
				}
			}
		}
	}

	// next, check global scopes
	return w.lookupGlobal(name, pos)
}

// lookupGlobal looks up a symbol exclusively in the global namespace and in the
// list of imported symbols.  It throws an error if the symbol is not defined.
func (w *Walker) lookupGlobal(name string, pos *report.TextPosition) (*depm.Symbol, bool) {
	// global symbols
	if sym, ok := w.chFile.Parent.SymbolTable[name]; ok {
		// add the global symbol to the list of dependencies
		w.deps[sym.Name] = struct{}{}

		return sym, true
	}

	// imported symbols
	if sym, ok := w.chFile.ImportedSymbols[name]; ok {
		// no need to mark it as a dependency: always declared first and not
		// defined within the package
		return sym, true
	}

	// universal symbols
	if sym, ok := w.uni.GetSymbol(name); ok {
		return sym, true
	}

	w.reportError(pos, "undefined symbol: `%s`", name)
	return nil, false
}

// lookupOperators retrieves the overloads for a particular operator.  It
// returns all the unique operator collections matching the passed in AST
// operator.  It reports an error if the lookup fails.
func (w *Walker) lookupOperators(aop ast.Oper) ([]*depm.Operator, bool) {
	var operators []*depm.Operator

	// global operators
	if op, ok := w.chFile.Parent.OperatorTable[aop.Kind]; ok {
		operators = append(operators, op)
	}

	// imported operators
	if op, ok := w.chFile.ImportedOperators[aop.Kind]; ok {
		operators = append(operators, op)
	}

	// universal operators
	if op, ok := w.uni.GetOperator(aop.Kind); ok {
		operators = append(operators, op)
	}

	if len(operators) == 0 {
		w.reportError(aop.Pos, "no defined overloads for operator: `%s`", aop.Name)
		return nil, false
	}

	return operators, true
}

// -----------------------------------------------------------------------------

// topScope gets the scope on the top of the scope stack assuming the stack is
// nonempty.
func (w *Walker) topScope() *Scope {
	return w.scopes[len(w.scopes)-1]
}

// pushScope pushes a new local scope onto the scope stack.
func (w *Walker) pushScope() {
	if len(w.scopes) > 0 {
		w.scopes = append(w.scopes, &Scope{
			Vars:      make(map[string]*depm.Symbol),
			LocalMuts: make(map[string]*int),

			// copy down parent data
			Func: w.scopes[len(w.scopes)-1].Func,
		})
	} else {
		w.scopes = append(w.scopes, &Scope{Vars: make(map[string]*depm.Symbol), LocalMuts: make(map[string]*int)})
	}
}

// pushFuncScope pushes a new local scope as the top scope of a function.
func (w *Walker) pushFuncScope(f *typing.FuncType, args []*ast.FuncArg) {
	localArgs := make([]*depm.Symbol, len(args))
	for i, arg := range args {
		localArgs[i] = &depm.Symbol{
			Name:        arg.Name,
			File:        w.chFile,
			DefPosition: nil, // never errored upon
			Type:        arg.Type,
			DefKind:     depm.DKValueDef,
			Mutability:  depm.NeverMutated,
			Public:      false,
		}
	}

	// no other data is copied down
	w.scopes = append(w.scopes, &Scope{
		Vars:           make(map[string]*depm.Symbol),
		LocalMuts:      make(map[string]*int),
		Func:           f,
		LocalArgs:      localArgs,
		IsFuncTopScope: true,
	})
}

// popScope pops a scope off the scope stack (assuming there are scopes to pop).
func (w *Walker) popScope() {
	topScope := w.topScope()

	// update local mutabilities
	for name, mutptr := range topScope.LocalMuts {
		*mutptr = topScope.Vars[name].Mutability
	}

	w.scopes = w.scopes[:len(w.scopes)-1]
}
