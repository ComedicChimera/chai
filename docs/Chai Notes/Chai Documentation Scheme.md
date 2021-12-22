## Idea
We need to find a way to teach and document Chai in an efficient and engaging manner.  A definitive reference of some form is an absolute must, but such a reference, unless presented in a very organized and approachable manner, will be hard for beginners.

## Proposal
We have two pieces of documentation:
1. **The Book**: The book will be the definitive reference for the Chai programming language.  This will be the primary source of documentation until the Tour (discussed next) is completed.
2. **The Tour**: This will be an interactive, online walkthrough (built around the playground) that gives a basic "tutorial introduction" to Chai and its features.  It will essentially be a series of panels documenting different features paired with an exercise to help the reader practice their skills.

## Book Layout
The book will be layed out in a series of chapters which are made up of sections.  This book will be hosted digitally on [chai-lang.dev](https://chai-lang.dev).  The layout of this book will be as follows:

### Introduction
The purpose of this small chapter is to introduce the premise of the book.  This section must mention that this book is written with the *standard Chai compiler* in mind (`chaic`).

This should mention the impetus for the design of the language as well as provide a formal introduction to the notation scheme used in the book.

### Chapters

**Lexical Structure**
Specifies in detail the lexical semantics Chai including: what different lexemes are, how the language is represented, the usage and placement of newlines and split-joins, etc.

**Program Structure**
Describes the high level structure of Chai project, beginning with modules and working all the way down to expressions.  This places each element of the language in the context in which it can be used.

**Type System**
Enumerates built-in types; discusses type equality and equivalency, nullability, and type casting; enumerates all valid casts.

**Numeric Expressions**
Describes numeric literals and basic numeric operations (floating-point and integral); covers all arithmetic expressions as well as bitwise operators and their usage on integral types; discusses complex and rational numbers and their associated expressions.

**Strings**
Describes string literals and string operations (indexing, slicing, etc); specifies string representation and UTF-8 encoding; explains the rune type and Unicode-point representation.

**Block Expressions**
Describes the high level structure of procedural expressions (ie. do blocks); discusses variables, assignment, and unary update statements.

**Conditional Logic**
Describes the boolean type, comparison operators, and logical operator; discusses the if/elif/else statement, exhaustivity, type consistency, and the nothing type.

**Pattern Matching**
Describes tuples, tuple unpacking, and tuple indexing; describes the construction of patterns; discusses the match and test-match expressions; expands upon array, list, and dictionary shape patterns.

**Sequences and Iterators**
Describes lists, dictionaries, and arrays in full detail; introduces indexing, slicing, and reference operators; includes a digression on value semantics; discusses the construction and function of iterators (including those on strings); discusses loops, generators, loop control flow statements, and after blocks.

**Functions**
Describes top-level defined functions, arguments, and function calling in detail; discusses the return statement and variadic arguments; discusses the `main` and `init` functions. 

**References**
Describes references, referencing, and dereferencing; expands upon heap allocation and garbage collection; explains pass-by-reference.

**Defined Types**
Describes structured types, algebraic types (including structured variants and enumeration specifying: eg. `| Variant = 2`), hybrid types, and type aliases; discusses type inheritance.

**Generics**
Discusses unconstrained and constrained type generics, size generics, and variadic generics and their relation to variadic arguments; explains generic type inference and type specifiers; discusses type unions and their usage as constraints; describes generic control flow constructs and typing rules; describes generic for loops (`for from`).

**Method Spaces**
Discusses method spaces and method calling; describes type classes, class implementation, virtual methods, and type class constraints.

**Lambdas**
Discusses lambdas, closures, function types, and partial application.

**Operators**
Describes operator definitions and overloading; discusses partial operator application.

**Monads**
Discusses the `Option` and `Result` types; describes the `Monad` type class; enumerates all monadic operators, their behaviors, and their relation to program control flow.

**Packages and Modules**
Discusses packages and the global namespace; discusses the import statement; explains import resolution; discusses modules; specifies the module schema.

**Concurrency**
Discusses strands, futures, and awaiting; introduces channels and guards; discusses strand coordination; introduces the event system; discusses volatility.

**Meta Directives**
Enumerates the various meta directives in the language and their semantics; explains the placement of such directives.

**System Interfacing**
Discusses C binding, relevant annotations, and elaborates upon the meta directives for C binding; explains how to link to static libraries, dynamic libraries, and other object files; discusses how to generate DLLs and static libraries; introduces inline assembly; discusses revelant interactions with the Chai runtime.

### Appendices
The appendices of the book will contain extra information that is necessary for purposes of completeness but mostly boils down to large lists and tables.  The appendices will proceed as follows:

**Appendix A: Operators**
Provides a definitive list of built-in operators and operator precedences.

**Appendix B: Annotations**
Enumerates all defined annotations and their purposes.  It also mentions which annotations allow functions to elide their body.

**Appendix C: Intrinsic Functions**
Enumerates all intrinsically-defined functions -- not including operators.

**Appendix D: Language Grammar**
Provides a formal language grammar for Chai marked up in EBNF with all exceptions to its formal rules (ie. whitespace) accounted for.




