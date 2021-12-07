package syntax

import (
	"chai/depm"
	"chai/report"
	"strings"
)

// isymbol represents an imported symbol.
type isymbol struct {
	// Name will be empty string if it is an operator symbol.
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
	importedSymbols := make(map[isymbol]struct{})
	var importedPkg *depm.ChaiPackage
	pkgImportName := ""

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
			importedSymbols[isymbol{Name: firstId.Value, OpKind: -1, Pos: firstId.Position}] = struct{}{}
			importedPkg, pkgImportName, ok = p.parseSymbolImport(importedSymbols)
			if !ok {
				return false
			}
		case NEWLINE, AS, DOT:
			// => package import
			importedPkg, pkgImportName, ok = p.parsePackageImport(firstId.Value, firstId.Position)
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
			importedSymbols[isym] = struct{}{}
			importedPkg, pkgImportName, ok = p.parseSymbolImport(importedSymbols)
			if !ok {
				return false
			}
		} else {
			return false
		}
	}

	// TODO: add the package as a dependency and declare imported symbols
	_ = importedPkg
	_ = pkgImportName

	return true
}

// package_import = pkg_path ['as' 'IDENTIFIER']
// NOTE: first ID should be parsed before this function is called and passed in.
func (p *Parser) parsePackageImport(moduleName string, moduleNamePos *report.TextPosition) (*depm.ChaiPackage, string, bool) {
	// TODO
	return nil, "", false
}

// symbol_import = isymbol {',' isymbol} 'from' pkg_path
// NOTE: first `isymbol` should be parsed before this function is called and
// passed in.
func (p *Parser) parseSymbolImport(importedSymbols map[isymbol]struct{}) (*depm.ChaiPackage, string, bool) {
	// TODO
	return nil, "", false
}

// pkg_path = 'IDENTIFIER' {'.' 'IDENTIFIER'}
// NOTE: first ID should be parsed before this function is called and passed in.
func (p *Parser) parsePkgPath(moduleName string, moduleNamePos *report.TextPosition) (*depm.ChaiPackage, bool) {
	// parse the subpath itself
	pkgSubPath := strings.Builder{}
	pathEndPos := moduleNamePos

	for p.got(COMMA) {
		if !p.next() {
			return nil, false
		}

		pkgSubPath.WriteRune('.')
		pkgSubPath.WriteString(p.tok.Value)
		pathEndPos = p.tok.Position

		if !p.assertAndNext(IDENTIFIER) {
			return nil, false
		}
	}

	// import a package based on the subpath
	if pkg, ok := p.importFunc(p.chFile.Parent.Parent, moduleName, pkgSubPath.String()); ok {
		return pkg, true
	}

	p.reportError(
		report.TextPositionFromRange(moduleNamePos, pathEndPos),
		"unable to import package `%s%s`",
		moduleName, pkgSubPath.String(),
	)
	return nil, false
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
		return isymbol{Name: "", OpKind: opTok.Kind, Pos: opTok.Position}, p.assertAndNext(RPAREN)
	}

	return isymbol{}, false
}
