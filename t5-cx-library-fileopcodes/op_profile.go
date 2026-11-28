//go:build !nocxos

package fileopcodes

import (
	"fmt"
	"os"

	"github.com/skycoin/skycoin-cx/t1-cx-runtime-profiling" // thin: only CPUProfileStartOnFile; no depth gain
	sandbox "github.com/skycoin/skycoin-cx/t1-cx-runtime-sandbox"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// OpStartProfile writes a CPU profile to a program path in the sandbox.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpStartProfile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	programPath, path, fileError := ast.ProgramFileSystemPathRead(prgrm, &inputs[0], &OsLibrary.OsLibraryPool)
	if fileError != nil {
		osProfileFailurePrint(fileError, inputs[0].StrBytesGet(prgrm), nil, true)
		return
	}
	profRate := int(inputs[1].I32Get(prgrm))
	if profRate == 0 {
		return
	}
	file, fileError := sandbox.FileCreate(prgrm.ProgramSandboxRoot, programPath)
	if fileError != nil {
		osProfileFailurePrint(fileError, inputs[0].StrBytesGet(prgrm), nil, true)
		return
	}
	stop, err := profiling.CPUProfileStartOnFile(file, profRate)
	if err != nil {
		osProfileFailurePrint(nil, nil, err, true)
		return
	}
	OsLibrary.OsLibraryProfileMng.OsProfileAdd(path, stop)
}

func OpStopProfile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	_, path, fileError := ast.ProgramFileSystemPathRead(prgrm, &inputs[0], &OsLibrary.OsLibraryPool)
	if fileError != nil {
		return
	}
	profiles := &OsLibrary.OsLibraryProfileMng
	slot := profiles.OsProfileFind(path)
	if slot < 0 {
		return
	}
	stop := profiles.OsProfileMngSlots[slot].OsProfileStop
	profiles.OsProfileMngSlots[slot] = OsProfile{OsProfilePath: stringpool.StringGidNil}
	if err := stop(); err != nil {
		osProfileFailurePrint(nil, nil, err, false)
	}
}

// osProfileFailurePrint prints why a CPU profile did not start or stop to
// stderr: the refused path in pathBytes, the failed file call or the Go
// error of starting or stopping it.
// VIOLATION STRHOT-359: B, Go gives the profile failure only as error text, printed to stderr, #412.
func osProfileFailurePrint(fileError *sandbox.FileError, pathBytes []byte, failure error, starting bool) {
	switch {
	case fileError != nil:
		fmt.Fprintln(os.Stderr, fileError.FileErrorTextOf(pathBytes))
	case starting:
		fmt.Fprintln(os.Stderr, "start CPU profile:", failure)
	default:
		fmt.Fprintln(os.Stderr, failure)
	}
}
