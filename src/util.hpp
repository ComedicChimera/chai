#ifndef UTIL_H_INCLUDED
#define UTIL_H_INCLUDED

#include <cstdint>

namespace chai {
    using u64 = std::uint64_t;
    using u32 = std::uint32_t;
    using i32 = std::int32_t;

    // concatVec concatenates one vector `b` onto the end of another vector `a`
    template<typename T>
    void concatVec(std::vector<T>& a, const std::vector<T>& b) {
        a.insert(a.end(), b.begin(), b.end());
    }
}

#endif 