## Principles
- Minimalist: no flashy log messages or redundant information.
- Silent: if compilation succeeds, there should be zero compiler output.
- Informative: error messages should actually tell the user what went wrong.
- Contextual: error messages should give some indication of what led to them.

## Error Kinds and Contexts
There are four categories of error can occur when compiling Chai code.

| Category | Purpose |
| -------- | ------- |
| fatal | some high level error in the configuration of the compiler |
| compile | error occurred somewhere in Chai source text (invalid code) |
| package | error occurred trying to load a package (eg. empty package) |
| internal | an unexpected error (bug) was detected in the compiler |

Errors can have a context which gives extra information about where the error occurred.  The primary form of context in Chai is the **generic context**.  This indicates the series of generic type substitutions that led to a specific error.  

## Message Formatting
All errors are prefixed by `error` and similarly all warnings are prefixed by `warning`

Fatal error messages are simply displayed as one-line output with no other information.

Compile errors begin with the file the error occurred in followed by the line and column the error begins on.  Then the error message itself is displayed.  The subsection of erroneous text is displayed and highlighted (using carets).  As an example:

```
file.chai:10:15: error: undefined symbol: `a`
10 | let x = 10 + a
   |              ^
```

Package errors give an indication of what package was failed to load either by name (preferred) or by path.

```
error: failed to load package `pkg`: package contains no source files

error: failed to load package at `project/1pkg`: `1pkg` is not a valid package name
```
