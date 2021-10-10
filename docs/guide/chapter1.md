# Basics

This chapter introduces the basics of the Chai programming language:
how to write and compile basic programs.

This guide assumes the reader already has basic knowledge of programming and
computer science (eg. variables, functions, loops, dependencies, etc.)

## Hello World

A simple *Hello world* program in Chai looks like so:

    import println from io.std

    def main() = println("Hello, world!")

There are a couple key parts to break down here:

1. The file begins with an **import statement** that retrieves a function
   `println` from a package at the path `io.std`.  `std` is a sub-package of the
   `io` module -- we will study what that really means in a later chapter.
2. The program begins with a **main function**, named `main` and defined using
   the `def` keyword.  This function must be included in every executable
   program and invoked at the beginning of program execution.
3. The body of the main function is just an expression; we use the `=` operator
   to bind an expression the result of a function.  In this case, it is an
   expression that returns `nothing`, which is a special type in Chai denoting
   "no value".
4. The `println` function call itself is fairly simple: we can `println` by its
   name using parentheses and pass it a single string argument.  Strings are
   collections of UTF-8 encoded text in Chai and are denoted with double quotes.

That is all the code necessary to say "Hello, world!".  However, we have a little
bit more work to do.  Firstly, let's assume we have a directory like so:

    hello/
        hello.chai

`hello.chai` is a file containing the source code outlined above -- the file
extension for Chai code is `.chai`. 

We need to create a **module** before we can compile our code.  This is an
organizational design choice Chai makes: modules are absolutely essential for
import resolution and dependency management so they MUST be included in every
project no matter how trivial.  We will learn about what a module actually is in
a later chapter but for now, just understand them as the method of
"architecting" Chai projects.

The good news is that modules are super easy to set up.  All we need to do is
run the following command from the command line while in our `hello` directory:

    chai mod init hello

This creates a new module called `hello` in the current directory.  You should
see a `chai-mod.toml` appear in the current directory:

    hello/
        chai-mod.toml  <--
        hello.chai

