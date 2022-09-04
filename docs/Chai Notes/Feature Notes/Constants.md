# Constants
Constants provide a convenient way to define constant values.

## Syntax
```
const const_name: type_label = constant_value


const const0: type0 = value0, const1: type1 = value1
```

## Semantics
Constants are considered values and handled similarly to variables.

They may never be mutated.

They may be defined locally or globally.

They must always have a type extension and an initializer.

When reference, they always take the form of `const` pointers.

## Compile-Time Constants
**Compile-Time constants** are declared like normal constants except they use the keyword `comptime`.  Compile-time constants must have compile-time constant values and are treated as r-values wherever they are encountered.  

```
comptime PI = 3.14159265 
```
