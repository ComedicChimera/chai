package syntax

import (
	"fmt"
	"reflect"
)

const _goalSymbol = "file"

// constructParsingTable takes in an input rule table and attempt to build a
// parsing table for it.  If it fails, a descriptive error is returned.  If it
// succeeds a full parsing table is returned.  NOTE: this function resolves
// shift-reduce conflicts in favor of SHIFT (will warn if it does this)!
func constructParsingTable(brt *BNFRuleTable) (*ParsingTable, error) {
	ptableBuilder := PTableBuilder{BNFRules: brt, firstSets: make(map[string][]int)}

	if err := ptableBuilder.build(); err != nil {
		return nil, err
	}

	return ptableBuilder.Table, nil
}

// LRItem represents an LR(0) item
type LRItem struct {
	// Rule refers to the rule number in the RuleTable not
	// the final rule number in the parsing table
	Rule int

	// DotPos refers to the index the dot is considered to
	// be placed BEFORE (so a dot at the end of the item
	// would have a dot pos == to the length of the rule)
	DotPos int
}

// LRItemSet represents a complete LR(1) state
type LRItemSet struct {
	// Items matches each LRItem with its lookaheads thereby rendering each
	// entry an LR(1) item.  The lookaheads are stored as map for fast merging
	Items map[LRItem]map[int]struct{}

	// Conns represents all of the possible progressions for a given item set
	Conns map[BNFElement]int
}

// PTableBuilder holds the state used to construct the parsing table
type PTableBuilder struct {
	BNFRules *BNFRuleTable
	Table    *ParsingTable

	ItemSets []*LRItemSet

	// allows us to memoize first sets by nonterminals
	firstSets map[string][]int
}

// build uses the current builds a full parsing table for a given rule table
func (ptb *PTableBuilder) build() error {
	// augment the rule set
	augRuleNdx := len(ptb.BNFRules.RulesByIndex)
	ptb.BNFRules.RulesByIndex = append(ptb.BNFRules.RulesByIndex,
		&BNFRule{ProdName: "_END_", Contents: []BNFElement{BNFNonterminal(_goalSymbol)}})
	ptb.BNFRules.RulesByProdName["_END_"] = []int{augRuleNdx}

	// create the starting LR(1) kernel
	startSet := &LRItemSet{Items: map[LRItem]map[int]struct{}{
		{Rule: augRuleNdx, DotPos: 0}: {EOF: struct{}{}},
	}}

	// create an initial LR(0) item with an LR(1) kernel for the start set
	ptb.closureOf(startSet)

	// calculate all of the LR(0) item sets
	ptb.nextLR0Items(startSet)

	// convert all of the LR(0) item sets to LR(1) item sets
	ptb.createLR1Items(startSet)

	// for i, set := range ptb.ItemSets {
	// 	for item := range set.Items {
	// 		name := ptb.BNFRules.RulesByIndex[item.Rule].ProdName

	// 		if name == "arg_decl" {
	// 			fmt.Printf("State %d: ", i)
	// 			ptb.printSet(set)
	// 			break
	// 		}
	// 	}
	// }

	// for i, set := range ptb.ItemSets {
	// 	fmt.Printf("State %d: ", i)
	// 	ptb.printSet(set)
	// }

	// the number of states (rows) will be equivalent to the number of item sets
	// since the merging has already occurred in generating the item sets
	ptb.Table = &ParsingTable{Rows: make([]*PTableRow, len(ptb.ItemSets))}

	return ptb.buildTableFromSets()
}

