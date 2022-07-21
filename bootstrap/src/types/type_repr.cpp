#include "types.hpp"

#include <format>

namespace chai {
    std::string IntegerType::repr() const {
        return std::format(
            "{}{}", 
            isUnsigned ? 'u' : 'i',
            m_size
        );
    }

    std::string FloatingType::repr() const {
        return std::format("f%d", m_size);
    }

    std::string PointerType::repr() const {
        std::string builder { "*" };
        
        if (isConst) {
            builder += "const ";
        }

        return builder + m_elemType->repr();
    }

    std::string FunctionType::repr() const {
        std::string builder;

        switch (paramTypes.size()) {
        case 0:
            builder += "()";
            break;
        case 1:
            builder += paramTypes[0]->repr();
            break;
        default:
            builder += '(';

            int i = 0;
            for (auto& paramType : paramTypes) {
                if (i > 0) {
                    builder += ", ";
                }

                builder += paramType->repr();

                i++;
            }

            builder += ')';
            break;
        }

        builder += " -> ";

        return builder + m_returnType->repr();
    }
}