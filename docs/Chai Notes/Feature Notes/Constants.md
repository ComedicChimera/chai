# Constants
Constants provide a convenient way to define compile-time constant values.

## Syntax
```
const const_name: type_label = constant_value


const const0: type0 = value0, const1: type1 = value1
```

## Semantics
Constants are considered values and handled similarly to variables.

They may never be mutated.

Their values must be compile-time constants (evaluable at compile-time: *more rigorous definition of what this means is required*).

They may be defined locally or globally.

They must always have a type extension and an initializer.

When referenced, they act like literal constants (eg. `&1`)