// nextLR0Items computes all of the connections for the current item set and
// recursively creates sets as necessary while adding them to an item graph
func (ptb *PTableBuilder) nextLR0Items(itemSet *LRItemSet) {
	// add our item set to the item set graph (assume it has not been added)
	ptb.ItemSets = append(ptb.ItemSets, itemSet)

	// calculate all of the goto kernels that are connected to the item set
	gotoKernels := make(map[BNFElement]*LRItemSet)
	for item := range itemSet.Items {
		bnfRule := ptb.BNFRules.RulesByIndex[item.Rule]

		// we can only apply goto if the dot is not at the end of the rule and
		// we are not dealing with an epsilon rule (and all rules that start
		// with epsilon will be epsilon rules due to how our rules generate)
		if item.DotPos < len(bnfRule.Contents) && bnfRule.Contents[item.DotPos].Kind() != BNFKindEpsilon {
			gotoItem := LRItem{Rule: item.Rule, DotPos: item.DotPos + 1}
			dottedElem := bnfRule.Contents[item.DotPos]

			// in both cases, simply create an empty map for to hold lookaheads
			if gotoKernel, ok := gotoKernels[dottedElem]; ok {
				// if a goto kernel already exists, add our item to the kernel
				gotoKernel.Items[gotoItem] = make(map[int]struct{})
			} else {
				// otherwise, create a new blank item set to hold our goto kernel
				gotoKernels[dottedElem] = &LRItemSet{Items: map[LRItem]map[int]struct{}{gotoItem: make(map[int]struct{})}}
			}
		}
	}

	// initialize our goto connections for the starting LR(0) item set
	itemSet.Conns = make(map[BNFElement]int)

	// determine whether or not a goto kernel already has a representative item
	// set in the item graph.  If it does, create a connection to the
	// preexisting set.  Otherwise, add the full item set calculated from the
	// goto kernel to the item graph.  Clean up the gotoKernels map as we go
	// since we no longer need to store its data.
mainloop:
	for elem, gotoKernel := range gotoKernels {
		// remove the goto kernels as we iterate (don't need the map entry now)
		delete(gotoKernels, elem)

		// calculate the closure of the goto kernel so we can compare and/or add it
		ptb.closureOf(gotoKernel)

		for i, otherItemSet := range ptb.ItemSets {
			// if two items have the same items, they will be equivalent
			if reflect.DeepEqual(gotoKernel.Items, otherItemSet.Items) {
				// create a connection to our preexisting item set
				itemSet.Conns[elem] = i
				continue mainloop
			}
		}

		// the connection index will be the next index (length) in the slice of
		// item sets if the item set is being newly added
		connIndex := len(ptb.ItemSets)

		itemSet.Conns[elem] = connIndex

		// calculating depth first vs. breadth first is effectively equivalent
		// here in terms of performance so we can just calculate from here
		ptb.nextLR0Items(gotoKernel)
	}
}

// closureOf calculates all of the items in LR(0) item set based on its item kernel
func (ptb *PTableBuilder) closureOf(itemSet *LRItemSet) {
	for addedMore := true; addedMore; {
		addedMore = false

		for item := range itemSet.Items {
			bnfRule := ptb.BNFRules.RulesByIndex[item.Rule]

			// nothing to calculate if the dot is at the end of rule
			if item.DotPos == len(bnfRule.Contents) {
				continue
			}

			// if we have a nonterminal, add all its rule to the item set
			if nt, ok := bnfRule.Contents[item.DotPos].(BNFNonterminal); ok {
				startingLength := len(itemSet.Items)

				for _, ruleRef := range ptb.BNFRules.RulesByProdName[string(nt)] {
					ruleItem := LRItem{Rule: ruleRef, DotPos: 0}

					itemSet.Items[ruleItem] = make(map[int]struct{})
				}

				// if the length changed, something was added
				if len(itemSet.Items) != startingLength {
					addedMore = true
				}
			}
		}
	}
}

// createLR1Items converts a given LR(0) item set to an LR(1) item set and
// converts all of its connected sets recursively.  Assumes that the input set
// has an LR(1) kernel (already has spontaneously generated lookaheads)
func (ptb *PTableBuilder) createLR1Items(itemSet *LRItemSet) {
	// begin by propagating lookaheads throughout the current item set to
	// convert it to a full LR(1) item set to that spontaneous generation is
	// possible to all connected item sets (otherwise, we can't proceed)
	ptb.propagateLookaheads(itemSet)

	lookaheadsChanged := false

	// next, go through every item in the item set and spontaneously generate
	// lookaheads for any connected kernel items generated from the src item
	for item, lookaheads := range itemSet.Items {
		bnfRule := ptb.BNFRules.RulesByIndex[item.Rule]

		// a connection will only exist if the dot is not at the end of the rule
		if item.DotPos < len(bnfRule.Contents) {
			dottedElem := bnfRule.Contents[item.DotPos]

			if dottedElem.Kind() == BNFKindEpsilon {
				continue
			}

			state := itemSet.Conns[dottedElem]

			connSet := ptb.ItemSets[state]

			// the matching kernel item will have the same rule with an
			// incremented dot position
			for connItem, connLookaheads := range connSet.Items {
				if connItem.Rule == item.Rule && item.DotPos == connItem.DotPos-1 {
					startLength := len(connLookaheads)

					// if the items match, spontaneously generate lookaheads for
					// the associated kernel item from its source item
					for l := range lookaheads {
						connLookaheads[l] = struct{}{}
					}

					if startLength != len(connLookaheads) {
						lookaheadsChanged = true
					}

					// there will only be one matching kernel item
					break
				}
			}
		}
	}

	// if no lookaheads changed, then we do not need to spontaneously generate
	// any new lookaheads for any connected sets and can just stop here
	if !lookaheadsChanged {
		return
	}

	// go through and recur to each connected set only after we have already
	// fully generated all of its spontaneous lookaheads (given it an LR(1)
	// kernel) so that propagation can occur correctly in connected set
	for _, state := range itemSet.Conns {
		ptb.createLR1Items(ptb.ItemSets[state])
	}
}

