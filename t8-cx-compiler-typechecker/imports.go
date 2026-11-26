package typechecker

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-syspackages" // thin: only IsBuiltinImport; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t6-cx-compiler-declarationextractor"
)

func ImportsParse(imports []declarationextractor.ImportDeclaration) error {

	// Make and add import packages to AST
	for _, imprt := range imports {
		// ImportDeclare adds the built-in packages under their own name.
		if syspackages.ImportBuiltin(imprt.ImportName) {
			continue
		}

		// Get Package
		pkgNameGid := astbuild.AST.CXProgramNameIntern(imprt.ImportName)
		pkg, err := astbuild.AST.PackageGet(pkgNameGid)

		// If package not in AST
		if err != nil || pkg == nil {

			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now a constructor call, site S1516.
			newPkg := ast.PackageMake(pkgNameGid)
			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now an arena insert, site S1517.
			pkgIdx := astbuild.AST.PackageAdd(newPkg)
			// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
			// Now an arena getter, site S1518.
			newPkg, err := astbuild.AST.PackageFromArrayGet(pkgIdx)

			if err != nil {
				return err
			}

			pkg = newPkg
		}
	}

	// Declare import in the correct packages
	for _, imprt := range imports {

		// Get Package
		pkgNameGid := astbuild.AST.CXProgramNameIntern(imprt.PackageID)
		pkg, err := astbuild.AST.PackageGet(pkgNameGid)

		// If package not in AST
		if err != nil || pkg == nil {

			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now a constructor call, site S1519.
			newPkg := ast.PackageMake(pkgNameGid)
			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now an arena insert, site S1520.
			pkgIdx := astbuild.AST.PackageAdd(newPkg)
			// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
			// Now an arena getter, site S1521.
			newPkg, err := astbuild.AST.PackageFromArrayGet(pkgIdx)

			if err != nil {
				return err
			}

			pkg = newPkg
		}

		astbuild.AST.SelectPackage(pkgNameGid)

		astbuild.ImportDeclare(astbuild.AST, imprt.ImportName, astbuild.AST.CXProgramSourceFileIntern(imprt.FileID), imprt.LineNumber)

	}

	return nil

}
