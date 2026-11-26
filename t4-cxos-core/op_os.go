//go:build !nocxos

package cxoscore

import (
	"os"

	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

func OpOsExit(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	exitCode := inputs[0].I32Get(prgrm)
	os.Exit(int(exitCode))
}
