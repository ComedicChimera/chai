## Idea
Type unions provide a way to define sets of constraints for generics.

## Syntax
```
union UnionName = type1 | type2 | type3 ...
```
Where `type1`, `type2`, and so on are either types or unions.

## Semantics
Unions may be used as constraints for generics like so:

```
func abs[T: Real](x: T) T
```

where `Real` is an example of a type union.

Unions are defined in the global scope as definitions.  They may be public or private.

When unions include other unions in their definitions, the "sub-unions" are merged into the defining union for purposes of type semantics.  For example:

```
union A = i32 | u32
union B = A | f32           # = i32 | u32 | f32
union C = i64 | f64 | u64
union D = C | D             # = i32 | u32 | f32 | i64 | f64 | u64
```

## Built-in Unions
Several unions are built into the language and imported as part of the prelude.  These unions are as follows:

```
union Int = i8 | i16 | i32 | i64 | u8 | u16 | u32 | u64
union Float = f32 | f64

union Rational = r64 | r128
union Complex = c64 | c128

union Real = Int | Float | Rational

union Num = Real | Complex

union Text = rune | string
```

