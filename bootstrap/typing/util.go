package typing

import "log"

// NothingType returns a new `nothing` type.
func NothingType() PrimType {
	return PrimType(PrimNothing)
}

// IsNothing returns if the given data type is equivalent to `nothing`.
func IsNothing(dt DataType) bool {
	return Equiv(dt, NothingType())
}

// boolType returns a new `bool` type.
func BoolType() PrimType {
	return PrimType(PrimBool)
}

// -----------------------------------------------------------------------------

// Simplify converts a type into its simplest form.  This removes all wrapping
// such as opaque types and type variables.  It also performs Nothing
// coalescion: all types which contains only Nothing evaluate to nothing. This
// is used to prepare a type for generation on the backend.
func Simplify(dt DataType) DataType {
	switch v := dt.(type) {
	case PrimType:
		return v
	case *RefType:
		return &RefType{ElemType: Simplify(v.ElemType)}
	case TupleType:
		{
			var simpleElemTypes []DataType
			for _, tElemType := range v {
				stElemType := Simplify(tElemType)
				if !IsNothing(stElemType) {
					simpleElemTypes = append(simpleElemTypes, stElemType)
				}
			}

			// handle degenerate tuples
			if len(simpleElemTypes) == 0 {
				return NothingType()
			} else if len(simpleElemTypes) == 1 {
				return simpleElemTypes[0]
			} else {
				return TupleType(simpleElemTypes)
			}
		}
	case *FuncType:
		{
			ft := &FuncType{
				IntrinsicName: v.IntrinsicName,
				ReturnType:    Simplify(v.ReturnType),
			}

			for _, arg := range v.Args {
				stype := Simplify(arg)
				if !IsNothing(stype) {
					ft.Args = append(ft.Args, stype)
				}
			}

			return ft
		}
	case *StructType:
		// Note: This form of simplify is used for reference to a type outside
		// the declaration.  In this case, we can just return the struct type as
		// is since it will simply correspond to a global definition look up.
		return v
	case *OpaqueType:
		return Simplify(*v.TypePtr)
	case *TypeVar:
		return Simplify(v.Value)
	case *AliasType:
		return Simplify(v.Type)

	}

	log.Fatalf("Simplification not yet supported for type: %s\n", dt.Repr())
	return nil
}

// SimplifyStructTypeDef simplifies a struct type definition.
func SimplifyStructTypeDef(st *StructType) DataType {
	var newFields []StructField
	newFieldsByName := make(map[string]int)

	for _, field := range st.Fields {
		stype := Simplify(field.Type)
		if !IsNothing(stype) {
			newFieldPos := len(newFields)

			newFields = append(newFields, StructField{
				Name:        field.Name,
				Type:        stype,
				Public:      field.Public,
				Initialized: field.Initialized,
			})

			newFieldsByName[field.Name] = newFieldPos
		}
	}

	// handle degenerate structs
	if len(newFields) == 0 {
		return NothingType()
	}

	return &StructType{
		NamedTypeBase: st.NamedTypeBase,
		Fields:        newFields,
		FieldsByName:  newFieldsByName,
	}
}
