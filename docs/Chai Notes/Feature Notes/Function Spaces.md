# Function Spaces
In Chai, a **function space** (or simply a **space**) is a specialized namespace bound to a specific type containing functions that are related to that type known as **bound functions** (often just referred to as the *functions* of a given space).  All types have a function space in which functions can be defined.  

You can access the space of a type using the `space` keyword followed by the type label for the parent type of the space.  This creates a block which is concluded by `end`.  Inside this block, you can define methods.

```
space List<string> {
	# functions go here
}
```

The function definitions are placed inside the space block.  There are two kinds of functions that can be defined inside a space:
- **Instance Functions**, more commonly called **Methods**.
- **Static Functions**, also called **Associated Functions**.

It should be noted that each individual "space block" is called a **(function) subspace** of the type.

Functions are uniquely defined within their respective spaces: two functions cannot be defined with the same name in the same space.  However, two functions can be defined with the same name in different spaces (eg. the `new` function is conventionally used to create a new instance of a given type)

## Static Functions
The first kind of bound function is defined simply as a normal function inside the subspace.

```
space List<string> {
	func new() List<string> = ...
}
```

The static function can then be called using the `.` operator on the type name.

```List.<string>.new()```

**Aside**: Generic inference allows the call above to shortened to `List.new()` in most cases.

The idea is that these functions are confined within their namespaces of their respective types.  Logically, this generally means that they are specifically defined for that type.  For example, a function `new` on its own is very non-specific: it could refer to almost anything.  By contrast, when `new` is defined as a static function of `List<string>`, it takes on the meaning of creating a new list of strings.  

## Instance Functions (Methods)
Methods are specialized functions which operate on a specific instance of a given type.  Methods use a special syntax (with the `method` keyword):

```
space List<string> {
	method push(elem: string) = ...
}
```

Methods can only be called on specific instances of list.  This is done using the `.` operator.  For example,

```
let list = ["abc", "def"];

# invocation of the `push` mode using the
list.push("ghi");
```

Methods can access the instance they are being called on using the the `self` keyword.

```
space List<string> {
	method push(elem: string) {
		self.grow(self.length + 1)
		self.buff[self.length] = elem
		self.length++
	}
}
```

Notice that in addition to manipulating values of the type, methods can also call other methods of the type.  Also, the `self` reference provides mutable access to properties (as shown above); however, it is not possible to assign directly to the `self` reference.

## Visibility
By default, all the bound functions of a subspace are private to the package they are defined in.  However, a bound function can be made public by prefixing it with the `pub` keyword.

```
space List<string> {
	pub func new() List<string> = ...
	
	pub method push(elem: string) {
		...
	}
}
```

You can also prefix a space block with the `pub` keyword which makes all functions defined in it public.

```
pub space List<string> {
	func new() List<string> = ...
	
	method push(elem: string) {
		...
	}
}
```

Notably, bound functions still must be unique within their global spaces even if they are not technically visible in all packages.  This is to prevent confusion: multiple bound functions with different behaviors defined across different packages makes for a horrible mess of a codebase.

## Generic Subspaces
Subspaces can also be defined generically using a generic space block.  This is done like so:

```
space<T> List<T> {
	...
}
```

Notice the angle brackets after the `space` keyword contain the actual type parameters.  The functions defined in this subspace will then be defined for all instances of the generic type list.

In fact, bound functions are associated with a given type via type pattern matching: the header type of the subspace is used as a pattern which the compiler will match against to see which bound functions are available.

For example, the following would bind functions onto all numeric types.

```
space<T: Num> T {
	...
}
```

Of course, the type parameter(s) are accessible as types within generic subspaces.

It should be mentioned that generic subspaces can make maintaining uniqueness more challenging.  For example,

```
space List<string> {
	func new() List<string>
}

space<T> List<T> {
	func new() List<T>
}
```

The latter space causes a uniqueness error since technically it redefines `new` in the space of `List<string>`.  This is not as readily apparent since the subspace headers are not the same.

This can be particularly troublesome if the spaces are defined across multiple packages. Unfortunately, there is no way to completely avoid this issue without compromising the flexibility of function spaces.  The best coping mechanism for this kind of complexity is to provide intuitive and detailed error messages when these uniqueness concerns arise.

## Importing Function Spaces
By default, when you import a package containing subspace definitions of a function space, those bound functions are also imported.

For example, this import brings in all the string utility methods provided by `stringutil`.

```
import stringutil
```

This is true even in cases where only single types are imported (even if the bound functions are not actually usable).

```
import File from io.fs
```

This import logic is acceptable because of global uniqueness: it is impossible for there to be a conflict between bound function names so we can just import them freely.  This is convenient for the end user and less effort for the compiler designer.