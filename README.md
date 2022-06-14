# Chai

*Chai* is a compiled, modern, multipurpose programming language designed for
speed without sacrifice.

It is strongly-typed, versatile, expressive, concurrent, and relatively easy to
learn. It boasts numerous new and old features and is designed to represent the
needs of any software developer.

Check out the official website to install, learn, use Chai: chai-lang.dev
(*insert link*)

*Note: The website is still in development, but the Markdown source for the
documentation can be found in the
[chai-lang.dev](//github.com/ComedicChimera/chai-lang.dev) repository.*

**Language IP and Source Code Copyright &copy; Jordan Gaines 2018-2021**

**This language is still in development!**

## Table of Contents

- [Features](#features)
- [Progress](#progress)
- [Building the Compiler](#building)
- [Compilation Pipeline](#pipeline)
- [Development](#development)
  * [Current Approach](#current-approach)
  * [Whirlwind](#whirlwind)

## <a name="features"> Features

- Strong, Static Typing
- Hindley-Milner Type Inference
- Generics
    * On functions and types
    * With powerful but intuitive constraint logic
- Type Classes
    * There are a special kind of higher-kinded generic that allows for
      method-based polymorphism.  In English, they replace regular classes and
      objects using generics based on functions bound to a given type (called
      methods in Chai).
- Lightweight Concurrency
- First-Class Functions 
    * Lambdas (anonymous functions)
    * Closures (lambdas that capture state around them)
    * Partial Function Application (able to call functions with only some of
      their arguments to compose a new function)
    * Operator Functions (trivially elevate operators into functions)
- Pattern Matching and Algebraic Typing
- Full Powered Expressions
    * If/Elif/Else in expressions
    * Full pattern matching (w/ fallthrough) in expressions
    * Loops in expressions (act as sequence generators)
    * Imperative blocks (you can write imperative code inside expressions)
- Versatile Built-in Collections
    * Lists (resizable collections of elements)
    * Dictionaries (ordered hash-maps)
    * Arrays (like mathematical vectors)
    * Multi-Dimensional Arrays (like mathematical matrices)
- Garbage Collection
- Directory-Based Package System 
    * All files in a directory are part of one shared namespace (called a package)
    * Packages are organized into modules for easy import resolution, build
      configuration, and dependency management
- Zero-Cost Abstractions
    * No type erasure
    * No boxing and unboxing
    * Method calls fully evaluated at compile-time
    * High-powered generics and algebraic types replace classes
- Fully Mutable State 
    * No functional programming-esque restrictions
- Monadic Error Handling 
    * No exceptions AND no ugly error handling after every function call
    * As a user, you won't actually have to deal with the headache-inducing idea
      of what a monad really is, but you will be able *intuitively* leverage the
      benefits of working with them -- they are reasonably comprehensible once
      they are presented a simpler form.

## <a name="progress"> Progress

This section details the progress made on the compiler as of now.  It is not
perfectly up to date but should help to give some idea of where we are.

*NOTE: This section refers to progress on the bootstrap compiler.*

### Target 1:

- [x] Function Definitions
- [x] Basic Annotations
- [x] Blocks and Variables
- [x] Function Calls
- [x] Bool, Int, Float, and Rune Literals
- [x] Graceful Shutdown
- [x] Naive Pointers

### Target 2:

- [x] Operator Definitions
- [x] Arithmetic Operators
- [x] Bitwise Operators
- [x] Conditional and Logical Operators
- [x] Const Pointers
- [x] Assignment
- [x] Constants
- [x] If/Elif/Else
- [x] While Loops
- [x] Break, Continue, and Return

### Target 3:

- [x] LLVM Debug + Metadata Bindings
- [x] Generate debug info for all constructs so far

*Note:* Debug info will be included on all constructs from here on.

### Target 4:

- [ ] Record Definitions
- [ ] The `.` Operator
- [ ] Aliases
- [ ] Newtypes
- [ ] Strings and String Literals

### Target 5:

- [ ] Import Statements and Import Resolution
- [ ] Visibility (`pub` keyword)
- [ ] The Prelude
- [ ] The Main Function (full implementation)
- [ ] Global Variables

### Target 6:

- [ ] Allocator (`malloc`, `realloc`, `free`)
- [ ] Garbage Collector (`gc_malloc`, `gc_realloc`)
- [ ] Safe Referencing and Dereferencing (escape analysis, etc.)
- [ ] Nullability and `core.unsafe`
- [ ] Proper Signal and Panic Handling

### Target 7:

- [ ] Type Generics
- [ ] Function and Operator Generics
- [ ] Buffer Types

### Target 8:

- [ ] Function Spaces
- [ ] Built-in String Space
- [ ] Built-in Buffer Space
- [ ] Type Classes
- [ ] Type Unions

### Target 9:

- [ ] Lists
- [ ] Dictionaries
- [ ] Sequences and Iterators
- [ ] For Loops
- [ ] Loop Generators

### Target 10:

- [ ] Tuples
- [ ] Tuple Pattern Matching
- [ ] Sum Types
- [ ] Sum Pattern Matching
- [ ] Hybrid Types
- [ ] Record Pattern Matching

### Target 11:

- [ ] `Monad`
- [ ] `Option` and `Result`
- [ ] Monadic Operators

### Target 12:

- [ ] `require` directive
- [ ] `cffi` package
- [ ] C Binding Compiler Flags
- [ ] `csrc` directive

## <a name="building"> Building the Compiler

TODO

## <a name="pipeline"> Compilation Pipeline

The compilation pipeline, that is the stages of compilation depend on the
iteration of the compiler but in general the flow is as follows:

    [Source Text]
       |
       V
    > Lexer
       |
       * - Tokens
       |
    > Parser
       |
       * - Untyped AST + Symbol Definitions
       |
    > Walker + Solver + Resolver
       |
       * - Typed AST
       |
    > Generator + LLVM
       |
       * - Object Code
       |
    > Native Linker
       |
       V
    [Executable]

## <a name="development"> Development

Chai has been a *very* long-term passion project of mine.  It is the language I
always wished that I had and has been both a joy and a total pain in the a$$ to
build, but that is the deal with software development.

I am a full-time student at Purdue University, first and foremost, so my time to
work on this language will be limited.  Furthermore, a lot of development
happens in my head and on my whiteboard -- playing with ideas and solving
problems. Just because there are no commits to this repository doesn't mean that
nothing is happening!

### <a name="current-approach"> The Current Approach

The compiler is being bootstrapped.  I am going to write a Python implementation
of the compiler using the LLVM C API with Python's `ctypes` module.  This
compiler may exclude some features not necessary to implement the compiler.
Once that is finished, a full Chai compiler will be implemented in Chai
rendering the language fully self-hosting. Eventually, the compiler will be able
to compile itself.

The `bootstrap` directory contains the Python implementation.  Once that
implementation is finished, the `compiler` directory will contain the actual
self-hosted (and final) version of the compiler.

### <a name="whirlwind"> Whirlwind

Chai has gone through many iterations and names.  It used to be called
*Whirlwind* and the old source repository for it can be found
[here](https://github.com/ComedicChimera/whirlwind).  Needless to say, it is no
longer in development.  This is technically the fifth implementation and fourth
version of this language, but it is the only one that I ever plan on releasing.
Much of the history of this language is development is saved on Github.  If you
are curious, the original repository for this language was *Syclone* and can be
found [here](https://github.com/ComedicChimera/SyClone).
