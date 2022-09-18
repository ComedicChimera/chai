# Pointers

**Table of Contents**

- [What is a Pointer?](#pointer-basics)
- [The Stack and the Heap](#stack-and-heap)
- [Dynamic Allocation](#dynamic-allocation)
- [Constant Pointers](#const-pointers)
- [Nullability](#nullability)

## <a name="pointer-basics"> What is a Pointer?

A **pointer** is an indirect reference to a value in memory.  The value it
points to is called its **content**.  To understand what this really means, we
need to have a better understanding of how computer memory works.  

All programs we write have memory associated with them.  This memory is where
all the data the program uses, everything from variables to the machine code of
your program, is stored.  

This memory can be thought of as an array of sorts: there are many individual
slots in which data can be stored.  These slots are identified by an **address**
which is a unique number that tells you where in the array to find the data:
an index.  For sake of simplicity, we will conceptualize these slots
as holding a single byte.

> This is not necessarily accurate, especially on a modern machine; however, it
> is not such a gross misrepresentation that the meaning of this thought
> exercise will be lost.  In fact, for the most part, when I think about memory,
> I view it as a giant byte array.

To get a more concrete sense for what this looks like, let's imagine storing a simple
32-bit integer (`i32`) in memory.  For example, let's store the number `364`.  If
we took a snapshot of the memory storing our integer, it would appear something like this:

| **Address** | **Content** |
| ----------- | ----------- |
| 0xab452cf0 | 0x6c |
| 0xab452cf1 | 0x01 |
| 0xab452cf2 | 0x00 |
| 0xab452cf3 | 0x00 |

> The address above is entirely arbitraryc.  A 64-bit machine has 8-byte
> addresses: that is 16 hex characters to represent an address!  So, I went with
> 32-bit addresses for readability. 

The least significant byte of integer is stored at the address 0xab452cf0 and
the most significant bytes come after it.

> This encoding of "least significant byte first" is called Little Endian. The
> reverse byte ordering, called Big Endian, is also possible.  It entirely
> depends on your machine which one will be used.

This is where pointers come in.  

TODO

## <a name="stack-and-heap"> The Stack and the Heap

TODO

## <a name="dynamic-allocation"> Dynamic Allocation

TODO

## <a name="const-pointers"> Constant Pointers

TODO

## <a name="nullability"> Nullability

TODO
