# Control Flow

This chapter introduces the basic control flow of Chai.  Some more advanced
aspects of control flow will be explained later.

## If, Elif, and Else

The most basic control flow in Chai is the **if statement**.  It runs the code
in its body if the condition in its head is true.  If statements begin with the
`if` keyword followed by a condition, a newline, the body of the if statement
and an `end`.

    if condition
        body
    end

> Indentation is never necessary in Chai, but it is strongly recommended to
> improve code readability.  Newlines, however, are required at the end of
> statements and at the beginning of blocks.

The condition can be any boolean.  For example,

    let x = random_value()
    
    if x % 2 == 0
        println("x is even")
    end

The body can be multiple statements or even an if statement inside another if:

    if x < 2 || x > -2
        if x == 0
            println("x is zero")
        end

        println("x is in [-2, 2]")
    end

If statements can also have an `else` condition which runs if the statement
itself is found out to be untrue.

    if x < 0
        println("x is negative")
    else
        println("x is a natural number")
    end

Notice that we only need one `end` for the full if/else tree.  This is true even
if our blocks contain multiple statements.

We can also add an arbitrary number of `elif` conditions which run if the previous
conditions in the tree fail.  These can be written with or without an `else`.

    if x % 3 == 0 && x % 5 == 0
        println("FizzBuzz")
    elif x % 3 == 0
        println("Fizz")
    elif x % 5 == 0
        println("Buzz")
    else
        println(x)

> The above program is logic for [Fizz Buzz](https://en.wikipedia.org/wiki/Fizz_buzz)

## If as an Expression

If blocks (with elif and else branches) can also be expressions.  They can be written
identically to regular if statements, recalling that the last value of any block is
the value yielded.

    let kind = if x % 2 == 0
        "even"
    else
        "odd"
    end

However, unlike regular blocks, expression blocks must be **exhaustive**.  This
means that they return a value for all possible input.  For if statements, this
essentially means they must have an else clause.

    let y = if x < 0
        -1
    end  # <-- ERROR: not exhaustive

As you might have noticed, the current notation for if blocks can be a bit
verbose at least in terms of whitespace.  As such, Chai allows all blocks
(expressions or not) to utilize an alternative syntax when they only have an
expression as their body.

We use the `->` followed by an expression instead of a newline and a block to
write more concise/single-line if conditions.

    let kind = if x % 2 == 0 -> "even" else -> "odd" end

Up until this point, we have been talking about if blocks as statements and
expressions as if they were different things, but outside of the exhaustivity
rule, they are identical.  This is because, semantically, an expression is a
statement.  So when you use if expressions as statements like we did in the
first section, you are really just using expressions.

The exhaustivity check really only comes into play because the value yielded by
the if expression is significant.  When you use it as a statement, the value
returned by each branch is of type `nothing`; therefore, it has no actual
semantic value.  Similar properties occur when if statements are used in
functions or expressions where `nothing` is the expected result.  In essence,
the exhaustivity check only applies when you use the yielded value.

However, there is one dimension of if expressions, we haven't discussed yet.
This is **type consistency**.  Much like exhaustivity, this rule applies when
the value of the if statement is used.  It asserts that the returned values of
the branches are of the same type.

    let result = if x < 34 -> "string" else -> 67 end  # TYPE MISMATCH ERROR

In the above code, the if branch has a yielded type of `string` but the else
branch has a yielded type of `i32`; therefore, we have violated type consistency
and will get an error.

If you are a little bit confused at this point, that is completely normal.  This
is one of the areas where Chai deviates from the programming language norm quite
a bit so even experienced programmers might be feeling a bit shaky on the rules.

The best tip I can give you is that Chai's compiler is designed to behave
logically.  If you use the value, the if expression must be exhaustive and type
consistent because: a) the if statement has to always give back a value for you
to use and b) that value must be of a type Chai can determine at compile time
because Chai is statically typed.  All the other rules exist to satisfy those
two conditions.  

## Header Variables

**Header variables** are a convenience feature in Chai that allows you define
variables in the header of if statements that will be bound to their bodies.

For example, consider you have a function that makes a network request and
returns some form a response, and you only want to process the response's data
if it is valid.

Normally, you would have to write code that looks like this:

    let resp = make_request()

    if is_valid(resp)
        do_something_with_data(resp)
    end

However, the `resp` variable isn't actually useful outside of the body if
the if statement.  This is where a header variable comes in.

Header variables are defined like regular variables except they are placed write
after the `if`.  They are closed by a semicolon followed by the actual condition
of the if.  Rewriting our previous code to use header variables looks like this:

    if let resp = make_request(); is_valid(resp)
        do_something_with_data(resp)
    end

    # `resp` is not defined out here :)

Another useful case for header variables is when you have some value
that is reused multiple times in a condition.  

    if (a + b * c) < 1 || (a + b * c) == 0
        ...
    end

This is an obvious waste of computation (especially if the expression is not
just trivial arithmetic).  The logical solution: header variables!

    if let k = a + b * c; k < 1 || k == 0
        ...
    end

It is worth noting that you should be careful not to overuse header variables:
if every if statement in your code base defined three or four header variables,
your code is probably going to be hard to read and maintain.  
