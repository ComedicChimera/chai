***NB: This section needs to be moved to a later chapter for more useful examples.***

## While Loops

A **while loop** is a control flow construct that allows us to repeat a block of
code while a condition is true.  In Chai, these loops begin with the keyword
`while` followed immediately by the condition.  After the condition, a block is
begun with a newline and concluded with `end`.

    while condition

    end

The `condition` can be any boolean value.  Like `if` expressions, you can place
as many statements or subblocks as you want within the loop's body.

    let i = 0
    while i < 10
        if i % 2 == 0
            println("i is even")
        end

        println(i)
        i++
    end

Loops are, like if expressions, actually expressions in Chai; however, their
behavior as an expression is significantly more complex than that of a simple
conditional expression so we will choose to only consider them as statements for
now.

### Break and Continue

Chai also provides two **control statements** for working with loops.  These
statements are `break`, which exits the loop immediately, and `continue` which
jumps to the next cycle of the loop (still rechecking the condition of course).

    while true
        if some_cond
            # exit the loop
            break
        end

        println("looping...")
    end

These statements are only valid within the context of a loop.

    # in regular code...
    break  # ERROR

Related to `break` and `continue` are **after blocks** which run based on how
the loop was exited.  These blocks begin with the keyword `after` and are placed
before the `end` of the loop (like the `elif` and `else` branches of the if
statement).  There are two variants: `after break` and `after continue`.  The
former is run only if the loop exits via breaking, and the latter runs only if
the loop exits normally (ie. without breaking).  Neither, both, or either can be
supplied at any time.  

Their primary purpose is to avoid meaningless flag variables.  For example, consider
we want to write a simple loop to ...