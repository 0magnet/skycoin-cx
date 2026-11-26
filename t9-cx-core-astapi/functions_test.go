package astapi_test

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes" // thin: tests use only Code, I16, I32; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t7-cx-runtime-libraries" // thin: tests use only MakeProgram, RegisterPackages; tests only, moving these tests up lets astapi drop to t5
	"github.com/skycoin/skycoin-cx/t9-cx-core-astapi"
)

func TestASTAPI_Functions(t *testing.T) {
	var cxprogram *ast.CXProgram
	t.Run("make program", func(t *testing.T) {
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Now a constructor call, site S1606.
		cxprogram = libraries.ProgramMake()
		libraries.PackagesRegister(cxprogram)
	})

	t.Run("add empty package", func(t *testing.T) {
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now an arena insert, site S1607.
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

	t.Run("add first input to function", func(t *testing.T) {
		err := astapi.FunctionNativeInputAdd(cxprogram, "main", "TestFunction", "inputOne", valuetypes.I32)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add second input to function", func(t *testing.T) {
		err := astapi.FunctionNativeInputAdd(cxprogram, "main", "TestFunction", "inputTwo", valuetypes.I32)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add first output to function", func(t *testing.T) {
		err := astapi.FunctionNativeOutputAdd(cxprogram, "main", "TestFunction", "outputOne", valuetypes.I16)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add second output to function", func(t *testing.T) {
		err := astapi.FunctionNativeOutputAdd(cxprogram, "main", "TestFunction", "outputTwo", valuetypes.I16)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})

	// t.Run("remove first input of function", func(t *testing.T) {
	// 	err := astapi.RemoveFunctionInput(cxprogram, "TestFunction", "inputOne")
	// 	if err != nil {
	// 		t.Errorf("want no error, got %v", err)
	// 	}
	// })

	// t.Run("remove first output of function", func(t *testing.T) {
	// 	err := astapi.RemoveFunctionOutput(cxprogram, "TestFunction", "outputOne")
	// 	if err != nil {
	// 		t.Errorf("want no error, got %v", err)
	// 	}
	// })

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})

	t.Run("remove function from package", func(t *testing.T) {
		err := astapi.PackageFunctionRemove(cxprogram, "main", "TestFunction")
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})
}
