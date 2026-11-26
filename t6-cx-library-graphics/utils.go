//go:build cxfx

package graphics

import (
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-interpreter" // thin: only Callback; no depth gain
)

// OpGlfwFuncI32I32 keeps the function the compiler resolved in the callback
// table and returns its handle for glfw.InvokeCallbackI32I32.
func OpGlfwFuncI32I32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, cxfxCallbackI32I32Add(ast.CXFunctionGid(inputs[0].I64Get(prgrm))))
}

func OpGlfwCallI32I32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	function := cxfxCallbackI32I32Get(inputs[0].I32Get(prgrm))
	cxfxCallbackInputsI32(0, inputs[1].I32Get(prgrm))
	cxfxCallbackInputsI32(1, inputs[2].I32Get(prgrm))
	interpreter.Callback(prgrm, function, cxfxState.CxfxCallbackInputs[:2])
}
