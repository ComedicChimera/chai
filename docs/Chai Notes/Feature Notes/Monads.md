# Monads
Monads are principle feature of Chai although they are implemented somewhat unconventionally.  They act as the primary method of error handling.

In essence, a **monad** is a construct that "redefines" sequential execution for a given type.  For example, the `Option` monad is defined such that code progresses if there is a value and exits or accumulates if their is not.  

Monads in Chai are also viral (though they can be cleansed using an explicit match statement) meaning the propagate into the return types of functions and expressions when monadic operators are used.

In general, monad semantics are invoked through the use of monadic operators.

## The Monad Type Class
Most formally, in Chai, a type is a **monad** (or is **monadic**) if it implements the `Monad` type class.  This type class is defines as follows:

```
class Monad<T> M {
	pub func unit(u: T) M<T>;

	pub method eval() Option<T>;
	
	pub method fail<R>() M<R>;
}
```

Each method serves a unique purpose within Chai's monadic system.  

- the `unit` bound function is equivalent to the `return` function in Haskell: it creates a new Monad in its "unit" state.
- the `eval` method attempts to extract the stored value.  It returns a tuple indicating success or failure (*note*: maybe revise this to be an `Option` type?)
- the `fail` method is used to propagate a fail state.  Essentially, if monadic chaining fails at any point, the `fail` method returns an appropriate "fail state" monad with the given inner type: this allows failure information such as errors to propagate between monads of different value types.
- the `exit` method is called at the end of a "monadic context" to facilitate cleanup.

The astute functional programmer will notice that the most core "functional feature" of the monad is missing from the type class: where is *monadic bind*?  Chai's `Monad` type class emulates the behavior of a monadic bind while decoupling it from higher order functions.  While this is theoretically anti-thetical to the very definition of a monad, you will find that the behavior is ostensibly the same when using monadic operators.  However, this decoupling allows Chai to more easily generate efficient monadic code without unwrapping higher order functions.

## Monadic Operators

TODO: `<-`, `?`, `>>=`, `then`, `catch`

## Dominance and Virality
TODO

## Context Blocks
A **context block** defines a new monadic sub context.  These are useful when you have different kinds of monads within each other but don't want to propagate the monad out: it essentially acts to "contain" monad virality within a finite context.

Context blocks begin with the `with` keyword followed by an initial monad binding.

```
with v <- m {
	...
}
```

Inside the context block, the monad which is the type of `m` is the dominant monad.  
