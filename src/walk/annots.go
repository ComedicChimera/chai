package walk

import (
	"chai/logging"
	"chai/syntax"
)

// Annotation represents an annotation (name and value)
type Annotation struct {
	Name    string
	NamePos *logging.TextPosition

	// Values stores leaves so we can keep their position
	Values []*syntax.ASTLeaf
}
