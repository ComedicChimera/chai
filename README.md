# Chai

*Chai* is a compiled, modern, multipurpose programming language designed for
speed without sacrifice.

It is strongly-typed, versatile, expressive, concurrent, and relatively easy to
learn. It boasts numerous new and old features and is designed to represent the
needs of any software developer.

Check out the official website to install, learn, use Chai: chai-lang.dev
(*insert link*)

*Note: The website is still in development so if you want to see what language
documentation exists, check out the `docs` directory.*

**Language IP and Source Code Copyright &copy; Jordan Gaines 2018-2021**

**This language is still in development!**

## Table of Contents

- [Features](#features)
- [Building the Compiler](#building)
- [Compilation Pipeline](#pipeline)
- [Development](#development)
  * [Current Approach](#current-approach)
  * [The Go Implementation](#go-impl)
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
- Intuitive Vectorization
    * Built-in `Array` type that allows for elementwise arithmetic operations
    that are [vectorized](https://en.wikipedia.org/wiki/SIMD)
    * Array types can be multi-dimensional allowing them to act like matrices
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

## <a name="building"> Building the Compiler

Navigate to the `compiler` directory and run the following command:

    cmake -S . -B out

## <a name="pipeline"> Compilation Pipeline

The compilation pipeline, that is the stages of compilation depend on the
iteration of the compiler but in general the flow is as follows:

1. Source Text
2. Untyped AST
3. Typed AST
4. Unoptimized MIR
5. Optimized MIR
6. LLVM IR

with some boundary blurring between the typed and untyped AST

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

Right now, the compiler is being written in C.  Why C?  Because, C is one of the
only languages with *working and up-to-date* LLVM bindings that doesn't drive me
absolutely insane.  I could go on a two hour long rant about all of my issues
with C++, but suffice it to say, I prefer C.  Python is too slow and the lack of
strong, static type system is irking for a project like a compiler.  OCAML is
an ugly-looking language that I couldn't be bothered to learn.  So, the compiler
is currently being written in C.

I am NOT changing languages again.  The only universe in which this compiler
gets "rewritten" if I choose to write an implementation in Chai -- ie.
bootstrapping.  That's it.  Mostly, because the effort working on a Chai
implementation would actually be useful -- testing the integrity of the
language, building out the standard library, and making it easier for new users
to contribute.  I am still unsure whether or not I actually have the patience to
do this or not, but if I do, I will likely only develop a partial C
implementation and then bootstrap a full implementation in Chai.  Although Chai
would theoretically be an *amazing* candidate to write a compiler in, it would
be a bit of a struggle to develop all of the libraries I would need to do it.
In theory, I would need to develop those libraries anyway but doing so with an
"alpha" (ie. only partially complete) version of the language presents its own
issues.  I will update this section as I make a more substantive decision on
this matter.

### <a name="go-impl"> The Go Implmentation

The few of you that had paid any attention to this repository previously may
have remembered a fairly substantial amount of work was done on a Golang
implementation of the Chai compiler.  You will have also noticed that this
implementation is no longer on the main branch of the repository.  In fact, it
is no longer being developed at all.

The reason for this is that I discovered that the LLVM bindings for Go don't
work -- at all.  I spent 2 months trying to get them to work and ultimately
concluded that there was nothing I could reasonably do to get them to work. I
didn't check that the bindings worked before I started building the compiler:
only that they existed.  Silly me thought that the *official LLVM bindings* for
Go that were included in the *main LLVM project repository* and listed on page
for supported Go bindings on the *official LLVM website* would work. They don't.
They wouldn't even install via `go get`, and when I tried to hack them into
project manually, I got assaulted with build errors that, after months of
wrangling with CGO (which for some reason only supports GCC), I concluded that
the only options would be to copy core system libraries such as `shell32.lib`
into the primary repository directory which for obvious reasons was no going to
work.  I needed the bindings to work on Windows: full stop -- they didn't.

I googled; I posted on Stack Overflow; I posted in the official LLVM Discord:
crickets -- no one knew a thing.  Short of messaging the creators of the
bindings manually on Github (all of whom appeared to be Linux developers), I
tried everything.  The primary other solution would have been to use a library to
generate LLVM source text, pass it to `llc` (which would entail a bunch of file
IO operations, parsing, and subprocess spawning), pass it to `opt` if I need to
optimize it, and then pass it to the assembler.  The performance cost, along
with the fact that the module that did the text-LLVM generation was not actually
an official LLVM package, discouraged me from pursuing this route -- having to
repeatedly reparse source text seemed foolish.  It may have also been possible
to use a tool like `protobuf` to pass data from the Go front-end to a C++
back-end, but that adds a whole bunch of added complexity: maintaining a
multi-language build pipeline and respository, maintaining a shared data schema,
running multiple applications together at runtime, and the time to serialize and
deserialize data passed between the languages.  In short, both options seemed
to incur a significant performance cost and a mountain of tech debt -- ultimately,
neither seemed worth it.  

This is all especially salient considering as I never really considered Golang a
"great" language to write a compiler in.  I picked it because it was a language
I knew, that I could work reasonably fast in, that had decent tooling, and most
importantly, had working LLVM bindings (or so I thought).  I was more than happy
to work in another language.  But, restarting for the fourth time, well, let's
just say it nearly drove me into a psychological abyss that would likely have
ended with an obituary.  I feel that I am really living out that old adage about
madness being the price of greatness.

### <a name="whirlwind"> Whirlwind

Chai has gone through many iterations and names.  It used to be called
*Whirlwind* and the old source repository for it can be found
[here](https://github.com/ComedicChimera/whirlwind).  Needless to say, it is no
longer in development.  This is technically the fifth implementation and fourth
version of this language, but it is the only one that I ever plan on releasing.
Much of the history of this language is development is saved on Github.  If you
are curious, the original repository for this language was *Syclone* and can be
found [here](https://github.com/ComedicChimera/SyClone).