# The Prelude
The prelude is the set of types and functions loaded into every module by default.

## Functions
```
assert(condition)
assert_m(condition, message)
abs(n)
divmod(n)
compare(c1, c2)
clamp(n, l, u)
ceil(n)
exit(code)
floor(n)
lerp(l, u, p)
min(a, b)
max(a, b)
norm(n)
panic(message)
round(n)
to_string(t)
strand(fn, args...)
sleep(ms)
todo()
unreachable()
```
## Aliases
```
rune = i32
byte = u8
isize = i64
usize = u64
```
## Types
```
string
Option<T>
Result<T, E>
List<T>
Map<K, V>
Ord
```
## Classes
```
Iter<T>
Seq<T>
Monad<T>
Show
Hash
Eq
Compare
```
## Unions
```
Text = rune | string
SInt = i8 | i16 | i32 | i64
UInt = u8 | u16 | u32 | u64
Int = SInt | UInt
Float = f32 | f64
Real = Int | Float
```
## (Compile-Time) Constants
```
MAX_I8
MAX_U8
MAX_I16
MAX_U16
MAX_I32
MAX_U32
MAX_I64
MAX_U64
MIN_F32
MAX_F32
MIN_F64
MAX_F64

HOST_OS
HOST_ARCH
HOST_TRIPLE
```
