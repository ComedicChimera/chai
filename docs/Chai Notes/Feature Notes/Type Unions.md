# Type Unions
Type unions provide a way to define sets of constraints for generics.

## Syntax
```
union UnionName = type1 | type2 | type3 ...;
```
Where `type1`, `type2`, and so on are either types or unions.

## Semantics
Unions may be used as constraints for generics like so:

```
func abs<T: Real>(x: T) T
```

where `Real` is an example of a type union.

Unions are defined in the global scope as definitions.  They may be public or private.

When unions include other unions in their definitions, the "sub-unions" are merged into the defining union for purposes of type semantics.  For example:

```
union A = i32 | u32;
union B = A | f32;           # = i32 | u32 | f32
union C = i64 | f64 | u64;
union D = C | D;             # = i32 | u32 | f32 | i64 | f64 | u64
```
