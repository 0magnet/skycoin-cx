package netopcodes

import (
	"github.com/skycoin/skycoin-cx/t1-cx-library-networkbatch"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// PackageRegister adds the net package to prgrm. The functions of sys/net
// are the opcodes of constants.OpcodeGroupNet, which NetOpcodeDispatch runs
// on the old loop and the vm runs itself, see docs/Networking.md.
func PackageRegister(prgrm *ast.CXProgram) {
	networkbatch.NetPrepare()

	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now a constructor call, site S1427.
	netPkg := ast.PackageMake(prgrm.CXProgramNameIntern("net"))
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now an arena insert, site S1428.
	prgrm.PackageAdd(netPkg)
}

// The constant codes of the net.Trigger* and net.Error* constants.
const (
	constNetTriggerFirst = iota + 0x2FFFF
	constNetErrorFirst   = constNetTriggerFirst + int(networkbatch.NetTriggerCount)
)

// ConstantsInit registers the net.Trigger* and net.Error* constants through
// constantAdd, actions.ConstI32Add in cx, which keeps the compiler out of
// the cx build core.
func ConstantsInit(constantAdd func(code int, name string, value int32)) {
	triggers := [networkbatch.NetTriggerCount]string{"None", "Data", "Accepted", "Connected", "Listening", "Closed", "Error", "Timeout", "Dropped"}
	for value, name := range triggers {
		constantAdd(constNetTriggerFirst+value, "net.Trigger"+name, int32(value))
	}
	errors := [networkbatch.NetErrorCount]string{"None", "Other", "Refused", "Reset", "Timeout", "Address", "AddressInUse", "SocketTableFull"}
	for value, name := range errors {
		constantAdd(constNetErrorFirst+value, "net.Error"+name, int32(value))
	}
}
