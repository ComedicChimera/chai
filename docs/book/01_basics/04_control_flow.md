# Control Flow

**Table of Contents**

- [Boolean Expressions](#conds)
- [If, Elif, and Else](#ifs)
- [Basic Loops](#loops)
- [Break and Continue](#bc)
- [Case Study: Guessing Game](#guessing)
- [If Expressions](#ifexprs)
- [Exercises](#exercises)

## <a name="conds"> Boolean Expressions

Chai's builtin *Boolean type* is identified by the type label `bool`.  The
constants `true` and `false` correspond to the true and false Boolean truth
values respectively.

Chai provides all six standard *comparison operators*:

| Operator | Meaning | Valid Operands |
| -------- | ------- | -------------- |
| `==` | Equals | Any two values of the same type |
| `!=` | Not Equals | Any two values of the same type |
| `<` | Less Than | Any two numbers or strings of the same type |
| `>` | Greater Than | Any two numbers or strings of the same type |
| `<=` | Less Than or Equal To | Any two numbers or strings of the same type |
| `>=` | Greater Than or Equals To | Any two numbers or strings of the same type |

Note that the inequality operators, `<`, `>`, `<=`, and `>=`, compare strings 
[lexicographically](https://en.wikipedia.org/wiki/Lexicographic_order).  
Additionally, all of the comparison operators have the same precedence and are
all lower precedence than all of the arithmetic operators.

Here are some example usages of these operators.

    5 < 4           // => false
    7.8 >= 4 + 2    // => true
    "abc" != "def"  // => true
    -4 < -5         // => false
    "12" > "134"    // => false
    1 == 0.999      // => false

Chai also provides three *logical operators*:

| Operator | Meaning | Arity |
| -------- | ------- | ----- |
| `!` | Logical NOT | 1 |
| `&&` | Logical AND | 2 |
| `||` | Logical OR | 2 |

All of these operators only work for boolean values.  `&&` and `||` are
the lowest precedence binary operators; however, `!` is higher precedence then
all of the binary operators.

Here are some example usages of these logical operators:

    true && false           // => false
    6 < 4 || 5 == 5         // => true
    !("abc" == "def")       // => true
    !(5 != 4) && "a" < "b"  // => false

The binary logical operators, `&&` and `||`, both use 
*short-circuit evaluation*.  This means that the if the truth value of the
result of the expression can be determined from only the first operand
expression, then the second operand expression will not be evaluated.  For
example,

    1 == 2 && 1 + 5 - 2 > -2

Because `1 == 2` evaluates to `false`, `&&` will not bother evaluating 
`1 + 5 - 2 > -2` since it knows the overall result of the expression is always
`false`.

## <a name="ifs"> If, Elif, and Else

**If statements** are a kind *block statement* which executes its body if its
condition is true.  They are used to allow code to execute only under certain
conditions.

If statements begin with the keyword `if` immediately followed by their
condition which is just a conditional expression.  After the condition comes the
if statement's body.  The body is simply the group of statements that will
execute if the condition is true.  The bodies of if statement are represented in
code as **blocks**.

A block is group of statements which can take one of two forms: one or more
statements wrapped in curly braces `{}` or a single statement with a `:`
beginning the block.  Here is that syntax illustrated with if statements:

    // First Form
    if condition {
        one;
        or;
        more;
        statements;
    }
    
    // Second Form
    if condition:
        statement; 

As an example of an if statement in action, here is a simple snippet of code
that prints if a user-inputted number is even.

    let n: i64;
    Scanf("%d", &n);

    if n % 2 == 0:
        Println(n, "is even");
    
Of course, this program would be much more useful if we could also indicate
whether the user's value is odd instead of just not doing anything.  Luckily,
if statements can also have an **else clause** which will run if the condition
is false.  

Else clauses begin with the `else` keyword followed by a block which is its
body.  The else clause is placed immediately after the end of its parent if
statement's body.  

Here is our "even" tester modified to make use of an else clause.

    let n: i64;
    Scanf("%d", &n);

    if n % 2 == 0:
        Println(n, "is even");
    else:
        Println(n, "is odd");

In addition to just an `else`, if statements can also make use of one or more
**elif clauses** which execute if their own condition is true *and* if the
condition of the if statement and those of all previous elif clauses are false.
It should be noted that when an if statement included elif clauses, the else
clause will only execute if the if condition and all the elif conditions are
false. Furthermore, if statements which use elif clauses need not have an else
clause.

The syntax for an elif clause is the same as for an if statement except elif
clauses use the keyword `elif` before their condition and must be placed after
the if statement but before the else clause if it exists.

Here is an example program making use of elif clauses which prints the hex codes
of some common colors.

    // `string` is the type label for strings.
    let color: string;

    // Using `Scanf` to read in a string instead of an integer:
    // we use the format specifier `%s` instead of `%d`.
    Scanf("%s", &color);  

    if color == "red":
        Println("ff0000")
    elif color == "green":
        Println("00ff00");
    elif color == "red":
        Println("0000ff");
    elif color == "white":
        Println("ffffff");
    elif color == "black":
        Println("000000");
    else:
        Println("unknown color");

## <a name="ifexprs"> If Expressions

In addition to regular if statements, Chai also provides **if expressions** which
allow you to select between different expressions depending on conditions.

Let's consider an example to demonstrate how if expressions work.

    let abs: i64;
    if n < 0:
        abs = -n;
    else:
        abs = n; 

In the above code, we are storing the absolute value of some number `n` into the
variable `abs`.  

We can make this code a lot more concise using an if expression.  If expressions look
exactly like regular if statements except their bodies are denoted `=>` followed by an
expression.  Here is the absolute value code above translated into an if expression:

    let abs = if n < 0 => -n else => n;

Note that, like with short-circuit evaluation, only the code that the if
expression selects will be evaluated: eg. if `n` is non-negative, then the `-n`
code won't run at all.

If expressions can also have `elif` clauses.  For example:

    print(n);  // `print` works like `Println` but doesn't add a newline.

    Println(
        if n % 10 == 1 && n % 100 != 11 => "st"
        elif n % 10 == 2 && n % 100 != 12 => "nd"
        elif n % 10 == 3 && n % 100 != 13 => "rd"
        else => "th"
    );

However, unlike regular if statements, if expressions must *always* have an else
clause. Else clauses are required to make the if expression **exhaustive**: they
ensure that it always evaluates to something.

## <a name="loops"> Basic Loops

**While loops** are a kind of block statement that repeats their body as long as
the condition in its header is true.  They allow for code to be repeatedly
executed based on a condition.

While loops begin with the keyword `while` followed by the condition of the loop.
After the condition, the body block is placed.  Here is an example of a program
which prints out the numbers from 1 to 10 using a while loop:

    let n = 1;
    while n <= 10 {
        Println(n);

        n++;
    }

As it turns out, the pattern above is rather common: create a variable, repeat
until a condition is true, and update the variable on each iteration.  So, Chai
provides another kind of loop called a **tripartite for loop** (or C-style for
loop as they aremore commonly known).  This loop just condenses the syntax above
into a line as opposed to spreading it over many lines.  

Tripartite for loops use the keyword `for` followed by the three key statements
in order separated by semicolons: declare, check, and update.  Here is the above
while loop rewritten as a tripartite for loop:

    for let n = 1; n <= 10; n++ {
        Println(n);
    }

The code does the exact same thing, but using a much more concise syntax.

> In a later chapter, we will actually see how to make this loop even shorter
> using ranges although it doesn't work for all kinds of tripartite for loops.

Note that both the variable declaration and the update statement can be elided
in a tripartite for loop although the semicolons need to remain.  For example, if
you wanted to be able to access the variable `n` after the loop ended:

    let n = 1;
    for ; n <= 10; n++ {
        Println(n);
    }

Now, `n` is not constrained to the scope of the loop.

The final kind of loop we are going to introduce in this chapter is the 
**do-while loop**.  Do-while loops work exactly like regular while loops except
they always execute their bodies at least once.  

Do-while loops use the keyword `do` followed by a block which is their body
followed by the keyword `while` and the condition of the loop.  For example,
here is some code which uses the fictious function `getChar` to read the next
character from some stream until no more characters can be returned.

    let c: rune;  // `rune` is the type used for characters in Chai.
    do {
        c = getChar();
        Println(c);
    } while c != -1;

Notice that unlike with the other block statements, a semicolon is required
after the condition of a do-while loop.

## <a name="bc"> Break and Continue

A **break statement** is used to exit out of a loop early.  

Break statements are denoted with the keyword `break`.

    while true {
        let input: i64;
        Scanf("%d", &input);

        if input == 0:
            // Exit the loop when input = 0.
            break;

        Println("n ** 2 =", n ** 2);
    }

A **continue statement** is used to continue to the next iteration of a loop
immediately: thereby skipping the remaining content of the body.  Notably, when
a continue statement is encountered, the loop condition is still checked to
determine when to continue looping: continue statements just skip the body.

Continue statements are denoted with the keyword `continue`.

    while true {
        let input: i64;
        Scanf("%d", &input);

        if input == 0:
            // Skip the `Println` when the input = 0.
            continue;

        Println("n ** 2 =", n ** 2);
    }

> A continue statement can often be interchanged for an else clause, but there
> are many situations when introducing an else clause would be very inconvenient
> and messy.  For example, if you many nested layers of control flow, it is
> often easier to use a continue statement rather than trying to use an else
> clause.

Continue statements also highlight the single notably difference between a
tripartite for loop and a while loop: continue will not skip the update
statement of a tripartite for loop, but it will skip the update statement of a
while loop.  For example:

    for let i = 0; i < 100; i++ {
        if i % 3 == 0:
            continue;

        Println("i =", i);
    }

The code above will behave as intended: skipping iterations where `i` is
divisible by 3. However, if we were to use the more verbose while loop here, you
would find that your code gets caught in an infinite loop unless you explicitly
placed an extra `i++` before the continue statement.

All loops in Chai can have an *else clause* which will run only if the loop exits
normally (ie. via its condition).  If the loop is broken out of using a break statement,
then the else clause will not run.  

As an example, if we use the fictious function, `getElement`, we can see where
the behavior of an else clause would be useful:

    for let i = 0; i < 10; i++ {
        if getElement(i) == 0 {
            Println("Found a zero!");
            break;
        }
    } else:
        Println("Did not find a zero.");

> In later chapters, the else clause will become very useful for searching and
> validating collections of data.

## <a name="guessing"> Case Study: Guessing Game

Now, let's take a look at a sample program which is going to put all the topics
we studied in this chapter together.

This program is going to be a simple guessing game which generates a random
number between 1 and 100 and then prompts the user to guess until they get the
right number.

We are going to put this program together piece by piece so we can make sure we
fully understand how it works.  

Let's start by adding in the boilerplate code that we covered in Chapters 1 and
2.

    // Our package declaration.
    package guesser;

    // Import our standard I/O utilities.
    import Println, Scanf from io.std;

    // Our main function.
    func main() {
        // TODO
    }

Now, the first challenge of is to actually generate a random number.  Luckily,
Chai's standard library provides us a package called `random` which contains a
function called `RandInt` that does exactly what we want.  First, let's import that
function.

    package guesser;

    import printlin, scan from io.std;

    // Import the `RandInt` function.
    import RandInt from random;

    func main() {
        // TODO
    }

Then, we can call `RandInt` with the the arguments `1` denoting our lower bound and
`101` denoting our exclusive upper bound to get our random number.

    package guesser;

    import printlin, scan from io.std;
    import RandInt from random;

    func main() {
        // Generate a random number between 1 and 100.
        let n = RandInt(1, 101);

        // TODO
    }

Now, let's set up the guessing loop.  For this, we are just going to use a simple
infinite loop: ie. `while true` and use a `break` to exit out when the user guesses
correctly.  We will also add the code to get the user's guess.

    package guesser;

    import Println, scan from io.std;
    import RandInt from random;

    func main() {
        // Generate a random number between 1 and 100.
        let n = RandInt(1, 101);

        // The main guessing loop.
        while true {
            // Prompt our user to enter the number.
            Println("Enter your guess (1 <= n <= 100):")

            // Read in user input.
            let input: i64;
            Scanf("%d", &input);

            // TODO
        }
    }

Now, we can use a simple if statement to check the user's guess.  We will break
out of the loop when the user guesses correctly.

    package guesser;

    import printlin, scan from io.std;
    import RandInt from random;

    func main() {
        // Generate a random number between 1 and 100.
        let n = RandInt(1, 101);

        // The main guessing loop.
        while true {
            // Prompt our user to enter the number.
            Println("Enter your guess (1 <= n <= 100):")

            // Read in user input.
            let input: i64;
            Scanf("%d", &input);

            // Check the user's guess.
            if input == n {
                // Correct answer => exit the loop.
                Println("Correct!");
                break;
            } elif input < n:
                // Answer is too low.
                Println("Too low!");
            else:
                // Answer is too high.
                Println("Too high!");
        }
    }

That's it!  A simple guessing game programmed using only the basic concepts we covered
in this chapter.  Compile and run it on your local system to give it a try!

## <a name="exercises"> Exercises

1. Implement a simple [Fizz Buzz](https://en.wikipedia.org/wiki/Fizz_buzz)
   program: print the first 100 fizz-buzz numbers.

2. Write a program to print the nth Fibonacci number with `n` is given by the
   user.

3. Write a program to calculate the square root of `n` where `n` is given by the
   user using [Newton's Method](https://en.wikipedia.org/wiki/Newton%27s_method).
