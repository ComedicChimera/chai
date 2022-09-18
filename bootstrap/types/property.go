package types

// Property represents a property of a type.
type Property struct {
	// The kind of property.  This must be one of the enumerated property kinds.
	Kind PropertyKind

	// The type of the property.
	Type Type

	// Whether the property is mutable.
	Mutable bool
}

// A kind of property.
type PropertyKind int

// Enumeration of property kinds.
const (
	PropStructField PropertyKind = iota
)

// getProperty attempts to return the property of typ named name.  If no
// matching property is found, this function returns `nil`.
func GetProperty(typ Type, name string) *Property {
	innerTyp := InnerType(typ)

	switch v := innerTyp.(type) {
	case *StructType:
		if field, ok := v.GetFieldByName(name); ok {
			return &Property{
				Kind: PropStructField,
				Type: field.Type,
				// Struct fields are always mutable if the struct is mutable.
				Mutable: true,
			}
		}
	case *PointerType:
		// Implement automatic dereferencing.
		return GetProperty(v.ElemType, name)
	}

	return nil
}
