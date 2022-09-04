# Defined Constraints
**Defined constraints** provide a way to define sets of constraints for generics.

## Syntax
```
constraint_def := 'constraint' 'IDENT' '=' constraint_set ';' ;
constraint_set := type_label {('|' | '&') type_label};
```

Where each of the type labels can be other constraints or classes.

## Semantics
Defined constraints may be used as constraints for generics like so:

```
func abs<T: Real>(x: T) T
```

where `Real` is an example of a defined constraints.

Constraints are defined in the global scope as definitions.  They may be public or private.

When unions include other constraints in their definitions, the "sub-constraints" are merged into the defining constraint for purposes of type semantics.  For example:

```
constraint A = i32 | u32;
constraint B = A | f32;           # = i32 | u32 | f32
constraint C = i64 | f64 | u64;
constraint D = C | D;             # = i32 | u32 | f32 | i64 | f64 | u64
```

When the `|` operator is used, it means that any of the sub-constraints must be satisfied in order for the overall constraint to be satisfied.  However, when the `&` operator is used, it means all the constraints must be satisfied.  By default, `&` has higher precedence than `|`.  This is particularly useful when working with classes:

```
constraint OrderedKey = Hash & Less;
```