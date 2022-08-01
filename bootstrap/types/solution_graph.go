package types

// typeVarNode represents a type variable node within the solution graph.
type typeVarNode struct {
	// The type variable represented by the node.
	Var *TypeVariable

	// The list of substitution nodes belonging to the type variable.
	Nodes []*subNode

	// Whether or not the default to the first substitution in the list of
	// substitution nodes if there is more than one possible substitution when
	// finalizing type deduction for the type variable.
	Default bool
}

// addSubstitution adds a substitution to a type variable node.
func (s *Solver) addSubstitution(parent *typeVarNode, sub substitution) *subNode {
	// Create the new substitution node.
	snode := &subNode{
		ID:     s.subIDCounter,
		Sub:    sub,
		Parent: parent,
	}
	s.subIDCounter++

	// Add it the graph and to parent.
	s.subNodes[snode.ID] = snode
	parent.Nodes = append(parent.Nodes, snode)

	// Return the newly created substitution node.
	return snode
}

/* -------------------------------------------------------------------------- */

// subNode represents a substitution node within the solution graph.
type subNode struct {
	// The unique ID of the substitution node.
	ID uint64

	// The substitution represented by the substitution node.
	Sub substitution

	// The type variable node of the type variable for which the node
	// represents a possible substitution.
	Parent *typeVarNode

	// The list of substitution nodes the node shares an edge with.
	Edges []*subNode
}

// IsOverload returns whether the node is an overload: whether its parent has
// more than one possible substitution.
func (snode *subNode) IsOverload() bool {
	return len(snode.Parent.Nodes) > 1
}

// AddEdge adds an edge between the node and another substitution node.
func (snode *subNode) AddEdge(osnode *subNode) {
	snode.Edges = append(snode.Edges, osnode)
	osnode.Edges = append(osnode.Edges, snode)
}
