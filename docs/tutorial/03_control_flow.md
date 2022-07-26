# Control Flow

**Table of Contents**

- [Conditional Expressions](#conds)
- [If, Elif, and Else](#ifs)
- [While and C-Style For Loops](#while)
- [Case Study: Guessing Game](#guessing)
- [Exercises](#exercises)

## <a name="conds"> Conditional Expressions

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
    scanf("%d", &n);

    if n % 2 == 0:
        println(n, "is even");
    
Of course, this program would be much more useful if we could also indicate
whether the user's value is odd instead of just not doing anything.  Luckily,
if statements can also have an **else clause** which will run if the condition
is false.  

Else clauses begin with the `else` keyword followed by a block which is its
body.  The else clause is placed immediately after the end of its parent if
statement's body.  

Here is our "even" tester modified to make use of an else clause.

    let n: i64;
    scanf("%d", &n);

    if n % 2 == 0:
        println(n, "is even");
    else:
        println(n, "is odd");

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

    // Using `scanf` to read in a string instead of an integer:
    // we use the format specifier `%s` instead of `%d`.
    scanf("%s", &color);  

    if color == "red":
        println("ff0000")
    elif color == "green":
        println("00ff00");
    elif color == "red":
        println("0000ff");
    elif color == "white":
        println("ffffff");
    elif color == "black":
        println("000000");
    else:
        println("unknown color");

## <a name="while"> While and C-Style For Loops

## <a name="guessing"> Case Study: Guessing Game

## <a name="exercises"> Exercises

1. Implement a simple [Fizz Buzz](https://en.wikipedia.org/wiki/Fizz_buzz)
   program: print the first 100 fizz-buzz numbers.

2. Write a program to print the nth Fibonacci number with `n` is given by the
   user.

3. Write a program to calculate the square root of `n` where `n` is given by the
   user using [Newton's Method](https://en.wikipedia.org/wiki/Newton%27s_method).
