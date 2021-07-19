package walk

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"chai/typing"
	"fmt"
)

// WalkPredicates walks the predicates (func bodies, initializers) of a file
func (w *Walker) WalkPredicates(root *sem.HIRRoot) {
	for _, def := range root.Defs {
		switch v := def.(type) {
		case *sem.HIRFuncDef:
			ft := v.DefBase.Sym().Type.(*typing.FuncType)

			// validate the argument initializers
			for name, argInit := range v.ArgumentInitializers {
				var expectedType typing.DataType
				for _, arg := range ft.Args {
					if arg.Name == name {
						expectedType = arg.Type
						break
					}
				}

				argBranch := convertIncompleteToBranch(argInit)
				if expr, ok := w.walkExpr(argBranch, true); ok {
					w.solver.AddSubConstraint(expectedType, expr.Type(), argBranch.Position())

					// solve the argument initializer immediately -- the only
					// context should be that of the initializer
					if w.solver.Solve() {
						v.ArgumentInitializers[name] = expr
					}
				}
			}

			bodyBranch := convertIncompleteToBranch(v.Body)

			// skip external or intrinsic functions
			if bodyBranch == nil {
				continue
			}

			// validate the function body
			if expr, ok := w.walkFuncBody(bodyBranch, ft); ok {
				v.Body = expr
			}
		case *sem.HIROperDef:
			bodyBranch := convertIncompleteToBranch(v.Body)

			// skip intrinsic operators
			if bodyBranch == nil {
				continue
			}

			// validate the operator body
			if expr, ok := w.walkFuncBody(bodyBranch, v.DefBase.Sym().Type.(*typing.FuncType)); ok {
				v.Body = expr
			}
		}
	}
}

// -----------------------------------------------------------------------------

// Many expression functions take a boolean parameter called `yieldsValue` which
// essentially indicates whether or not the block yields a meaningful value.
// This is used to facilitate the behavior of if-blocks, match-blocks, etc. that
// can yield multiple different, ununifiable types on different branches because
// their value is not used.

// walkFuncBody walks a function body (`expr`)
func (w *Walker) walkFuncBody(branch *syntax.ASTBranch, fn *typing.FuncType) (sem.HIRExpr, bool) {
	// handle function context management
	w.pushFuncContext(fn)
	defer w.popExprContext()

	// handle control frame management
	w.pushControlFrame(FKFunc)
	defer w.popControlFrame()

	// a function that returns nothing effectively yields no meaningful value
	yieldsValue := true
	if pt, ok := fn.ReturnType.(typing.PrimType); ok && pt == typing.PrimKindNothing {
		yieldsValue = false
	}

	// walk the function body
	if hirExpr, ok := w.walkExpr(branch, yieldsValue); ok {
		// constraint the return type of the block if the function yields a
		// value and it the body does not have any unconditional control flow
		if yieldsValue && w.hasNoControlEffect() {
			w.solver.AddSubConstraint(fn.ReturnType, hirExpr.Type(), branch.Position())
		}

		// run the solver at the end of the function body
		if !w.solver.Solve() {
			return nil, false
		}

		// if we reach here, then the body was walked successfully
		return hirExpr, true
	} else {
		return nil, false
	}
}

// walkExprList walks an `expr_list` node.  It assumes that the expressions must
// yield a value and cannot cause control flow changes (only context in which
// `expr_list` is used)
func (w *Walker) walkExprList(branch *syntax.ASTBranch) ([]sem.HIRExpr, bool) {
	exprs := make([]sem.HIRExpr, branch.Len()/2+1)
	for i, item := range branch.Content {
		// only branch is `expr`
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			if expr, ok := w.walkExpr(itembranch, true); ok {
				exprs[i/2] = expr
			} else {
				return nil, false
			}
		}
	}

	return exprs, true
}

// -----------------------------------------------------------------------------

// walkExpr walks an expression node and returns a HIRExpr
func (w *Walker) walkExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	exprBranch := branch.BranchAt(0)

	switch exprBranch.Name {
	case "do_block":
		return w.walkDoBlock(exprBranch, yieldsValue)
	case "simple_expr":
		return w.walkSimpleExpr(exprBranch, yieldsValue)
	case "block_expr":
		return w.walkBlockExpr(exprBranch, yieldsValue)
	}

	// unreachable
	return nil, false
}

// walkSimpleExpr walks a simple expression (`simple_expr`)
func (w *Walker) walkSimpleExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	coreExpr := branch.BranchAt(0)
	if coreExpr.Name == "core_expr" {
		// TODO: handle then clauses

		// core expressions must yield a value if there is a then clause
		return w.walkCoreExpr(coreExpr, yieldsValue || branch.Len() == 2)
	} else if coreExpr.Name == "lambda" {
		// TODO
	}

	// unreachable
	return nil, false
}

