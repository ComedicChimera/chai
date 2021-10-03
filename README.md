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
- [Compilation Pipeline](#pipeline)
- [Development](#development)
  * [Current Approach](#current-approach)
  * [The Go Implementation](#go-impl)
  * [Whirlwind](#whirlwind)

## <a name="features"> Features

TODO

## <a name="building"> Building the Compiler

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

TODO: This is still being determined fully ATM -- I am still deciding whether or
not I want to bootstrap the compiler or not.  Doing so has several benefits, but
also several drawbacks.  It is all TBD.

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