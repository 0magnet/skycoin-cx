package main

import (
	"strings"
	"testing"

	lower "github.com/skycoin/skycoin-cx/t4-cx-compiler-lower"
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-interpreter"
)

const operationRefusedSource = `package main

import "sys/os"

func main() {
	test(os.Close(-1), false, "os.Close of a handle that is not open")
}
`

// An embedder that runs a program without core.ExecutorRun, as the Android
// playground did, ran a file opcode its run did not enable (D119).
func TestOperationRefusedOutsideExecutorRun(t *testing.T) {
	prgrm := serializeCompile(t, "refused.cx", operationRefusedSource)
	if prgrm.CXProgramOperationRefused() == 0 {
		t.Fatal("os.Close is not refused without --enable=file")
	}
	image, failures := lower.Program(prgrm)
	if len(failures) != 1 || failures[0].FailureReason != lower.ReasonOperationRefused || len(image.ImageFuncs) != 0 {
		t.Errorf("lower.Program = %d functions, failures %+v, want only ReasonOperationRefused", len(image.ImageFuncs), failures)
	}
	err := interpreter.RunCompiled(prgrm, 0, nil)
	if err == nil || !strings.Contains(err.Error(), "CX_RUNTIME_LIBRARY_DISABLED, library file is not enabled") {
		t.Errorf("interpreter.RunCompiled = %v, want the refusal of the file library", err)
	}
}
