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
func NewWalker(chFile *depm.ChaiFile) *Walker {
	return &Walker{
		chFile: chFile,
		solver: typing.NewSolver(chFile.Context),
	}
}

// WalkDef walks a single definition of a source file.  This function allows
// the caller to control when a given definition is evaluated as well as
// what part of it to support generic evaluation.
func (w *Walker) WalkDef(def ast.Def) bool {
	// set the map of dependencies to match the definition being walked
	w.deps = def.Dependencies()

	switch v := def.(type) {
	case *ast.FuncDef:
		return w.walkFuncLike(v.Signature, v.Args, v.Body)
	case *ast.OperDef:
		return w.walkFuncLike(v.Op.Signature, v.Args, v.Body)
	}

	// clear the map of dependencies
	w.deps = nil

	// TODO: other definitions
	return false
}

// walkFuncLike walks a function like (ie. a function or an operator: semantics
// are the same for both from an analysis perspective at this point).
func (w *Walker) walkFuncLike(signature *typing.FuncType, args []ast.FuncArg, body ast.Expr) bool {
	// nil body => nothing to walk => all good
	if body == nil {
		return true
	}

	// push a scope for the function
	w.pushFuncScope(signature, args)

	// make sure the scope is popped before we exit
	defer w.popScope()

	// walk the function body expression
	if !w.walkExpr(body) {
		return false
	}

	// add a constraint to the body's return value to ensure that is matches
	// the function's return value if the function actually returns a value.
	if !signature.ReturnType.Equiv(typing.PrimType(typing.PrimNothing)) {
		w.solver.Constrain(signature.ReturnType, body.Type(), body.Position())
	}

	// type solve the function body
	if !w.solver.Solve() {
		return false
	}

	// update the constancy of the function arguments
	for i, argSym := range w.topScope().LocalArgs {
		// cannot become `Immutable`, but if they are never mutated, then they
		// become constant.
		args[i].Constant = argSym.Mutability == depm.NeverMutated
	}

	// body has been checked -- good to go
	return true
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
	// TODO: local symbol imports

	// global symbol table
	if sym, ok := w.chFile.Parent.SymbolTable[name]; ok {
		// add the global symbol to the list of dependencies
		w.deps[sym.Name] = struct{}{}

		return sym, true
	}

	w.reportError(pos, "undefined symbol: `%s`", name)
	return nil, false
}

// lookupOperator retrieves the overloads for a particular operator. It reports
// an error if the lookup fails.
func (w *Walker) lookupOperator(aop *ast.Oper) (*depm.Operator, bool) {
	// TODO: local operators

	if op, ok := w.chFile.Parent.OperatorTable[aop.Kind]; ok {
		return op, true
	}

	// TODO: get the operator name?
	w.reportError(aop.Pos, "no defined overloads for operator: `%s`", aop.Name)
	return nil, false
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
			Vars: make(map[string]*depm.Symbol),

			// copy down parent data
			Func: w.scopes[len(w.scopes)-1].Func,
		})
	} else {
		w.scopes = append(w.scopes, &Scope{Vars: make(map[string]*depm.Symbol), LocalMuts: make(map[string]*int)})
	}
}

// pushFuncScope pushes a new local scope as the top scope of a function.
func (w *Walker) pushFuncScope(f *typing.FuncType, args []ast.FuncArg) {
	localArgs := make([]*depm.Symbol, len(args))
	for i, arg := range args {
		localArgs[i] = &depm.Symbol{
			Name:        arg.Name,
			PkgID:       w.chFile.Parent.ID,
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
	// update local mutabilities
	for name, mutptr := range w.topScope().LocalMuts {
		*mutptr = w.topScope().Vars[name].Mutability
	}

	w.scopes = w.scopes[:len(w.scopes)-1]
}
