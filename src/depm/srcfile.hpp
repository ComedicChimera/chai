#ifndef SRCFILE_H_INCLUDED
#define SRCFILE_H_INCLUDED

#include <string>

namespace chai {
    // SrcFile represents a Chai source file and local namespace. It contains
    // the AST for that source file
    class SrcFile {
        std::string filepath;
    };
}



#endif 