// walkCoreExpr walks a `core_expr` node
func (w *Walker) walkCoreExpr(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	// awaited := false
	var root sem.HIRExpr
	for _, item := range branch.Content {
		if itembranch, ok := item.(*syntax.ASTBranch); ok {
			if itembranch.Name == "or_expr" {
				// if the branch has a length greater than one, then it either
				// has a suffix, an await, or both -- all of which imply it must
				// yield value
				if opRoot, ok := w.walkBinOperatorApp(itembranch, yieldsValue || branch.Len() > 1); ok {
					root = opRoot
				} else {
					return nil, false
				}
			} else /* core_expr_suffix */ {
				firstLeaf := itembranch.LeafAt(0)
				switch firstLeaf.Kind {
				case syntax.COLON:
					// type proposition
					if dt, ok := w.walkTypeLabel(itembranch.BranchAt(2)); ok {
						w.solver.AddEqConstraint(dt, root.Type(), itembranch.Position())
					} else {
						return nil, false
					}
				case syntax.AS:
					// type cast
					if dt, ok := w.walkTypeLabel(itembranch.BranchAt(1)); ok {
						w.solver.AddTypeAssertion(typing.AssertCast, root.Type(), dt, itembranch.Position())
						root = &sem.HIRCast{
							ExprBase: sem.NewExprBase(dt, root.Category(), root.Immutable()),
							Root:     root,
						}
					} else {
						return nil, false
					}

				}
			}
		} else {
			// only token is `await`
			// awaited = true
		}
	}

	return root, true
}

