package astapi

import (
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// PackagesNameListGet returns all names of all packages in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
//
// Example:
// We have this CX Program:
// 0.- Package: main
// 1.- Package: secondpackage
// 2.- Package: thirdpackage
//
// We use PackagesNameListGet(cxprogram).
// The Result will be the list of package names:
// []string{"main", "secondpackage", "thirdpackage"}
func PackagesNameListGet(cxprogram *ast.CXProgram) (list []string) {
	for _, pkgIdx := range cxprogram.Packages {
		// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
		// Now an arena getter, site S1610.
		pkg, err := cxprogram.PackageFromArrayGet(pkgIdx)
		if err != nil {
			panic(err)
		}

		list = append(list, cxprogram.CXProgramNameText(pkg.Name))
	}
	return list
}

// EmptyPackageAdd adds an empty package in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// pkgName - the name of the package to be added.
//
// Example:
// We have an empty program.
// We use EmptyPackageAdd(cxprogram, "main").
// The Result will be:
// 0.- Package: main
//
// Note the new main package added to the cx program.
func EmptyPackageAdd(cxprogram *ast.CXProgram, pkgName string) error {
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now a constructor call, site S1611.
	pkg := ast.PackageMake(cxprogram.CXProgramNameIntern(pkgName))
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now an arena insert, site S1612.
	cxprogram.PackageAdd(pkg)
	return nil
}
