//go:build !nocxos

package osopcodes

import (
	"time"

	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func OpTimeUnixMilli(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I64Set(prgrm, makeTimestamp())
}

func OpTimeUnixNano(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I64Set(prgrm, time.Now().UnixNano())
}

func OpTimeSleep(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	time.Sleep(time.Duration(inputs[0].I32Get(prgrm)) * time.Millisecond)
}
