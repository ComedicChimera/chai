#ifndef _TYPES_H_
#define _TYPES_H_

#include "chaic.hpp"

namespace chai {
    // UnitType represents a unit primitive type.
    class UnitType : public Type {
        // Private constructor used by TypeStore.
        friend class TypeStore;
        UnitType() {}

    public:
        inline TypeKind kind() const override { return TypeKind::UNIT; }

        inline size_t size() const override { return 1; }

        inline std::string repr() const override { return "()"; }

    protected:
        inline bool internalEquals(Type* other) const override { return other->kind() == TypeKind::UNIT; }
    };

    // BoolType represents a boolean primitive type.
    class BoolType : public Type {
        // Private constructor used by TypeStore.
        friend class TypeStore;
        BoolType() {}

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

        // Private constructor used by TypeStore.
        friend class TypeStore;

        // Creates a new integer type of the given size and signedness.
        IntegerType(size_t size, bool isUnsigned)
        : m_size(size)
        , isUnsigned(isUnsigned)
        {}

    
    public:
        // Whether the integer type is unsigned.
        bool isUnsigned;

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

        // Private constructor used by TypeStore.
        friend class TypeStore;

        // Creates a new floating-point type of the given size.
        FloatingType(size_t size)
        : m_size(size)
        {}

    public:
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
        // Private constructor used by TypeStore.
        friend class TypeStore;

        // Creates a new pointer type to the `elemType` with `isConst` constancy.
        PointerType(Type* elemType, bool isConst)
        : elemType(elemType)
        , isConst(isConst)
        {}

    public:
        // The size of a pointer on the desired target.
        static size_t pointerSize;

        // The element type of the pointer.
        Type* elemType;

        // Whether the pointer is constant.
        bool isConst;

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

    inline size_t PointerType::pointerSize;

    // FunctionType represents a function type.
    class FunctionType : public Type {
    public:
        // The parameter types of the function.
        std::vector<Type*> paramTypes;
        
        // returnType returns a pointer to the function's return type.
        Type* returnType;

        // Creates a new function with the parameter types `paramTypes` and
        // return type `returnType`.
        FunctionType(std::vector<Type*>& paramTypes, Type* returnType)
        : paramTypes(std::move(paramTypes))
        , returnType(returnType)
        {}

        inline TypeKind kind() const override { return TypeKind::FUNCTION; }

        inline size_t size() const override { return PointerType::pointerSize; }

        std::string repr() const override;

    protected:
        bool internalEquals(Type* other) const override;
    };
}

#endif