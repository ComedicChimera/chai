# Collections

Chai provides 4 fundamental collections as enumerated below.

**NOTE**: As of 1/30/22, vectors and matrices are no longer a part of the core language.  They may be readded with their own syntax as part of the math library at a later date.  They are not essential to the language's release and serve only to bloat and complicate the language (given the likely lack of adequate vector support upon release).  TLDR: we are not trying to compete with Matlab, and vectors and matrices have very little use to regular programmers on a day to day basis.

## List
Lists are resizable, homogenous collections of elements.  They use the `List[T]` type.
Their literal is written as: `[a, b, c, ...]` (using brackets). 

Lists have standard value semantics.

They are sequences and indexable by numbers.

## Dict

Dictionaries are resizable collections of homogenous key-value pairs.  They use the `Dict[K, V]` type. 

Their literals are written as follows:
```
{"orange": 2}  # the string "orange" is mapped to `2`

{"apple": 6.7, "plum": -2.31}  # "apple" -> 6.7, "plum" -> -2.31
```

The `:` separates keys and values `,` separates pairs.

Thry are indexable by their key value.  Keys must be unique and hashable.

Dictionaries have standard value semantics and are sequences.

They are also ordered.

## Buff
Buffers are fixed-size, homogenous collections of elements.  They use the `Buff[T]` type.  Their literal is written as: `{a, b, c, ...}` (using curly braces).

They are sequences and indexable by numbers.

Unlike lists, they do *not* have value semantics.  They essentially work like arrays in C (with a bit more safety ie. memory safety, bounds checks, etc).  They are generally meant to be used in cases where you want to "fill" a buffer up with data such as chunks from a stream.  

Although they are fixed size, their size is *not* encoded in their type (again, more like C array pointers).

## Tuple
Tuples are fixed-size, fixed-location, heterogenous "collections" of elements.  They use a type label of `(T1, T2, ...)` where `Ti` denotes the type of a specific element.  As mentioned above, elements have a specific position within the tuple.  

Their literals are written as: `(a, b, ...)`.  Note that tuples must contain at least two elements.  A literal written as `()` corresponds to the `nothing` type, the `(a)` syntax is simply a sub-expression.

They are not sequences and are not indexable by numbers.  Instead they use a special `.` syntax to access their elements (eg. `tuple.0` corresponds to the first element).  Tuples are more like structs then collections.

They are immutable, meaning you can't modify their elements.