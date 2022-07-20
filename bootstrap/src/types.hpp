#ifndef _TYPES_H_
#define _TYPES_H_

#include <string>
#include <vector>
namespace chai {
    // Enumeration of different kinds of types.
    enum class TypeKind {
        UNIT,
        BOOL,
        INTEGER,
        FLOATING,
        POINTER,
        FUNCTION,
    };

    // Type represents a Chai data type.
    class Type {
    public:
        // kind returns type's kind.
        virtual TypeKind kind() const = 0;

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

        /* ------------------------------------------------------------------ */

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

        /* ------------------------------------------------------------------ */

        // internalCastFrom returns whether you can cast the type `other` to
        // this type. This method need only return whether a cast is strictly
        // possible (ie. it doesn't need to check if two types are equal) and
        // doesn't handle inner types.  It should not be called from outside the
        // implementation of `castFrom`.
        inline virtual bool internalCastFrom(Type* other) const { return false; }
    };

    /* -------------------------------------------------------------------------- */

    // UnitType represents a unit primitive type.
    class UnitType : public Type {
    public:
        inline TypeKind kind() const override { return TypeKind::UNIT; }

        inline size_t size() const override { return 1; }

        inline std::string repr() const override { return "()"; }

    protected:
        inline bool internalEquals(Type* other) const override { return other->kind() == TypeKind::UNIT; }
    };

    // BoolType represents a boolean primitive type.
    class BoolType : public Type {
    public:
        inline TypeKind kind() const override { return TypeKind::BOOL; }

        inline size_t size() const override { return 1; }

        inline std::string repr() const override { return "bool"; }

    protected:
        inline bool internalEquals(Type* other) const override { return other->kind() == TypeKind::BOOL; }

        inline bool internalCastFrom(Type* other) const override {
            // int to bool
            return other->kind() == TypeKind::INTEGER;
        }
    };

    // IntegerType represents an integer primitive type.
    class IntegerType : public Type {
        // The size of the integer type in bytes.
        size_t m_size;
    
    public:
        // Whether the integer type is unsigned.
        bool isUnsigned;

        // Creates a new integer type of the given size and signedness.
        IntegerType(size_t size, bool isUnsigned)
        : m_size(size)
        , isUnsigned(isUnsigned)
        {}

        inline TypeKind kind() const override { return TypeKind::INTEGER; }

        inline size_t size() const override { return m_size; }

        std::string repr() const override;

    protected:
        bool internalEquals(Type* other) const override;

        inline bool internalCastFrom(Type* other) const override {
            // int to bool, int to float
            return other->kind() == TypeKind::BOOL || other->kind() == TypeKind::FLOATING;
        }
    };

    // FloatingType represents a floating-point primitive type.
    class FloatingType : public Type {
        // The size of the float type in bytes.
        size_t m_size;

    public:
        // Creates a new floating-point type of the given size.
        FloatingType(size_t size)
        : m_size(size)
        {}

        inline TypeKind kind() const override { return TypeKind::FLOATING; }

        inline size_t size() const override { return m_size; }

        std::string repr() const override;

    protected:
        bool internalEquals(Type* other) const override;

        inline bool internalCastFrom(Type* other) const override {
            // float to int
            return other->kind() == TypeKind::INTEGER;
        }
    };

    // PointerType represents a pointer type.
    class PointerType : public Type {
        // The element type of the pointer.
        std::unique_ptr<Type> m_elemType;

    public:
        // The size of a pointer on the desired target.
        static size_t pointerSize;

        // Whether the pointer is constant.
        bool isConst;

        // Creates a new pointer type to the `elemType` with `isConst` constancy.
        PointerType(Type* elemType, bool isConst)
        : m_elemType(elemType)
        , isConst(isConst)
        {}

        // elemType returns a pointer to the element type of the pointer.
        inline Type* elemType() const { return m_elemType.get(); }

        inline TypeKind kind() const override { return TypeKind::POINTER; }

        inline size_t size() const override { return pointerSize; }

        std::string repr() const override;

    protected:
        bool internalEquals(Type* other) const override;

        inline bool internalCastFrom(Type* other) const override {
            // TODO: remove bitcasting later
            return other->kind() == TypeKind::POINTER;
        }
    };

    // FunctionType represents a function type.
    class FunctionType : public Type {
        // The return type of the function.
        std::unique_ptr<Type> m_returnType;

    public:
        // The parameter types of the function.
        std::vector<std::unique_ptr<Type>> paramTypes;

        // Creates a new function with the parameter types `paramTypes` and
        // return type `returnType`.
        FunctionType(std::vector<std::unique_ptr<Type>>&& paramTypes, Type* returnType)
        : paramTypes(paramTypes)
        , m_returnType(returnType)
        {}

        // returnType returns a pointer to the function's return type.
        inline Type* returnType() const { return m_returnType.get(); }

        inline TypeKind kind() const override { return TypeKind::FUNCTION; }

        inline size_t size() const override { return PointerType::pointerSize; }

        std::string repr() const override;

    protected:
        bool internalEquals(Type* other) const override;
    };
}

#endif