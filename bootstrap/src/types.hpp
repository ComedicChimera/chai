#ifndef _TYPES_H_
#define _TYPES_H_

// Type represents a Chai data type.
class Type {
public:
    // size returns the size of the type in bytes.
    virtual size_t size() const = 0;

    // repr returns the string represents of the type.
    virtual std::string repr() const = 0;

    /* ---------------------------------------------------------------------- */

    // innerType returns the "inner type" of self.  For most types, this is just
    // an identity function; however, for types such as untyped and aliases
    // which essentially just wrap other types, this method unwraps them.
    inline virtual Type* innerType() { return this; }

    // align returns the alignment of the type in bytes.
    inline virtual size_t align() const { return size(); }

    // isNullable returns whether the type is nullable.
    inline virtual bool isNullable() const { return true; }

    /* ---------------------------------------------------------------------- */

    // equals returns whether this type is equal the type `other`.
    inline bool equals(Type* other) { 
        return innerType()->internalEquals(other->innerType());
    }

    // castFrom returns whether the type `other` can be cast to this type.
    inline bool castFrom(Type* other) {
        return innerType()->internalCastFrom(other->innerType());
    }

    // bitSize returns the size of the type in bits.
    inline size_t bitSize() const { return size() * 8; }

    // bitAlign returns the alignment of the type in bits.
    inline size_t bitAlign() const { return align() * 8; }

protected:
    // internalEquals returns whether this type is equal to the type `other`.
    // This method doesn't handle inner types and should not be called from
    // outside the implementation of `equals`.
    virtual bool internalEquals(Type* other) const = 0;

    // internalCastFrom returns whether you can cast the type `other` to this
    // type. This method need only return whether a cast is strictly possible
    // and doesn't handle inner types.  It should not be called from outside the
    // implementation of `castFrom`.
    virtual bool internalCastFrom(Type* other) const = 0;
};

#endif