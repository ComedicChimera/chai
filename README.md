# Chai

*Chai* is a compiled, modern, multipurpose programming language designed for
speed without sacrifice.

It is strongly-typed, versatile, expressive, concurrent, and relatively easy to
learn. It boasts numerous new and old features and is designed to represent the
needs of any software developer.

Check out the official website to install, learn, use Chai: chai-lang.dev
(*insert link*)

**Language IP and Source Code Copyright &copy; Jordan Gaines 2018-2021**

**This language is still in development!**

## Table of Contents

- [Features](#features)
- [Building the Compiler](#building)
  * [Stage 0 Interpreter](#stage-0i)
- [Compilation Pipeline](#pipeline)
  * [Interpreter](#interpret-pipeline)
  * [Compiler Iterations](#compile-pipeline)
- [Development](#development)
  * [Current Approach](#current-approach)
  * [The Go Implementation](#go-impl)
  * [Whirlwind](#whirlwind)

## <a name="features"> Features

TODO

## <a name="building"> Building the Compiler

This section documents how to build different iterations of the compiler:

### <a name="stage-0i"> Stage 0 Chai Interpreter

The Stage 0 interpreter is written in Haskell, but requires LLVM IR mainly for
the purpose of allowing the Stage 0 compiler to access LLVM functions. Python is
also required to run the build script.

- [Haskell 2010](https://www.haskell.org/)
- [cabal](https://www.haskell.org/cabal/)
- [LLVM v12.0.1](https://llvm.org/)
- [zlib v1.2.11](https://zlib.net/)
- [libxml2](http://xmlsoft.org/)
- [Python 3.8](https://python.org)

Navigate to the `compiler/stage0i` directory.  Start by running the command
below to install all the interpreter's dependencies.

    cabal update
    cabal install

To actually build the interpreter, you will need: to do a bit more work:

Firstly, you will need to specify the path to your *prebuilt* version of LLVM,
built to whatever configuration (eg. `Release` or `Debug`).  The `llvm-dir` is
the path to that prebuilt version: note that you will still need to supply the
source code for LLVM (eg. the main LLVM include directory).  

Furthermore, you will need to supply a path to the LLVM binary directory of your
build: ie. where the build script can find `llvm-config` to get all the
libraries it needs to link.

Finally, you should note that suitable versions of the libraries `libxml2` and
`zlib` should be supplied.  By default, the build script checks for these in the
`vendor` directory.  These can also be installed in other location provided your
compiler can find them.  However, on Windows, they are both required. Also, note
that the `zlib` library file is called `z.lib` NOT `zlib.lib`.  LLVM expects it
which that name -- I have no idea why at the normal version of `zlib` ships with
the latter, incorrect, name.

Once you have all that, run the `build.py` script like so:

    python build.py [-llvm-dir <LLVM parent directory>] [-llvm-bin-dir-suffix <subpath to the directory containing llvm-config>] [-vendor-dir <vendor directory>]

This will produce a binary in a folder call `dist_newstyle` that is the
standalone binary for the Stage 0 Chai interpreter.

### The Stage 0 Chai Compiler 

TODO

## <a name="pipeline"> Compilation Pipeline

The compilation pipeline, that is the stages of compilation depend on the
iteration of the compiler but in general the flow is as follows:

1. Source Text
2. Untyped AST
3. Typed AST
4. Unoptimized MIR
5. Optimized MIR
6. LLVM IR

Some notably mutations depend on the iteration:

### <a name="interpret-pipeline"> Interpreter

The interpreter obviously doesn't target LLVM IR, but instead interprets the
optimized Chai MIR directly.

The interpreter's directory structure is as follows:

| Directory | Purpose |
| --------- | --------- |
| `Syntax` | Lexing, Parsing, AST representation |
| `Depm` | Dependency management, global symbol resolution |
| `Semantic` | Type solving and expression validation |
| `MIR` | Lowering AST to MIR, Optimizing MIR, MIR representation |
| `Interpret` | Interpreter |

### <a name="compile-pipeline"> Compiler Iterations

Because the compiler have a little bit more side-effect freedom than the Haskell
interpreter.  Therefore, sometimes stages are blurred a bit more (ie. the AST
may be partially typed exiting the parser).

TODO: compiler directory structure

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

After some fiddling around with different tools and languages, I ultimately
realized that no language could adequately meet my needs but my own.  So,
I have decided to incrementally bootstrap Chai.  The idea goes like this:

1. Write an interpreter for a simpler subset of Chai (called Stage 0 Chai) in
   Haskell.
2. Write a compiler in Stage 0 Chai for Stage 0 Chai.  Run the interpreter on
   the compiler with itself as input -- thereby compiling the Stage 0 Compiler
   with itself.
3. Write a compiler for Chai (fully-implemented, called Stage 1) in Stage 0 Chai
   (borrowing from the compiler for Stage 0 Chai to avoid rewriting a bunch of
   the compiler unnecessarily) and have it compile itself (proving that the Stage 1
   compiler can compile Stage 0 Chai).
4. Write a compiler for Stage 1 Chai in Stage 1 Chai using the previously
   implement Stage 1 Compiler.  Have it compile itself to make the compiler
   fully self-hosting.  This compiler will also borrow heavily from the Stage 1
   compiler written in Stage 0 Chai.

At each stage, the compiler self-validated by compiling itself (which can be reliably
checked against the version compiled by the previous iteration).  No previous version
is destroyed, allowing for more validation.

Furthermore, at each compiler iteration, more of the standard library is built
and/or expanded/revised.  This allows the library to grow incrementally with the
compiler and adds further validation to ensure the compiler is working properly.
It also means that when the final compiler is finished, Chai will have a
reasonably substantive standard library already -- leaving on a few key modules
unimplemented. Even better, this library will already be reasonably well-tested
since it will be in use in the very compiler that is compiling it.

Finally, note that at each stage, the language itself isn't broken or an earlier
version deprecated, but rather extended.  For example, the first version won't
support type classes, but when they are added, the old codebase should still
compile.  The library APIs may be updated to use them, but in such way that no
old code is broken.  

This leads to a somewhat novel directory structure in the `bootstrap` directory
of:

| Directory | Compiler Iteration |
| --------- | ------------------ |
| `stage0i` | Stage 0 Chai interpreter |
| `stage0_0c` | Stage 0 Chai compiler written in Stage 0 Chai |
| `stage1_0c` | Stage 1 Chai compiler written in Stage 0 Chai |
| `stage1_1c` | Stage 1 Chai compiler written in Stage 1 Chai |

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
to work in another language.  

After this, I briefly explored writing the whole compiler in C++ (which is the
native language of LLVM).  However, I determined that doing so would be far too
painful, time consuming, and error-ridden to justify.  So I ultimately scrapped
the pure C++ implementation in favor of polylingual implmentation.  Although I
had originally decided against this, after trying my hand at the pure C++
approach, I realized that trying to write a frontend in either C++ or Go was
truly untenable. 

### <a name="whirlwind"> Whirlwind

Chai has gone through many iterations and names.  It used to be called
*Whirlwind* and the old source repository for it can be found
[here](https://github.com/ComedicChimera/whirlwind).  Needless to say, it is no
longer in development.  This is technically the fifth implementation and fourth
version of this language, but it is the only one that I ever plan on releasing.
Much of the history of this language is development is saved on Github.  If you
are curious, the original repository for this language was *Syclone* and can be
found [here](https://github.com/ComedicChimera/SyClone).