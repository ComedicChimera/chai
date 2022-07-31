# The Chai Type Deduction Algorithm

Here follows a somewhat complete description of Chai's type deduction algorithm.
This algorithm is only proven sound and complete in so far as it has stood
the test of time as Chai's type deduction algorithm and repeatedly arrived
at the correct conclusions in a variety of different cases.

This algorithm is complex.  Some might dare to call it convoluted.  However,
the problem it solves is equally complex and convoluted so I consider this
algorithm a success.  It could do with some optimization though.

## The General Approach

Chai's type deduction algorithm is loosely based on the Algorithm J for the
Hindley-Milner type system.  However, it has been extended and redesigned to
support generalized type overloading.

The algorithm works by considering a **solution graph** comprised of **type
variable nodes** and **substitution nodes**.  The type variable nodes represent
the undetermined types of the given solution context (eg. inside a given
function).  The substitution nodes represent the possible values for those
undetermined type variables.  Each type variable node has a number of
substitution nodes associated with it that represent the possible types this
node can have.  

These substitution nodes can be determined either before unification involving
that type variable begins or during the unifications involving the type
variable: type variable nodes which are **complete** can no longer has
substitutions added to them.

All substitution nodes have **edges** which represent relationships between
substitution nodes.  More precisely, if two substitution nodes, A and B, share
an edge, then substitution A is valid if and only if substitution B is valid.

The principle mechanism of Chai's type deduction algorithm is **unification**:
the process by which the solver attempts to make two types equal by unification.
This process fails if the two types cannot be made equal (eg. are of a different
shape or represent different primitive types).  The wrinkle to this algorithm
that Chai's type solver adds (extending from Algorithm J) is that the
unification algorithm also takes into account the substitution that led to the
unification: eg. if we are unifying a type against a substitution of a type
variable, different behavior may occur.  This substitution node is called the
**unification root** and may not be present if the unification is occurring at
the top level (ie. between two types which are not currently substitutions of
any type variable).  The exact formulation of the unification algorithm can be
inferred from the implementation below using the context and terminology
established above.