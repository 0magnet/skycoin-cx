//go:build !nocxos

package cxos

import (
	"fmt"
	"os"

	sandbox "github.com/skycoin/skycoin-cx/t1-cx-runtime-sandbox"
	profiling "github.com/skycoin/skycoin-cx/t1-profiling" // thin: only CPUProfileStartOnFile; no depth gain
	ast "github.com/skycoin/skycoin-cx/t3-ast"
)

// VIOLATION: state outside State. Owner: t5-cxos; moves to OsFileMng in phase 1.
var openProfiles = make(map[string]func() error)

// OpStartProfile writes a CPU profile to a program path in the sandbox.
// VIOLATION STRHOT-139: B, hands a path or text to Go or the OS, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpStartProfile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	programPath, fileError := ast.ProgramFileSystemPathRead(prgrm, &inputs[0])
	if fileError != nil {
		fmt.Fprintln(os.Stderr, fileError.FileErrorMessage)
		return
	}
	profRate := int(inputs[1].I32Get(prgrm))
	if profRate == 0 {
		return
	}
	file, fileError := sandbox.FileCreate(prgrm.ProgramSandboxRoot, programPath)
	if fileError != nil {
		fmt.Fprintln(os.Stderr, fileError.FileErrorMessage)
		return
	}
	stop, err := profiling.CPUProfileStartOnFile(file, profRate)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	openProfiles[programPath.Text()] = stop
}

// VIOLATION STRHOT-140: B, hands a path or text to Go or the OS, #412.
func OpStopProfile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	programPath, fileError := ast.ProgramFileSystemPathRead(prgrm, &inputs[0])
	if fileError != nil {
		return
	}
	stop, ok := openProfiles[programPath.Text()]
	if !ok {
		return
	}
	delete(openProfiles, programPath.Text())
	if err := stop(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
