package resolve

import (
	"chai/logging"
	"chai/sem"
	"chai/syntax"
	"fmt"
)

// Definition represents a top level definition as it has been extracted from
// the program source.  These definitions have NOT been walked yet.
type Definition struct {
	// Name is the "resolving" name -- name top level definitions will depend on
	Name string

	// NamePos is the position of the name
	NamePos *logging.TextPosition

	// SrcFile is the file this definition is defined in
	SrcFile *sem.ChaiFile

	// Definition properties
	Public      bool
	Annotations map[string]*sem.Annotation

	// AST is the full AST branch of the definition
	AST *syntax.ASTBranch
}

// extractDefinition extracts `Definition` from a single definition node (ie.
// the subnode of a `definition` or `pub_definition` node) and adds it to the
// appropriate resolution list.
func (r *Resolver) extractDefinition(srcfile *sem.ChaiFile, branch *syntax.ASTBranch, public bool) bool {
	def := &Definition{
		Annotations: make(map[string]*sem.Annotation),
		SrcFile:     srcfile,
		Public:      public,
	}

	// remove all the layers around the core definition itself
	if branch.Name == "annotated_def" || branch.Name == "internal_annotated_def" {
		if annots, ok := r.walkAnnotations(srcfile, branch.BranchAt(0)); ok {
			branch = branch.BranchAt(1)
			def.Annotations = annots
		} else {
			return false
		}
	}

	if branch.Name == "pub_def" {
		def.Public = true
		branch = branch.BranchAt(1)
	}

	if branch.Name == "def_core" {
		branch = branch.BranchAt(0)
	}

	def.AST = branch

	switch branch.Name {
	case "type_def", "class_def", "cons_def":
		// these definitions are all independent and have their names as their
		// second element -- TODO: handle closed type definitions
		def.Name = branch.LeafAt(1).Value
		def.NamePos = branch.LeafAt(1).Position()
		r.independents = append(r.independents, def)
	case "variable_decl":
		r.globalVars = append(r.globalVars, def)
	default:
		// all other definitions kinds are dependent
		r.dependents = append(r.dependents, def)
	}

	return true
}

// walkAnnotations walks the annotations (`annotation` branch) of a definition
func (r *Resolver) walkAnnotations(srcfile *sem.ChaiFile, branch *syntax.ASTBranch) (map[string]*sem.Annotation, bool) {
	annots := make(map[string]*sem.Annotation)

	for _, item := range branch.Content {
		// only branch is `annot_single`
		if branch, ok := item.(*syntax.ASTBranch); ok {
			name := ""
			for _, elem := range branch.Content {
				// all sub nodes are leaves
				leaf := elem.(*syntax.ASTLeaf)
				if leaf.Kind == syntax.IDENTIFIER {
					name = leaf.Value

					if _, ok := annots[name]; ok {
						logging.LogCompileError(
							srcfile.LogContext,
							fmt.Sprintf("multiple sets of values provided for annotation `%s`", name),
							logging.LMKAnnot,
							leaf.Position(),
						)

						return nil, false
					}

					annots[name] = &sem.Annotation{
						Name:    name,
						NamePos: leaf.Position(),
					}
				} else if leaf.Kind == syntax.STRINGLIT {
					annots[name].Values = append(annots[name].Values, leaf)
				}
			}
		}
	}

	return annots, true
}
