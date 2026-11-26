package opcodes

import (
	"strconv"

	// GCREACH: t3-ast owns the heap and the GC roots. This file reads and writes values
	// and allocates nothing, see docs/GarbageCollector.md.
	"github.com/skycoin/skycoin-cx/t3-ast"
)

func OpBoolPrint(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	printFormattedLine(prgrm, strconv.AppendBool(formatBuffer(prgrm), inputs[0].BoolGet(prgrm)))
}

func OpBoolEqual(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0 := inputs[0].BoolGet(prgrm) == inputs[1].BoolGet(prgrm)
	outputs[0].BoolSet(prgrm, outV0)
}

func OpBoolUnequal(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0 := inputs[0].BoolGet(prgrm) != inputs[1].BoolGet(prgrm)
	outputs[0].BoolSet(prgrm, outV0)
}

func OpBoolNot(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0 := !inputs[0].BoolGet(prgrm)
	outputs[0].BoolSet(prgrm, outV0)
}

func OpBoolAnd(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	inpV0 := inputs[0].BoolGet(prgrm)
	inpV1 := inputs[1].BoolGet(prgrm)
	outputs[0].BoolSet(prgrm, inpV0 && inpV1)
}

func OpBoolOr(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	inpV0 := inputs[0].BoolGet(prgrm)
	inpV1 := inputs[1].BoolGet(prgrm)
	outputs[0].BoolSet(prgrm, inpV0 || inpV1)
}
