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
  * [Compiler](#compiler-prog)
  * [Standard Library](#std-lib-prog)
- [Building the Compiler](#building)
- [Compilation Pipeline](#pipeline)
- [Development](#development)
  * [Current Approach](#current-approach) <-- Update?
  * [The Go Implementation](#go-impl) <-- Remove Note?
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

## <a name="progress"> Progress

This section details the progress made on the compiler as of now.  It is not
perfectly up to date but should help to give some idea of where we are.

*NB: Alpha Chai is the name for the reduced subset of Chai being implemented for bootstrapping purposes.*

### <a name="compiler-prog"/> Compiler

- [ ] Alpha Chai in Python <--
- [ ] Alpha Chai in Alpha Chai
- [ ] Chai in Chai

See the Trello for more detailed info on the progress of Step 1

### <a name="std-lib-prog"/> Standard Library

*These are only the planned features of Alpha Chai*

- [ ] Builtin Types
  * [ ] `Array`
  * [ ] String
  * [ ] `List`
  * [ ] `Dict`
  * [ ] `Iter` 
- [ ] Runtime
  * [ ] Program Startup
  * [ ] Graceful Exit
  * [ ] Global Initializers
  * [ ] Signaling
  * [ ] Allocator
  * [ ] Garbage Collector
  * [ ] Argc and Argv
- [ ] Standard I/O
  * [ ] `puts`
  * [ ] `puti`
  * [ ] `putf`
- [ ] String Manipulation
  * [ ] `StringBuilder`
  * [ ] `starts_with`
  * [ ] `ends_with`
  * [ ] `contains`
  * [ ] `trim_left`
  * [ ] `trim_right`
  * [ ] `parse_int`
  * [ ] `parse_float`
  * [ ] `to_string`
- [ ] File I/O
  * [ ] Open and Close
  * [ ] Read File
  * [ ] Write File
  * [ ] Create File
  * [ ] Basic Path Manipulation
  * [ ] Walk Directory
  * [ ] Create Directory
- [ ] TOML
  * [ ] Parsing
  * [ ] Serialization
  * [ ] Deserialization (needs reflection?)

## <a name="building"> Building the Compiler

TODO

## <a name="pipeline"> Compilation Pipeline

The compilation pipeline, that is the stages of compilation depend on the
iteration of the compiler but in general the flow is as follows:

1. Source Text
2. Typed AST
3. Chai MIR
4. LLVM IR

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
of the compiler using the Python's LLVM bindings -- this compiler will compile a
simple subset of Chai called Alpha Chai.  Once that is finished, a full Chai
compiler will be implemented in Chai rendering the language full self-hosting.
Eventually, the compiler will be able to compile itself.

The `bootstrap` directory contains the Python implementation.  Once that
implementation is finished, the `compiler` directory will contain the actual
self-hosted (and final) version of the compiler.

I tried briefly writing it in other languages (after the Go impl), but
ultimately Python is simply the best language that I am comfortable with to
write a compiler in.  But, for numerous reasons, Python is obviously not a
candidate for the final compiler.  It will allow me to rapidly assemble a
working compiler (I was able to completely catch up with a C implementation that
took me over a week to write in one day of Python coding) that I can then use to
make my language self-hosting.

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