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
- [Development](#development)
  * [The Go Implementation](#go-impl)
  * [Whirlwind](#whirlwind)

## <a name="features"> Features

## <a name="building"> Building the Compiler

The compiler is written in C++ and requires several dependencies:

- [LLVM v12.0.1](https://llvm.org/)
- [CMake v3.20.0 or later](https://cmake.org/)
- [zlib v1.2.11](https://zlib.net/)
- [libxml2](http://xmlsoft.org/)
- A C++ compiler and toolchain

To build the repository, first, navigate to the source directory and call the
CMake script like so:

    cmake -S . -B out -DLLVM_DIR=path_to_llvm_build -DLLVM_BIN_DIR_SUFFIX=relative_path_to_bin

You will need to specify the path to your *prebuilt* version of LLVM, built to
whatever configuration (`Release` or `Debug`).  The `LLVM_DIR` is the path to
that prebuilt version: note that you will still need to supply the source code
for LLVM (eg. the main LLVM include directory).  

Furthermore, you will need to supply a path to the LLVM binary directory of your
build: ie. where CMake can find `llvm-config` to get all the libraries it needs
to link.

Finally, you should note that suitable versions of the libraries `libxml2` and
`zlib` should be supplied.  By default, the CMake file checks for these in the
`vendor/lib` directory.  These can also be installed in other location provided
your compiler can find them.  However, on Windows, they are both required. Also,
note that the `zlib` library file is called `z.lib` NOT `zlib.lib`.  LLVM
expects it which that name -- I have no idea why at the normal version of `zlib`
ships with the latter, incorrect, name.

Once you have run CMake, you will get a either a Makefile or Visual Studio
solution depending if you are on Windows.  This should be trivial to build --
the CMake file takes care of all configuration for you.

## <a name="development"> Development

Chai has been a *very* long-term passion project of mine.  It is the language I
always wished that I had and has been both a joy and a total pain in the a$$ to
build, but that is the deal with software development.

I am a full-time student at Purdue University, first and foremost, so my time to
work on this language will be limited.  Furthermore, a lot of development
happens in my head and on my whiteboard -- playing with ideas and solving
problems. Just because there are no commits to this repository doesn't mean that
nothing is happening!

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
to work in another language.  I wanted to use Haskell, but the `llvm-hs` package
at the time of writing this only supports LLVM 9 which is obviously
dissatisfactory (and reeks of possible tech debt).  In conclusion, I pivoted to
the native language of LLVM, C++, and confirmed before beginning my brave new
enterprise, that I could get the library to build on Windows which, after some
trial and error, it eventually did.  Plus, C++ is a well-tested, well-tooled,
relatively modern and powerful language that is, above all, performant: a
not-half-bad choice to write a compiler in.

### <a name="whirlwind"> Whirlwind

Chai has gone through many iterations and names.  It used to be called
*Whirlwind* and the old source repository for it can be found
[here](https://github.com/ComedicChimera/whirlwind).  Needless to say, it is no
longer in development.  This is technically the fifth implementation and fourth
version of this language, but it is the only one that I ever plan on releasing.
Much of the history of this language is development is saved on Github.  If you
are curious, the original repository for this language was *Syclone* and can be
found [here](https://github.com/ComedicChimera/SyClone).