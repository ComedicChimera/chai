# Chai

Chai is a compiled, modern, multipurpose programming language designed for speed
without sacrifice.

It is strongly-typed, versatile, expressive, concurrent, and relatively easy to
learn. It boasts numerous new and old features and is designed to represent the
needs of any software developer.

Check out the official website to install, learn, use Chai: [chai-lang.dev](TODO)

***Language IP and Source Code Copyright &copy; Jordan Gaines 2018-2021***

***NOTE: This language is still in development! This is the fourth version of the compiler and language.***

***New Home of [Whirlwind](https://github.com/ComedicChimera/whirlwind)***

## Building the Chai Compiler

The Chai compiler requires a number of tools to build it.  In particular,

- Go v1.16 or later
- CMake (for LLVM)
- Git (to install and use LLVM)

Additionally, for Windows, you will need:

- Visual C++ Tools (`MSBuild` should be in your PATH)

For Unix platforms, you will need

- GNU Make

Because Chai was originally developed on Windows, the build script is a
Powershell script named `cacao.ps1`.  This script automates a number of common
operations and provides a simple interface to set up and build Chai.  Ideally,
you should not really do anything related to setting up or build Chai manually.

The `cacao` script accepts several subcommands which are documented below.

| Command | Purpose | Additional Arguments |
| ------- | ------- | -------------------- |
| `setup` | Setup and install all the necessary variables and components to build Chai | *none* |
| `build` | Build a Chai debug binary to the `bin` directory | *none* |
| `release` | Create a release build of Chai in the `bin` directory for the target platform | `target-platform` |

If you are just looking to quickly build from source, the following sequence of
commands is sufficient.

    git clone github.com/ComedicChimera/chai
    cd chai
    ./cacao setup
    ./cacao release [your-platform]

Note: This will take nearly an eternity so you probably want to put on an
Netflix show or something while it builds :).







