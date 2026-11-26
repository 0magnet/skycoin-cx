//go:build !nocxos && (!cxcore || cxcorelibraryfile)

package vm

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t0-cxfiles"
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// fileOpcodeRun runs the file opcode op on the block at offset, decision
// D119. A build with nocxos or a core without cxcorelibraryfile leaves this
// file out, see vm_file_off.go.
func fileOpcodeRun(op ir.Opcode, memory []byte, offset uint32) {
	switch op {
	case ir.OpcodeFileClose:
		valuetypes.BoolWrite(memory, valuetypes.Pointer(offset+4), cxfiles.FileHandleClose(valuetypes.I32Read(memory, valuetypes.Pointer(offset))))
	case ir.OpcodeFileSeek:
		position := cxfiles.FileHandleSeek(valuetypes.I32Read(memory, valuetypes.Pointer(offset)), valuetypes.I64Read(memory, valuetypes.Pointer(offset+4)), valuetypes.I32Read(memory, valuetypes.Pointer(offset+12)))
		valuetypes.I64Write(memory, valuetypes.Pointer(offset+16), position)
	case ir.OpcodeFileSync:
		valuetypes.BoolWrite(memory, valuetypes.Pointer(offset+4), cxfiles.FileHandleSync(valuetypes.I32Read(memory, valuetypes.Pointer(offset))))
	}
}
