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
		case "linkname":
			if len(annot.Values) != 1 {
				w.logError(
					"annotation @linkname requires the link name of the symbol as an argument",
					logging.LMKAnnot,
					annot.NamePos,
				)

				return false, false
			}

			needsBody = false
		case "extern", "intrinsic", "dllimport":
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
		default:
			w.logWarning(
				fmt.Sprintf("annotation `%s` has no predefined meaning on functions", annot.Name),
				logging.LMKAnnot,
				annot.NamePos,
			)
		}
	}

	return needsBody, true
}

// validateOperatorAnnotations validates the annotations for an operator.  It
// returns a string indicating the name of the intrinsic operator this operator
// definition implements if one is supplied -- this string will be empty if the
// operator is not intrinsic and therefore requires a body.  It also returns a
// boolean indicating whether or not the annotations were valid.
func (w *Walker) validateOperatorAnnotations(annots map[string]*sem.Annotation) (string, bool) {
	intrinsicName := ""

	for _, annot := range annots {
		switch annot.Name {
		case "intrinsic_op":
			if len(annot.Values) != 1 {
				w.logError(
					"annotation `intrinsic_op` takes an operator generator name as an argument",
					logging.LMKAnnot,
					annot.NamePos,
				)

				return "", false
			}

			intrinsicName = annot.Values[0].Value
		case "inline", "unsafe", "tailrec", "constexpr", "introspect", "inplace":
			// NOTE: @introspect lets you access private properties
			if len(annot.Values) != 0 {
				w.logError(
					fmt.Sprintf("annotation `%s` does not take arguments", annot.Name),
					logging.LMKAnnot,
					annot.Values[0].Position(),
				)

				return "", false
			}
		default:
			w.logWarning(
				fmt.Sprintf("annotation `%s` has no predefined meaning on operators", annot.Name),
				logging.LMKAnnot,
				annot.NamePos,
			)
		}
	}

	return intrinsicName, true
}
