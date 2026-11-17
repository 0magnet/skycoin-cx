package astapi_test

import (
	"testing"

	"github.com/skycoin/skycoin-cx/n001_types"
	cxconstants "github.com/skycoin/skycoin-cx/n004_constants"
	cxast "github.com/skycoin/skycoin-cx/n006_ast"
	cxinit "github.com/skycoin/skycoin-cx/n018_init"
	parsingcompletor "github.com/skycoin/skycoin-cx/n020_cxparsingcompletor"
	"github.com/skycoin/skycoin-cx/n022_astapi"
)

func TestASTAPI_Expressions(t *testing.T) {
	var cxprogram *cxast.CXProgram

	// Needed for AddNativeExpressionToFunction
	// because of dependency on cxast.OpNames
	parsingcompletor.InitCXCore()

	t.Run("make program", func(t *testing.T) {
		cxprogram = cxinit.MakeProgram()
	})

	t.Run("add empty package", func(t *testing.T) {
		err := astapi.AddEmptyPackage(cxprogram, "main")
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add empty function to package", func(t *testing.T) {
		err := astapi.AddEmptyFunctionToPackage(cxprogram, "main", "TestFunction")
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add input to function", func(t *testing.T) {
		err := astapi.AddNativeInputToFunction(cxprogram, "main", "TestFunction", "inputOne", types.I32)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add output to function", func(t *testing.T) {
		err := astapi.AddNativeOutputToFunction(cxprogram, "main", "TestFunction", "outputOne", types.I16)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add first expression to function", func(t *testing.T) {
		err := astapi.AddNativeExpressionToFunction(cxprogram, "TestFunction", cxconstants.OP_ADD)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add second expression to function", func(t *testing.T) {
		err := astapi.AddNativeExpressionToFunction(cxprogram, "TestFunction", cxconstants.OP_SUB)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("add an expression to function on a line number", func(t *testing.T) {
		err := astapi.AddNativeExpressionToFunctionByLineNumber(cxprogram, "TestFunction", cxconstants.OP_DIV, 1)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.PrintProgram()
	})

	t.Run("remove expression from a function", func(t *testing.T) {
		err := astapi.RemoveExpressionFromFunction(cxprogram, "TestFunction", 2)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})

	t.Run("print program", func(t *testing.T) {
		cxprogram.PrintProgram()
	})
}
