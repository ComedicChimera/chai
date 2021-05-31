package syntax

import (
	"fmt"
)

func (ptb *PTableBuilder) printSet(set *LRItemSet) {
	fmt.Print("{ ")

	ptb.printItems(set.Items)

	fmt.Println("} connects to")

	for source, state := range set.Conns {
		switch v := source.(type) {
		case BNFTerminal:
			fmt.Printf("    %d -> ", v)
		case BNFNonterminal:
			fmt.Printf("    %s -> ", v)
		}

		fmt.Println(state)
	}

	fmt.Println()
}

func (ptb *PTableBuilder) printItems(items map[LRItem]map[int]struct{}) {
	for item, lookaheads := range items {
		ptb.printLR0Item(item)

		fmt.Print(", ")

		for lookahead := range lookaheads {
			fmt.Printf("%d/", lookahead)
		}

		fmt.Print("; ")
	}
}

func (ptb *PTableBuilder) printLR0Item(item LRItem) {
	rule := ptb.BNFRules.RulesByIndex[item.Rule]

	fmt.Printf("%s -> ", rule.ProdName)

	for i, ruleItem := range rule.Contents {
		if i == item.DotPos {
			fmt.Print(".")
		}

		switch v := ruleItem.(type) {
		case BNFTerminal:
			fmt.Printf("%d ", v)
		case BNFNonterminal:
			fmt.Printf("%s ", v)
		case BNFEpsilon:
			fmt.Printf("'' ")
		}
	}

	if item.DotPos == len(rule.Contents) {
		fmt.Print(". ")
	}
}
