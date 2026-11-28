// Package terminalopcodes holds the old loop form of the terminal library,
// the opcodes of opcodetable.OpcodeGroupTerminal (decision D142). The vm and
// the native runtime run their own forms on the same t0-cx-library-terminal.
package terminalopcodes

import (
	assert "github.com/skycoin/skycoin-cx/t0-cx-core-assert"
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t0-cx-library-terminal"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// terminalBufferView returns the bytes of the ui8 array value, in place in
// program memory.
func terminalBufferView(prgrm *ast.CXProgram, value *ast.CXValue) []byte {
	signature := value.TypeSignature
	if signature.Type != ast.TypeArrayAtomic || signature.TypeGet(prgrm) != valuetypes.UI8 {
		terminalBufferAssertFailure()
	}
	return prgrm.Memory[value.Offset : value.Offset+signature.ArraySizeGet(prgrm)]
}

// terminalBufferAssertFailure ends a program that passes a buffer that is
// not a ui8 array.
// STRHOT REPORTER: holds the assert text of a buffer that is not a ui8 array, #412.
func terminalBufferAssertFailure() {
	assert.PreconditionTrue(false, "a terminal buffer must be a ui8 array, such as var buffer [256]ui8")
}

func OpOsArgCount(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, terminal.ArgCount())
}

func OpOsArgRead(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, terminal.ArgRead(inputs[0].I32Get(prgrm), terminalBufferView(prgrm, &inputs[1])))
}

func OpOsInputLineRead(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, terminal.InputLineRead(terminalBufferView(prgrm, &inputs[0])))
}
