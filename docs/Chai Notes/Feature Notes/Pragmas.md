# Pragmas
Pragmas provide a way to control the compiler more precisely.  They are sort of like preprocessor directives in C (except Chai doesn't actually use a preprocessor).

A general rule for all pragmas is that they will only execute once per compilation, and they are always executed before any code is generated.

## Syntax
All pragmas begin with a `$` followed by the directive name (which is a normal) identifier.  Eg.

`$ warn "deprecated"`

From there, pragmas deviate quite a bit in what syntax they use for their arguments.

## Locality
Pragmas can be defined as either **global** or **local**.  **Global pragmas** affect the entire file they are defined in whereas **local pragmas** only affect the function or relevant code block they are defined in.

The **locality** of a pragma is determined entirely based on where the pragma is placed.  Pragmas which are placed at the file level are considered global.  By contrast, pragmas which are placed inside of other definitions are considered local.

## Conditional Execution
The compiler is not guaranteed to execute a particular pragma at compile time.  The following conditions are must be met for any pragma to be executed:

1. The pragma must be well formed.
2. The file must be parseable up until the pragma.
3. No pragma placed before the pragma blocks compilation.

 For example, if the pragma is placed after a `require` pragma, then it will only execute if the `require` is satisfied.

### Local Pragma Execution
In addition to the above conditions, local pragmas require that the file be fully compileable (ie. syntactically and semantically valid) in order to be executed.  Furthermore, they have special execution requirements depending on where they are defined.

##### Inside Functions (or Methods)
A pragma placed inside a function will execute under the following conditions:

1. The function is called within user code.
2. If the pragma is placed within generic control flow, the control flow branch containing the pragma is evaluated at compile-time.

##### Inside Types (Class, Union, or Space) Definitions
A pragam placed inside a type definition will only execute if that type is used within user source code.

## Builtin Pragmas
### Require
The require directive specifies a condition of compilation.  The directive identifier is `require`.  It accepts a condition that if false will prevent compilation.

**Examples:**
```
$ require false  # file will never compile

$ require os == "windows"  # conditional compilation based on OS

$ require arch == "amd64"  # conditional compilation based on architecture
```

*Note*: More advanced conditions may be supported in the future

### C Source
The C source directive will allow for the direct inclusion of C source code in Chai files.  The directive identifier is `csrc`.  It accepts a string (can be multi-line) of C code to include.  

For more information and examples of this directive, see [[C Bindings]].

### Warn
The warn directive produces an arbitrary warning message.  The directive identifier is `warn`.  It accepts a string warning message.

**Examples:**
```
$ warn "This package is deprecated."

def some_func()
	$ warn "This function is unsafe."
end
```

### Error
The error directive produces an arbitrary compilation error.  The directive identifier is `error`.  It accepts a string error message.

**Examples:**
```
$ error "Unsupported platform."

def todo()
	$ error "Unimplemented function"
end
```

### Inline Assembly.
The inline assembly directive will allow users to write assembly code inside their functions.  The directive name is asm.  It accepts an assembly string (can be muli-line) to run.  This directive can NOT be used at the file level.

## Future Pragma Ideas
These directives may be implemented at some point in Chai's future.

- Floating-point environment configuration
- Unsafe code designators (eg. turn of bounds checking for arrays)
- C-style preprocessor symbols 