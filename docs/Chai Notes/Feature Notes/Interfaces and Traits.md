# Interfaces and Traits
Interfaces and traits are the two key mechanisms of class-based polymorphism.

Interfaces act as runtime polymorphism over classes, and traits act as compile-time polymorphism over classes.  Both have specific uses.

## Interfaces
An **interface** is a data type that acts a polymorphic interface to the associated functions of a class.  

Interface types are created by wrapping a class into an interface instance in a process called **boxing***.  This wrapped type is known as the **underlying type** of the interface.  

Interface polymorphism happens at runtime: given any arbitrary interface type, it is not possible to determine what its underlying type.  However, runtime mechanisms can be used to test the underlying type of the interface.  Furthermore, interface instances can be cast to their underlying type in a process called **unboxing** or **downcasting**.  It should be noted that attempting to cast an interface to a type which is not its true underlying type will result in a runtime panic: interface unboxing can fail at a runtime.  

### Interface Declarations
Interfaces are declared using the `interf` keyword followed by the name and type parameters of the interface:

```
interf Iter<T> 
```

After this, the body of the interface is defined inside curly braces.  This body is composed of associated function definitions which come in two flavors:

1. **Abstract**: they don't have a body and must be implementing by a deriving class.
2. **Virtual**: they do have a body and may be overridden by a deriving class.

Here is an example of an abstract method:

```
interf Greeter {
	method Greet() string;
}
```

Here is an example of a virtual class operator:

```
interf Collection<T> {
	// -- snip --

	oper ([]) (ndx: isize) T = self.Get(ndx);
}
```

### Implementing Interfaces
A class which **implements** an interface is then considered a subtype of that interface type and may be implicitly coerced into it.  This is the singular context in which implicit coercion occurs in Chai.  However, it must be noted that this implicit coercion can *only* happen between a *pointer to the deriving class type* and the interface *not* between the raw class type and the interface.

The criterion for implementation are two fold:

1. The class must declare that implements the interface.
2. The class must provide implementations of all abstract functions of an interface.

Notice that interfaces in Chai are *not* duck-typed: they must be implemented explicitly.  To declare that a class implements an interface, an `is` clause is included in the class definition like so:

```
// `is` clause: is Seq<T>`
class List<T> is Collection<T> {
	// -- snip --

	// Iter() method implementation:
	method Iter() Iter<T> = &ListIter{/* -- snip -- */};
}
```

A given class can implement multiple interfaces by using a comma-separated list of interfaces in its `is` clause.

### Interface Inheritance
Interfaces can inherit from other interfaces in which case they gain all of the abstract and virtual functions of the interfaces they inherit from.  Like with implementation, interface inheritance is denoted using an `is` clause in the definition.

```
interf MutCollection<T> is Collection<T> {
	// Addition abstract function:
	method Set(ndx: isize, item: T);
}
```

Interfaces can inherit from multiple parent interfaces.

Interfaces can also "redefine" any of the functions of their parent interfaces.  Such redefinitions may not change the signature of the functions but may convert the function between abstract and virtual status.  That is to say, a deriving interface can make the abstract functions of its parent(s) virtual by providing implementations or make the virtual functions of its parent(s) abstract by redefining without an implementation.

## Traits
A **trait** is a constraint that is satisfied only when all the functions of the trait are contained in the class of its implementing type.

Although they sound quite different, traits and interfaces are similar constructs with a couple key differences:

1. Traits use the keyword `trait` instead of `interf`.
2. Traits are fully evaluated at compile-time rather than runtime.
3. Traits are type constraints rather than as types.
4. Traits constrain the input types of associated functions monomorphically.

That last point is perhaps the most important.  Traits are constructed in such a way that as opposed to having a single polymorphic form like interface, they are reduced into numerous monomorphic forms via generic substitution at compile-time.  This seems abstract but has a very concrete meaning.

### Trait Declarations
Traits are defined using the `trait` keyword followed by the name of the trait and its type parameters.  This is the same as with interfaces.  However, before we begin the body, traits also define a **trait parameter** which is a placeholder type parameter in which the implementing type will be substituted.

The rest of the trait body is the same as that of an interface except the trait parameter can be used as a type throughout the trait.

```
// A is our trait parameter.
trait Add A {
	// Signature: (A, A) -> A which is what we want to express.
	oper (+) (other: A) A;
}
```

Notice that trait parameters give us some extra flexibility in describing the associated functions of the deriving class: we can make much more granular assertions about the signatures of their associated functions.  In the above example, instead of just saying "`Add` means you can take any two `Add` and add them to create a new `Add`", we instead assert that "`Add` means you can take two of the same `Add` and add them to produce another of the same `Add`" which has a very different meaning.

### Implementing Traits
Traits are implemented identically to interfaces: their implementation is explicit and based entirely on whether the sets of functions match up.  They even use an `is` clause like interfaces.

```
// `is Add` => trait implementation.
class Vec2 is Add {
	// -- snip --

