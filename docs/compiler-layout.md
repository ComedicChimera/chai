# Compiler Layout

This document describes the layout of the Chai compiler and how it compiles your
code (at a high-level).

## Package Table

Here is a list of the all the packages in the Chai compiler listed with their
purpose.

| Package | Purpose |
| ------- | ------- |
| `chai/build` | coordinates the compilation process |
| `chai/cmd` | launches the `chai` application and handles the CLI |
| `chai/common` | stores common symbols and functions used throughout the compiler |
| `chai/logging` | displays information to the user |
| `chai/mods` | implements types and functions for processing modules |
| `chai/resolve` | resolves and rearranges symbols for resolution |
| `chai/sem` | implements types and functions for semantic analysis and HIR |
| `chai/syntax` | scans and parses Chai source code |
| `chai/typing` | defines and implements Chai's type system and inferencer |
| `chai/walk` | performs semantic analysis and HIR generation |