// walkBinOperatorApp walks a binary operator application
func (w *Walker) walkBinOperatorApp(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	switch branch.Name {
	case "unary_expr":
		return w.walkUnaryOperatorApp(branch, yieldsValue)
	case "comp_expr":
		// if the branch is length 1, then there is no operator application
		// being performed and we can just recur downward
		if branch.Len() == 1 {
			return w.walkBinOperatorApp(branch.BranchAt(0), yieldsValue)
		}

		// comparisons in Chai work like those in Python: comparison operators
		// will apply to create multiple relations to the same value.  For
		// example, `a < b < c` is interpreted as one comparison: `b` is between
		// `a` and `c` rather than as separate comparisons.  This leads to
		// cleaner code that reads more like mathematical notation

		// comparisons contains a list of all the comparisons performed broken
		// down indivually: eg. `a < b < c` splits into `a < b` and `b < c`.
		// These will all be anded together at the end
		var comparisons []sem.HIRExpr

		// op is the operator currently being applied
		var op *sem.Operator

		// lhs is the left operand of the current operator
		var lhs sem.HIRExpr

		for i, item := range branch.Content {
			switch v := item.(type) {
			case *syntax.ASTLeaf:
				// only leaf is operator
				var ok bool
				op, ok = w.lookupOperator(v.Kind, 2)
				if !ok {
					w.logMissingOpOverload(v.Value, 2, v.Position())
					return nil, false
				}
			case *syntax.ASTBranch:
				// only branch is sub node; operator applications imply that
				// expression must yield a value (in order to be operated upon)

				// if `lhs` is `nil`, then we are collecting the first operand
				if lhs == nil {
					if expr, ok := w.walkBinOperatorApp(branch, true); ok {
						lhs = expr
					} else {
						return nil, false
					}
				} else /* we are collecting a right operand */ {
					if expr, ok := w.walkBinOperatorApp(branch, true); ok {
						// first, we build the operator application
						opApp := w.makeOperatorApp(
							op,
							[]sem.HIRExpr{lhs, expr},
							// we just want to highlight the current comparison
							// since they don't associate like other binary
							// expressions
							syntax.TextPositionOfSpan(branch.Content[i-2], v),
							[]*logging.TextPosition{
								branch.Content[i-2].Position(),
								v.Position(),
							},
						)

						// then, add it to comparisons
						comparisons = append(comparisons, opApp)

						// now, the current expression (rhs) is going to become
						// the lhs of the next expression; we will always read
						// in operator first and then the following expression
						// will be interpreted as the rhs
						lhs = expr
					} else {
						return nil, false
					}
				}
			}
		}

		// if there is only one comparison, that is what we return; no
		// anding/combining required
		if len(comparisons) == 1 {
			return comparisons[0], true
		}

		// otherwise, we sequentially and all the expressions together and
		// return the final result -- this accomplishes the desired behavior of
		// `a < b < c` translating as `(&& (< a b) (< b c))`
		var andChain sem.HIRExpr

		// get the and operator as we will use it repeatedly here
		andOp, ok := w.lookupOperator(syntax.AND, 2)
		if !ok {
			w.logError(
				"operator `&&` has no binary overload to perform multi-comparisons",
				logging.LMKOperApp,
				branch.Position(),
			)
		}

		for i, item := range comparisons {
			// if there is no and chain set up yet, the first comparison becomes
			// the root of the chain
			if andChain == nil {
				andChain = item
			} else {
				// otherwise, we just combine the two operators with an and and
				// have that be the new root of the chain -- bubbling outward
				andChain = w.makeOperatorApp(
					andOp,
					[]sem.HIRExpr{andChain, item},
					// Position Formula Explanation:
					// (< a b) (< b c) (< c d)
					//    0       1       2
					// => i = 1, 2
					// a < b < c < d
					//   ^ ^
					// i +/- 1
					// a < b < c < d
					// ^   ^
					//   ^   ^
					// i * 2
					// a < b < c < d
					// ^       ^
					//     ^       ^
					syntax.TextPositionOfSpan(branch.Content[(i-1)*2], branch.Content[(i+1)*2]),
					// using same logical formula as one above to get sub-expressions
					[]*logging.TextPosition{
						syntax.TextPositionOfSpan(branch.Content[(i-1)*2], branch.Content[i*2]),
						syntax.TextPositionOfSpan(branch.Content[i*2], branch.Content[(i+1)*2]),
					},
				)
			}
		}

		return andChain, true
	default:
		// if the branch is length 1, then there is no operator application
		// being performed and we can just recur downward
		if branch.Len() == 1 {
			return w.walkBinOperatorApp(branch.BranchAt(0), yieldsValue)
		}

		// op is operator currently being applied
		var op *sem.Operator

		// rootOperand to the operator may be made up of previous operator
		// applications -- eg. `1 + 2 + 3` -> `(+ (+ 1 2) 3)`
		var rootOperand sem.HIRExpr

		for i, item := range branch.Content {
			switch v := item.(type) {
			case *syntax.ASTLeaf:
				// only leaf is the operator
				var ok bool
				op, ok = w.lookupOperator(v.Kind, 2)
				if !ok {
					w.logMissingOpOverload(v.Value, 2, v.Position())
					return nil, false
				}
			case *syntax.ASTBranch:
				// only branch is sub node; operator applications imply that
				// expression must yield a value (in order to be operated upon)
				if expr, ok := w.walkBinOperatorApp(v, true); ok {
					if op == nil {
						// if the operator is `nil`, we haven't collected it yet
						// and therefore can just store our expression in the
						// root (this is the first operand)
						rootOperand = expr
					} else {
						rootOperand = w.makeOperatorApp(
							op,
							[]sem.HIRExpr{rootOperand, expr},
							// exprPos is all of the expression up to and
							// include the current rhs
							syntax.TextPositionOfSpan(branch, v),
							[]*logging.TextPosition{
								// left expression is from the root of the
								// branch up until the current operator
								syntax.TextPositionOfSpan(branch, branch.Content[i-2]),
								v.Position(),
							},
						)
					}
				} else {
					return nil, false
				}
			}

		}

		return rootOperand, true
	}
}