// propagateLookaheads propagates lookaheads from an LR(1) item kernel to the
// remaining LR(0) items to convert the input item set into a full LR(1) item by
// making repeated passes through the item set and propagating lookaheads to all
// items as necessary until no meaningful propagations occur
func (ptb *PTableBuilder) propagateLookaheads(itemSet *LRItemSet) {
	for propagated := true; propagated; {
		propagated = false

		for item, itemLookaheads := range itemSet.Items {
			// if we have no lookaheads, then we have nothing to propagate (yet)
			if len(itemLookaheads) == 0 {
				continue
			}

			bnfRule := ptb.BNFRules.RulesByIndex[item.Rule]

			// if our dot is at the end of the rule, there will be nothing to
			// propagate to (ie. no connected nonterminals)
			if item.DotPos == len(bnfRule.Contents) {
				continue
			}

			// only if we have a nonterminal, will we have something to propagate to
			if nt, ok := bnfRule.Contents[item.DotPos].(BNFNonterminal); ok {
				// calculate the lookaheads that we will pass on to the next item
				var lookaheads map[int]struct{}

				if item.DotPos == len(bnfRule.Contents)-1 {
					lookaheads = itemLookaheads
				} else {
					firstSet := ptb.first(bnfRule.Contents[item.DotPos+1:])

					// remove all the epsilons from the first set
					n := 0
					for _, l := range firstSet {
						// -1 == epsilon
						if l != -1 {
							firstSet[n] = l
							n++
						}
					}

					// if any epsilons were removed, the desired length of the
					// new slice will change so we can use it to test if there
					// were an epsilons in the first set
					epsilonsRemoved := n != len(firstSet)
					firstSet = firstSet[:n]

					lookaheads = make(map[int]struct{})
					for _, first := range firstSet {
						lookaheads[first] = struct{}{}
					}

					if epsilonsRemoved {
						for itemLookahead := range itemLookaheads {
							lookaheads[itemLookahead] = struct{}{}
						}
					}
				}

				// pass the lookaheads on to all associated rules
				for destItem, destLookaheads := range itemSet.Items {
					// associated rules will have the same name
					if ptb.BNFRules.RulesByIndex[destItem.Rule].ProdName == string(nt) {
						// combine the lookaheads of the destination item and
						// the current item and only consider this action a
						// propagation of the lookaheads change (the length of
						// the lookahead set changes if lookaheads are added)
						startingLength := len(destLookaheads)

						for lookahead := range lookaheads {
							destLookaheads[lookahead] = struct{}{}
						}

						if len(destLookaheads) != startingLength {
							propagated = true
						}
					}
				}
			}
		}
	}
}

