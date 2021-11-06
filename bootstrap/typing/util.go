package typing

// InnerType extracts the inner type value of a data type. This value can then
// be used for things like cast checking.  For example, it extracts the value
// for a type variable.
func InnerType(dt DataType) DataType {
	switch v := dt.(type) {
	case *TypeVar:
		return InnerType(v.Value)
	default:
		// no inner type to extract
		return v
	}
}
