#ifndef SCANNER_H_INCLUDED
#define SCANNER_H_INCLUDED

#include <string>
#include <sstream>

namespace chai {
    class Scanner {
        int line, col;
        std::stringstream ss;

    public:
        Scanner(const std::string& filePath);
    };
}

#endif