	// Notice that we replace `A` with its concrete value here.
	oper (+) (other: Vec2) Vec2 = Vec2{self.x+other.x, self.y+other.y};
}
```

Note that you can also implement multiple traits just like interfaces, and you can even implement traits and interfaces simultaneously!

### Using Traits
Traits are used a type constraints *not* types.  So, to make use of them, we need to use generics.  Here is a quick example:

```
func Sum<S: Seq, A: Add>(seq: S<A>) A {
	let it = seq.Iter();

	let sum = it.Next();

	for item in it {
		sum += item;
	}

	return sum;
}
```

Note that the use of traits is *necessary* to gain access to associated functions inside generics: you can't use the `+` operator on a generic type parameter without explicitly constraining it to be `Add` (or some trait which inherits `Add`) first.

### Trait Inheritance
Traits can also inherit from one another like interfaces.  They have the same notions of abstract and virtual functions as interfaces.  They also inherit via explicit `is` clauses just like interfaces.

The only real difference is how the trait parameters are handled: when a trait derives from another trait, all the trait functions it inherits have their trait parameter replaced by the trait parameter of the deriving trait inside the deriving trait.  The behavior is obvious, but important to mention.

The consequence of them above is that you can inherited functions from virtual functions in the deriving trait seamlessly as you can with interfaces: no casting/generics required.

## Extensions
Both traits and interfaces can have addition *virtual functions* added to them via **extensions**.  This is a way of adding additional functionality (eg. the `sequtil` module) without having to define a whole new trait or interface.

It is important to note that extensions can *only* contain virtual functions.  Additional abstract functions can *never* be added outside of the initial trait or interface declaration.

Extensions use the keyword `extend` followed by the interface or trait declaration header followed by a body containing some number of virtual functions to add.  Here is an example of extending a trait and extending an interface:

```
extend trait Trait T {
	func ExtendedFunc() T = ...
}

extend interf Interf {
	func ExtendedFunc() Interf = ...
}
```

Extensions can happen in other packages as well, but in that case must be imported.

## Visibility
Visibility for traits and interfaces works the same as for classes.  It is worth mentioning though that if a trait or interface contains any abstract functions with are not public, then no class outside the package defining the trait or interface will be able to implement the trait or interface.

Furthermore, the visibility of implemented functions must match up: if a function is public in the trait or interface, then it must be public in the deriving class and vice versa.  This principle also applies to overriding in trait/interface inheritance. 

## Key Idea: Prefer Traits to Interfaces
Traits to describe common behavior.  Interfaces to encapsulate/anonymize common behavior.

Traits tend to have significantly better performance than interfaces so where it is possible, use traits instead of interfaces.  That being said, there are some very common and important situations in which the limitations of traits make it better to use interfaces instead.  Similarly, there are also many situations in which the limitations of interfaces make traits more practical.

Generally, interfaces are used when you want to "merge" all the children into a single notion: the children are no longer individuals but rather abstracted into one idea.  For example, if you want to store many different kinds of the children in the same list, then you probably want to use an interface.  Conversely, if you just want a way of grouping/sharing common behavior, then traits probably make more sense.

As an example, look at the differences between `Seq<T>` and `Iter<T>`.  `Seq<T>` is generally just a way to group things you can iterate over: you rarely have a need to "anonymize" different sequences.  It is a pattern of behavior rather than an amorphous concept.  So, we use a trait for `Seq<T>`.  Conversely, we almost never want to consider the individual kinds of `Iter<T>`, and we want to "merge" all iterators into one common type since they fundamentally all do the same thing with no "unique" functionality between them.  So, we use an interface for `Iter<T>`.

Ultimately, they are tools in your toolbox to use in the way that makes the most sense.  Neither is "better" than the other.  

