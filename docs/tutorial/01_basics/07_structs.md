# Structs

**Table of Contents**

- [What is a Struct?](#struct-intro)
- [Spread Initialization](#spread-init)
- [Default Initializer](#default-init)
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

So then, how do we use this new data structure we have created?  We can start by
creating a `Vec2` using a **struct literal**:

    let u = Vec2{x=2, y=3};

Struct literals begin with the name of struct we want to create followed by a
pair of curly braces.  Inside these curly braces, we initialize each of the
fields in the struct.

We can leave some fields empty if we want them to just assume their null values.

    let v = Vec2{x=6};  // y = 0

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

## <a name="spread-init"/> Spread Initialization

TODO

## <a name="default-init"/> Default Initializers

TODO

## <a name="composite-structs"/> Composite Structs

TODO

## <a name="exercises"> Exercises

1. Write a struct for a 3D vector and implement a function to calculate the
   [cross product](https://en.wikipedia.org/wiki/Cross_product) between two 3D
   vectors.