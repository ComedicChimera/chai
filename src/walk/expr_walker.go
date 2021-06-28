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

			// validate the function body
			if expr, ok := w.walkFuncBody(convertIncompleteToBranch(v.Body), ft); ok {
				v.Body = expr
			}
		case *sem.HIROperDef:
			// validate the operator body
			if expr, ok := w.walkFuncBody(
				convertIncompleteToBranch(v.Body),
				v.DefBase.Sym().Type.(*typing.FuncType)); ok {

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

// walkFuncBody walks a function body (`decl_func_body`)
func (w *Walker) walkFuncBody(branch *syntax.ASTBranch, fn *typing.FuncType) (sem.HIRExpr, bool) {
	// handle function context management
	w.pushFuncContext(fn)
	defer w.popExprContext()

	// maps branch of len 1 and 2 to 0, maps branch of len 3 to 1 -- always
	// giving us the correct branch for the expression
	exprBranch := branch.BranchAt(branch.Len() / 2)

	// a function that returns nothing effectively yields no meaningful value
	yieldsValue := true
	if pt, ok := fn.ReturnType.(typing.PrimType); ok && pt == typing.PrimKindNothing {
		yieldsValue = false
	}

	// walk the function body
	var hirExpr sem.HIRExpr
	switch exprBranch.Name {
	case "block_expr":
		if e, ok := w.walkBlockExpr(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	case "simple_expr":
		if e, ok := w.walkSimpleExpr(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	case "do_block":
		if e, ok := w.walkDoBlock(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	case "expr":
		if e, ok := w.walkExpr(exprBranch, yieldsValue); ok {
			hirExpr = e
		} else {
			return nil, false
		}
	}

	// constraint the return type of the block if the function yields a value
	if yieldsValue {
		w.solver.AddSubConstraint(fn.ReturnType, hirExpr.Type(), branch.Position())
	}

	// run the solver at the end of the function body
	if !w.solver.Solve() {
		return nil, false
	}

	// if we reach here, then the body was walked successfully
	return hirExpr, true
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
					// TODO
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
		// TODO: handle triple comparison expressions
		return nil, false
	default:
		// if the branch is length 1, then there is no operator application
		// being performed and we can just recur downward
		if branch.Len() == 1 {
			return w.walkBinOperatorApp(branch.BranchAt(0), yieldsValue)
		}

		// otherwise, we have an operator application

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
					w.logError(
						fmt.Sprintf("operator `%s` has no binary overloads", v.Value),
						logging.LMKName,
						v.Position(),
					)

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
							syntax.TextPositionOfSpan(branch, v),
							[]*logging.TextPosition{
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
	return nil, false
}

// makeOperatorApp creates a new operator application
func (w *Walker) makeOperatorApp(oper *sem.Operator, operands []sem.HIRExpr, exprPos *logging.TextPosition, opsPos []*logging.TextPosition) sem.HIRExpr {
	// generalize the operator into a generic function; start by generating
	// constraints representing the most general form of the operator
	qCons := make([]typing.DataType, len(oper.Overloads[0].Quantifiers))
	for _, overload := range oper.Overloads {
		// we can blindly merge quantifiers here since we know the merges are
		// valid (by the fact that the overloads are defined at all)
		for i, q := range overload.Quantifiers {
			if qCons[i] == nil {
				// copy constraint sets so append doesn't modify them
				qCons[i] = q.QType.Copy()
			} else {
				// merge constraint sets as necessary assuming the preexisting
				// constraint set can be freely modified
				qConsSet, okOld := qCons[i].(*typing.ConstraintSet)
				newQConsSet, okNew := q.QType.(*typing.ConstraintSet)

				if okOld && okNew {
					qConsSet.Set = append(qConsSet.Set, newQConsSet.Set...)
				} else if okOld {
					qConsSet.Set = append(qConsSet.Set, q.QType)
				} else if okNew {
					qCons[i] = &typing.ConstraintSet{
						SrcPackageID: w.SrcFile.Parent.ID,
						Set:          append([]typing.DataType{qCons[i]}, newQConsSet.Set...),
					}
				} else {
					qCons[i] = &typing.ConstraintSet{
						SrcPackageID: w.SrcFile.Parent.ID,
						Set:          []typing.DataType{qCons[i], q.QType},
					}
				}
			}
		}
	}

	// create our type variables based on those constraint sets
	tvars := make([]*typing.TypeVariable, len(qCons))
	for i, q := range qCons {
		tvar := w.solver.CreateTypeVar(nil, func() {
			w.logError(
				fmt.Sprintf("unable to determine overload used for `%s` operator", oper.Name),
				logging.LMKTyping,
				exprPos,
			)
		})

		// this constraint should never be errored upon since it comes first out
		// of all constraints on this type variable
		w.solver.AddSubConstraint(q, tvar, nil)
		tvars[i] = tvar
	}

	// constrain the argument types appropriately
	for i, n := range oper.ArgsForm {
		w.solver.AddSubConstraint(tvars[n], operands[i].Type(), opsPos[i])
	}

	// create the function type that will be used for the operator
	fargs := make([]*typing.FuncArg, len(operands))
	for i, n := range oper.ArgsForm {
		// don't need to handle by-reference here
		fargs[i] = &typing.FuncArg{Type: tvars[n]}
	}

	// intrinsic name and boxed don't need to be set here; they can't be because
	// we don't know what overload we are dealing with yet
	ft := &typing.FuncType{
		Args:       fargs,
		ReturnType: tvars[oper.ReturnForm],
	}

	return &sem.HIROperApply{
		ExprBase: sem.NewExprBase(ft.ReturnType, sem.RValue, false),
		Oper:     oper,
		Operands: operands,
		OperFunc: ft,
	}
}
