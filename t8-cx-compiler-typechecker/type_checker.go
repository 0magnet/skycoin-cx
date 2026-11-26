package typechecker

import (
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader" // thin: only File; no depth gain
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-printer" // thin: only CompileErrorsPrint; no depth gain
	"github.com/skycoin/skycoin-cx/t6-cx-compiler-declarationextractor"
	cxinit "github.com/skycoin/skycoin-cx/t7-init" // thin: only MakeProgram, RegisterPackages; could take the program as an argument, tier 8 to 7
)

func AllDeclarationsParse(files []*packageloader.File, imports []declarationextractor.ImportDeclaration, globals []declarationextractor.GlobalDeclaration, structs []declarationextractor.StructDeclaration, funcs []declarationextractor.FuncDeclaration) error {

	// Make AST if not made yet
	// VIOLATION: init chaining. Init 7-8 again when actions.AST is nil, called
	// from AllDeclarationsParse. The caller should pass the program in.
	if astbuild.AST == nil {
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1526 S1527.
		astbuild.AST = cxinit.ProgramMake()
		cxinit.PackagesRegister(astbuild.AST)
	}
	defer printer.CompileErrorsPrint(astbuild.AST)

	err := ImportsParse(imports)
	if err != nil {
		return err
	}

	err = StructsParse(files, structs)
	if err != nil {
		return err
	}

	err = GlobalsParse(files, globals)
	if err != nil {
		return err
	}

	err = FuncHeadersParse(files, funcs)
	if err != nil {
		return err
	}

	return nil
}
