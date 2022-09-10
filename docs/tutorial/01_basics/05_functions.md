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

    func name(parameters) returnType

The parameters are defined like variables.  Each of the parameters are separated
by commas and begins with the name of the parameter variable followed by a colon
and a type label.  For example, here is a function to accepts an `i64` called `a`
and an `f64` called `b` and returns an `i64`.

    func myFunc(a: i64, b: f64) i64

If we have multiple parameters of the same type, we can group them together
under one type label like we do with variables:

    func add(a, b: u32) u32

If the function doesn't return anything, then we can simply elide the return
type.

    func myPrint(s: string)

> Eliding the return type is equivalent to specifying the type of `unit` which
> we will discuss in more detail later in the chapter.

Of course, a function isn't all that useful without a body.  There are two
different kinds of ways we can specify a function body in Chai.  The first is to
use a **block body**.  A block body is simply a series of statement comprising
the body of the function enclosed in braces.

    func myFunc() {
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

    func printSqrt(n: f64) {
        if n < 0 {
            Println("error: negative number");
            return
        }

        // `sqrt` is defined elsewhere in the program
        Println(sqrt(n));
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
don't seem to return anything: for example, the `Println` function doesn't need
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
        Println("Hello!");
    }

    // Long/expanded form.
    func greet() unit {
        Println("Hello!");
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

A **scope** is a region of visibility: it defines an area in which one or more
symbols, such as variables, is accessible.  For example, the body of a function
defines a scope in which the parameters and local variables of that function are
visible.  

There are generally two kinds of scopes in Chai: **local scopes** and 
**global scopes**.  Symbols defined in local scopes only exist for part of the
lifetime of the program.  For example, after a function returns, all the local
symbols of that function "cease to exist": they are no longer accessible.  By
contrast, symbols defined global scopes are visible for the lifetime of the
program: they always exist although some can change.  For example, all functions
are defined in the global scope of the current package.

### Subscopes

Scopes can be nested within each other.  For example, the local scopes of
functions are nested within the global scope.  Scopes nested within other scopes
are called **subscopes** of that scope.

Furthermore, we have already on several occasions created nested local scopes
within the local scopes of functions.  As an example, whenever you create a
block such as the body of an if statement, a new scope is created:

    func fn() {
        let x = 5;  // Variable `x` in local scope of `fn`.
    
        if x < 10 {
            let y = "test";  // Variable `y` in local scope of if statement.

            // -- SNIP --
        } // Variable `y` goes out of scope.
    } // Variable `x` goes out of scope.

We can also create subscopes arbitrary using a **basic block** (also simply
called a *subscope*).  Basic blocks are contained in curly braces and do nothing
other than create a new scope.
    
    let x = 5;

    // Begin a new subscope.
    {
        let y = 10;
    } // Variable `y` goes out of scope here.

### Shadowing

Variables can also **shadow** each other.  This happens when two variables are
declared with the same name in two different scopes.  For example:

    let x = 10;

    {
        let x = 3.14;

        Println(x);  // What happens here?
    }

When this happens, Chai will always used the value in the most immediate
enclosing scope.  So, in the example above, `3.14` will be printed since the
variable storing `3.14` is defined in the most immediate scope.  

This rule applies even if the shadowing variable is defined one or more scopes
above its usage.

    let x = 10;

    {
        let x = 3.14;

        {
            Println(x);  // Prints `3.14`.
        }
    }

However, shadowing will not cause variables to outlive their scope.  This is because
in order for one variable to shadow another, we have to be in or below the scope
the shadowing variable is defined in:

    let x = 10;

    {
        let x = 3.14;

        Println(x);  // Prints `3.14`.
    } // Shadowing variable `x` goes out of scope here.

    // Previously shadowed variable `x` is now visible again.

    Println(x);  // Prints `10`.

Shadowing can be a bit quirky at times so we generally recommend you don't use
it unless you have a good reason: it can often be difficult to keep track of
what variable is being used especially if scopes are nested quite deeply.

## <a name="recursion"> Recursion

Chai, like most modern programming languages, supports 
[recursion](https://en.wikipedia.org/wiki/Recursion_(computer_science)) out of
the box: Chai functions can call themselves with little difficulty.

A common recursive algorithm is the 
[factorial function](https://en.wikipedia.org/wiki/Factorial).  We can implement
it in Chai quite easily.

First, let's start with the function signature:

    func factorial(n: i64) i64

The `factorial` function takes an integer as input and returns an integer as
output.

> We won't concern ourselves with non-natural numbers in this exercise.

There are two cases to the factorial function: $n < 2$ and $n \ge 2$.  In the
first case, $1$ is always returns since $0! = 1! = 1$.  

    func factorial(n: i64) i64 =
        if n < 2 => 1 else => ...;

Then, we will consider the recursive case.  We know that $n! = n \cdot (n - 1)!$.
This can be easily expressed in Chai like so:

    func factorial(n: i64) i64 =
        if n < 2 => 1 else => n * factorial(n - 1);

Just like that, we have implemented a recursive factorial function.

You can test on it on some sample inputs to make sure it works.

## <a name="exercises"> Exercises

1. Write a function which computes the average of two floating-point numbers.
2. Write an implementation of the 
[Ackermann function](https://en.wikipedia.org/wiki/Ackermann_function).


