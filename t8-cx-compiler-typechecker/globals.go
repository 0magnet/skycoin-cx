package typechecker

import (
	"regexp"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader" // thin: only File; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t6-cx-compiler-declarationextractor"
)

// Parse Globals
// - takes in globals from cx/cmd/declaration_extractor
// - adds globals to AST
func GlobalsParse(files []*packageloader.File, globals []declarationextractor.GlobalDeclaration) error {
	prgrm := astbuild.AST

	// Range over global declarations and parse
	for _, global := range globals {

		// Get Package
		pkgNameGid := astbuild.AST.CXProgramNameIntern(global.PackageID)
		pkg, err := astbuild.AST.PackageGet(pkgNameGid)

		// If package not in AST
		if err != nil || pkg == nil {

			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now a constructor call, site S1510.
			newPkg := ast.PackageMake(pkgNameGid)
			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now an arena insert, site S1511.
			pkgIdx := astbuild.AST.PackageAdd(newPkg)
			// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
			// Now an arena getter, site S1512.
			newPkg, err := astbuild.AST.PackageFromArrayGet(pkgIdx)

			if err != nil {
				return err
			}

			pkg = newPkg
		}

		astbuild.AST.SelectPackage(pkgNameGid)

		// Read File
		source, err := SourceBytesGet(files, global.FileID)
		if err != nil {
			return err
		}

		// Extract Declaration from file
		reGlobalDeclaration := regexp.MustCompile(`var\s+(\w*)\s+([\*\[\]\w\.]+)(?:\s*\=\s*[\s\S]+\S+){0,1}`)
		globalDeclaration := source[global.StartOffset : global.StartOffset+global.Length]
		globalTokens := reGlobalDeclaration.FindSubmatch(globalDeclaration)

		// add Global as CX Argument to CX Package
		_, globalArg := ast.ArgumentMake(prgrm, prgrm.CXProgramNameIntern(global.GlobalVariableName), prgrm.CXProgramSourceFileIntern(global.FileID), global.LineNumber)
		globalArg.Offset = valuetypes.InvalidPointer
		globalArg.Package = ast.CXPackageIndex(pkg.Index)
		globalArgIdx := ast.CXArgumentIndex(globalArg.Index)

		pkg.GlobalAdd(astbuild.AST, globalArgIdx)

		var declarationSpecifier *ast.CXArgument

		declarationSpecifier, err = DeclarationSpecifierParse(prgrm, globalTokens[2], prgrm.CXProgramSourceFileIntern(global.FileID), global.LineNumber, declarationSpecifier)
		if err != nil {
			return err
		}

		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1515.
		astbuild.GlobalInPackageDeclare(astbuild.AST, nil, astbuild.AST.CXArgFromArrayGet(globalArgIdx), declarationSpecifier, nil, false)

	}
	return nil
}
