package syntax

import (
	"chai/ast"
)

// parseFile parses the top level of a file.  This is the grammatical start
// symbol for the parser.
func (p *Parser) parseFile() ([]ast.Def, bool) {
	// skip leading newlines
	if !p.newlines() {
		return nil, false
	}

	// TODO: directive parsing

	// import parsing
	for p.got(IMPORT) {
		if !p.parseImportStmt() {
			return nil, false
		}

		if !p.newlines() {
			return nil, false
		}
	}

	// parse definitions until we get an EOF
	var defs []ast.Def
	annotations := make(map[string]string)
	inPubBlock := false
	for !p.got(EOF) {
		pubDef := inPubBlock

		switch p.tok.Kind {
		case END:
			// ending of public blocks
			if inPubBlock {
				if !p.advance() {
					return nil, false
				}

				inPubBlock = false
			} else {
				p.reject()
				return nil, false
			}
		case ANNOTSTART:
			// annotations
			if _annots, ok := p.parseAnnotations(); ok {
				annotations = _annots
			} else {
				return nil, false
			}

			// annotations then visibility then definition
			fallthrough
		case PUB:
			// we have to add this extra check in case we fall through from
			// annotations
			if p.got(PUB) {
				// visibility
				if !p.next() {
					return nil, false
				}

				if p.got(NEWLINE) {
					// public block
					if !p.advance() {
						return nil, false
					}

					inPubBlock = true
					continue
				} else {
					// single public definition => mark the next definition as
					// public
					pubDef = true
				}

			}

			// visibility then definition
			fallthrough
		default:
			// assume definition
			if def, ok := p.parseDefinition(annotations, pubDef); ok {
				defs = append(defs, def)

				// assert newlines after a definition
				if !p.got(EOF) && !p.assertAndAdvance(NEWLINE) {
					return nil, false
				}
			} else {
				return nil, false
			}
		}

		// reset annotations
		if len(annotations) != 0 {
			annotations = make(map[string]string)
		}
	}

	// check for missing closing ends on public blocks
	if inPubBlock {
		p.reject()
		return nil, false
	}

	return defs, true
}

// -----------------------------------------------------------------------------

// annotations = '@' (annot_elem | '[' annot_elem ']') 'NEWLINE'
// annotation = 'IDENTIFIER' ['(' 'STRING_LIT' ')']
func (p *Parser) parseAnnotations() (map[string]string, bool) {
	if !p.assertAndNext(ANNOTSTART) {
		return nil, false
	}

	annotations := make(map[string]string)
	expectingMultiple := false

	if p.got(LBRACKET) {
		expectingMultiple = true
		if !p.advance() {
			return nil, false
		}
	}

	// parse annotations
	for {
		// collect annotation pieces
		if !p.assertAndNext(IDENTIFIER) {
			return nil, false
		}

		idTok := p.lookbehind

		value := ""
		var valueTok *Token
		if p.got(LPAREN) {
			// skip newlines after the opening lparen
			if !p.newlines() {
				return nil, false
			}

			if !p.want(STRINGLIT) {
				return nil, false
			}

			value = p.tok.Value
			valueTok = p.tok

			if !p.advance() || !p.assertAndNext(RPAREN) {
				return nil, false
			}
		}

		// check that the annotation is not duplicated
		if _, ok := annotations[idTok.Value]; ok {
			p.errorOn(idTok, "multiple values specified for annotation `%s`", idTok.Value)
			return nil, false
		}

		// validate special annotation arity
		if specialCfg, ok := specialAnnotations[idTok.Value]; ok {
			if specialCfg.ExpectsValue && value == "" {
				p.errorOn(idTok, "the special annotation `%s` requires a value", idTok.Value)
				return nil, false
			} else if !specialCfg.ExpectsValue && value != "" {
				p.errorOn(valueTok, "the special annotation `%s` doesn't take a value", idTok.Value)
				return nil, false
			}
		}

		// store new annotation
		annotations[idTok.Value] = value

		if expectingMultiple && p.got(COMMA) {
			if !p.advance() {
				return nil, false
			}
		} else {
			break
		}
	}

	// make sure multi-annotation are closed and end the annotation
	if expectingMultiple && !p.assertAndNext(RBRACKET) || !p.assertAndNext(NEWLINE) {
		return nil, false
	}

	return annotations, true
}

// specialAnnotConfig specifies what usages are legal for special annotations.
type specialAnnotConfig struct {
	// Usage is a string indicating what type of code object can use this
	// annotation.  Valid values are: `function`, `operator`
	Usage string

	ExpectsValue bool
}

var specialAnnotations = map[string]specialAnnotConfig{
	"callconv":  {Usage: "function", ExpectsValue: true},
	"intrinsic": {Usage: "function", ExpectsValue: false},
	"entry":     {Usage: "function", ExpectsValue: false},
	"extern":    {Usage: "function", ExpectsValue: false},
	"inline":    {Usage: "function", ExpectsValue: false},

	"intrinsicop": {Usage: "operator", ExpectsValue: true},
}

// validateAnnotUsage checks that the given list of annotations is being used on
// the correct type of code object (eg. you can use `intrinsic` on an operator).
// It assumes the parser is currently positioned on a suitable token to error
// upon.
func (p *Parser) validateAnnotUsage(annotations map[string]string, usage string) bool {
	for name := range annotations {
		if specialCfg, ok := specialAnnotations[name]; ok {
			if specialCfg.Usage != usage {
				p.rejectWithMsg("the special annotation `%s` can only be applied to %s", name, specialCfg.Usage)
				return false
			}
		}
	}

	return true
}

// definition = func_def | oper_def | type_def | space_def | var_def | const_def
func (p *Parser) parseDefinition(annotations map[string]string, public bool) (ast.Def, bool) {
	switch p.tok.Kind {
	case DEF:
		// func_def
		return p.parseFuncDef(annotations, public)
	case OPER:
		// oper_def
		return p.parseOperDef(annotations, public)
	case TYPE:
		// type_def
		return p.parseTypeDef(annotations, public)
	case SPACE:
		// TODO: space_def
	case LET:
		// global_var_def
		{
			vd, ok := p.parseVarDecl(true, annotations, public)
			return vd.(*ast.VarDecl), ok
		}
	case CONST:
		// TODO: const_def
	}

	p.reject()
	return nil, false
}
