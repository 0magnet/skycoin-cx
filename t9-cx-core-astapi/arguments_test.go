package astapi_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"  // thin: tests use only Code, I16, I32; no depth gain
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable" // thin: tests use only OpAdd, OpDiv, OpSub; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	cxinit "github.com/skycoin/skycoin-cx/t7-init"                         // thin: tests use only MakeProgram, RegisterPackages; tests only, moving these tests up lets astapi drop to t5
	parsingcompletor "github.com/skycoin/skycoin-cx/t8-cxparsingcompletor" // thin: tests use only InitCXCore; tests only, could init with t7-init and t4-opcodes
	"github.com/skycoin/skycoin-cx/t9-cx-core-astapi"
)

func TestASTAPI_Arguments(t *testing.T) {
	var cxprogram *ast.CXProgram

	// Needed for FunctionNativeExpressionAdd
	// because of dependency on cxast.OpNames
	parsingcompletor.CXCoreInit()

	t.Run("make program", func(t *testing.T) {
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Now a constructor call, site S1585.
		cxprogram = cxinit.ProgramMake()
		cxinit.PackagesRegister(cxprogram)
	})

	t.Run("add empty package", func(t *testing.T) {
		// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
		// Now an arena insert, site S1586.
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

	t.Run("add expression to function", func(t *testing.T) {
		err := astapi.FunctionNativeExpressionAdd(cxprogram, "TestFunction", opcodetable.OpAdd)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add first input to expression", func(t *testing.T) {
		buf := new(bytes.Buffer)
		var num int32 = 5
		binary.Write(buf, binary.LittleEndian, num)
		err := astapi.ExpressionLiteralInputAdd(cxprogram, "main", "TestFunction", buf.Bytes(), valuetypes.I16, 1)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add second input to expression", func(t *testing.T) {
		err := astapi.ExpressionNativeInputAdd(cxprogram, "main", "TestFunction", "y", valuetypes.I16, 1)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add output to expression", func(t *testing.T) {
		err := astapi.ExpressionNativeOutputAdd(cxprogram, "main", "TestFunction", "z", valuetypes.I16, 1)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})

	t.Run("make input of an expression a pointer", func(t *testing.T) {
		err := astapi.ExpressionInputPointerMake(cxprogram, "TestFunction", 1, 1)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("make output of an expression a pointer", func(t *testing.T) {
		err := astapi.ExpressionOutputPointerMake(cxprogram, "TestFunction", 1, 0)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})

	// t.Run("remove input from an expression", func(t *testing.T) {
	// err := astapi.RemoveInputFromExpression(cxprogram, "TestFunction", 0)
	// if err != nil {
	// 	t.Errorf("want no error, got %v", err)
	// }
	// })

	// t.Run("remove output from an expression", func(t *testing.T) {
	// 	err := astapi.RemoveOutputFromExpression(cxprogram, "TestFunction", 0)
	// 	if err != nil {
	// 		t.Errorf("want no error, got %v", err)
	// 	}
	// })

	t.Run("print program", func(t *testing.T) {
		cxprogram.ProgramPrint()
	})

	t.Run("add global", func(t *testing.T) {
		_, arg := ast.GlobalMake(cxprogram, cxprogram.CXProgramNameIntern("testGlobal"), valuetypes.I16, ast.SourceFileGidNil, -1)

		// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
		// Now an arena getter, site S1589.
		cxpackage, err := cxprogram.PackageFromArrayGet(cxprogram.Packages[cxprogram.CxIrState.CxIrStateNameMain])
		if err != nil {
			panic(err)
		}

		// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
		// Now an arena insert, site S1590.
		argIdx := cxprogram.CXArgInArrayAdd(arg)
		cxpackage.GlobalAdd(cxprogram, argIdx)
	})

	t.Run("get accessible i16 args", func(t *testing.T) {
		args, err := astapi.FunctionAccessibleArgsByTypeGet(cxprogram, "main", "TestFunction", valuetypes.I16)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}

		got := map[string]bool{}
		for _, arg := range args {
			t.Logf("Arg (name,type)=(%v,%v)\n", cxprogram.CXProgramNameText(arg.Name), valuetypes.Code(arg.Meta))
			got[cxprogram.CXProgramNameText(arg.Name)] = true
		}
		for _, want := range []string{"testGlobal", "y"} {
			if !got[want] {
				t.Errorf("want accessible i16 arg %s, got %v", want, got)
			}
		}
	})
}
