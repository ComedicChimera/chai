#include "types.hpp"

namespace chai {
    bool IntegerType::internalEquals(Type* other) const {
        if (other->kind() == TypeKind::INTEGER) {
            auto* it = dynamic_cast<IntegerType*>(other);
            return m_size == it->m_size && isUnsigned == it->isUnsigned;
        }

        return false;
    }

    bool FloatingType::internalEquals(Type* other) const {
        if (other->kind() == TypeKind::FLOATING) {
            auto* ft = dynamic_cast<FloatingType*>(other);
            return m_size == ft->m_size;
        }

        return false;
    }

    bool PointerType::internalEquals(Type* other) const {
        if (other->kind() == TypeKind::POINTER) {
            auto* pt = dynamic_cast<PointerType*>(other);
            return isConst == pt->isConst && m_elemType->equals(pt->elemType());
        }

        return false;
    }

    bool FunctionType::internalEquals(Type* other) const {
        if (other->kind() == TypeKind::FUNCTION) {
            auto* ft = dynamic_cast<FunctionType*>(other);

            if (paramTypes.size() != ft->paramTypes.size())
                return false;

            if (!m_returnType->equals(ft->returnType()))
                return false;

            for (int i = 0; i < paramTypes.size(); i++) {
                if (!paramTypes[i]->equals(ft->paramTypes[i].get()))
                    return false;
            }
            
            return true;
        }

        return false;
    }
}