// walkUnaryOperatorApp walks a unary operator application
func (w *Walker) walkUnaryOperatorApp(branch *syntax.ASTBranch, yieldsValue bool) (sem.HIRExpr, bool) {
	// result will store the result expression of walking the unary operator
	var result sem.HIRExpr

	// prefixOp is the prefix operator being applied
	var prefixOp *sem.Operator

	// prefix operators take precedence over postfix operators so this loop
	// works fine; we will wrap prefix applications in postfix applications
	for _, item := range branch.Content {
		switch v := item.(type) {
		case *syntax.ASTLeaf:
			// handle reference operators
			switch v.Kind {
			case syntax.AMP:
				// assert that the root type is not a reference
				w.solver.AddTypeAssertion(
					typing.AssertNonRef,
					result.Type(),
					nil,
					syntax.TextPositionOfSpan(branch, item),
				)

				// return the HIRIndirect
				return &sem.HIRIndirect{
					ExprBase: sem.NewExprBase(&typing.RefType{ElemType: result.Type()}, sem.RValue, false),
					Root:     result,
				}, true
			case syntax.STAR:
				// create a type variable to house the element type of the root
				elemType := w.solver.CreateTypeVar(nil, func() {})

				// constraint the root type to be equal to a reference to the
				// element type variable: `root == &{_}`
				w.solver.AddEqConstraint(result.Type(), &typing.RefType{ElemType: elemType}, branch.Position())

				// return the HIRDeref; note that it is an LValue not an RValue
				return &sem.HIRDereference{
					ExprBase: sem.NewExprBase(elemType, sem.LValue, false),
					Root:     result,
				}, true
			}

			// only leaf is operator
			op, ok := w.lookupOperator(v.Kind, 1)
			if !ok {
				w.logMissingOpOverload(v.Value, 1, v.Position())
				return nil, false
			}

			// if there is no result yet, this is a prefix operator
			if result == nil {
				prefixOp = op
			} else /* postfix operator */ {
				result = w.makeOperatorApp(
					op,
					[]sem.HIRExpr{result},
					branch.Position(),
					// len 2 => expr at position 0 (no prefix operator)
					// len 3 => expr at position 1 (prefix operator)
					[]*logging.TextPosition{branch.Content[(branch.Len()-1)/2].Position()},
				)
			}
		case *syntax.ASTBranch:
			// if the branch contains more than one element, there is an
			// operator and the expression must yield a value
			if atomExpr, ok := w.walkAtomExpr(v, yieldsValue || branch.Len() > 1); ok {
				// if there is a prefix operator, we apply it here
				if prefixOp != nil {
					result = w.makeOperatorApp(
						prefixOp,
						[]sem.HIRExpr{atomExpr},
						syntax.TextPositionOfSpan(branch, v),
						[]*logging.TextPosition{v.Position()},
					)
				} else /* no prefix operator */ {
					result = atomExpr
				}
			} else {
				return nil, false
			}
		}
	}

	return result, true
}

// makeOperatorApp creates a new operator application
func (w *Walker) makeOperatorApp(oper *sem.Operator, operands []sem.HIRExpr, exprPos *logging.TextPosition, opsPos []*logging.TextPosition) sem.HIRExpr {
	// generalize the operator into a generic function; start by generating type
	// variables representing the unknown arguments and return type of the
	// operator
	tvars := make([]*typing.TypeVariable, oper.Arity+1)
	for i := range tvars {
		tvars[i] = w.solver.CreateTypeVar(nil, func() {
			w.logError(
				fmt.Sprintf("unable to determine matching overload for `%s` operator", oper.Name),
				logging.LMKTyping,
				exprPos,
			)
		})
	}

	// then create the overloads for each of those type variables
	overloadValues := make(map[int][]typing.DataType)
	for _, overload := range oper.Overloads {
		switch v := overload.Signature.(type) {
		case *typing.FuncType:
			for i, arg := range v.Args {
				tvarID := tvars[i].ID

				if vals, ok := overloadValues[tvarID]; ok {
					overloadValues[tvarID] = append(vals, arg.Type)
				} else {
					overloadValues[tvarID] = []typing.DataType{arg.Type}
				}
			}

			lastTvarID := tvars[len(tvars)-1].ID
			if vals, ok := overloadValues[lastTvarID]; ok {
				overloadValues[lastTvarID] = append(vals, v.ReturnType)
			} else {
				overloadValues[lastTvarID] = []typing.DataType{v.ReturnType}
			}

			// TODO: handle generic types
		}
	}

	// generate all the type constraints and overloads
	for i, tvar := range tvars {
		// apply our overload to the type variable
		w.solver.AddOverload(tvar, overloadValues[tvar.ID]...)

		// constrain any corresponding argument types appropriately
		if i < len(opsPos) {
			w.solver.AddSubConstraint(tvar, operands[i].Type(), opsPos[i])
		}
	}

	// add all the overload correspondences; do this after all overloads have
	// been added -- can't have correspondences between non-existent overloads
	for _, tvar := range tvars {
		for _, item := range tvars {
			if item.ID != tvar.ID {
				w.solver.AddOverloadCorrespondence(tvar.ID, item.ID)
			}
		}
	}

	// create the function type that will be used for the operator
	fargs := make([]*typing.FuncArg, len(operands))
	for i := range fargs {
		// don't need to handle by-reference here
		fargs[i] = &typing.FuncArg{Type: tvars[i]}
	}

	// intrinsic name and boxed don't need to be set here; they can't be because
	// we don't know what overload we are dealing with yet
	ft := &typing.FuncType{
		Args:       fargs,
		ReturnType: tvars[len(tvars)-1],
	}

	return &sem.HIROperApply{
		ExprBase: sem.NewExprBase(ft.ReturnType, sem.RValue, false),
		Oper:     oper,
		Operands: operands,
		OperFunc: ft,
	}
}
