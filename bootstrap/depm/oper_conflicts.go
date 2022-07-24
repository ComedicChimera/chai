package depm

import (
	"chaic/common"
	"chaic/report"
	"chaic/types"
)

// CheckOperatorConflicts checks a package for operator definition conflicts.
// Any conflicts it encounters are reported.
func CheckOperatorConflicts(pkg *ChaiPackage) {
	// TODO: handle scripts, imports, and prelude

	// Check the package global namespace.
	for _, operators := range pkg.OperatorTable {
		for _, operator := range operators {
			searchForConflicts(pkg, operator.OpRepr, operator.Overloads)
		}
	}
}

// searchForConflicts searches a list of operator overloads for conflicts.
func searchForConflicts(pkg *ChaiPackage, opRepr string, overloads []*common.OperatorOverload) {

	for i, a := range overloads[:len(overloads)-1] {
		for _, b := range overloads[i+1:] {
			if conflictsWith(a.Signature, b.Signature) {
				// We report the overload `b` conflicts with overload `a` since
				// `a` occurs in the overload table before `b`.
				chFile := pkg.Files[b.FileNumber]
				report.ReportCompileError(
					chFile.AbsPath,
					chFile.ReprPath,
					b.DefSpan,
					"multiple overloads for %s defined with conflicting signatures: %s v. %s\n",
					opRepr,
					a.Signature.Repr(),
					b.Signature.Repr(),
				)
			}
		}
	}
}

// conflictsWith returns whether two overload signatures conflict. This function
// assumes that these two signatures are associated with the same operator and
// have the same arity.
func conflictsWith(a, b types.Type) bool {
	if aft, ok := a.(*types.FuncType); ok {
		if bft, ok := a.(*types.FuncType); ok {
			for i, aparam := range aft.ParamTypes {
				if !types.Equals(aparam, bft.ParamTypes[i]) {
					return false
				}
			}

			return true
		}
	}

	// TODO: handle generics
	report.ReportICE("invalid types for operator signatures: %s and %s", a.Repr(), b.Repr())
	return true
}
