package lower

import (
	"chai/depm"
	"chai/ir"
	"strconv"
)

// Lowerer converts a package into an IR bundle.
type Lowerer struct {
	pkg *depm.ChaiPackage
	b   *ir.Bundle

	// globalPrefix is the string prefix prepended to all global symbols of this
	// package to prevent linking errors and preserve package namespaces.
	globalPrefix string

	// currfile is the current file being lowered.
	currfile *depm.ChaiFile

	// localFunc is the reference to the current function being compiled.
	localFunc *ir.FuncDecl

	// scopes is the scope stack of local symbols.  This is the mechanism used
	// to map the high-level scope structure of Chai onto the flat, low level
	// scope structure of the IR.
	scopes []map[string]symbolValue

	// localIdentCounter is used to generate local identifier names.
	localIdentCounter int

	// internedStrings stores the values of the string literal values interned
	// in the current IR bundle.
	internedStrings map[string]*ir.GlobalVar
}

// symbolValue is a data type used to store the value of a locally or
// globally defined symbol so that it can be accessed during IR generation.
type symbolValue struct {
	Value ir.Value

	// IsMutable indicates whether or not this symbol value is a variable
	// pointer that must be loaded before it can be used.
	IsMutable bool
}

// NewLowerer creates a new lowerer for the given package.
func NewLowerer(pkg *depm.ChaiPackage) *Lowerer {
	return &Lowerer{
		pkg:             pkg,
		globalPrefix:    pkg.Parent.Name + pkg.ModSubPath + ".",
		b:               ir.NewBundle(),
		internedStrings: make(map[string]*ir.GlobalVar),
	}
}

// Lower runs the lowerer.
func (l *Lowerer) Lower() *ir.Bundle {
	for _, file := range l.pkg.Files {
		l.currfile = file

		for _, def := range file.Defs {
			l.lowerDef(def)
		}
	}

	return l.b
}

// -----------------------------------------------------------------------------

// pushScope pushes a new local scope.
func (l *Lowerer) pushScope() {
	l.scopes = append(l.scopes, make(map[string]symbolValue))
}

// popScope pops a local scope.
func (l *Lowerer) popScope() {
	l.scopes = l.scopes[:len(l.scopes)-1]
}

// defineLocal defines a new local variable.
func (l *Lowerer) defineLocal(name string, value ir.Value, mut bool) {
	l.scopes[len(l.scopes)-1][name] = symbolValue{
		Value:     value,
		IsMutable: mut,
	}
}

// lookup looks up a symbol value from the local scope.
func (l *Lowerer) lookup(name string) symbolValue {
	// start by traversing local scopes (in reverse order for shadowing)
	for i := len(l.scopes) - 1; i > 0; i-- {
		if sv, ok := l.scopes[i][name]; ok {
			return sv
		}
	}

	// TODO: symbol imports

	// assume it is in the global symbol: return a global ident. we need to add
	// a prefix here since the default lookup won't have one
	gname := l.globalPrefix + name
	gsym := l.b.SymTable[gname]
	return symbolValue{
		&ir.GlobalIdentifier{
			ValueBase: ir.NewValueBase(gsym.Typ),
			Name:      gname,
		},
		gsym.Decl.Section() != ir.SectionText, // global variables are always mutable
	}
}

// -----------------------------------------------------------------------------

// internString creates a new global variable to hold the string's data and
// a new global struct to represent the string value which contains that
// inner pointer.  This
func (l *Lowerer) internString(data string) ir.Value {
	// check for duplicate string entries
	if gv, ok := l.internedStrings[data]; ok {
		return &ir.GlobalIdentifier{
			ValueBase: ir.NewValueBase(ir.PointerType{gv.Typ}),
			Name:      gv.Name,
		}
	}

	// new global variable to store the data
	dataVar := &ir.GlobalVar{
		Name: l.globalPrefix + "__istr.data." + strconv.Itoa(len(l.internedStrings)),
		Val: ir.ConstString{
			Val: data,
		},
	}
	dataVar.Typ = dataVar.Val.Type()
	l.b.SymTable[dataVar.Name] = &ir.IRSymbol{
		Linkage: ir.Private,
		Typ:     ir.PointerType{dataVar.Typ},
		Decl:    dataVar,
	}

	// TODO: new global variable for the structure (struct literal)

	// TODO: mark the string as interned and return a pointer to the
	// global string literal.

	return nil
}
