package walk

import "chai/syntax"

// WalkDef walks a core definition node (eg. `type_def` or `variable_decl`)
func (w *Walker) WalkDef(branch *syntax.ASTBranch, public bool, annots map[string]*Annotation) bool {
	return false
}
