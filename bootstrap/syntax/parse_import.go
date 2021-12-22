package syntax

import (
	"chai/depm"
	"chai/report"
	"strings"
)

// isymbol represents an imported symbol.
type isymbol struct {
	// Name will contain the operator string if this symbol is an imported
	// operator.
	Name string

	// OpKind will be `-1` if this is an named symbol.
	OpKind int

	Pos *report.TextPosition
}

// import_stmt = 'import' (package_import | symbol_import)
func (p *Parser) parseImportStmt() bool {
	if !p.assertAndNext(IMPORT) {
		return false
	}

	// parse and collect the import statement
	importedSymbols := make(map[string]isymbol)
	var importedPkg *depm.ChaiPackage
	pkgImportName := ""
	var pkgImportNamePos *report.TextPosition

	if p.got(IDENTIFIER) {
		// save the first ID: don't know what it is yet
		firstId := p.tok
		if !p.next() {
			return false
		}

		// could be package import or symbol import
		var ok bool
		switch p.tok.Kind {
		case COMMA, FROM:
			// => symbol import
			importedSymbols[firstId.Value] = isymbol{Name: firstId.Value, OpKind: -1, Pos: firstId.Position}
			importedPkg, pkgImportNamePos, ok = p.parseSymbolImport(importedSymbols)
			pkgImportName = importedPkg.Name
			if !ok {
				return false
			}
		case NEWLINE, AS, DOT:
			// => package import
			importedPkg, pkgImportName, pkgImportNamePos, ok = p.parsePackageImport(firstId.Value, firstId.Position)
			if !ok {
				return false
			}
		default:
			p.reject()
			return false
		}
	} else if p.got(LPAREN) {
		// only be symbol import: parse the first isymbol and then feed it to
		// parse symbol_import
		if isym, ok := p.parseISymbol(); ok {
			importedSymbols[isym.Name] = isym
			importedPkg, pkgImportNamePos, ok = p.parseSymbolImport(importedSymbols)
			pkgImportName = importedPkg.Name
			if !ok {
				return false
			}
		} else {
			return false
		}
	}

	// add the package as a dependency
	var chPkgImport depm.ChaiPackageImport
	if _cpi, ok := p.chFile.Parent.ImportedPackages[importedPkg.ID]; ok {
		chPkgImport = _cpi
	} else {
		chPkgImport = depm.ChaiPackageImport{
			Pkg:       importedPkg,
			Symbols:   make(map[string]*depm.Symbol),
			Operators: make(map[int]*depm.Operator),
		}
		p.chFile.Parent.ImportedPackages[importedPkg.ID] = chPkgImport
	}

	// handle imported symbols and operators
	if len(importedSymbols) > 0 {
		// add each imported symbol and check for name collisions
		for _, isym := range importedSymbols {
			if isym.OpKind == -1 {
				// symbol imports
				if p.chFile.ImportCollides(isym.Name) {
					p.reportError(isym.Pos, "multiple symbols imported with name `%s`", isym.Name)
					return false
				}

				p.chFile.ImportedSymbols[isym.Name] = &depm.Symbol{
					Name:  isym.Name,
					PkgID: importedPkg.ID,
					// the DefPosition is set to the isymbol's position for
					// error handling in resolver; this will be updated with the
					// actual symbol's position later
					DefPosition: isym.Pos,
					DefKind:     depm.DKUnknown,
					Mutability:  depm.NeverMutated,
					Public:      true,
				}
			} else {
				// operator imports

				// for operators, we simply add a single, untyped overload for
				// our respective package to tell the resolver where to load
				// overloads from.  It is likely that there may be conflicts
				// that arise later on, but we won't handle them now.
				overload := &depm.OperatorOverload{
					// Signature is `nil` to mark this as resolving
					Context: &report.CompilationContext{
						ModName:    importedPkg.Parent.Name,
						ModAbsPath: importedPkg.Parent.AbsPath,
						// FileRelPath will be determined later
					},
					// Similar to symbols, operators use the import position as
					// their definition position until resolved
					Position: isym.Pos,
					Public:   true,
				}

				if op, ok := p.chFile.ImportedOperators[isym.OpKind]; ok {
					op.Overloads = append(op.Overloads, overload)
				} else {
					p.chFile.ImportedOperators[isym.OpKind] = &depm.Operator{
						OpName:    isym.Name,
						Overloads: []*depm.OperatorOverload{overload},
					}
				}
			}
		}
	} else {
		// check for name collisions
		if p.chFile.ImportCollides(pkgImportName) {
			p.reportError(pkgImportNamePos, "multiple symbols imported with name `%s`", pkgImportName)
			return false
		}

		// add as visible package
		p.chFile.VisiblePackages[pkgImportName] = importedPkg
	}

	return true
}

