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

	// controlStack is the stack of control flow frames (ie. functions, if
	// statements, etc) that are generated within a given expression.  This
	// frames are used to keep track of control flow effects.  This stack
	// doesn't necessarily align with the expr context stack (eg. do blocks) so
	// it has to be stored separately
	controlStack []*ControlFrame

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

// pushScopeContext pushes a standard scope expression context
func (w *Walker) pushScopeContext() {
	w.exprContextStack = append(w.exprContextStack, &ExprContext{
		LoopContext:  w.currExprContext().LoopContext,
		MatchContext: w.currExprContext().MatchContext,
		FuncContext:  w.currExprContext().FuncContext,
		Scope:        make(map[string]*sem.Symbol),
	})
}

// pushFuncContext pushes a function context to the expr context stack
func (w *Walker) pushFuncContext(fn *typing.FuncType) {
	// we don't propagate loop flags into sub-functions
	w.exprContextStack = append(w.exprContextStack, &ExprContext{
		FuncContext:  fn,
		FuncArgScope: true,
		Scope:        make(map[string]*sem.Symbol),
	})
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

// -----------------------------------------------------------------------------

// ControlFrame represents a frame in which control flow is applied
type ControlFrame struct {
	// ControlKind is the control flow kind that applied in this frame.  This
	// should be one of the enumerated control flow kinds.  Defaults to
	// `CFUnknown`
	ControlKind int

	// FrameKind indicates the top of control flow enclosing the block.  Should
	// be one of the enumerated frame kinds.
	FrameKind int
}

// Enumeration of control flow kinds
const (
	CFNone    = iota // No change to control flow
	CFReturn         // Return from function
	CFLoop           // Change the control flow of a loop
	CFMatch          // Changes the control flow of a match
	CFPanic          // Panic and exit block
	CFUnknown        // Control flow effect is not yet known
)

// Enumeration of control frame kinds
const (
	FKFunc  = iota // Function control frame
	FKLoop         // Loop frame
	FKMatch        // Match Frame
	FKIf           // If Frame
)

// pushControlFrame pushes a new control flow frame onto the stack
func (w *Walker) pushControlFrame(frameKind int) {
	w.controlStack = append(w.controlStack, &ControlFrame{
		ControlKind: CFUnknown, FrameKind: frameKind,
	})
}

// popControlFrame pops a control flow frame off the control flow stack and
// updates any enclosing frames with the effect of that popping
func (w *Walker) popControlFrame() {
	topFrame := w.controlStack[len(w.controlStack)-1]
	w.controlStack = w.controlStack[:len(w.controlStack)-1]

	if len(w.controlStack) > 0 {
		// only want to propagate control flow if the control frame itself
		// doesn't stop it from bubbling; this switch will exit the function
		// early if the top frame catches the control flow change
		switch topFrame.FrameKind {
		case FKFunc:
			// function catches all
			return
		case FKLoop:
			if topFrame.ControlKind == CFLoop {
				// loop control caught by loop
				return
			}
		case FKMatch:
			if topFrame.ControlKind == CFMatch {
				// match control caught by match
				return
			}
		}

		// if we reach here, an update should occur
		w.updateControl(topFrame.ControlKind)
	}
}

// hasNoControlEffect returns whether or not the top most control frame has a
// control effect.  This is used for testing whether or not value constraints
// need to be added to the results of blocks
func (w *Walker) hasNoControlEffect() bool {
	topControlKind := w.controlStack[len(w.controlStack)-1].ControlKind

	return topControlKind == CFNone || topControlKind == CFUnknown
}

// updateControl updates the control kind of the enclosing control frame
func (w *Walker) updateControl(kind int) {
	topFrame := w.controlStack[len(w.controlStack)-1]

	if topFrame.FrameKind == FKMatch || topFrame.FrameKind == FKIf {
		// if the control frame is branched, the control kind is updated to be
		// the "lowest" control flow effect:P eg. if there is a `break` on one
		// branch and a `return` on another, then the control flow kind is
		// `break` since break is the lowest result of the code path
		switch topFrame.ControlKind {
		case CFNone:
			// no consistent control flow effect (eg. empty branch)
			break
		case CFReturn, CFUnknown, CFPanic:
			// everything overrides `return` (highest control flow), `panic`
			// (tied with return), and `unknown` (default value -- no known
			// control flow yet)
			topFrame.ControlKind = kind
		case CFLoop:
			if kind == CFNone {
				// none overrides all control flow
				topFrame.ControlKind = kind
			} else if kind == CFMatch {
				// only control flow that *could* override loop is match, but
				// only if the match is enclosed by the loop (ie. match is the
				// inner context)
				for i := len(w.exprContextStack) - 1; i >= 0; i-- {
					ctx := w.exprContextStack[i]

					if !ctx.MatchContext {
						// match context exited: match is inner
						topFrame.ControlKind = CFMatch
						break
					} else if !ctx.LoopContext {
						// loop context exited before match: loop is inner
						break
					}
				}
			}
		case CFMatch:
			if kind == CFNone {
				// none overrides all control flow
				topFrame.ControlKind = kind
			} else if kind == CFLoop {
				// only control flow that *could* override match is loop, but
				// only if the loop is enclosed by the match (ie. match is the
				// inner context)
				for i := len(w.exprContextStack) - 1; i >= 0; i-- {
					ctx := w.exprContextStack[i]

					if !ctx.MatchContext {
						// match context exited: match is inner
						break
					} else if !ctx.LoopContext {
						// loop context exited before match: loop is inner
						topFrame.ControlKind = CFLoop
						break
					}
				}
			}
		}
	} else if topFrame.ControlKind == CFUnknown {
		// for sequential control frames (FKFunc, FKLoop), the kind is only
		// updated once since any control flow will automatically exit the
		// sequential block
		topFrame.ControlKind = kind
	}
}
