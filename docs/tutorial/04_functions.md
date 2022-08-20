# Functions

**Table of Contents**

- [Defining Basic Functions](#defining)
- [The Unit Type](#unit)
- [Constant Parameters](#consts)
- [Scoping](#scoping)
- [Recursion](#recursion)

## <a name="defining"> Defining Basic Functions

We have been using functions for a while, and now its our turn to create our own!

A **function definition** begins with the `func` keyword followed by the name of
the function.  After the name comes a pair of closing parentheses.  Any parameters
the function takes are defined inside these parentheses.  Finally, the end of
the function definition is its return type.

    func name(parameters) return_type

The parameters are defined like variables.  Each of the parameters are separated
by commas and begins with the name of the parameter variable followed by a colon
and a type label.  For example, here is a function to accepts an `i64` called `a`
and an `f64` called `b` and returns an `i64`.

    func my_func(a: i64, b: f64) i64

If we have multiple parameters of the same type, we can group them together
under one type label like we do with variables:

    func add(a, b: u32) u32

If the function doesn't return anything, then we can simply elide the return
type.

    func my_print(s: string)

> Eliding the return type is equivalent to specifying the type of `unit` which
> we will discuss in more detail later in the chapter.

Of course, a function isn't all that useful without a body.  There are two
different kinds of ways we can specify a function body in Chai.  The first is to
use a **block body**.  A block body is simply a series of statement comprising
the body of the function enclosed in braces.

    func my_func() {
        statement1
        statement2
        ...
    }

> Note that we cannot use the `:` short-hand syntax for a single statement block
> with functions.

We can return values from functions using the **return statement**.  Return
statements begin with keyword `return` followed by whatever we want to return.

    func add(a, b: i64) i64 {
        return a + b;
    }

Return statements cause the function to exit as soon as it is executed.  

    func abs(n: f64) f64 {
        if n < 0:
            return -n;

        // We only reach here if `n` >= 0.
        return n;
    }

All functions that specify a return type (ie. return something) must always
return a value via a `return` statement.  For example, if we left off the
`return n` in the `abs` function, we would get an error:

    func abs(n: f64) f64 {
        if n < 0:
            return -n;

        // ERROR: missing return statement
    }

Note that the function doesn't have to explicitly end with a return statement:
as long as all possible paths of execution return a value, we are in the clear.

    func abs(n: f64) f64 {
        if n < 0:
            return -n;
        else:
            return n;

        // We don't need a return statement here: the code above always returns
    }

In functions that don't return anything, we simply elide the value in the return
statement: this allows us to still exit early if the function either doesn't
need to or cannot continue.

    func print_sqrt(n: f64) {
        if n < 0 {
            println("error: negative number");
            return
        }

        // `sqrt` is defined elsewhere in the program
        println(sqrt(n));
    }

Notice that functions which don't return anything don't need to have a return
statement at the end.

Now that we talked about block bodies, we can take a look at the other way to
specify a function body: an **expression body**. 

Expression bodies begin with an `=` followed whatever we want to return from
the function and are ended with a semicolon.  For example, we can rewrite the
`add` function from early much more concisely using an expression body:

    func add(a, b: i64) i64 = a + b;

Notice that expression bodies are really just a short-hand for a common kind of
block body: `= expr;` is equivalent to `{ return expr; }`

## <a name="unit"> The Unit Value

In the previous section, we distinguished between functions that return something
and functions that don't.  However, this distinction is actually incorrect.

In Chai, all expressions must yield a value.  Since function calls are a type of
expression, all function calls must yield value which in turns means that all
functions must return some value.

However, as we have already seen, there are plenty of cases where functions
don't seem to return anything: for example, the `println` function doesn't need
to return any value to us.

This is where the **unit value** comes in.  The unit value represents a
"useless" or "empty" value.  Functions which don't return a meaningful result,
return the unit value by default.

The unit value is denoted by an empty pair of parentheses: `()`.  As you might
guess, the unit value is of the **unit type** which is denoted explicitly by the
type label `unit`. 

    let x: unit = ();

When we elide the return type of a function, Chai implicitly adds the type of
`unit`.  Similarly, when we write an empty return statement, Chai adds in the
missing `()`.  Finally, Chai will also add in a `return ()` at the end of any
function which returns the unit value and doesn't otherwise return.  To
demonstrate this, let's consider two identical definitions of the function
`greet`:

    // Short/standard form.
    func greet() {
        println("Hello!");
    }

    // Long/expanded form.
    func greet() unit {
        println("Hello!");
        return ();
    }

Chai will always allow you to be explicitly with your usage of the unit type,
but will almost always safely infer the "missing" values and types if you
choose to be more concise.

As a final note, you might be wondering if there is a performance cost to all
the unit value business.  The answer is fortunately no.  The unit type and value
exist primarily on a semantic level: it will generally have little to no
presence in the executable produced by the Chai compiler.  In fact, in most
cases, the unit value has a size of zero: a fact which will prove most useful
later on.

## <a name="scoping"> Scoping

TODO: sub-scopes, shadowing, parameters v local variables

## <a name="recursion"> Recursion

TODO: introduction to recursion, discussions of tail recursion?

## <a name="exercises"> Exercises

1. Write a function which computes the average of two floating-point numbers.
2. Write an implementation of the 
[Ackermann function](https://en.wikipedia.org/wiki/Ackermann_function).