// package_import = pkg_path ['as' 'IDENTIFIER']
// NOTE: first ID should be parsed before this function is called and passed in.
func (p *Parser) parsePackageImport(moduleName string, moduleNamePos *report.TextPosition) (*depm.ChaiPackage, string, *report.TextPosition, bool) {
	// parse the package path
	pkg, pkgNamePos, ok := p.parsePkgPath(moduleName, moduleNamePos)
	if !ok {
		return nil, "", nil, false
	}

	// read in a rename as necessary
	if p.got(AS) {
		if !p.next() {
			return nil, "", nil, false
		}

		renameTok := p.tok
		if !p.assertAndNext(IDENTIFIER) {
			return nil, "", nil, false
		}

		return pkg, renameTok.Value, renameTok.Position, true
	}

	return pkg, pkg.Name, pkgNamePos, true
}

// symbol_import = isymbol {',' isymbol} 'from' pkg_path
// NOTE: first `isymbol` should be parsed before this function is called and
// passed in.
func (p *Parser) parseSymbolImport(importedSymbols map[string]isymbol) (*depm.ChaiPackage, *report.TextPosition, bool) {
	// parse remaining symbol imports
	for p.got(COMMA) {
		if !p.next() {
			return nil, nil, false
		}

		if isym, ok := p.parseISymbol(); ok {
			if _, ok := importedSymbols[isym.Name]; ok {
				p.reportError(isym.Pos, "symbol `%s` imported multiple times", isym.Name)
				return nil, nil, false
			}

			importedSymbols[isym.Name] = isym
		}
	}

	// parse the package path
	if !p.assertAndNext(FROM) {
		return nil, nil, false
	}

	modNameTok := p.tok
	if !p.assertAndNext(IDENTIFIER) {
		return nil, nil, false
	}

	pkg, pkgNamePos, ok := p.parsePkgPath(modNameTok.Value, modNameTok.Position)
	if !ok {
		return nil, nil, false
	}

	return pkg, pkgNamePos, true
}

// pkg_path = 'IDENTIFIER' {'.' 'IDENTIFIER'}
// NOTE: first ID should be parsed before this function is called and passed in.
func (p *Parser) parsePkgPath(moduleName string, moduleNamePos *report.TextPosition) (*depm.ChaiPackage, *report.TextPosition, bool) {
	// parse the subpath itself
	pkgSubPath := strings.Builder{}
	pathEndPos := moduleNamePos

	for p.got(COMMA) {
		if !p.next() {
			return nil, nil, false
		}

		pkgSubPath.WriteRune('.')
		pkgSubPath.WriteString(p.tok.Value)
		pathEndPos = p.tok.Position

		if !p.assertAndNext(IDENTIFIER) {
			return nil, nil, false
		}
	}

	// import a package based on the subpath
	if pkg, ok := p.importFunc(p.chFile.Parent.Parent, moduleName, pkgSubPath.String()); ok {
		return pkg, pathEndPos, true
	}

	p.reportError(
		report.TextPositionFromRange(moduleNamePos, pathEndPos),
		"unable to import package `%s%s`",
		moduleName, pkgSubPath.String(),
	)
	return nil, nil, false
}

// isymbol = 'IDENTIFIER' | '(' operator ')'
func (p *Parser) parseISymbol() (isymbol, bool) {
	if p.got(IDENTIFIER) {
		return isymbol{Name: p.tok.Value, OpKind: -1, Pos: p.tok.Position}, p.next()
	}

	if !p.assertAndNext(LPAREN) {
		return isymbol{}, false
	}

	if opTok, ok := p.parseOperator(); ok {
		return isymbol{Name: opTok.Value, OpKind: opTok.Kind, Pos: opTok.Position}, p.assertAndNext(RPAREN)
	}

	return isymbol{}, false
}
