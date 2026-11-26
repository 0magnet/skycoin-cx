package astapi_test

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"  // thin: tests use only Code, I16, I32; no depth gain
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable" // thin: tests use only OpAdd, OpDiv, OpSub; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	cxinit "github.com/skycoin/skycoin-cx/t7-init"                         // thin: tests use only MakeProgram, RegisterPackages; tests only, moving these tests up lets astapi drop to t5
	parsingcompletor "github.com/skycoin/skycoin-cx/t8-cxparsingcompletor" // thin: tests use only InitCXCore; tests only, could init with t7-init and t4-opcodes
	"github.com/skycoin/skycoin-cx/t9-cx-core-astapi"
)

func TestASTAPI_Expressions(t *testing.T) {
	var cxprogram *ast.CXProgram

	// Needed for FunctionNativeExpressionAdd
	// because of dependency on cxast.OpNames
	parsingcompletor.CXCoreInit()

	t.Run("make program", func(t *testing.T) {
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Now a constructor call, site S1598.
		cxprogram = cxinit.ProgramMake()
		cxinit.PackagesRegister(cxprogram)
	})

	t.Run("add empty package", func(t *testing.T) {
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now an arena insert, site S1599.
		err := astapi.EmptyPackageAdd(cxprogram, "main")
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add empty function to package", func(t *testing.T) {
		err := astapi.PackageEmptyFunctionAdd(cxprogram, "main", "TestFunction")
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add input to function", func(t *testing.T) {
		err := astapi.FunctionNativeInputAdd(cxprogram, "main", "TestFunction", "inputOne", valuetypes.I32)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add output to function", func(t *testing.T) {
		err := astapi.FunctionNativeOutputAdd(cxprogram, "main", "TestFunction", "outputOne", valuetypes.I16)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add first expression to function", func(t *testing.T) {
		err := astapi.FunctionNativeExpressionAdd(cxprogram, "TestFunction", opcodetable.OpAdd)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add second expression to function", func(t *testing.T) {
		err := astapi.FunctionNativeExpressionAdd(cxprogram, "TestFunction", opcodetable.OpSub)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add an expression to function on a line number", func(t *testing.T) {
		err := astapi.FunctionNativeExpressionAddAtLine(cxprogram, "TestFunction", opcodetable.OpDiv, 1)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})

	t.Run("remove expression from a function", func(t *testing.T) {
		// OBJMODEL: state.CXExpressionMng.CXExpressionRelease(gid), phase 4 #346.
		// Now a release, site S1601.
		err := astapi.FunctionExpressionRemove(cxprogram, "TestFunction", 2)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})
}
