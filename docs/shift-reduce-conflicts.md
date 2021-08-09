# Shift-Reduce Conflicts

This file documents the known shift-reduce conflicts in the Chai grammar that
are auto-resolved in favor of shift.

## Type Patterns

It is technically possible for a type pattern to be ambiguous if it used in a
place where the type pattern would completely illogical.  For example, if you
use a type pattern inside a slice, there is ambiguity surrounding the colon.

    list[t is x: T]

Is `T` the end bound of the slice or the type?  Is `x` a type or an identifier?
In this context, it is impossible to tell.  Chai currently resolves this as `x`
is an identifier and `T` is the type.  However, since this is practically a
non-existent use case for `is` and for slicing, there is really no reason to
apply any major adjustments to the syntax to "handle" this conflict.
