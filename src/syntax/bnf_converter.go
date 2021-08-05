package syntax

import (
	"fmt"
)

// BNFRuleTable is a data structure used to represent the BNF grammar as it is
// converted to a LALR(1) parsing table
type BNFRuleTable struct {
	// represent the expanded set of BNF grammar rules
	RulesByIndex []*BNFRule

	// map production names to the list of rules they are related to (for
	// fast-lookup by production name)
	RulesByProdName map[string][]int
}

// BNFRule is a data structure representing a single BNF rule in the expanded
// grammar (derived from original grammar).  These rules contain only terminals,
// nonterminals, and epsilons.  All other structure is expanded out.
type BNFRule struct {
	// ProdName is the name of the production the rule belongs to
	ProdName string

	// Contents are all of the BNF elements contained in the specific rule
	Contents []BNFElement
}

// BNFElement is a single element of a BNF rule
type BNFElement interface {
	Kind() int // should be one of the enumerated BNF kinds
}

// Different kinds of BNF elements
const (
	BNFKindTerminal = iota
	BNFKindNonterminal
	BNFKindEpsilon
)

// BNFTerminal is a terminal used in a BNF rule (int = token value)
type BNFTerminal int

// Kind of BNFTerminal is BNFKindTerminal
func (BNFTerminal) Kind() int {
	return BNFKindTerminal
}

// BNFNonterminal is a nonterminal used in a BNF rule (string = prod name)
type BNFNonterminal string

// Kind of BNFNonterminal is BNFKindNonterminal
func (BNFNonterminal) Kind() int {
	return BNFKindNonterminal
}

// BNFEpsilon is just an empty byte that stores no meaningful value
type BNFEpsilon struct{}

// Kind of BNFEpsilon is BNFKindEpsilon
func (BNFEpsilon) Kind() int {
	return BNFKindEpsilon
}

// Expander is a struct used to hold the shared state used as the EBNF grammar
// is expanded into the BNF grammar (growing ruletable and anon-name counter)
type Expander struct {
	table *BNFRuleTable

	// used to generate unique anonymous production names
	anonCounter int

	// used to add the suffix to anonymously generated names
	currentProdName string
}

// expandGrammar expands and generates the BNF RuleTable from a given EBNF
// grammar.  NB: source grammar should be disposed of after this function is run!
func expandGrammar(g Grammar) *BNFRuleTable {
	e := &Expander{table: &BNFRuleTable{RulesByProdName: make(map[string][]int)}}

	for name, prod := range g {
		e.currentProdName = name
		e.expandProduction(name, prod)
	}

	// reduce the capacity of the slice (normally, it is nearly double the length!)
	e.table.RulesByIndex = append([]*BNFRule(nil), e.table.RulesByIndex[:len(e.table.RulesByIndex)]...)

	return e.table
}

// expandProduction expands a given production by the given name and adds
// its rules to the ruleTable as well as any anonymous rules that pop up
func (e *Expander) expandProduction(name string, contents Production) {
	// alternators only occur in isolation (ie. here => no problem)
	if contents[0].Kind() == GKindAlternator {
		for _, ebnfRule := range contents[0].(*AlternatorElement).groups {
			bnfRuleContents := e.expandGroup(ebnfRule)

			// we can create multiple rules for the same production
			e.addRule(name, bnfRuleContents)
		}

		return
	}

	// if we do not have alternator => expand like a regular group
	fullProdBnfRuleContents := e.expandGroup(contents)

	// if our group contains nothing but an anonymous nonterminal => inline that
	// nonterminal (convert its rules into the current production's rules)
	if len(fullProdBnfRuleContents) == 1 {
		if nt, ok := fullProdBnfRuleContents[0].(BNFNonterminal); ok && nt[0] == '$' {
			e.inlineAnonNonterminal(name, nt)
			return
		}
	}

	// otherwise, assume that the group is acceptable as our main rule
	e.addRule(name, fullProdBnfRuleContents)
}

