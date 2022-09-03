# Variables and Constants

**Table of Contents**

- [Variable Declarations](#vars)
- [Assignment](#assignment)
- [Constants](#constants)
- [Case Study: Adder](#adder)

## <a name="vars"> Variable Declarations

A **variable** named location in memory used to store values.

In Chai, you can declare variables using the `let` keyword like so:

    let x = 10;
    println(x);  // Prints 10.

The type of `x` is inferred based on the value it is initialized with.

You can specify the type of a variable using a **type extension** like so:

    let pi: f64 = 3.14159265;

If you specify a type extension, you don't have to explicitly initialize the
variable: it will be initialized to its **null value** by default (which is zero
for all numeric types).

    let y: i32;  // Implicitly: `y = 0`.

> This kind of variable declaration only works for types which are *nullable*.
> We will see what this means when we encounter our first non-nullable type.

You can also declare multiple variables at once if you separate their
declarations using commas:

    let a = x * 4, b = y + 6 - x;

## <a name="assignment"> Assignment

All variables declared with `let` are *mutable* meaning we can assign to them.
Assignment in Chai is done using the `=` operator like so:

    a = b;

You can assign to multiple variables at once by separating the variables and
expressions using commas. 

    x, y = 5, 6;  // Equivalent to `x = 5, y = 6`.

Chai fully evaluates the right-hand side expressions before it assigns them to
the left-hand side.  This means that you can trivially swap the values of two
variables using multi-assignment without having to use a temporary variable.

    a, b = b, a;  // Swaps a and b's values.

Often, we want to apply an operator to between the variable's value and a new
value.  For example, if wanted to add 2 to the variable `x`, we would write:

    x = x + 2;

Because this kind of operation is so common, Chai provides a short-hand for
statements like the one above called **compound assignment**:

    x += 2;  // Equivalent to `x = x + 2`.

Compound assignment can be performed with any of the binary arithmetic and
bitwise operators:

    x *= 2;   // Double x.
    y /= 4;   // Divide y by 4.
    a <<= 1;  // Left-shift a by 1.

Furthermore, we can apply compound assignment between multiple values:

    x, y **= 2, 3;  // Square x and cube y.

Finally, Chai offers one more bit of short-hand for adding and subtracting
one from a variable's value which happens quite often.  This short-hand comes
in the form of the **increment** and **decrement** statements.  

They are written by placing either `++` for increment or `--` for decrement
after the variable we want to mutate.  That is:

    // Increment x by 1: equivalent to `x += 1`.
    x++;

    // Decrement y by 1: equivalent to `y -= 1`.
    y--;

> Unlike in programming languages like C, increment and decrement are statements
> in Chai: they cannot be used as expressions.  Furthermore, there is no notion
> of "postfix" vs. "prefix" incrementing and decrementing: the operator is
> always placed after the the value being mutated.

## <a name="constants"> Constants

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
