# Classes
A **class** is an extension upon a defined type which adds some number of **associated functions** to that type.  They are similar to the OOP notion of a class with a few key differences namely:

1. Classes are extensions upon defined types not unique types themselves.
2. Classes do not have constructors or destructors.
3. The visibility of associated functions is set a package-level rather than a class level.
4. Classes cannot inherit from one another.

All class definitions begin with the `class` keyword followed by the name of the class and any type parameters the class accepts:

```
class MyClass

class List<T>
```

After this, the body of the class is enclosed in curly braces.  Class bodies are comprised of three things:

1. The base type.
2. Associated functions.
3. Class variables

All classes define a sort of "namespace" in which all of the associated functions and class variables live.  

## The Base Type
All classes have a **base type** which is the actual data the class is structured around.  This base type can be any defined type: either a struct or an enum.  The base type must be defined before any associated functions of the class in the class's primary definition.

The syntax for defining a base type is the same as the regular type definition except the name and type parameters of the type definition are elided: the name and type parameters are specified in the class definition.  Here are two examples of base type definitions:

```
class List<T> {
	struct {
		arr: []T;
		len: isize;
	}
}

class Option<T> {
	exposed enum =
		| Some(T)
		| None
		;
}
```

## The Associated Functions
There are three flavors of associated functions which are as follows:

1. Static associated functions called **static functions**.
2. Instance associated functions called **methods** or **instance functions**.
3. Operator associated functions called **class operators**.

### Static Functions
Static functions are functions which do not accept an instance of the class.  They act as "utility functions" defined on the namespace of the class.  For example, one common use of static function is to fulfill the role of a constructor, namely a static function called `New`.

Static functions are declared like regular functions within the class definition.

```
class List<T> {
	// -- snip --

	func New() List<T> = ...
}
```

They can then be called on the type as if it were a namespace:

```
let list = List.<i64>.New();
```

### Methods
Methods are functions which operate on an instance of the class.  They fulfill the role of "typical" OOP-style methods.  

They are declared like regular functions but instead of using the `func` keyword, they use the `method` keyword.

```
class List<T> {
	// -- snip --

	method Push(item: T) {
		// -- snip --
	}
}
```

Methods can access the instance they are operating on through the special value `self` which is visible to them implicitly.  `self` is always a pointer to the class itself.

```
class Vec2 {
	struct {
		x, y: i64;
	}

	func Zero() Vec2 = Vec2{x: 0, y: 0};

	method Magnitude() i64 = Sqrt(self.x ** 2 + self.y ** 2);
}
```

Methods can then be called on instances of the class using the `.` operator:

```
list.Push(10);
```

Note that the `.` operator will work even for pointers to the class:

```
let list = &[1, 2, 3];

list.Push(4);  // This works too.
```

### Class Operators
Class operators are like methods except they are identifier by an operator symbol rather than a name.  They are Chai's primary way of performing operator overloading.  

They are declared using the `oper` keyword followed by the operator symbol in parentheses and the signature of the operator.  They also have access to `self`.

```
class Vec2 {
	// -- snip --

	oper (+) (b: Vec2) Vec2 = Vec2{self.x + b.x, self.y + b.y};
}
```

They can be "called" by simply applying the operator in a manner matching the signateu of the class operator:

```
let a = Vec2{4, 5};
let b = Vec2{6, 7};

let c = a + b;  // Call to the `+` class operator.
```

Unlike named associated functions, multiple class operators can be defined for the same operator symbol provided they do not **conflict**.  Operators are said to conflict if they have the same number of non-variadic parameters which are all of equal types.

```
oper (+) (a: i64) i64 = ...

// CONFLICT!
oper (+) (b: i64) i64 = ...
```

In this way, classes can defined multiple **overloads** of the same operator.

There are also restrictions of the signatures of class operators: they must match the arity of the operator they are overloading.  For example, the `+` operator must always accept exactly one argument.  However, the `-` operator may either accept 0 arguments (unary `-`) or 1 argument (binary `-`)

## Class Variables
Class variables are variables whose namespace is constrained to a particular class: they are accessible just like static functions by applying the `.` operator the class name.

They are declared exactly like regular global variables just inside the class definition instead of the global namespace.  Note that both `const` and `comptime` variables can be used as class variables.

```
class User {
	let idCounter: i64 = 0;

	comptime tableName: string = "users";

	// -- snip --
}
```

They can be declared at any point in the class including before the base type.

## Visibility
The visibility semantics for classes work mostly the same for both the class type itself and the base type.  You can prefix the class definition with `pub` to export the class type, and if you have a struct base type, you can prefix its fields with `pub` to export them.  

Associated functions behave exactly like regular functions: prefixing them with `pub` will make them visibile outside of the class's package.  The same goes for class variables.

## Class Extensions
It is common for class implementations to span multiple files or even multiple packages.  In this case, you can use the `extend` syntax to declare extensions to associated functions of a class.  Note that class extensions may *not* contain a base type definition.

**a.chai:**
```
class Greeter {
	struct {
		name: string;
	}

	func New(name: string) Greeter = Greeter{name};
}
```

**b.chai:**
```
extend class Greeter {
	method Greet() string = "Hello, %s!".Format(self.name);
}
```

Note that exactly one partial class definition must contain the base type definition.

Classes may be extended from outside of their enclosing package using this mechanism as well.  Note that not all associated functions may be visible outside of the defining package.  In this case, you will only be call and implement [[Interfaces and Traits]] based on the functions that are visible to you.

