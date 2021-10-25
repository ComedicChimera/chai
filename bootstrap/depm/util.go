package depm

// GenerateIDFromPath generates an ID from an absolute path.
func GenerateIDFromPath(abspath string) uint {
	return 0
}

// IsValidIdentifier returns whether or not a given string would be a valid
// identifier (module name, package name, etc.).
func IsValidIdentifier(idstr string) bool {
	if idstr[0] == '_' || ('a' <= idstr[0] && idstr[0] <= 'z') || ('A' <= idstr[0] && idstr[0] <= 'Z') {
		for _, c := range idstr[1:] {
			if c == '_' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
				continue
			}

			return false
		}

		return true
	}

	return false
}
