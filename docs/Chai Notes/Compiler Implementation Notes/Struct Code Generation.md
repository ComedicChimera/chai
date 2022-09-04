# Struct Code Generation
This file outlines the rules used to determine how to generate the appropriate LLVM IR for structs.  The logic outlined herein may also be useful for the implementation of other "struct-like" objects such as arrays, enums, and class objects.

Notably, the real complexity of this problem arises from:
1. The constraints of the target system.
2. The goal of generating efficient target code.

While we do not (yet) want to overcomplicate things with the full scope of possible struct code
generation optimizations, we also do not want a completely naive code generation algorithm for something as essential as structs.

Generally, we break structs into two categories: **small** and **large**.  Small structs are $\le 2 \cdot \text{PointerSize}$.  Large structs are $> 2 \cdot \text{PointerSize}$. 

## Small Structs
### Type Representation
Small structs are represented directly as LLVM struct types with no wrapping.  For example, the small struct:

```
struct Vec2 {
	x, y: i32;
}
```

compiles to the LLVM type definition:

```llvm
%Vec2 = type {i32, i32}
```

### Variable Storage
Small structs are treated as simple values will stored in variables:

```llvm
; let s = Vec2{x=2, y=3};
%s = alloca %Vec2
store {i32 2, i32 3}, %Vec2* %s
```

### Field Access
We access the fields of small structs using `extractvalue` like so:

```llvm
; s.x
%value_s = load %Vec2, %Vec2* %s
%s.x = extractvalue %Vec2 %value_s, 0
```

This ends up being the same as just using GEP in all the experimenting that I did.  Since GEP is more difficult in this case, we just use `extractvalue`.  

### Use in Functions
We can pass small structs just as values to functions like so:

```llvm
; f(s)
%value_s = load %Vec2, %Vec2* %s
call void @f(%Vec2 %value_s)
```

They are also returned as values:

```llvm
; func zero() Vec2
define %Vec2 @zero()
```

And received as parameters like normal values:

```llvm
; func f(v: Vec2)
define void @f(%Vec2 %v) {
entry:
	%param_v = alloca %Vec2
	store %Vec2 %v, %Vec2* %param_v

	; ...

	ret void
}
```

## Large Structs
### Type Representation
Large structs compile to a corresponding struct type in LLVM in a similar manner to small structs.  For example,

```
struct Vec3 {
	x, y, z: i64;
}
```

will compile to:

```llvm
%Vec3 = type {i64, i64, i64}
```

However, the one wrinkle is that large structs are generally wrapped in a pointer when they are used.  In most cases, the type label `Vec3` will compile as `%Vec3*`.  We will see what ramifications this has later on.

### Variable Storage
When stored in a variable, large structs still mainain one level of indirection.  For example, the code:

```
let v3 = Vec3{x=1, y=2, z=3}
```

compiles as:

```llvm
%v3 = alloca %Vec3

; First field access optimization
%ptr_v3.x = bitcast %Vec3* to i64*
store i64 1, %i64* %ptr_v3.x

%ptr_v3.y = getelementptr inbounds %Vec3, %Vec3* %v3, i64 0, i32 1
store i64 2, %Vec3* %ptr_v3.y

%ptr_v3.z = getelementptr inbounds %Vec3, %Vec3* %v3, i64 0, i32 2
store i64 3, %Vec3* %ptr_v3.z
```

As you can see, the variable is still of type `%Vec3*` not `%Vec3**`.  However, this means that when storing a large struct into a large struct variable, we use a `memcpy` instead of a store.  For example, the code:

```
v3 = other_v3;
```

compiles as:

```llvm
%v3_i8ptr = bitcast %Vec3* %v3 to i8*
%other_v3_i8ptr = bitcast %Vec3* %other_v3 to i8*

call void @llvm.memcpy.p0i8.p0i8.i64(i8* %v3_i8ptr, i8* %other_v3_i8ptr, i64 24, i1 false)
```

### Field Access
For large structs, field accessing always compiles a GEP operation unless we are accessing the first field in which case we use a bitcast.

The code demonstrating creating a `Vec3` struct literal adequately demonstrates this.

### Use in Functions
Large structs are always passed as pointer values so when they are passed to functions, they must be copied.

```llvm
; f(v3)
%v3_copy = alloca %Vec3
%0 = bitcast %Vec3* %v3 to i8*
%1 = bitcast %Vec3* %v3_copy to i8*
call void @llvm.memcpy.p0i8.p0i8.i64(i8* %1, i8* %0, i64 24, i1 false)

call void @f(%Vec3* %v3_copy)
```

They are not copied; however, when received in as parameters: no stack allocation is needed.

Returning large structs is done by an "return pointer".  The caller allocates space for the returned struct on the stack and passes the pointer to it to the callee.  The callee then stores the returned struct into the passed in pointer.  The callee itself simply returns void.  

For example, the code below

```
func one() Vec3 {
	return Vec3{x=1, y=1, z=1};
}

func main() {
	let v3 = one();

	// -- snip --
}
```

compiles to:

```llvm
define void @one(%Vec3* sret(%Vec3) %return) {
entry:
	%return.x = bitcast Vec3* %return to i64*
	store i64 1, i64* %return.x

	%return.y = getelementptr inbounds %Vec3, %Vec3* %return, i64 0, i32 1
	store i64 1, i64* %return.y

	%return.z = getelementptr inbounds %Vec3, %Vec3* %return, i64 0, i32 2
	store i64 1, i64* %return.z

	ret void
}

define external void @main() {
entry:
	%v3 = alloca Vec3
	call void @one(%Vec3* %v3)

	; ...

	ret void
}
```

### Copy Elision
For large structs, a new mechanism comes into play to generate more efficient code: copy elision.  We have already seen it a few times.  In essence, it revolves around removing calls to `memcpy` where they aren't really needed.

Here are the rules:
1. When a large R-value struct is passed to a function as an argument, the copy is elided.
2. When a large R-value struct is stored into an L-value, the copy is elided.
3. If a large struct that is returned from a called function is stored directly into an L-value, then the destination pointer is used as its output pointer (no allocation and later copying needed).
4. When a large struct is returned from by the current function, the holder of the struct value returned is always replaced with the output variable (no copying into the output variable).