Chai uses [TOML](https://toml.io) to mark up its module files.  If you were to
open this file, you would actually see quite a lot of information which, again,
we will discuss in more detail later.  Just leave it be for now as the compiler
selects sensible defaults for us when configuring the module.

Now, to the moment we've all been waiting for, actually building the program.
Doing so is actually quite simple, just use the `build` command.

    chai build .

Notice that we pass in a directory instead of a source file.  This is once again
because Chai organizes everything by directories and modules rather than
individual source files.  You will find that this feature is actually super
useful as your projects grow larger than a single file.

Now, if we check our directory structure, it should now look like this:

    hello/
        out/
            hello.exe  <-- our binary
            hello.pdb  <-- a debugging file for GDB
        chai-mod.toml
        hello.chai

> If you are on a Unix platform, your executable will be named accordingly.

Chai creates the executable in the `out` directory by default.  You can change
this in the module file if you want to a later time.

For now, we will leave it as is and invoke our executable directly:

    > ./out/hello.exe
    Hello, world!

Yay!  Our simple program works as expected.  

Finally, it is worth mentioning that we can streamline our build process a bit
if we just want to run the program.  We can just use the `run` command.

    > chai run .
    Hello, world!

This will compile our binary, run it, and then remove any trace of the binary's
existence for your convenience.

## Comments

**Comments** are text that is ignored by the compiler.  In Chai, comments do
not have ANY special behavior outside of just being comments. 

Line comments begin with hash symbol and continue until the end of the line.

    # I am a line comment

Block comments begin with `#!` and end with `!#`.  They can go over multi-lines
or be nested within a single line.

    #! I
    am a
    block
    comment !#

    some_code #! I am an nested block comment !# some_more_code

Comments are absolutely a valuable tool that should be used frequently, but not
excessively.  

## Types and Operators

If we want to do a bit more than just print "Hello, world!", we will need to
engage in a brief discussion of Chai's type system.  Namely, Chai provides us
with several basic types:

| Label | Meaning |
| ----- | ------- |
| `i8` | 8-bit signed integer |
| `i16` | 16-bit signed integer |
| `i32` | 32-bit signed integer |
| `i64` | 64-bit signed integer |
| `u8` | 8-bit unsigned integer |
| `u16` | 16-bit unsigned integer |
| `u32` | 32-bit unsigned integer |
| `u64` | 64-bit unsigned integer |
| `f32` | 32-bit float |
| `f64` | 64-bit float |
| `string` | UTF-8 encoded string |
| `bool` | boolean |
| `rune` | 32-bit, unicode code point |
| `nothing` | empty/no value |

You will notice that Chai considers strings to be a primitive type and that it
operates on Unicode by default.  We will talk in more detail about strings in
a later section, but for now, let's focus on the numbers.

All integers and floats in Chai have fixed sizes and their size is apart of
their name.  On every platform, `i32` will be a 32 bit integer.  This tends to
ensure that your API has more standard behavior on all systems.

### Number Literals

In Chai, standard "integer" numbers can be used for *any* of the numeric types.
For example, if you write the number `42` in your program, it can be inferred to
be an integer or a float of any size depending on how it is used.  Note that
this does not mean that it doesn't have a single, concrete numeric type; rather,
the type inferencer applies no initial constraints outside of it being a number
to the literal.  This is important because Chai does NOT perform or support ANY
"type coercion" (ie. where the compiler implicitly converts between two types
without an explicit cast).  Thus, the type inferencer has to make up for that
loss by being more general with its initial constraints.

However, some literals will have more specific constraints applied to them.  For
example, the following are all floats:

    4.5
    0.141
    6.626e-34
    3E8

however, their size is not constrained.

Similarly, the following are all guaranteed to be some form of integer:

    5u   # unsigned
    67l  # long (64 bit)
    87ul # unsigned + long => `u64`

Note that if a type for a literal cannot be determined from context (which
happens surpisingly often), then the compiler will pick a sensible default:
generally one of the 32 or 64 bit forms.  If you want a specific type, then you
should state it explicitly.

### Arithmetic

Arithmetic in Chai works similarly to any other programming language.  Below are
the full list of builtin arithmetic operators.  These work for all numbers.

| Operator | Operation |
| -------- | --------- |
| `+` | Add two numbers |
| `-` | Subtract two numbers OR negate one number |
| `*` | Multiply two numbers |
| `/` | Divide two numbers and produce a floating point result |
| `//` | Divide two numbers and produce a floored, integer result |
| `%` | Find the remainder of a division operation |
| `**` | Raise a number to a non-negative, integer power |

Notice that there are two division operators in Chai: one for floating point
division and one for integer division.  This is to avoid random casts being
littered all over your program.  Note that both operators work for both kinds of
input (integers and floats).

Chai also supports parentheses and applies standard operator precedence rules
for arithmetic (ie. exponents, multiplication and division, addition and
subtraction -- performed left-to-right for ties of precedence).

    4 + 5              # => 9
    (65 * 0.8) // 2    # => 26
    0.5 ** 2           # => 0.25
    (3.14 + 2.72) * 64 # => 375.04
    5 % 3              # => 2
    -10 / (6 - 3)      # => -3.333...

### Type Casting

Type casting allows for the conversion of one type into another.  These are
performed using the `as` keyword.  All casts will fail at compile-time if the
cast is invalid.

    5.4 as i32      # => 5
    12  as f64      # => 12.0
    "string" as f64 # COMPILE ERROR

Note that casts can cause data to be lost during conversion (eg. `5.4` to `i32`
essentially floors it). 

### Booleans

**Booleans** are a common fundamental type in Chai that used to represent a
true/false value.  Their literals are `true` and `false`.

Several operators are defined on booleans, called **logical operators**:

| Operator | Operation |
| -------- | --------- |
| `&&` | Logical AND |
| `||` | Logical OR |
| `!` | Logical NOT |

> Both logical AND and logical OR support short-circuit evaluation.

These operators behave as standard 
[boolean logic](https://en.wikipedia.org/wiki/Boolean_algebra) operators.

Several other operators are used to produce boolean values, called **comparison
operators**:

| Operator | Operation |
| -------- | --------- |
| `==` | True if both values are equal |
| `!=` | True if both values are not equal |
| `<` | True if the LHS is less than the RHS |
| `>` | True if the LHS is greater than the RHS |
| `<=` | True if the LHS is less than or equal to the RHS |
| `>=` | True if the LHS is greater than or equal to the RHS |

Both `==` and `!=` are defined for all values of the same type.  However, the
other comparison operators are only defined for numbers and runes by default.

Here are some examples of these operators:

    5 > 3           # => true
    "hi" == "hello" # => false
    7.6 <= -8.1     # => false      
