# Variables

This chapter introduces variables, assignment, scoping, and blocks.  Chai's
system for handling variables and state is fairly intuitive once you understand
the syntax and the rules.

It is worth noting that from this point forward, the main function will be
elided: it can be assumed that all code that is not itself a top-level
definition is enclosed in an appropriate function (normally the main function).
Additionally, all code samples can be assumed to have access to `println`.
Furthermore, made up functions may occasionally be used: eg. `random_value()`
which doesn't actually exist in the language (at least not in that form). Just
use common sense when reading the documentation.

## Blocks and Statements

Up until now, the only Chai code we have dealt with has been in the form of
definitions and simple expressions.  In order to work with variables, we need to
discuss "procedural" code: namely, blocks.

A **block** is a set of **statements** that execute sequentially.  Blocks begin
with the keyword `do` followed by a newline and are ended by the `end` keyword.

    do
        statement1
        statement2
        ...
    end

Notice that Chai uses newlines to delimit statements.  

In Chai, blocks are expressions where the last statement (or expression) is the
returned value of the block.  

    do
        statement1
        statement2
        (5 + 4)  # <-- returned value of the block
    end

> The parentheses on the final expression are required: this is not the case for
> most statements and atomic expressions, but for more complex expressions it
> is.

Because blocks are expressions, we can place them after the `=` in a function
definition to allow for multiple statements to be executed as part of the
function.

    def main() = do
        println("First")
        println("Second")
    end

This pattern happens so often that functions (and most other block statements as
we will see later) allow you to elide the `= do` to denote that the function has
a block body.

    def main()
        println("First")
        println("Second")
    end

## Defining Variables

**Variables** are defined with the `let` keyword followed by a name and an
**initializer**.

    let x = 5

You can specify the type of a variable explicitly using a **type extension**;
however, these are rarely required as the compiler can deduce the type of a
variable in the vast majority of cases.

    let x: i32 = 5

You can, however, elide the initializer if you provide a type extension.  This
will cause the variable to be **null initialized** -- it will be assigned a
default value (which for numbers is always `0`).

    let x: i32

You can also declare multiple variables at once, even if they are of different
types, by separating the declarations with commas.

    let x = 5, y = 1.2 * 45
    let s: string, b = -12, c: bool = true || false

If you define multiple variables with the same name (in the same scope), you
will get an error.

    let x = 5
    let x = 7  # COMPILE ERROR

Variables, once defined, can then be used by name in expressions and other
definitions.

    let a = b + x
    b + 5 * x - a

If you use a variable that is not defined, you will also get an error.

    z + 4  # COMPILE ERROR

Finally, you can initialize multiple variables at once with the same value like
so:

    let v1, v2 = 1

    # same can be done with types
    let s1, s2, s3: string

## Assignment

All variables in Chai are **mutable** meaning their values can be changed.
**Assignment** is an operation that changes the value of a variable.

Assignment is performed with the `=` operator.

    let x = 5
    
    # -- SNIP --
    
    x = 10

You can assign to multiple variables at once provided that the number of
variables on the left matches the number of values on the right.

    let a, b: i32

    a, b = 5 * x, -7 - x

Variables are assigned based on their left to right ordering.  In the above
expression, `a` is bound to `5 * x` and `b` is bound to `-7 - x`. 

Chai fully evaluates the right side an assignment before performing the actual
assignment operation.  This means that you can trivially swap two variables
in one line of code:

    a, b = b, a  # b and a are both fully evaluated before they are reassigned here

Chai also supports **compound assignment operators**.  Essentially, these operators
provide a concise alternative to the pattern `var = var op expr`.  For example:

    a = a + b

    # this is exactly equivalent to the code above
    a += b

Most of Chai's operators have compound assignment forms including all of the
arithmetic operators and all of the binary logical operators (ie. `&&` and
`||`).

    let x, y, z = 5

    x *= y + z
    z -= 12

These operators can also be applied to multiple values at once.

    x, y *= z, -z

Finally, Chai supports the **increment** (`++`) and **decrement** (`--`)
operators.  These operators add one or remove one for a number variable.

    x++  # add one to x
    z--  # subtract one from z

> Chai exclusively supports these operators as statements -- they cannot be used
> inside simple expressions like in C and C++

## Shadowing

**Shadowing** is the process by which one variable can take the place of another
in a lower scope.  A **scope** is simple the range of existence of a variable.
All blocks define new scopes.

For example, the variable `a` is shadowed in all lower scopes.

    let a = "hello"

    do
        let a = 5

        println(a + 2)  # prints `7`
    end

    println(a)  # prints `hello`

In general, we recommend that avoid using shadowing: it exists because it is
sometimes a very convenient behavior but can be confusing if used in excess.



