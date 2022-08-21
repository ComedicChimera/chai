package types

import (
	"fmt"
	"strings"
)

// buildTraceback builds the type traceback used in a type error message.
func (s *Solver) buildTraceback(sb *strings.Builder, tnodes map[uint64]*typeVarNode) {
	// If there are no type variables involved in the type comparison, then
	// there is no need to print any traceback.
	if len(tnodes) == 0 {
		return
	}

	// Print the traceback header.
	sb.WriteString("\ninvolving the following undetermined type variables:\n\n")

	// For each variable, ...
	for _, tnode := range tnodes {
		// Print the leading indentation.
		sb.WriteString("  ")

		// Begin the variable's section with its name.
		sb.WriteString(tnode.Var.Name)
		sb.WriteRune('\n')

		// Create the initial set of overloads that we are going to prune down
		// as we build the traceback.
		var initialOverloads []Type
		for _, node := range tnode.Nodes {
			initialOverloads = append(initialOverloads, node.Sub.Type())
		}

		// Add the prune set contents in reverse order to the initial overloads
		// to get the full initial set of values for the type.  This is done in
		// reverse order so that we can quickly slice them off the end when we
		// are slowly paring down the list in the traceback.
		for i := len(tnode.PruneSets) - 1; i >= 0; i-- {
			for _, pnode := range tnode.PruneSets[i].Pruned {
				initialOverloads = append(initialOverloads, pnode.Sub.Type())
			}
		}

		// Print the first (declaration) line of the traceback.
		sb.WriteString(fmt.Sprintf("    -> (line: %d, col: %d): is ", tnode.Var.Span.StartLine, tnode.Var.Span.StartCol))
		if tnode.Known {
			sb.WriteString("known")
		} else {
			sb.WriteString("inferred")
		}
		sb.WriteString(" to be one of:\n")

		// Print the initial overload list.
		buildOverloadList(sb, initialOverloads)

		// For each prune set, print the inference the type solver made.
		for _, ps := range tnode.PruneSets {
			// Print the header of the traceback entry.
			sb.WriteString(fmt.Sprintf("    -> (line: %d, col: %d) is then inferred to be one of:\n", ps.Span.StartLine, ps.Span.StartCol))

			// Remove the overloads that were "pared" out.
			initialOverloads = initialOverloads[:len(initialOverloads)-len(ps.Pruned)]

			// Print the new overload list.
			buildOverloadList(sb, initialOverloads)
		}
	}

	// Write the final line of the traceback.
	sb.WriteString("\nresulting in the error:\n")
}

// buildOverloadList builds an overload list used in a type traceback.
func buildOverloadList(sb *strings.Builder, overloads []Type) {
	// Print the leading indentation (7 spaces) of the first line.
	sb.WriteString("       ")

	// Keep track of the length of the line so we know when to wrap.
	lineLen := 0

	// Print each overload.
	for i, overload := range overloads {
		// Get the overload's representative string.
		overloadRepr := overload.Repr()

		// If the line with the overload repr included would be too long,
		if lineLen+len(overloadRepr) > 64 {
			// Print a newline and indent for the next line of the list.
			sb.WriteString("\n       ")

			// Reset the line length.
			lineLen = 0
		}

		// Print the overload repr and update the line length.
		sb.WriteString(overloadRepr)
		lineLen += len(overloadRepr)

		// If we need a separator, ...
		if i != len(overloads)-1 {
			// Add the separator comma and update the line length.
			sb.WriteString(", ")
			lineLen += 2
		}
	}

	// Add the final newline of the overload list.
	sb.WriteRune('\n')
}
