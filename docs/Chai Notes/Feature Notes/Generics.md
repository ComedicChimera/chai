# Generics
*Note*: right now this is just unorganized garbage I ported over from Evernote -- it will be cleaned up in the future.

## Variadic Generics
- Place a `...` before the type name (can still have contraints)
	- Eg. `MyType<...T: Seq>`
- Variadic generics can be expanded into other types using `...`
	- Eg. `(...T)` expands into a tuple of the n-values of `T`.
- Variadic arguments:
	- `println<...T: Show>(...args: T)` (automatically unpacks `T` for each argument)
	- Passed as a tuple (to allow a mixture of types)
		- Use generic control flow handle them
- Variadic generics cannot be used without being expanded.

## Generic Control Flow
Generic control flow is a specialized type of control flow that acts as a form of meta-programming based on the type system itself.  It allows for customized code generation based on generic evaluation.

This can only be used inside of generic definitions.

### Generic If
- Syntax: `if type _ is _`
	- Eg. `if type T is i32`
- Allows for pattern matching:
	- Eg. `if type T is (K, V)`

### Generic Match
- Syntax: `match type _`
- Cases just match on types.
- Allow for case pattern matching.
	- No need to extract variables that correspond to type being matched over since with generics we now the variable is simply that type: we just want our code to be able to react to it conditionally :) -- ie. no `v := x.(type)` is necessary
	
### Generic Loops
- Syntax: `for<T> item from struct_or_tuple`
- Allows for iteration over tuple and struct fields
	- Eg. for the `println` function:
```
func println<...T: Show>(...args: T) {
	for<T> arg from args {
		...
	}
}
```
- These *can* be used outside of a generic context.
- They can produce sequences or tuples :)

### Sizeof Operator
- Syntax: `sizeof(T)`
- Returns the size in bytes of a type (like in C)
- Expanding variadic generics inside of `sizeof` gives the number of variable types
	- Eg. `sizeof(...T)` where `T` consists of 3 types is `3`