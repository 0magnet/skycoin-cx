//go:build !noregexp

package regexp

import (
	"github.com/skycoin/skycoin-cx/t0-types" // thin: only Read_str, STR, Write_str; no depth gain
	"github.com/skycoin/skycoin-cx/t3-ast"
)

// RegisterPackage adds the regexp package and its Regexp struct to prgrm. The
// regexp opcodes are rows of constants.OpcodeDefinitionTable.
func RegisterPackage(prgrm *ast.CXProgram) {
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now a constructor call, site S1410.
	regexpPkg := ast.PackageMake(prgrm.CXProgramNameIntern("regexp"))
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now an arena insert, site S1411.
	pkgIdx := prgrm.PackageAdd(regexpPkg)
	// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
	// Now an arena getter, site S1412.
	regexpPkg, _ = prgrm.PackageFromArrayGet(pkgIdx)

	// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
	// Now a constructor call, site S1413.
	regexpStrct := ast.StructMake(prgrm.CXProgramNameIntern("Regexp"))
	_, regexpArg := ast.ArgumentMake(prgrm, prgrm.CXProgramNameIntern("exp"), ast.SourceFileGidNil, 0)
	regexpArg.TypeSet(types.STR).PackageSet(regexpPkg)
	regexpStrct.FieldAdd(prgrm, regexpArg)
	regexpPkg.StructAdd(prgrm, regexpStrct)
}