// first calculates the first set of a given slice of a BNF rule (can be used
// recursively) NOTE: should not be called with an empty slice (will fail)
func (ptb *PTableBuilder) first(ruleSlice []BNFElement) []int {
	if nt, ok := ruleSlice[0].(BNFNonterminal); ok {
		var firstSet []int

		// start by accumulating all firsts of nonterminal (inc. epsilon) to
		// start with (in next block).  NOTE: string(nt) should be "free"

		// check to see if the nonterminal first set has already been calculated
		if mfs, ok := ptb.firstSets[string(nt)]; ok {
			firstSet = make([]int, len(mfs))

			// copy so that our in-place mutations of the first set don't affect
			// the memoized version (small cost but ultimately trivial)
			copy(firstSet, mfs)
		} else {
			for _, rRef := range ptb.BNFRules.RulesByProdName[string(nt)] {
				// r will never be empty
				ntFirst := ptb.first(ptb.BNFRules.RulesByIndex[rRef].Contents)

				firstSet = append(firstSet, ntFirst...)
			}

			// memoize the base first set (copied for same reasons as above)
			mfs := make([]int, len(firstSet))
			copy(mfs, firstSet)
			ptb.firstSets[string(nt)] = mfs
		}

		// if there are no elements following a given production, then any
		// epsilons will remain in the first set (as nothing follows)
		if len(ruleSlice) == 1 {
			return firstSet
		}

		// if there are elements that follow the current element in our rule,
		// then we need to check for and remove epsilons in the first set (as
		// they may not be necessary)
		n := 0
		for _, f := range firstSet {
			if f != -1 {
				firstSet[n] = f
				n++
			}
		}

		// if the length has changed, epsilon values were removed and therefore,
		// we need to consider the firsts of what follows our first element as
		// valid firsts for the rule (ie. Fi(Aw) = Fi(A) \ { epsilon} U Fi(w))
		if n != len(firstSet) {
			// now we can trim off the excess length
			firstSet = firstSet[:n]

			// can blindly call first here because we already checked for rule
			// slices that could result in a runtime panic (ie. will be empty)
			firstSet = append(firstSet, ptb.first(ruleSlice[1:])...)
		}

		return firstSet
	} else if _, ok := ruleSlice[0].(BNFEpsilon); ok {
		// convert epsilon into an integer value (-1) and apply Fi(epsilon) =
		// {epsilon }.  NOTE: epsilons only occur as solitary rules (always
		// valid)
		return []int{-1}
	}

	// apply Fi(w) = { w } where w is a terminal
	return []int{int(ruleSlice[0].(BNFTerminal))}
}

// buildTableFromSets attempts to convert the itemset graph into a parsing table
// and returns a descriptive error if the operation is unsuccessful
func (ptb *PTableBuilder) buildTableFromSets() error {
	// used to keep track of which PTableRules have already been generated so
	// that any rules that will become redundant can be mapped to already
	// existing rules (ie. rules that have the same name and size)
	ptableRules := make(map[string][]int)

	for i, itemSet := range ptb.ItemSets {
		row := &PTableRow{Actions: make(map[int]*Action), Gotos: make(map[string]int)}
		ptb.Table.Rows[i] = row

		// calculate the necessary actions based on our connections
		for conn, state := range itemSet.Conns {
			switch v := conn.(type) {
			case BNFTerminal:
				row.Actions[int(v)] = &Action{Kind: AKShift, Operand: state}
			case BNFNonterminal:
				row.Gotos[string(v)] = state
			}
		}

		// place any reduce actions in the table as necessary
		for item, lookaheads := range itemSet.Items {
			bnfRule := ptb.BNFRules.RulesByIndex[item.Rule]

			// dot is at the end of a rule/epsilon rule => reduction time
			if item.DotPos == len(bnfRule.Contents) || reflect.DeepEqual(bnfRule.Contents[0], BNFEpsilon{}) {
				// determine the correct reduction rule
				reduceRule := -1

				// if we are reducing by the _END_ rule, we are accepting
				// meaning, we don't actually need to store or calculate a rule,
				// but if we are reducing by anything else then we do
				if bnfRule.ProdName != "_END_" {
					// convert the BNF rule into a ptable rule
					convertedRule := &PTableRule{Name: bnfRule.ProdName}
					if bnfRule.Contents[0].Kind() == BNFKindEpsilon {
						convertedRule.Count = 0
					} else {
						convertedRule.Count = len(bnfRule.Contents)
					}

					if addedRules, ok := ptableRules[convertedRule.Name]; ok {
						for _, ruleRef := range addedRules {
							if ptb.Table.Rules[ruleRef].Count == convertedRule.Count {
								reduceRule = ruleRef
								break
							}
						}

						// if rule is still -1, it has not been added => not redundant
						if reduceRule == -1 {
							reduceRule = len(ptb.Table.Rules)
							ptb.Table.Rules = append(ptb.Table.Rules, convertedRule)

							// add our rule to the ptableRules entry for the production name
							ptableRules[convertedRule.Name] = append(ptableRules[convertedRule.Name], reduceRule)
						}
					} else {
						// if it hasn't even been added to ptableRules, we need to
						// create it and create an entry for it in ptableRules.
						reduceRule = len(ptb.Table.Rules)
						ptb.Table.Rules = append(ptb.Table.Rules, convertedRule)

						ptableRules[convertedRule.Name] = []int{reduceRule}
					}
				}

				// attempt to add all of the corresponding reduce actions
				for lookahead := range lookaheads {
					if action, ok := row.Actions[lookahead]; ok {
						// action already exists.  If it is a shift action,
						// we warn but do not error.  If it is another reduce
						// action, we print out the error and return that
						// the table construction was unsuccessful.
						if action.Kind == AKShift {
							fmt.Printf("Shift/Reduce Conflict Resolved Between `%s` and `%d`. \nRule: ", bnfRule.ProdName, lookahead)
							ptb.printLR0Item(item)
							fmt.Println("\nOriginal State Shift Is:")
							ptb.printItems(ptb.ItemSets[action.Operand].Items)

							fmt.Print("\n\n")

							// Shift-Reduce Conflict -- all accounted for
						} else if action.Kind == AKReduce {
							oldRule, newRule := ptb.Table.Rules[action.Operand], ptb.Table.Rules[reduceRule]

							if *oldRule == *newRule {
								// if the rules create the same tree and consume
								// the same amount of tokens, we don't care if
								// they contain different elements => they are
								// effectively equal (so no real conflict
								// exists: they pop the same number of states
								// and refer to the same column in the GOTOs)
								continue
							} else if oldRule.Count == 0 && newRule.Count == 0 {
								// no real conflict exists between two epsilon
								// rules of different names since the parser
								// simply discards all empty trees.  However,
								// we need to resolve this conflict in favor of
								// the rule whose GOTO state has the most
								// possible actions so that we don't lose
								// actions (randomly).  Since this conflict only
								// occurs when two epsilon rules occur in
								// sequence with one another, we know that the
								// state with more actions will contain the
								// state with fewer actions (shared lookaheads)

								oldGotoActionCount := ptb.getActionCount(row.Gotos[oldRule.Name])
								newGotoActionCount := ptb.getActionCount(row.Gotos[newRule.Name])

								// only if the new goto state's action count
								// is higher, do we need to update anything
								if newGotoActionCount > oldGotoActionCount {
									action.Operand = reduceRule
								}

								continue
							}

							return fmt.Errorf("reduce/reduce conflict between `%s` and `%s`", oldRule.Name, newRule.Name)
						}

						// ACCEPT collisions should never happen
					} else if reduceRule == -1 {
						// if our reduce rule is -1, we are reducing the goal
						// symbol which means we should accept (no test for EOF)
						row.Actions[lookahead] = &Action{Kind: AKAccept}
					} else {
						row.Actions[lookahead] = &Action{Kind: AKReduce, Operand: reduceRule}
					}
				}
			}
		}
	}

	return nil
}

