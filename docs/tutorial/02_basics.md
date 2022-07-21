# The Basics

This chapter of the tutorial will walk you through all the basic topics you need
to get started writing some "useful" programs.  Although it is longer, it mainly
focuses on topics most programmers are already familiar with so it shouldn't be
that much of a bear.

**Table of Contents**
- [Comments](#comments)
- [Numbers, Types, and Arithmetic](#numbers)
- [Variables and Constants](#vars)
- [Case Study: Adder](#adder)
- [Conditional Control Flow](#ifs)
- [While and C-Style For Loops](#while)
- [Case Study: Guessing Game](#guessing)
- [Exercises](#exercises)

## <a name="comments"> Comments

A comment is a piece of source text ignored by the compiler: they are commonly
used for documentation.  Chai has two basic styles comment: 

1. Line Comments
2. Block Comments

**Line comments** begin with `//` and go until a newline is encountered:

    // I am a line comment.

**Block comments** begin with `/*` and go until `*/` is encountered.  They can
span multiple lines, a single line, or a small subsection of line.

    /* I am a block comment. */ 
    /* I
    am
    also
    a
    block
    comment */  
    code /* Sneaky comment */ more code

Comments will be included throughout this tutorial and are the primary mechanism
for documenting Chai code.

## <a name="numbers"> Numbers, Types, and Arithmetic

Chai provides 10 basic numeric types which are as follows:

| Label | Meaning |
| ----- | ------- |
| `i8` | signed, 8-bit integer |
| `i16` | signed, 16-bit integer |
| `i32` | signed, 32-bit integer |
| `i64` | signed, 64-bit integer |
| `u8` | unsigned, 8-bit integer |
| `u16` | unsigned, 16-bit integer |
| `u32` | unsigned, 32-bit integer |
| `u64` | unsigned, 64-bit integer |
| `f32` | IEEE-754, 32-bit floating-point number |
| `f64` | IEEE-754, 64-bit floating-point number |

### Numeric Literals

There are several kinds of numeric literals supported by Chai.

The first are **number literals** which can be any one of the numeric types:
they represent numbers generally without being confined to a specific type.

They appear as a whole numbers and can be written in decimal, binary,
hexadecimal, or octal.  They can also contain `_` at any point in the literal
for digit separation (excluding the prefix).

    // Decimal
    1
    42
    1_000_000
    123_45

    // Binary
    0b1
    0b1010
    0b1011_0110
    0b1_011_01_01101

    // Hexadecimal
    0xa
    0xFF
    0x4b3Ad
    0x42_0a_b1_2f

    // Octal
    0o5
    0o123
    0o763_234
    0o24_3543_342

The second category of numeric literals are **integer literals**.  These are
literals that can only have integer types.  The only way to denote such literals
is to include a literal suffix: `u` (value is unsigned) and/or `l` (value is
long: 64 bits). Here are some examples:

    23l        // Long
    123ul      // Unsigned Long
    0b101u     // Unsigned
    0xff_aelu  // Unsigned Long

It should be noted that if an integer literal is too large to represented by any
supported integer type, a compile error will occur.

The final kind of numeric literals are **floating literals**.  These are
literals that can only have floating-point types.  These literals must include
either a decimal and/or an exponent.  They can be denoted in either decimal
(using `e`/`E` for the exponent) or hexadecimal (using `p`/`P` for the
exponent).

    // Decimal
    3.141_592_65
    6.626e-23
    1E9
    1.0
    1.4_14

    // Hexadecimal
    0x1.01
    0x1fap12
    0x6b.13P84D

### Arithmetic Operators

Chai supports all the common arithmetic operator: `+`, `-`, `*`, `/`, and `%` as
well as `**`, the power operator.  It also supports standard operator precedence
for these operators and the use of parentheses.  Here are some example
arithmetic expressions.

    2 + 3            // => 5
    5.6 * 2.1        // => 11.76
    -2 * (5 + 4.2)   // => -18.4
    65 % 4           // => 1
    (1.4 - 1.5) % 1  // => 0.1
    2 ** 2           // => 4
    (1 + 4.0) ** -1  // => 0.2

    2 ** 2 ** 3      // => 256 (** is right associative)

    5 / 2    // => 2   (integer division)
    5 / 2.0  // => 2.5 (floating division)

All binary Chai arithmetic operators accept values *of the same type*.  Since we
are only using literals above, this is not a problem since Chai simply infers a
compatible type for all the literals involved.  However, if we have two
incompatible literals, we will get a type error.

    // type error: no overload of `+` accepts (untyped unsigned int literal, untyped floating literal)
    0b10u + 0xff.2a 

> We will break down more granularly what that type error actually means when we
> talk about operator overloading in a later chapter.

Chai also supports all the standard bitwise operators for integers: `~` (bitwise
complement), `&` (bitwise AND), `|` (bitwise OR), `^` (bitwise XOR), and `<<`
(left shift), and `>>` (right shift).

    ~(~1 << 2)      // => 0b111
    0b10 & 0b11     // => 0b10
    0b111 ^ ~0b101  // => 0b101

    -8 >> 1   // => -4 (arithmetic right shift for signed types) 
    
### Type Casting

Chai is *strongly* and *statically* typed which means that all types must be known
at compile-time and values cannot change from one type to another.  Furthermore,
Chai generally does not perform _any_ implicit conversions.  This means that if you
want to convert from one type to another, you need to do so explicitly using
a **type cast**.

As you might surmise, a **type cast** converts one type into another.  Casting is
done with `as` operator like so:

    2.3 as i64

By default any numeric type can be cast to any other numeric type.  However, casts
such as the one above may result in data loss, rounding, or truncation depending
on the types and values involved in the type cast.

Note that type casting has lower precedence than any of the arithmetic operators:

    2 / 4 as f32    // result is cast to f32 => 0
    2 / (4 as f32)  // 4 is cast to f32 => 0.5

## <a name="vars"> Variables and Constants



## <a name="adder"> Case Study: Adder

## <a name="ifs"> Conditional Control Flow

## <a name="while"> While and C-Style For Loops

## <a name="guessing"> Case Study: Guessing Game

## <a name="exercises"> Exercises

1. Implement a simple [Fizz Buzz](https://en.wikipedia.org/wiki/Fizz_buzz)
   program: print the first 100 fizz-buzz numbers.

2. Write a program to print the nth Fibonacci number with `n` is given by the
   user.

3. Write a program to calculate the square root of `n` where `n` is given by the
   user using [Newton's Method](https://en.wikipedia.org/wiki/Newton%27s_method).