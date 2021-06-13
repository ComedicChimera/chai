package walk

import (
	"chai/logging"
	"chai/sem"
	"fmt"
)

// validateFuncAnnotations takes in a map of annotations for a function,
// validates them, and returns whether or not the function must have a body
// (first bool, second bool is success or fail).
func (w *Walker) validateFuncAnnotations(annots map[string]*sem.Annotation) (bool, bool) {
	needsBody := true

	for _, annot := range annots {
		switch annot.Name {
		case "dllimport":
			if len(annot.Values) != 1 {
				w.logError(
					"annotation @dllimport requires the name of symbol to import as an argument",
					logging.LMKAnnot,
					annot.NamePos,
				)

				return false, false
			}

			needsBody = false
		case "extern", "intrinsic":
			needsBody = false
			fallthrough
		case "inline", "unsafe", "tailrec", "constexpr", "introspect":
			// NOTE: @introspect lets you access private properties
			if len(annot.Values) != 0 {
				w.logError(
					fmt.Sprintf("annotation `%s` does not take arguments", annot.Name),
					logging.LMKAnnot,
					annot.Values[0].Position(),
				)

				return false, false
			}
		}
	}

	return needsBody, true
}
