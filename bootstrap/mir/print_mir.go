package mir

import (
	"strings"
)

// Repr returns the full textual representation of MIR package.
func (bundle *MIRBundle) Repr() string {
	sb := strings.Builder{}

	for _, external := range bundle.Externals {
		sb.WriteString("extern ")
		sb.WriteString(external.Repr())
		sb.WriteString(";\n")
	}

	sb.WriteRune('\n')

	for _, forward := range bundle.Forwards {
		sb.WriteString("forward ")
		sb.WriteString(forward.Repr())
		sb.WriteString(";\n")
	}

	sb.WriteRune('\n')

	for _, impl := range bundle.Functions {
		if impl.Def.Public() {
			sb.WriteString("pub ")
		} else {
			sb.WriteString("priv ")
		}

		sb.WriteString(impl.Def.Repr())

		// TODO: print body

		sb.WriteString("\n\n")
	}

	return sb.String()
}
