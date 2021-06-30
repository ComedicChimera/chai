package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/typing"
)

// Walker is the construct responsible for performing semantic analysis on files
// as both the top level and expression level
type Walker struct {
	// SrcFile is the file this walker is walking
	SrcFile *sem.ChaiFile

	// exprContextStack stores the contextual values and flags used for an
	// expression.  The stack allows us to push and pop contexts at will as
	// traverse the tree.  Eg. when we enter a loop, we push a loop context and
	// pop it when we exit.
	exprContextStack []*ExprContext

	// solver is the solver used for performing type checking and inference
	// within this source file
	solver *typing.Solver
}

// NewWalker creates a new walker for a given file
func NewWalker(f *sem.ChaiFile) *Walker {
	return &Walker{
		SrcFile: f,
		solver:  typing.NewSolver(f.LogContext),
	}
}

// -----------------------------------------------------------------------------

// ExprContext stores all the contextual variables within the body of a given
// function, method, or lambda.
type ExprContext struct {
	// FuncContext stores the current enclosing function or lambda to
	// facilitate parameter lookups, return type checkings, etc.
	FuncContext *typing.FuncType

	// FuncArgScope indicates whether or not the enclosing context scope is
	// the scope in which the function parameters are defined -- support
	// shadowing
	FuncArgScope bool

	// LoopContext indicates whether or not `break` and `continue` are usable
	LoopContext bool

	// MatchContext indicates whether or not `fallthrough` is usable
	MatchContext bool

	// Scope contains all the symbols local to this expression
	Scope map[string]*sem.Symbol
}

// currExprContext gets the current expression context
func (w *Walker) currExprContext() *ExprContext {
	return w.exprContextStack[len(w.exprContextStack)-1]
}

// pushFuncContext pushes a function context to the expr context stack
func (w *Walker) pushFuncContext(fn *typing.FuncType) {
	// we don't propagate loop flags into sub-functions
	w.exprContextStack = append(w.exprContextStack, &ExprContext{FuncContext: fn, FuncArgScope: true})
}

// pushLoopContext pushes the context inside a loop
func (w *Walker) pushLoopContext() {
	// we know this will always be called inside an enclosing context so we can
	// safely use `currExprContext` to propagate context flags and values down
	// as we need them
	w.exprContextStack = append(w.exprContextStack, &ExprContext{
		LoopContext:  true,
		MatchContext: w.currExprContext().MatchContext,
		FuncContext:  w.currExprContext().FuncContext,
		Scope:        make(map[string]*sem.Symbol),
	})
}

// pushLoopContext pushes the context inside a matchj
func (w *Walker) pushMatchContext() {
	// we know this will always be called inside an enclosing context so we can
	// safely use `currExprContext` to propagate context flags and values down
	// as we need them
	w.exprContextStack = append(w.exprContextStack, &ExprContext{
		LoopContext:  w.currExprContext().LoopContext,
		MatchContext: true,
		FuncContext:  w.currExprContext().FuncContext,
		Scope:        make(map[string]*sem.Symbol),
	})
}

// popExprContext pops the top element off the expr context stack
func (w *Walker) popExprContext() {
	if len(w.exprContextStack) == 0 {
		logging.LogFatal("attempted to pop from empty expression context stack")
	}

	w.exprContextStack = w.exprContextStack[:len(w.exprContextStack)-1]
}
