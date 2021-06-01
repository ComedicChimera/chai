package common

import "hash/fnv"

// GenerateIDFromPath takes an absolute path and converts it into a numeric ID;
// this is used by modules and packages to generate their unique IDs
func GenerateIDFromPath(abspath string) uint {
	h := fnv.New32a()
	h.Write([]byte(abspath))
	return uint(h.Sum32())
}
