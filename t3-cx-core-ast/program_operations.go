package ast

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// CXProgramOperationRefused returns the lowest library opcode that prgrm
// uses and its run does not enable, and 0 when there is none. Every way to
// run a program checks it, core.ExecutorRun, execute.RunCompiled and
// lower.Program, so no executor runs a refused program (decision D119).
func (prgrm *CXProgram) CXProgramOperationRefused() int {
	refused := 0
	for index := 0; index < prgrm.CXProgramFunctionSlotCount(); index++ {
		fn := prgrm.CXProgramFunctionSlotLive(index)
		if fn == nil {
			continue
		}
		code := fn.AtomicOPCode
		if code == 0 || opcodetable.OpcodeLibraryOf(code) == 0 || prgrm.ProgramOperationsEnabled.OpcodeSetContains(code) {
			continue
		}
		if refused == 0 || code < refused {
			refused = code
		}
	}
	return refused
}

// CXProgramOperationRefusalLine is the line a program refused for opcode
// code prints, the runtime error CX_RUNTIME_LIBRARY_DISABLED without a
// source position. It names the library when the run enables nothing of it,
// and the operation otherwise.
func (prgrm *CXProgram) CXProgramOperationRefusalLine(code int) string {
	library := opcodetable.OpcodeLibraryOf(code)
	message := RuntimeLibraryDisabledMessage(library)
	if prgrm.ProgramOperationsEnabled.OpcodeSetLibraryAny(library) {
		name := opcodetable.LibraryOperationName(code)
		message = "operation " + name + " is not enabled, enable it with --enable=" + name
	}
	return "error: " + ErrorString(valuetypes.CXRuntimeLibraryDisabled) + ", " + message
}
