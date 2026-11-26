package actions

import (
	"fmt"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-cx-core-syspackages" // thin: only CheckImport, GuiPackage, IsDefaultPackage; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// PackageDeclare() switches the current package in the program.
//
// Input arguments description:
//
//	prgrm - a CXProgram that contains all the data and arrays of the program.
//	pkgName - name of the package to declare.
func PackageDeclare(prgrm *ast.CXProgram, pkgName string) {
	pkgNameGid := prgrm.CXProgramNameIntern(pkgName)
	// Add a new package to the program if it's not previously defined.
	if _, err := prgrm.PackageGet(pkgNameGid); err != nil {
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now a constructor call, site S1088.
		pkg := ast.PackageMake(pkgNameGid)
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now an arena insert, site S1089.
		prgrm.PackageAdd(pkg)
	}

	_, err := prgrm.SelectPackage(pkgNameGid)
	if err != nil {
		panic(err)
	}
}

// ImportDeclare()
//
// Input arguments description:
//
//	prgrm - a CXProgram that contains all the data and arrays of the program.
//	impName - name of the import to declare.
//	currentFile - the current cx source code file.
//	lineNo - the current line number from the source code.
func ImportDeclare(prgrm *ast.CXProgram, impName string, currentFile ast.SourceFileGid, lineNo int) {
	ImportDeclareAt(prgrm, impName, currentFile, lineNo, prgrm.CXProgramCompileErrorColumn(currentFile, lineNo))
}

// ImportDeclareAt is ImportDeclare of the path at column, which a local
// named like the package reports as the import it shadows (#486).
func ImportDeclareAt(prgrm *ast.CXProgram, impName string, currentFile ast.SourceFileGid, lineNo int, column int) {
	// Make sure we are inside a package
	pkg, err := prgrm.CurrentPackageGet()
	if err != nil {
		// FIXME: Should give a relevant error message
		panic(err)
	}

	if err := syspackages.ImportCheck(impName); err != nil {
		prgrm.CXProgramCompileErrorAdd(ast.CompileErrorCodeMessage, currentFile, lineNo, ast.CompileErrorOperands{prgrm.CXProgramCompileErrorText(err.Error())})
		return
	}
	if name, found := strings.CutPrefix(impName, syspackages.SysPrefix); found && syspackages.GuiPackage(name) && !syspackages.GuiBuild {
		prgrm.CXProgramCompileErrorAdd(ast.CompileErrorCodeImportGuiBuildMissing, currentFile, lineNo, ast.CompileErrorOperands{prgrm.CXProgramCompileErrorText(impName)})
		return
	}

	// Checking if it's a package in the CX workspace by trying to find a
	// slash (/) in the impName.
	// We start backwards and we stop if we find a slash.
	slashFound := false
	c := len(impName) - 1
	for ; c >= 0; c-- {
		if impName[c] == '/' {
			slashFound = true
			break
		}
	}
	ident := ""
	// If the `impName` has a slash, then we need to strip
	// everything behind the slash and the slash itself.
	if slashFound {
		ident = impName[c+1:]
	} else {
		ident = impName
	}

	identGid := prgrm.CXProgramNameIntern(ident)
	prgrm.CXProgramImportPositionAdd(pkg.Name, identGid, ast.CXProgramImportPosition{ImportFile: currentFile, ImportLine: lineNo, ImportColumn: column})

	// If the package is already imported, then there is nothing more to be done.
	if _, err := pkg.ImportGet(prgrm, identGid); err == nil {
		return
	}

	// If the package is already defined in the program, just add it to
	// the importing package.
	if imp, err := prgrm.PackageGet(identGid); err == nil {
		pkg.ImportAdd(prgrm, imp)
		return
	}

	// All packages are read during the first pass of the compilation.  So
	// if we get here during the 2nd pass, it's either a core package or
	// something is panic-level wrong.
	if syspackages.PackageDefault(ident) {
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now a constructor call, site S1090.
		imp := ast.PackageMake(identGid)
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now an arena insert, site S1091.
		impIdx := prgrm.PackageAdd(imp)
		// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
		// Now an arena getter, site S1092.
		newImp, err := prgrm.PackageFromArrayGet(impIdx)
		if err != nil {
			panic(err)
		}
		pkg.ImportAdd(prgrm, newImp)

		prgrm.CurrentPackage = ast.CXPackageIndex(pkg.Index)
	} else {
		// This should never happen.
		panic(fmt.Sprintf("error: %s:%d: unknown error when trying to read package '%s'", prgrm.CXProgramSourceFileText(currentFile), lineNo, ident))
	}
}
