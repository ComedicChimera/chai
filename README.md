# Chai

*Chai* is a compiled, modern, multipurpose programming language designed for
speed without sacrifice.

It is strongly-typed, versatile, expressive, concurrent, and relatively easy to
learn. It boasts numerous new and old features and is designed to represent the
needs of any software developer.

Check out the official website to install, learn, use Chai: chai-lang.dev
(*insert link*)

*Note: The website is still in development, but the Markdown source for the
documentation can be found in [docs/tutorial](/docs/tutorial/)

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

- [x] Target 1: Basic Compilation
    - Functions (void return)
    - Variables (no assignment)
    - Function Calling
    - Numeric and Character Literals
    - Type Casting
    - Graceful Shutdown
- [x] Target 2: Expressions and Control Flow
    - Return Statements
    - Assignment
    - Arithmetic Operators
    - Bitwise Operators
    - Comparison Operators
    - Boolean Literals
    - Logical Operators
    - Increment and Decrement
    - If/Elif/Else
    - While Loops
    - C-Style For Loops
    - Break and Continue
- [x] Target 3: Structs
    - Struct Types
    - Struct Declarations
    - Struct Literals
    - Property Accessing
    - Property Mutating
- [ ] Target 4: Arrays and Strings
    - Array Types
    - Array Literals
    - Array Indexing
    - Fast For Each Loops for Arrays
    - String Literals
    - String Indexing
- [ ] Target 5: Packages and Imports
    - Import Statements
    - Package Declarations
    - Import Resolution
    - Visibility/The Public Modifier
    - Global Variables
    - The Prelude (the `core` package)
    - Runtime Separation (the `runtime` package, `__chai_main` and `__chai_init`)
- [ ] Target 6: Runtime and Dynamic Memory
    - Proper Signal Handling
    - The Allocator
    - The Garbage Collector
    - Allocation with the `&` Operator
    - Escape Analysis
    - Nullability Semantics
    - Dynamic Array Allocation (`make` expressions)
- [ ] Target 7: Debug Info
    - Module-Level Metadata
    - Function Metaata
    - Struct Metadata
    - Scope Metadata
    - Positional Data
    - Debug Declarations
    - Debug Assignment

*NOTE: Other targets TBD*
 
Optimizations To Implement at Some Point:

- Copy Elision
- RVO

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

TODO: update when approach is decided

### <a name="whirlwind"> Whirlwind

Chai has gone through many iterations and names.  It used to be called
*Whirlwind* and the old source repository for it can be found
[here](https://github.com/ComedicChimera/whirlwind).  Needless to say, it is no
longer in development.  This is technically the fifth implementation and fourth
version of this language, but it is the only one that I ever plan on releasing.
Much of the history of this language is development is saved on Github.  If you
are curious, the original repository for this language was *Syclone* and can be
found [here](https://github.com/ComedicChimera/SyClone).
