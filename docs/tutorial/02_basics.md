# The Basics

**Table of Contents**
- [Comments](#comments)
- [Numbers, Types, and Arithmetic](#numbers)
- [Variables and Constants](#vars)
- [Case Study: Adder](#adder)

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

### Variable Declarations

In Chai, you can declare variables using the `let` keyword like so:

    let x = 10;
    println(x);  // prints 10

The type of `x` is inferred based on the value it is initialized with.

You can specify the type of a variable using a **type extension** like so:

    let pi: f64 = 3.14159265;

If you specify a type extension, you don't have to explicitly initialize the
variable: it will be initialized to its **null value** by default (which is zero
for all numeric types).

    let y: i32;  // y = 0

> This kind of variable declaration only works for types which are *nullable*.
> We will see what this means when we encounter our first non-nullable type.

You can also declare multiple variables at once if you separate their
declarations using commas:

    let a = x * 4, b = y + 6 - x;

### Assignment

All variables declared with `let` are *mutable* meaning we can assign to them.
Assignment in Chai is done using the `=` operator like so:

    a = b;

You can assign to multiple variables at once by separating the variables and
expressions using commas. 

    x, y = 5, 6;  // x = 5, y = 6;

Chai fully evaluates the right-hand side expressions before it assigns them to
the left-hand side.  This means that you can trivially swap the values of two
variables using multi-assignment without having to use a temporary variable.

    a, b = b, a;  // swaps a and b's values

Often, we want to apply an operator to between the variable's value and a new
value.  For example, if wanted to add 2 to the variable `x`, we would write:

    x = x + 2;

Because this kind of operation is so common, Chai provides a short-hand for
statements like the one above called **compound assignment**:

    x += 2;  // equivalent to `x = x + 2`

Compound assignment can be performed with any of the binary arithmetic and
bitwise operators:

    x *= 2;   // double x
    y /= 4;   // divide y by 4
    a <<= 1;  // left-shift a by 1

Furthermore, we can apply compound assignment between multiple values:

    x, y **= 2, 3;  // square x, cube y

### Constants

Sometimes, we don't want to allow our variables to mutable.  Thus, Chai provides
a special type of variable called a **constant** which is *immutable*.
Constants are declared in the same way as variables except the `const` keyword
is used instead of `let`.

    const c: i64 = 2;

    const d = x ** 4;

> The expressions used to initialize constants do NOT need to be compile-time
> constant (ie. constexpr for you C++ devs).  Chai provides a different
> mechanism for those kinds of constants called *uniforms* that we will see
> later.

Because constants are immutable, we can't assign to them:

    c = 3;  // COMPILE ERROR

## <a name="adder"> Case Study: Adder

To cement all the concepts have discussed so far together, we are going to take
a look at a simple example program which makes use of variables and arithmetic.

This program is a simple *adder* which has the user input two numbers and prints
out their sum.  Notably, this program will also demonstrate how to do basic user
input in Chai.

Let's take a look at the whole program and then we will break down how it works.

    package adder;

    import println, scanf from io.std;

    func main() {
        let a, b: i64;  

        println("Enter the two numbers:");

        scanf("%d %d", &a, &b);

        println(a, "+", b, "=", a + b);
    }

A simple run of the program might appear as follows:

    > ./adder
    Enter the two numbers:
    4 5
    4 + 5 = 9

The first two lines of the program should already be familiar to you.  The
primary wrinkle is that we are importing a second function from `io.std` called
`scanf` which we will later use to allow the user to input numbers.

Then, we have a fairly standard main function which contains the actual
machinery of our program.

> As an aside, all code that actually "runs" must be contained within a function
> or an initializer.  Chai does not allow free-floating code in the global
> scope.  The code snippets presented out of contents in this tutorial are just
> that: snippets from larger programs.  

The first line of our main function:

    let a, b: i64;

is a fairly standard variable declaration with which you should already be
familiar. However, it does demonstrate a little bit a syntax that I didn't
explicitly show off earlier: we can declare multiple variables of the same type
using one type extension.  In this case, `a` and `b` are both of type `i64`.

The second line is just a call to `println` which we already studied in the
first chapter.

The third line is probably the most "enigmatic" line of the bunch:

    scanf("%s %s", &a, &b);

`scanf` is a special function which is part of Chai's *formatted I/O* library.
In essence, it allows us to read input from a stream, like standard in, in
accordance with something called **format string** which describes what that
input should look like.

The format string in this case is the passed as the first argument to `scanf`.
The `%d` denote places where user input is expected: ie. where the we expect the
user to input the numbers: the `d` means we expect an integer to be entered.
Everything outside that is not the `%d` is the text that `scanf` expects the
user to enter around their input: the "punctuation" if you will.  In this case,
we have a single space which means that we expect the user to enter a space
between their inputted values. Putting this all together, we can say that the
format string `"%d %d"` states that we expect the user to input two integers
with a space between them.  

The next two arguments indicate the locations where we want it to store the
values it reads in.  In this case, we want `scanf` to store its results inside
the variables `a` and `b`.  The `&` in front of `a` and `b` denotes that we
are giving `scanf` the locations of `a` and `b` rather than their values.

More precisely, `&` is called the *indirection operator* it is used to create a
*pointer* to a value.  Pointers are a very important topic in Chai that we will
cover in much greater detail later.  For now, you can think of them as
representing the "locations" of values rather than the values themselves.

Putting all these pieces together, we can then say that the line:

    scanf("%d %d", &a, &b);

reads two integers from standard in (ie. the command-line) separated by a space
into the variables `a` and `b`.

The final line of our main function:

    println(a, "+", b, "=", a + b);

is just another call to `println` with the wrinkle that we are giving it more
than one argument.  When `println` is given multiple arguments, it simply
prints out each of its arguments in order separated by spaces.  

That's it: the full *adder* program broken down line by line.  In the next
chapter, we are going to look at how to control the flow of our program using
conditional logic and loops.
