package syntax

// Grammar is defined to be a set of named productions
type Grammar map[string]Production

// Production itself is simply a set of grammatical elements
type Production []GrammaticalElement

// Used to designate the different kinds of grammatical constructs
const (
	GKindAlternator = iota
	GKindRepeat
	GKindSuite // group that can be in an optional suite (? ... ?)
	GKindGroup
	GKindOptional
	GKindTerminal
	GKindNonterminal
)

// GrammaticalElement represents a piece of the grammar once it is serialized
// into an object
type GrammaticalElement interface {
	Kind() int
}

// Terminal grammatical element (int representing token kind)
type Terminal int

// Nonterminal grammatical element (string representing name of production)
type Nonterminal string

// GroupingElement represents all of the other grouping grammatical elements
// (eg. groups, optionals, repeats, etc.)
type GroupingElement struct {
	kind     int
	elements []GrammaticalElement
}

// NewGroupingElement creates a new grouping element
func NewGroupingElement(kind int, elems []GrammaticalElement) *GroupingElement {
	return &GroupingElement{kind: kind, elements: elems}
}

// Kind of a terminal is GKindTerminal
func (Terminal) Kind() int {
	return GKindTerminal
}

// Kind of a nonterminal is GKindNonterminal
func (Nonterminal) Kind() int {
	return GKindNonterminal
}

// Kind returns the kind of grouping elements based on the stored kind member
// variable
func (g *GroupingElement) Kind() int {
	return g.kind
}

// AlternatorElement represents a grammatical alternator storing a slice of the
// subgroups it alternates between
type AlternatorElement struct {
	groups [][]GrammaticalElement
}

// NewAlternatorElement create a new alternator element from some number of
// groups efficiently
func NewAlternatorElement(groups ...[]GrammaticalElement) *AlternatorElement {
	return &AlternatorElement{groups: groups}
}

// Kind of alternator is GKindAlternator
func (AlternatorElement) Kind() int {
	return GKindAlternator
}

// PushFront pushes a group onto the front of alternator element
func (ae *AlternatorElement) PushFront(group []GrammaticalElement) {
	ae.groups = append([][]GrammaticalElement{group}, ae.groups...)
}