// expandGroup converts a group into a (set of) BNF rules (can be a production)
func (e *Expander) expandGroup(group []GrammaticalElement) []BNFElement {
	// a group will be the same size as its original production
	ruleContents := make([]BNFElement, len(group))

	// NOTE: all items that create anonymous productions will also insert a
	// nonterminal reference to that production at their position in the rule
	for i, item := range group {
		switch item.Kind() {
		case GKindTerminal:
			// the terminal kind -1 denotes an epsilon
			if item.(Terminal) == -1 {
				ruleContents[i] = BNFEpsilon{}
			} else {
				ruleContents[i] = BNFTerminal(item.(Terminal))
			}
		case GKindNonterminal:
			ruleContents[i] = BNFNonterminal(item.(Nonterminal))
		// group just creates a new anonymous production for the group unless
		// the group can be inlined (contains no alternators and one element)
		case GKindGroup:
			subGroup := item.(*GroupingElement).elements

			if len(subGroup) == 1 && subGroup[0].Kind() != GKindAlternator {
				ruleContents[i] = e.expandGroup(subGroup)[0]
			} else {
				anonName := e.getAnonName()
				e.expandProduction(anonName, subGroup)
				ruleContents[i] = BNFNonterminal(anonName)
			}
		// creates a new anonymous production with an epsilon rule
		case GKindOptional:
			anonName := e.getAnonName()

			e.expandProduction(anonName, item.(*GroupingElement).elements)
			e.epsilonRule(anonName)

			ruleContents[i] = BNFNonterminal(anonName)
		// same process as optionals but adds a reference to itself at the end
		// of all of the normal rules ({ F } => E' where E' -> F E' | epsilon)
		case GKindRepeat:
			anonName := e.getAnonName()

			e.expandProduction(anonName, item.(*GroupingElement).elements)
			ntRef := BNFNonterminal(anonName)

			for _, ruleRef := range e.table.RulesByProdName[anonName] {
				e.table.RulesByIndex[ruleRef].Contents = append(e.table.RulesByIndex[ruleRef].Contents, ntRef)
			}

			// add epsilon rule after self-references (so it doesn't get one)
			e.epsilonRule(anonName)

			ruleContents[i] = ntRef
		// suite groups can split into two rules: one where the group is wrapped
		// in a suite and one where it isn't.  The group itself is lowered into
		// anonymous production unless it is length 1 (for efficiency).  Either
		// way, a new anonymous production is created for the suite group and a
		// nonterminal reference is inserted in its place in the given rule.
		case GKindSuite:
			subGroup := item.(*GroupingElement).elements
			var subGroupRef BNFElement

			if len(subGroup) == 1 && subGroup[0].Kind() != GKindAlternator {
				subGroupRef = e.expandGroup(subGroup)[0]
			} else {
				anonName := e.getAnonName()
				e.expandProduction(anonName, subGroup)
				subGroupRef = BNFNonterminal(anonName)
			}

			suiteAnonName := e.getAnonName()
			e.addRule(suiteAnonName, []BNFElement{subGroupRef})
			e.addRule(suiteAnonName, []BNFElement{
				BNFTerminal(INDENT), subGroupRef, BNFTerminal(DEDENT),
			})

			ruleContents[i] = BNFNonterminal(suiteAnonName)
		}
		// all other kinds should not occur here (alternator handled above, flags are unused rn)
	}

	return ruleContents
}

// addRule adds a new rule to the current rule set from the given production
func (e *Expander) addRule(prodName string, contents []BNFElement) {
	if ruleRefs, ok := e.table.RulesByProdName[prodName]; ok {
		e.table.RulesByProdName[prodName] = append(ruleRefs, len(e.table.RulesByIndex))
	} else {
		e.table.RulesByProdName[prodName] = []int{len(e.table.RulesByIndex)}
	}

	e.table.RulesByIndex = append(e.table.RulesByIndex, &BNFRule{ProdName: prodName, Contents: contents})
}

// inlineAnonNonterminal converts the rules of an anonymous nonterminal into
// those of a named nonterminal (to reducing the number of rules in the table)
func (e *Expander) inlineAnonNonterminal(newName string, nonterminal BNFNonterminal) {
	ntName := string(nonterminal)

	// assume the nonterminal is already defined
	ruleRefs := e.table.RulesByProdName[ntName]
	delete(e.table.RulesByProdName, ntName)
	e.table.RulesByProdName[newName] = ruleRefs

	for _, r := range ruleRefs {
		e.table.RulesByIndex[r].ProdName = newName
	}
}

// epsilonRule creates an epsilon rule for the current production (by name)
func (e *Expander) epsilonRule(name string) {
	e.addRule(name, []BNFElement{BNFEpsilon(struct{}{})})
}

// getAnonName gets a new anonymous production name for the current production
func (e *Expander) getAnonName() string {
	e.anonCounter++
	return fmt.Sprintf("$%d-%s", e.anonCounter, e.currentProdName)
}
