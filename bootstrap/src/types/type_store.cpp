#include "type_store.hpp"

namespace chai {
    TypeStore::TypeStore() 
    : m_i8Type(1, false)
    , m_u8Type(1, true)
    , m_i16Type(2, false)
    , m_u16Type(2, true)
    , m_i32Type(4, false)
    , m_u32Type(4, true)
    , m_i64Type(8, false)
    , m_u64Type(8, true)
    , m_f32Type(4)
    , m_f64Type(8)
    {}

    PointerType* TypeStore::pointerType(Type* elem, bool isConst) {
        // We can use the address of the element type as the basis for the
        // lookup key for the pointer: because all types are stored globally,
        // any two element types which are "equal" should have the same address.
        // Furthermore, on 32-bit systems, the upper bits of the 64-bit integer
        // are completely unused, and on 64-bit systems, as far as I know, the
        // most significant bit is not used either (eg. on Windows the valid
        // virtual address space spans from 0x000'0000 to 0x7ff'ffff) so we can
        // use the most significant bit to indicate constancy.  This may not be
        // the "best" solution long term, but it is efficient and works for
        // right now.
        auto key = (std::uint64_t)(elem) & ((std::uint64_t)(isConst) << 63);

        auto iter = m_pointerStore.find(key);

        if (iter == m_pointerStore.end()) {
            auto ptr = new PointerType(elem, isConst);
            m_pointerStore.emplace(std::make_pair(key, ptr));
            return ptr;
        }

        return iter->second.get();
    }
}