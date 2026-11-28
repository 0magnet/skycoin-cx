package vm

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t0-cx-library-terminal"
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// terminalOpcodeRun runs the terminal opcode op on the block at offset,
// decision D142. A buffer is the address and size the lowering wrote, empty
// when it lies outside memory.
func terminalOpcodeRun(op ir.Opcode, memory []byte, offset uint32) {
	switch op {
	case ir.OpcodeTerminalArgCount:
		valuetypes.I32Write(memory, valuetypes.Pointer(offset), terminal.ArgCount())
	case ir.OpcodeTerminalArgRead:
		buffer := terminalBuffer(memory, offset+4)
		valuetypes.I32Write(memory, valuetypes.Pointer(offset+4+2*machinePointerSize), terminal.ArgRead(valuetypes.I32Read(memory, valuetypes.Pointer(offset)), buffer))
	case ir.OpcodeTerminalInputLineRead:
		buffer := terminalBuffer(memory, offset)
		valuetypes.I32Write(memory, valuetypes.Pointer(offset+2*machinePointerSize), terminal.InputLineRead(buffer))
	}
}

// terminalBuffer returns the ui8 array whose address and size are at offset.
func terminalBuffer(memory []byte, offset uint32) []byte {
	address := uint64(valuetypes.PtrRead(memory, valuetypes.Pointer(offset)))
	size := uint64(valuetypes.PtrRead(memory, valuetypes.Pointer(offset+machinePointerSize)))
	if address+size > uint64(len(memory)) {
		return nil
	}
	return memory[address : address+size]
}
