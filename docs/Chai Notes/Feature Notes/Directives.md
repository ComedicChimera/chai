# Directives
**Directives** provide a way to control the compiler more precisely.  They are sort of like preprocessor directives in C (except Chai doesn't actually use a preprocessor).

A general rule for all directives is that they will only execute once per compilation, and they are always executed before any code is generated.

## Syntax
All directives begin with a `#` followed by the directive name (which is a normal) identifier.  Eg.

`#warn "deprecated"`

From there, directives deviate quite a bit in what syntax they use for their arguments.

## Locality
Directives can be defined as either **global** or **local**.  **Global directives** affect the entire file they are defined in whereas **local directives** only affect the function or relevant code block they are defined in.

The **locality** of a directive is determined entirely based on where the directive is placed.  Directives which are placed at the file level are considered global.  By contrast, directives which are placed inside of other definitions are considered local.

## Conditional Execution
The compiler is not guaranteed to execute a particular directive at compile time.  The following conditions are must be met for any directive to be executed:

1. The directive must be well formed.
2. The file must be parseable up until the directive.
3. No directive placed before the directive blocks compilation.

 For example, if the directive is placed after a `require` directive, then it will only execute if the `require` is satisfied.

### Local Directive Execution
In addition to the above conditions, local directives require that the file be fully compileable (ie. syntactically and semantically valid) in order to be executed.  Furthermore, they have special execution requirements depending on where they are defined.

##### Inside Functions (or Methods)
A directive placed inside a function will execute under the following conditions:

1. The function is called within user code.
2. If the directive is placed within generic control flow, the control flow branch containing the directive is evaluated at compile-time.

##### Inside Types (Class, Constraint, or Space) Definitions
A pragam placed inside a type definition will only execute if that type is used within user source code.

## Builtin Directives
### Require
The require directive specifies a condition of compilation.  The directive identifier is `require`.  It accepts a condition that if false will prevent compilation of all code which occurs after the require.

**Examples:**
```
#require false  # file will never compile

#require OS == "windows"  # conditional compilation based on OS

#require ARCH == "amd64"  # conditional compilation based on architecture
```

*Note*: More advanced conditions may be supported in the future

### C Import
The C source directive will allow for the direct inclusion of C source code in Chai files.  The directive identifier is `cimport`.  It accepts a string (can be multi-line) of C code to include. 

For more information and examples of this directive, see [[C Bindings]].

### Warn
The warn directive produces an arbitrary warning message.  The directive identifier is `warn`.  It accepts a string warning message.

**Examples:**
```
#warn "This package is deprecated."

func some_func() {
	#warn "This function is unsafe."
}
```

### Error
The error directive produces an arbitrary compilation error.  The directive identifier is `error`.  It accepts a string error message.

**Examples:**
```
#error "Unsupported platform."

func todo() {
	#error "Unimplemented function"
}
```

### Static Assert
The static assert directive runs a static assertion at compile-time.  The directive identifier is `static_assert`.  It accepts a condition and an optional error message.

**Examples:**
```
#static_assert ARCH != "arm"
```
### If/Elif/Else
The if/elif/else directive allows more precise conditional compilation than require.  The directive identifier is `if` and it accepts a condition followed by a code-block.  It can also have `elif` and `else` clauses.  Note that as with all other directives, the conditions must be compile-time constant.

```
#if OS == "windows" {
	// Handle Windows
} else {
	// Handle other OSs
}
```

### Inline Assembly.
The inline assembly directive will allow users to write assembly code inside their functions.  The directive identifier is `asm`.  It accepts an assembly string (can be muli-line) to run.  This directive can NOT be used at the file level.

## Future Directive Ideas
These directives may be implemented at some point in Chai's future.

- Floating-point environment configuration
- Unsafe code designators (eg. turn of bounds checking for arrays)
- C-style preprocessor symbols 
