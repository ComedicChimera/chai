# Shift-Reduce Conflicts

This file documents the two known shift-reduce conflicts in the Chai grammar
that are auto-resolved in favor of shift.

## Nobreak Clauses

Because the bodies of loops can themselves contain loops as expressions with no
newline delimiter, the "association" of the nobreak block is technically in
flux.

For example,

    while cond -> while cond2 -> expr nobreak -> expr

What loop does that `nobreak` clause belong to?  From a purely syntactic
standpoint, it is completely ambiguous.

However, this ambiguity can be trivially resolved by simply declaring that the
`nobreak` clause associates with the innermost loop (as is logical).  So
following this rule, the `nobreak` clause associates with the inner while loop.

## If/Elif/Else Chains

An almost identical issue occurs with `if` statements, specifically the `elif`
and `else` clauses that follow them.  

For example,

    if cond1 -> if cond2 -> expr else -> expr

Which `if` does the `else` associate with?  The innermost or outermost?  The
answer is once again to define that the `elif` and `else` clauses always
associate inwards.  Now we know that this `else` belongs to the innermost `if`.