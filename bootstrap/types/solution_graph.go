package types

import "chaic/report"

// pruneSet is a set of nodes that pruned by a particular unification.
type pruneSet struct {
	// The nodes that were pruned.
	Pruned []*subNode

	// The span of source text causing the unification that pruned them.
	Span *report.TextSpan
}

// typeVarNode represents a type variable node within the solution graph.
type typeVarNode struct {
	// The type variable represented by the node.
	Var *TypeVariable

	// The list of substitution nodes belonging to the type variable.
	Nodes map[uint64]*subNode

	// The order to select the default substitutions in if no type can be
	// inferred for the variable.  If this list is empty, then the variable
	// cannot default.
	DefaultOrder []uint64

	// Whether this nodes types were entirely inferred or given as overloads.
	Known bool

	// The list of substitution nodes pruned from this type variable.  These are
	// grouped in order of the unification the pruned them.
	PruneSets []pruneSet

	// The list of property constraints applied to the type variable node.
	Properties []propertyConstraint
}

// First returns the first remaining substitution node (if any) of a type variable.
func (tnode *typeVarNode) First() *subNode {
	for _, snode := range tnode.Nodes {
		return snode
	}

	return nil
}

// Default selects the first remaining node in the default order.
func (tnode *typeVarNode) Default() *subNode {
	for _, id := range tnode.DefaultOrder {
		if snode, ok := tnode.Nodes[id]; ok {
			return snode
		}
	}

	return nil
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

	// Add it the graph and to its parent.
	s.subNodes[snode.ID] = snode
	parent.Nodes[snode.ID] = snode

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

// pruneSubstitution removes a substitution node from the solution graph as well
// as all substitutions which share an edge with it.
func (s *Solver) pruneSubstitution(snode *subNode, pruning map[uint64]struct{}) {
	// First, mark the current node as already being pruned so that we don't try
	// to prune it multiple times and get caught in a cycle while pruning the
	// graph.
	pruning[snode.ID] = struct{}{}

	// Go through each edge of the current substitution node and prune them if they
	// are not already being pruned.
	for _, edge := range snode.Edges {
		if _, ok := pruning[edge.ID]; !ok {
			s.pruneSubstitution(edge, pruning)
		}
	}

	// Remove the substitution from the substitution graph.
	delete(s.subNodes, snode.ID)

	// Remove the substitution from its parent type variable's list of nodes.
	delete(snode.Parent.Nodes, snode.ID)

	// Add it to the approriate prune set in its parent type variable.
	if len(snode.Parent.PruneSets) == 0 {
		snode.Parent.PruneSets = append(snode.Parent.PruneSets, pruneSet{
			Pruned: []*subNode{snode},
			Span:   s.currSpan,
		})
	} else {
		topSet := snode.Parent.PruneSets[len(snode.Parent.PruneSets)-1]
		if topSet.Span == s.currSpan {
			topSet.Pruned = append(topSet.Pruned, snode)
		} else {
			snode.Parent.PruneSets = append(snode.Parent.PruneSets, pruneSet{
				Pruned: []*subNode{snode},
				Span:   s.currSpan,
			})
		}
	}
}
