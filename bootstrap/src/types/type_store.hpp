#ifndef _TYPE_STORE_H_
#define _TYPE_STORE_H_

#include <cstdint>

#include "types.hpp"

namespace chai {
    // TypeStore is the global memory manager for types: generally, types cannot
    // easily be assigned to a single owner, and because they are so prolific,
    // trying to use a shared_ptr tends to be a pain.  Plus, primitive types
    // don't change between source files and occur so often that it makes sense
    // to create one global instance of each and just reuse them.  All othet
    // types are kept allocated in a global vector because they tend to occur
    // across multiple packages and need to persist for the full-lifetime of the
    // dependency graph.  So, it makes sense to just keep them globally
    // allocated.
    class TypeStore {
        // Global primitive type references.
        UnitType m_unitType;
        BoolType m_boolType;
        IntegerType m_i8Type;
        IntegerType m_u8Type;
        IntegerType m_i16Type;
        IntegerType m_u16Type;
        IntegerType m_i32Type;
        IntegerType m_u32Type;
        IntegerType m_i64Type;
        IntegerType m_u64Type;
        FloatingType m_f32Type;
        FloatingType m_f64Type;

        // Type store for pointer types.
        std::unordered_map<std::uint64_t, std::unique_ptr<PointerType>> m_pointerStore;

    public:
        // Creates a new default type store.
        TypeStore();

        /* ------------------------------------------------------------------ */

        // unitType returns a `unit` type.
        inline UnitType* unitType() { return &m_unitType; }

        // boolType returns a `bool` type.
        inline BoolType* boolType() { return &m_boolType; }

        // i8Type returns an `i8` type.
        inline IntegerType* i8Type() { return &m_i8Type; }

        // u8Type returns a `u8` type.
        inline IntegerType* u8Type() { return &m_u8Type; }

        // i16Type returns an `i16` type. 
        inline IntegerType* i16Type() { return &m_i16Type; }

        // u16Type returns a `u16` type.
        inline IntegerType* u16Type() { return &m_u16Type; }

        // i32Type returns an `i32` type.
        inline IntegerType* i32Type() { return &m_i32Type; }

        // u32Type returns an `u32` type.
        inline IntegerType* u32Type() { return &m_u32Type; }
        
        // i64Type returns an `i64` type.
        inline IntegerType* i64Type() { return &m_i64Type; }
        
        // u64Type returns an `u64` type.
        inline IntegerType* u64Type() { return &m_u64Type; }
        
        // f32Type returns an `f32` type.
        inline FloatingType* f32Type() { return &m_f32Type; }
        
        // f64Type returns an `f64` type.
        inline FloatingType* f64Type() { return &m_f64Type; }

        /* ------------------------------------------------------------------ */

        // pointerType creates a new pointer type.
        PointerType* pointerType(Type* elemType, bool isConst);
        
    };

    // typeStore is the global type store used for compilation.
    inline TypeStore typeStore;
}

#endif