# Structs

**Table of Contents**

- [What is a Struct?](#struct-intro)
- [Value Semantics](#value-semantics)
- [Spread Initialization](#spread-init)
- [Mull Values and Default Initializers](#default-init)
- [Composite Structs](#composite-structs)
- [Exercises](#exercises)

## <a name="struct-intro"/> What is a Struct?

A **struct** is a data structure that organizes its data into **fields** where
each field corresponds to a named data entry within the struct.  That is a bit
of a mouthful so let's make it a bit more concrete.

Suppose that we wanted to represent a vector comprised of 2 entries: `x` and `y`.
We could use 2 separate integers to denote this idea, but then our data would be
decoupled and would be hard to work with.  For example, consider a function which adds
to vectors.  How would you write one:

    func addVec2(x1, y2, x2, y2: i64) ? = ?

What do we return?  We can't return two values at once (yet).  And, look how
long the signature is.  Then consider creating functions to calculate the dot
product of two vectors or to generate the Cartesian unit vectors.  We would
quickly find that our "two separate integer" representation is very inconvenient
and difficult to reuse.

This is where structs come in.  They allow us to create a new type that couples the
two integer together so we can treat them as one value rather than two separate
values.

Structs are declared using the `struct` keyword followed by the name of the struct.

    // A new struct named `Vec2`.
    struct Vec2

Now, we need to define the values that comprise the struct.  We could simply use
a data structure that associated each value with a position.  But, this is not
very convenient as we would have to remember where in the data structure each
data item is.  

This is where the notion of fields comes in.  Fields allow us to give names to
the types that comprise our data structure.

Since vectors are comprised of two integer values, `x` and `y`, we will go ahead
and create two fields to represent that like so:

    struct Vec2 {
        x: i64;
        y: i64;
    }

The fields of our struct are enclosed in curly braces.  Each field entry begins with
its name followed by a type label.  We end the field entry with a semicolon.  

You might have already guessed, but we can make this declaration a little
shorter by combining the two parameters into one since they have same type (like
we did with parameters).

    struct Vec2 {
        x, y: i64;
    }

### Struct Literals

So then, how do we use this new data structure we have created?  We can start by
creating a `Vec2` using a **struct literal**:

    let u = Vec2{x=2, y=3};

Struct literals begin with the name of struct we want to create followed by a
pair of curly braces.  Inside these curly braces, we initialize each of the
fields in the struct.

We can leave some fields empty if we want them to just assume their null values.

    let v = Vec2{x=6};  // y = 0

We can also specify field values positionally like so:

    let r = Vec2{2, 3};

This is generally not recommended as it can make it difficult to determine what
fields are being specified.

### Accessing Fields with the Dot Operator

We can access these fields using the `.` operator followed by the name of the
field.

    Println(u.x, u.v);  // Prints `2 3`.

We can also mutate fields using the `.` operator on the left-side of assignment.

    v.y = 10;

    Println(v.x, v.y);  // Prints `6 10`.

So, now the we understand the basics of structs, let's revisit `addVec2` and use
our new struct to implement that function.  First, we need to know the type label
for our `Vec2`.  Luckily, it's really easy to remember: it's just `Vec2`.  Now,
we can write our function:

    func addVec2(a, b: Vec2) Vec2 = Vec2{x=a.x+b.x, y=a.y+b.y};

Now, we can call it just like any normal function:

    let w = addVec2(u, v);

    Println(w.x, w.y);  // Prints `8 13`.

## <a name="value-semantics"/> Value Semantics

Unlike some programming languages, structs in Chai obey strict
**value semantics**.  In essence, this means that structs behave like discrete
values rather than as references to data.  

To understand what this means, let's consider a simple example
using our `Vec2` struct from before.

    let u = Vec2{x=10, y=12}
    let v = u;

    v.x++;

    Println(u.x); 
    Println(v.x);

To some, the behavior here might seem obvious: the first `Println` call will
print `10` and the second will print `11`.  If you run this program, you will
see that this is indeed what happens.

However, if you are coming from a language like Java or Python, you might expect
both values to be `11`.  This assumption is incorrect.  The reason is value
semantics.

In this context, because structs obey value semantics, the struct `v` is explicitly
created as a copy of `u` rather than as a reference to the same data as `u`.  

This behavior is what we mean by "structs behave like values".  It is identical
to the situation where you create an integer variable using another integer variable's
value, you wouldn't expect changing one of the variables to affect the other's value.
This is because you are thinking of integers as values.  

Value semantics are critical but very intuitive part of Chai.  They are really
so simple that in another world this section wouldn't be needed.  Unfortunately,
many modern languages, especially object-oriented ones, have muddied the waters
by conflating structs with objects and introducing unpredictable reference
semantics.  So, this section is needed to avoid any confusion from programmers
coming from those languages.

## <a name="spread-init"/> Spread Initialization

Often, we want to be able to create structs inline from other structs we have already
created.  For example, consider we have the struct:

    struct User {
        id: i64;
        name: string;
        age: i32;
        email: string;
    }

What if we wanted to create a copy of that user with a different `id`?  Normally,
we would need to do something like:

    let userCopy = oldUser;
    userCopy.id = newID;

This is fine for this simple example but can be really tedious when we need to
do this with many structs at once or with many different fields.  

Luckily, Chai offers a more concise solution to this problem: 
**spread initialization**. Spread initialization allows us to create new struct
while populating all the fields we don't specify with the values from a
previous struct.  Using spread initialization, we can condense the previous two
lines of code into one:

    let userCopy = User{...oldUser, id=newID};

As you can see, the syntax is quite easy: simply write a regular struct literal
where the first entry is `...` followed by the struct we want to use for spread
initialization.

## <a name="default-init"/> Null Values and Default Initializers

Structs, like numbers, have a null value associated with them which represents
the default value of that struct.  For structs, the null value is always equivalent
to a struct literal with no specified fields.  For example, the null value for our
`Vec2` struct is always equivalent to `Vec2{}`.  

We can create a "null" struct using a simple variable declaration:

    let zero: Vec2;

We can also access the null value explicitly using the special value `null`.

    let zero: Vec2 = null;

However, you must be very careful when using this syntax as if you don't make
the type obvious to the compiler, you will get an error.

By default, all fields in a struct which are not explicitly initialized will be
given their null value.  For example, if we have a struct `v2` initialized like
so:

    let v2 = Vec2{x=2};

The `y` field of `v2` will default to the value zero.  

We can change the value a field defaults to using a **default initializer**.
Default initializers are specified in the struct definition as initializers on
the fields that we want to set a specific default value for. 

    struct User {
        // Default initializer for `id`.
        id: i64 = getNewUserID();
        
        name, email: string;
        age: i32;
    }

Notice that these initializers look just like variable initializers, but, in
constrast to variables, you must *always* specify the type of a struct field
even if you provide an initializer.

Now, when we create a new `User` struct, the `id` field will default to whatever
value is returned by `getNewUserID`.  

One very important thing to mention is that the initializers are run each time
the struct is initialized *NOT* once at the beginning of the program.  This
means that the `getNewUserID` function will be called *every time* the struct is
initialized without a specified value for `id`.

## <a name="composite-structs"/> Composite Structs

Structs can be created from other structs via mechanism known as
**struct composition**.  Such structs are called **composite structs**.

To create a composite struct, we first need "base" struct to build our
composite struct from:

    struct Entity {
        position: Vec2;
        dimensions: Vec2;
        health: i64 = 10;
    }

`Entity` will be our base struct.  We can then use our `Entity` struct to
create other more specific kinds of Entity using struct composition.

We can denote a composite struct by first placing a colon after its name
followed by the struct we want to act as its base.

    struct MovingEntity : Entity {
        velocity: Vec2;
    }

The struct `MovingEntity` is now a composite struct deriving from `Entity`. This
means that `MovingEntity` has *all* of the fields of `Entity` in addition to the
fields it explicitly declares.

    let me = MovingEntity{position=Vec2{2, 3}, velocity=Vec2{5, 0}};

    Println(me.position.x, me.velocity.x);

It is also possible to have multiple base structs by listing multiple bases after
the colon separated by commas:

    struct A {
        a: string;
    }

    struct B {
        b: i64;
    }

    // AB also has fields `a` and `b` from its parent structs.
    struct AB : A, B {
        c: f64;
        d: string;
    }

Composite structs are *not* interchangable for their parent structs: you can't use
a `MovingEntity` in place of an `Entity` or vice versa.  

> Much later on, we will look ways you can achieve traditional OOP-style
> "inheritance", where the child can be passed as an instance of the parent,
> using something called a class object.

## <a name="exercises"> Exercises

1. Write a struct for a 3D vector and implement a function to calculate the
   [cross product](https://en.wikipedia.org/wiki/Cross_product) between two 3D
   vectors.