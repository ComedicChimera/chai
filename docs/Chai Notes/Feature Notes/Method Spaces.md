# Method Spaces
In Chai, a **method space** (or simply a *space*) is a specialized namespace bound to a specific type containing functions that are related to that type known as **methods**.  All types have a function space in which methods can be defined.  

You can access the space of a type using the `space` keyword followed by the type label for the parent type of the space.  This creates a block which is concluded by `end`.  Inside this block, you can define methods.

```
space List[string]
	# methods go here
end
```

The methods definitions are placed inside the space block.  Methods come in two varieties:
- **Static** methods
- **Instance** methods or **modes** 

It should be noted that each individual "space block" is called a **(method) subspace** of the type.

Methods are uniquely defined within their respective spaces: two methods cannot be defined with the same name in the same space.  However, two methods can be defined with the same name in different spaces (eg. the `new` method is conventionally used to create a new instance of a given type)

## Static Methods
The first kind of method is defined simply as a normal function inside the subspace.

```
space List[string]
	def new() List[string] = ...
end
```

The static method can then be called using the `.` operator on the type name.

```List.[string].new()```

**Aside**: Generic inference allows the call above to shortened to `List.new()` in most cases.

The idea is that these method are confined within their namespaces of their respective types.  Logically, this generally means that they are specifically defined for that type.  For example, a function `new` on its own is very non-specific: it could refer to almost anything.  By contrast, when `new` is defined as a static method of `List[string]`, it takes on the meaning of creating a new list of strings.  

## Instance Methods (Modes)
Instance methods are specialized methods which operate on a specific instance of a given type.  These methods are also knowns as the *modes* of the type.  These methods use a special syntax (with the `mode` keyword):

```
space List[string]
	mode push(elem: string) = ...
end
```

Modes can only be called on specific instances of list.  This is done using the `.` operator.  For example,

```
let list = ["abc", "def"]

# invocation of the `push` mode using the
list.push("ghi")
```

Modes can access the instance they are being called on using the the `this` keyword.

```
space List[string]
	mode push(elem: string)
		this.grow(this.len() + 1)
		this.buff[this.len()] = elem
		this.length++
	end
end
```

Notice that in addition to manipulating values of the type, modes can also call other modes of the type.  Also, the `this` reference provides mutable access to properties (as shown above); however, it is not possible to assign directly to the `this` reference.

## Visibility
By default, all the methods of a subspace are private to the package they are defined in.  However, a method can be made public by prefixing it with the `pub` keyword.

```
space List[string]
	pub def new() List[string] = ...
	
	pub mode push(elem: string)
		...
	end
end
```

You can also prefix a space block with the `pub` keyword which makes all methods defined in it public.

```
pub space List[string]
	def new() List[string] = ...
	
	mode push(elem: string)
		...
	end
end
```

Notably, methods still must be unique within their global spaces even if they are not technically visible in all packages.  This is to prevent confusion: multiple methods with different behaviors defined across different packages makes for a horrible mess of a codebase.

## Generic Subspaces
Subspaces can also be defined generically using a generic space block.  This is done like so:

```
space[T] List[T]
	...
end
```

Notice the brackets after the `space` keyword contain the actual type parameters.  The methods defined in this subspace will then be defined for all instances of the generic type list.

In fact, methods are bound to a given type via type pattern matching: the header type of the subspace is used as a pattern which the compiler will match against to see which methods are available.

For example, the following would bind methods onto all numeric types.

```
space[T: Num] T
	...
end
```

Of course, the type parameter(s) are accessible as types within generic subspaces.

It should be mentioned that generic subspaces can make maintaining uniqueness more challenging.  For example,

```
space List[string]
	def new() List[string]
end

space[T] List[T]
	def new() List[T]
end
```

The latter space causes a uniqueness error since technically it redefines `new` in the space of `List[string]`.  This is not as readily apparent since the subspace headers are not the same.

This can be particularly troublesome if the spaces are defined across multiple packages. Unfortunately, there is no way to completely avoid this issue without compromising the flexibility of method spaces.  The best coping mechanism for this kind of complexity is to provide intuitive and detailed error messages when these uniqueness concerns arise.

## Static Methods of Non-Named Types
One advantage of method spaces is that they can be defined on *any* type not just named types (or "classes" as in some languages such as Java). 

In this case of modes, this enhancement is obvious and easy to implement.  For example, strings has a number of useful modes defined for them such as `.upper` or `.split`.  

However, this becomes challenging in the case of static methods for all "non-named" types.   For example, consider you had a static method for integers that would return its maximum value.  How would you call such a static method in normal code?  For named types, we can simply use their name (eg. the `new` method):

```List.new()```

However, for a type like `i32` whose name is a keyword, this becomes more difficult.  It becomes far more difficult when we consider more complex non-named types such as functions or tuples.  Clearly, static methods for these types provide a unique syntactic challenge.

*Below are some proposed solutions to this problem:*

Perhaps, a unique syntax like `space[i32]` could allow you to access the static methods?  They main drawback is verbosity which somewhat defeats the purpose of such static methods.

Another proposition is simply to disallow static methods for all non-named types.  This seems a bit short-sighted but is likely not to cause problems for most users of the language.


