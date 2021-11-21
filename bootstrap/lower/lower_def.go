package lower

import (
	"chai/ast"
	"chai/ir"
	"chai/typing"
	"log"
)

// lowerDef lowers an AST definition to an IR definition.
func (l *Lowerer) lowerDef(def ast.Def) {
	switch v := def.(type) {
	case *ast.FuncDef:
		l.lowerFunc(v)
	}
}

// lowerFunc lowers a function definition.
func (l *Lowerer) lowerFunc(afunc *ast.FuncDef) {
	// convert arguments
	var iargs []ir.FuncArg
	for _, arg := range afunc.Args {
		// nothing arguments don't compile
		if arg.Type.Equiv(typing.PrimType(typing.PrimNothing)) {
			continue
		}

		// TODO: byref arguments
		if arg.ByRef {
			log.Fatalln("by reference arguments aren't supported yet")
		}

		iargs = append(iargs, ir.FuncArg{
			Name: arg.Name,
			Typ:  l.lowerType(arg.Type),
		})
	}

	// handle annotations, linkage, and determine the function name
	var name string
	callConv := ir.ChaiCC
	inline := hasAnnot(afunc.Annotations(), "inline")

	linkage := ir.Private
	if afunc.Public() {
		linkage = ir.Public
	}

	// external to source package => no name mangling
	if hasAnnot(afunc.Annotations(), "dllimport") {
		name = afunc.Name
		linkage |= ir.DllImport
		linkage |= ir.External
	} else if hasAnnot(afunc.Annotations(), "dllexport") {
		name = afunc.Name
		linkage |= ir.DllExport
	} else if hasAnnot(afunc.Annotations(), "extern") {
		name = afunc.Name
		linkage |= ir.External
	} else if hasAnnot(afunc.Annotations(), "entry") {
		// entry has a special name that goes to the linker so we don't want it
		// mangled and we must force it to be public
		name = afunc.Name
		linkage = ir.Public
	} else {
		name = l.globalPrefix + afunc.Name
	}

	if ccName, ok := afunc.Annotations()["callconv"]; ok {
		switch ccName {
		case "win64":
			callConv = ir.Win64CC
		default:
			log.Fatalln("unsupported calling convention")
		}
	}

	// create and append the function decl
	fdecl := &ir.FuncDecl{
		Name:       name,
		Args:       iargs,
		ReturnType: l.lowerType(afunc.Signature.ReturnType),
		Inline:     inline,
		CallConv:   callConv,
	}

	l.b.SymTable[name] = &ir.IRSymbol{
		Linkage: linkage,
		Decl:    fdecl,
	}

	// lower the function body and add a function definition as necessary
	if afunc.Body != nil {
		// set the local function
		l.localFunc = fdecl

		// reset the local identifier counter
		l.localIdentCounter = 0

		// push a new local scope and set it to pop once completed
		l.pushScope()
		defer l.popScope()

		// TODO
	}
}

// -----------------------------------------------------------------------------

// hasAnnot checks if a map of annotations contains an annotation
func hasAnnot(annots map[string]string, annot string) bool {
	_, ok := annots[annot]
	return ok
}