// addRule converts a BNF rule into a usable rule (assumes rule is not redundant)
// and adds it to the rule set.  Returns the a reference to the new rule.
// func (ptb *PTableBuilder) addRule(bnfRule *BNFRule) int {
// 	// We know the next rule will be at the end of rules so we can get the rule
// 	// reference preemtively from the length
// 	reduceRule := len(ptb.Table.Rules)

// 	var count int
// 	if _, ok := bnfRule.Contents[0].(BNFEpsilon); ok {
// 		// epsilon rules consume no tokens to make their tree
// 		count = 0
// 	} else {
// 		count = len(bnfRule.Contents)
// 	}

// 	ptb.Table.Rules = append(ptb.Table.Rules,
// 		&PTableRule{Name: bnfRule.ProdName, Count: count})

// 	return reduceRule
// }

// getActionCount calculates the number of actions that are produced by a given
// item set (used in epsilon reduce/reduce conflict resolution)
func (ptb *PTableBuilder) getActionCount(state int) int {
	// if have already calculated the state, we can just count its actions
	if state < len(ptb.Table.Rows) {
		row := ptb.Table.Rows[state]

		if row != nil {
			return len(row.Actions)
		}
	}

	// otherwise, we have to calculate it from the base item set (no need to
	// mess up/complicate row computation by computing states out of order)
	itemSet := ptb.ItemSets[state]
	shiftActionCount := 0

	// FORMULA: (based on knowledge of table construction)
	// # of terminal connections + # of unique lookaheads = # of actions

	for conn := range itemSet.Conns {
		if conn.Kind() == BNFKindTerminal {
			shiftActionCount++
		}
	}

	// can't really assume size very easily
	totalLookaheads := make(map[int]struct{})

	for _, lookaheads := range itemSet.Items {
		for lookahead := range lookaheads {
			totalLookaheads[lookahead] = struct{}{}
		}
	}

	return shiftActionCount + len(totalLookaheads)
}
