package types

// unifyResult represents the complex result of a unification.
type unifyResult struct {
	// Whether or not the unification was successful.
	Unified bool

	// The map of the IDS of substitution nodes visited during unification
	// associated with whether they should be pruned once the equivalency
	// assertion which causes unification completes (true => prune).
	Visited map[uint64]bool

	// The set of IDs of nodes that we given substitutions by this unification.
	Completes map[uint64]struct{}
}

// Both combines two unification results in such a way that both must hold
// true in order for the yielded unification to hold true.
func (ur *unifyResult) Both(other *unifyResult) {
	ur.Unified = ur.Unified && other.Unified

	for k, ov := range other.Visited {
		if v, ok := ur.Visited[k]; ok {
			ur.Visited[k] = v || ov
		} else {
			ur.Visited[k] = ov
		}
	}

	for k := range other.Completes {
		ur.Completes[k] = struct{}{}
	}
}

// Either combines two unification results in such a way that either unification
// can hold true for the yielded unification result to hold true.
func (ur *unifyResult) Either(other *unifyResult) {
	ur.Unified = ur.Unified || other.Unified

	for k, v := range ur.Visited {
		if ov, ok := other.Visited[k]; ok {
			ur.Visited[k] = v && ov
		} else {
			delete(ur.Visited, k)
		}
	}

	for k := range ur.Completes {
		if _, ok := other.Completes[k]; !ok {
			delete(ur.Completes, k)
		}
	}
}

/* -------------------------------------------------------------------------- */

// unify attempts two make two types equal by substitution.  If a root node is
// given, (ie. is not nil), then this equality must be contingent upon the given
// root substitution node holding.
func (s *Solver) unify(root *subNode, lhs, rhs Type) *unifyResult {
	// Check for right-hand type variables since we type switch down the left.
	if rtv, ok := rhs.(*TypeVariable); ok {
		// Make sure we don't unify a type variable with itself.
		if ltv, ok := lhs.(*TypeVariable); ok {
			if ltv.ID == rtv.ID {
				return &unifyResult{Unified: true}
			}
		}

		return s.unifyTypeVar(root, rtv, lhs)
	}

	// From here on, RHS can never be a type variable.

	// Match the shape of the types first.
	switch v := lhs.(type) {
	case *TypeVariable:
		return s.unifyTypeVar(root, v, rhs)
	case *PointerType:
		if rpt, ok := rhs.(*PointerType); ok && v.Const == rpt.Const {
			return s.unify(root, v.ElemType, rpt.ElemType)
		}
	case *FuncType:
		if rft, ok := rhs.(*FuncType); ok && len(v.ParamTypes) == len(rft.ParamTypes) {
			var result *unifyResult
			for i, paramType := range v.ParamTypes {
				presult := s.unify(root, paramType, rft.ParamTypes[i])

				if i == 0 {
					result = presult
				} else {
					result.Both(presult)
				}

				if !result.Unified {
					return result
				}
			}

			if result == nil {
				return s.unify(root, v.ReturnType, rft.ReturnType)
			} else {
				result.Both(s.unify(root, v.ReturnType, rft.ReturnType))
				return result
			}
		}
	case PrimitiveType:
		if rpt, ok := rhs.(PrimitiveType); ok {
			return &unifyResult{Unified: v == rpt}
		}
	}

	// Default to false if the shapes don't match.
	return &unifyResult{Unified: false}
}

// unifyTypeVar attempts to make a type variable equal to the given type by
// substitution.  If a root node is given (ie. is not nil), then this equality
// must be contingent upon the given root substitution node holding.
func (s *Solver) unifyTypeVar(root *subNode, tvar *TypeVariable, typ Type) *unifyResult {
	tnode := s.typeVarNodes[tvar.ID]

	var result *unifyResult

	// Handle complete type variables.
	if _, ok := s.completes[tvar.ID]; ok {
		// Go through each substitution of the given type variable.
		for i, subNode := range tnode.Nodes {
			// Unify the substitution with the inputted type making the
			// substitution the root of unification.
			subResult := s.unify(subNode, subNode.Sub.Type(), typ)

			// Merge the substitution unification result with the overall
			// unification result.
			if i == 0 {
				result = subResult
			} else {
				result.Either(subResult)
			}

			if subResult.Unified { // Substitution unification succeeded.
				// Indicate that this substitution node should not be pruned.
				result.Visited[subNode.ID] = false

				// If there is a root, add an edge between this substitution and
				// that unification root.
				if root != nil {
					root.AddEdge(subNode)
				}
			} else { // Substitution unification failed.
				// Indicate that this substitution should be pruned.
				result.Visited[subNode.ID] = true
			}
		}
	} else { // Handle incomplete type variables.
		// Mark the type variable as completed by this unification.
		result = &unifyResult{Unified: true, Completes: map[uint64]struct{}{tvar.ID: {}}}

		// Add a substitution of the given type to the type variable if it does not
		// already have a substitution which is equal to the given type.
		var matchingSubNode *subNode
		for _, subNode := range tnode.Nodes {
			if sresult := s.unify(subNode, subNode.Sub.Type(), typ); sresult.Unified {
				result.Both(sresult)
				matchingSubNode = subNode
				break
			}
		}

		if matchingSubNode == nil {
			matchingSubNode = s.addSubstitution(tnode, basicSubstitution{typ: typ})
		}

		// Add an edge between the root and the matching substitution.
		if root != nil {
			root.AddEdge(matchingSubNode)
		}
	}

	// If we are not dealing with an overload substitution, update the global
	// list of completed type variable with the completed type variable of the
	// overall result.
	if root == nil || !root.IsOverload() {
		for k := range result.Completes {
			s.completes[k] = struct{}{}
		}
	}

	// Return the generated unification result.
	return result
}
