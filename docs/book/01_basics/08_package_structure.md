# Package Structure

**Table of Contents**

- [Multi-File Packages](#multi-file)
- [Global Variables](#global-vars)
- [Compile-Time Constants](#comptime-consts)
- [Package Initializers](#pkg-inits)

## <a name="multi-file"> Multi-File Packages

A **package** is a collection of Chai source files which share a common global
namespace.  Packages are the basic Chai translation unit: a package is smallest
unit of code that Chai can compile as a standalone binary.

So far, we have only dealt with single file packages; however, packages can have
multiple files.  Packages occupy a single directory rather than a single file.
All of the Chai source files in a directory are considered part of that package.
This is why we needed to create a new directory to house our tutorial code since
a package *requires* a directory to contain its contents.

All files in a package must have the same package declaration.  This means that
if we have multiple files in the same package, they must specify the same
package path: we couldn't have one file that declares the package path as `a`
and another one that declares it as `b.a`.  

As mentioned in the definition of a package, all the files in a package share a
global namespace.  This means that all the symbols declared in the global scope
of each file are visible in *all* the files.

For example, consider a package made up of two files: `a.chai` and `b.chai`.  If
`a.chai` declared a function called `foo` and `b.chai` declared a function
called `baz`, then we would be able to call `foo` from `b.chai` and `baz` from
`a.chai`.

Take a moment to let that sink in and really understand it.  It is a relatively
simple idea, but it can be a bit unintuitive for programmers that have never
used a package-based language before.

### Imported Symbols

There is one little wrinkle when it comes to package namespacing which we
haven't discussed yet: imported symbols.  These are symbols like `Println` and
`Scanf` that we bring in from other packages.

Unlike all other "global" symbols, imported symbols are *NOT* shared between
files in a package.  Although they technically exist in a file's global scope,
they remain file-local.  So, we need to import `Println` in every file that
uses it, not just once in the whole package.

The reason for this is that not all files will require the same imports: you will
often find that many packages are only used by one or two files of your package
while the rest don't need those imported symbols.  Making imports local helps
reduce clutter and makes symbols easier to find in your packages.

## <a name="global-vars"> Global Variables

Variables can be declared globally as well as locally.  Such variables are
called **global variables**.  They are declared in the global namespace of our
package and exist for the entire runtime of our program.

We can create a global variable by simply declared a regular variable inside
the global namespace of our package:

    package example;

    let myGlobal: i64 = 10;

    func main() {
        // -- SNIP --
    }

The only difference is that global variables must have an explicit type: we
cannot elide the `i64` on `myGlobal` like we could if it were a local variable.

We can access global variables from within any function in our program.

    func incrementMyGlobal() i64 {
        myGlobal++;
        return myGlobal;
    }

As mentioned above, global variables persist for the entire lifetime of our program.
Thus, whenever we call `incrementMyGlobal`, we will get a different value:

    func main() {
        Println(incrementMyGlobal());  // Prints `11`.

        Println(incrementMyGlobal());  // Prints `12`.
    }

Global variables can also be declared as constant.

    const PI: f64 = 3.14159265;

Just like local constants, global constants can't be mutated.

    func main() {
        PI = 1.414;  // ERROR
    }

Global variables are very useful as they can allow us to easily share state
between functions.  However, this does not mean that they should be used
frivolously: excessive use of global variables is *very* bad programming
practice and should be avoided.  As we move into later chapters, you will find
many alternatives to global variables that will be much more "idiomatic".  That
being said, they do still have many uses and are an important tool in your Chai
toolbox.

## <a name="comptime-consts"> Compile-Time Constants

There is a special class global constant called a **compile-time constant**.
Compile-time constants have the unique property that there values are entirely
known at compile-time.  This allows for several useful optimizations and can
give our programs a substantive performance boost over regular global
constants.

We can declare compile-time constants like global constants except we use the
`comptime` keyword instead of the `const` keyword.

    comptime PI: f64 = 3.14159265;

Unlike regular global constants, compile-time constants *must* be initialized.
Furthermore, their initialization expressions must be
**compile-time evaluable**.  This means that must fall within a specific subset
of Chai operations which can the compiler can perform determistically to arrive
at a specific, single value for the constant.

In general, the following expressions are considered compile-time evaluable:

- Number, rune, boolean, and string literals.
- The constant `null`.
- Other compile-time constants.
- Arithmetic, comparison, and logical operations using built-in operators.
- Struct literals with compile-time evaluable initializers.
  * If they have default initializers which are not compile-time evaluable,
    those initializers must be "overridden" by an explicit initializer.

They are several other expressions that are also compile-time evaluable:
consult the appendices for a full a list.

## <a name="pkg-inits"> Package Initializers

A **package initializer** is a special function which is run whenever the
package is initialized.  These initializers run before `main` and will be run in
any package that is imported as part of your project not just your root package
(as with `main`).

Their explicit purpose is to initialize our package's global state.  They are NOT
meant to be our program's entry point.

Declaring a package initializer is easy: we just declare a special function named `init`
that takes no arguments and returns no value.

    func init() {
        // Package Initialization code
    }

When our program starts up, the `init` function will be called, and we can use it to
prepare our package's global state.
