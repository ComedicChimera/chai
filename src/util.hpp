#ifndef UTIL_H_INCLUDED
#define UTIL_H_INCLUDED

#include <cstdint>

namespace chai {
    using u32 = std::uint32_t;
    using i32 = std::int32_t;

    // getID generates a new ID for use by a given module or package
    u32 getID() {
        static u32 idCounter = 0;
        return idCounter++;
    }

    // concatVec concatenates one vector `b` onto the end of another vector `a`
    template<typename T>
    void concatVec(std::vector<T>& a, const std::vector<T>& b) {
        a.insert(a.end(), b.begin(), b.end());
    }
}

#endif 