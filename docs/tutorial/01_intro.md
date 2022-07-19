# Hello World

Welcome to the Chai programming language tutorial.  These documents will serve
as the primary resource for Chai until the Chai website it completed.

This tutorial is intended for programmers who already have a basic knowledge of
computer science: many concepts which are generally familiar to most programmers
will be glossed over fairly quickly.

**NOTE:** As of right now, most of these documents are provisional.

**Contents:**

- [Installation](#install)
- [Hello World](#hello)
- [Running a Program](#run)
- [Compiler Options](#opts)
- [Exercises](#exercises)

## <a name="install"> Installation

TODO

## <a name="hello"> Hello World

The very first program one writes when learning a new language is a 
*Hello World* program whose task is simple: print the phrase "Hello, world!" to
the console.

The *Hello World* program in Chai appears as follows:

    package hello;

    import println from io.std;

    func main() {
        println("Hello, world!");
    }

Let's break this program down piece by piece.  Starting with the first line:

    package hello;

This line is called a **package declaration**.  All Chai files must begin with
one.  Package declarations are used by the language to resolve imports and group
files together.  

The last identifier of a package declaration must always be the name of the
directory containing the file: so if we were to place this file in a directory
called `a`, then `a` would be the identifier declared in the package name. We
will talk more about package declarations later.

The next line of the program,

    import println from io.std;

is called an **import statement**.  It is used to import symbols from other
packages.  In this case, we are importing the function `println` from the
package `io.std` which is part of the **standard library**: a set of packages
that ship with every Chai installation and provide basic functionality needed to
write useful programs.  The `io.std` package contains all the tools used to
perform standard I/O: ie. reading and writing to the console.

Finally, we get to the meat of the program:

    func main() {
        println("Hello, world!");
    }

The code above defines a function called `main` which accepts no argument and
returns the unit value (ie. no useful value).  `main` is a special function
which contains the code that should be run immediately after your program
starts: all runnable programs must include a function called `main` defined in
the same form as above.

The body of main is contained in curly braces and contains a single statement:

    println("Hello, world!");

This statement calls the `println` function we imported earlier with a single
argument: `"Hello, world!"`.  The value `"Hello, world!"` is called a 
**string literal**: it is used to represent literal text in user source code.
In Chai, all string literals begin and end with double quotes and are UTF-8
encoded.  We will see more about how to work with strings in later chapters.

Finally, we add a semicolon at the end of line to conclude the statement.
This is because Chai is not at all whitespace sensitive so we need to include
a semicolon to tell it that we have finished our statement.

## <a name="run"> Running a Program

Now that you understand how our *Hello World* program works, we need to run
that program.  

First, create a directory called `hello` in a convenient location on your
system.  As discussed above, the directory must be called `hello` so as to be
compatible with the package declaration in our program.  If you want to use a
different directory name, then you will need to update your package declaration
as well.

Next, create the source file `hello.chai` inside this directory. Chai source
files must end with the extension `.chai`; however, unlike with the directory
name, the name of the source file is unimportant.

Finally, open up a shell of choosing and navigate to the `hello` directory.
Assuming you have installed Chai correctly, you should then be able to run the
following command to compile your source file into an executable:

    chaic .

You will see a new executable appears in your current directory called `out`
(with a platform specific extension as necessary).

You can then run this executable from the command-line and see the following
output:

    > ./out
    Hello, world!

Congratulations: you just wrote, compiled, and ran your very first Chai program!

Now, before we move on, there is one really important thing to notice here: when
invoking the Chai compiler (called `chaic`), we didn't pass the name of the file
we wanted to compile but rather the directory.  

This highlights a key difference between Chai and many other languages: Chai
considers the smallest possible translation unit to be a package: all the Chai
source files in a directory are compiled together into a single object before
being linked into an executable.  So, when we call the Chai compiler, we give it
the directory of the package we want it to compile not the individual source
file.

## <a name="opts"> Compiler Options

As you might guess the Chai compiler accepts many options and flags some of
which are going to be very useful for you going forward.

The first option is denoted `-o` (or `--outpath`) and is used to specify the
path to output the resulting executable (or other output) to.

For example, if we want to compile our *Hello World* program to an executable
called `hello` instead of `out`, we could use the `-o` option as follows:

    chaic -o hello .

If you run this command, you will see a new executable is produced called
`hello`.

The next useful option is the `-d` (or `--debug`) flag.  If included, this
flag will prompt the Chai compiler to produce debug information with your
executable so that you can use a tool like `gdb` to debug it.  

Another useful option is the `--loglevel` option.  By default, this option is
set to `verbose` which means that the Chai compiler may produce some
helpful (or unhelpful as the case may be) status messages as a part of
compilation.  If this messages annoy you, you can instead set the Chai compiler
to only display errors and warnings by setting the log level to `warn` like so:

    chaic -o hello --loglevel warn .

You can also set it to only display errors if you want to hide those pesky
warning messages by setting the log level to `error`.

    chaic -o hello --loglevel error .

And, of course, you can set it to display no output at all using the log level
`silent`.

The final option that you may find useful even this early on is the `-m` (or
`--outmode`) option.  This option is used to specify the kind of output you want
the compiler to produce.  By default, it is set to `exe` which will cause the
compiler to output an executable.  

However, if you wanted to instead produce individual object files, you can set
the output mode `obj`.  With this option set, the Chai compiler will create a
directory called `out` and populate it with the individual object files of your
project.  Note that if you use the `-o` option with this mode set, you will need
to specify a directory to output to instead of `out` not a file. 

Here is an example of using the option to output object files to a directory
called `obj`.

    chaic -o obj -m obj .

Chai generates object files in a standard format (eg. ELF on Linux, COFF on
Windows) so you should be able to use a linker provided by a tool like GCC to
link them. However, be warned that may encounter link errors related to system
utilities if you try to link with these object files directly.  It is generally
not recommended that you do this unless you have a specific reason to.

**TODO:** Should we add the capability to pass object files to Chai like `gcc`.

There are several other output modes you can see by running:

    chaic --help

These modes enable you to set the Chai compiler to output static libraries,
dynamic libraries, and even Assembly files.  Since these topics are a bit more
complicated, we will avoid them for now, but it is good to know the option
exists.

## <a name="exercises"> Exercises

As you read through this tutorial, I highly recommend that you play around with
Chai code and try to write some small programs as you go along.  This will help
you to get a concrete sense for how to actually use Chai instead of just
thinking about it in the abstract.

To that end, most of the chapters in this guide, will include one or more
*exercises* at the end of the chapter in a section like this one.  Most of these
exercises will be fairly short programming problems designed to help you test
your Chai skills.  Any coding solutions to the exercises will be provided in the
`solutions` directory.

Here is the exercise for this chapter:

1. Modify the *Hello World* program to print "Hello, Chai!" instead of "Hello,
   world!".
