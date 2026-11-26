//go:build cxcorelibrarynet

package main

import (
	"os"

	"github.com/skycoin/skycoin-cx/t1-cx-library-networkbatch"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-library-netopcodes"
)

// coreNetReplayEnv names the replay record a built program runs its net
// library from, since a built program takes no cx flags. See
// docs/Networking.md.
const coreNetReplayEnv = "CX_NET_REPLAY"

// coreLibraryNetLinked says that the net library is linked into this core.
const coreLibraryNetLinked = true

// The net library is linked into this core, see coreLibraryRegisters.
func init() {
	coreLibraryRegisters = append(coreLibraryRegisters, netopcodes.PackageRegister)
	if path := os.Getenv(coreNetReplayEnv); path != "" {
		networkbatch.NetRunMode, networkbatch.NetRunRecordPath = networkbatch.NetModeReplay, path
	}
}

// coreOpcodeDispatchNet runs an opcode of the net library.
func coreOpcodeDispatchNet(prgrm *ast.CXProgram, code int, inputs []ast.CXValue, outputs []ast.CXValue) bool {
	return netopcodes.NetOpcodeDispatch(prgrm, code, inputs, outputs)
}
