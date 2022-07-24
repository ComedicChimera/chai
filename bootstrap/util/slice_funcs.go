package util

// Contains returns whether the given slice contains the given element.
func Contains[T comparable](slice []T, elem T) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}

	return false
}

// Map applies a function to the given slice and returns the transformed slice.
func Map[T, R any](slice []T, f func(T) R) []R {
	mSlice := make([]R, len(slice))

	for i, elem := range slice {
		mSlice[i] = f(elem)
	}

	return mSlice
}
