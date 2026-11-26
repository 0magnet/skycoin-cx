package opcodes

import (
	"bytes"

	"github.com/skycoin/skycoin-cx/t0-cx-core-floatparse"
	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
	// GCREACH: t3-ast owns the heap and the GC roots. This file allocates strs, which
	// adds to the GC pressure, see docs/GarbageCollector.md.
	"github.com/skycoin/skycoin-cx/t3-ast"
)

// The str opcodes read strs as views into program memory and write new strs
// straight into it, with no Go string in between (#412).

func OpStrToI8(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalInt(inputs[0].StrBytesGet(prgrm), 8)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].I8Set(prgrm, int8(outV0))
}

func OpStrToI16(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalInt(inputs[0].StrBytesGet(prgrm), 16)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].I16Set(prgrm, int16(outV0))
}

func OpStrToI32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalInt(inputs[0].StrBytesGet(prgrm), 32)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].I32Set(prgrm, int32(outV0))
}

func OpStrToI64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalInt(inputs[0].StrBytesGet(prgrm), 64)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].I64Set(prgrm, outV0)
}

func OpStrToUI8(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalUint(inputs[0].StrBytesGet(prgrm), 8)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].UI8Set(prgrm, uint8(outV0))
}

func OpStrToUI16(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalUint(inputs[0].StrBytesGet(prgrm), 16)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].UI16Set(prgrm, uint16(outV0))
}

func OpStrToUI32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalUint(inputs[0].StrBytesGet(prgrm), 32)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].UI32Set(prgrm, uint32(outV0))
}

func OpStrToUI64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, ok := parseDecimalUint(inputs[0].StrBytesGet(prgrm), 64)
	if !ok {
		panic(types.CXRuntimeError)
	}
	outputs[0].UI64Set(prgrm, outV0)
}

func OpStrToF32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, result := floatparse.FloatParseBytes(inputs[0].StrBytesGet(prgrm), 32)
	if result != floatparse.FloatParseResultOK {
		panic(types.CXRuntimeError)
	}
	outputs[0].F32Set(prgrm, float32(outV0))
}

func OpStrToF64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outV0, result := floatparse.FloatParseBytes(inputs[0].StrBytesGet(prgrm), 64)
	if result != floatparse.FloatParseResultOK {
		panic(types.CXRuntimeError)
	}
	outputs[0].F64Set(prgrm, float64(outV0))
}

// VIOLATION VMALLOC-022: appending grows the format buffer past its size, once, #407.
func OpStrPrint(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	printFormattedLine(prgrm, append(formatBuffer(prgrm), inputs[0].StrBytesGet(prgrm)...))
}

func OpStrConcat(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].StrConcatSet(prgrm, inputs[0].StrBytesGet(prgrm), inputs[1].StrBytesGet(prgrm))
}

func OpStrSubstr(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	str := inputs[0].StrBytesGet(prgrm)
	begin := inputs[1].I32Get(prgrm)
	end := inputs[2].I32Get(prgrm)
	if begin < 0 || end < begin || int(end) > len(str) {
		panic(types.CXRuntimeSliceIndexOutOfRange)
	}
	outputs[0].StrBytesSet(prgrm, str[begin:end])
}

func OpStrIndex(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, int32(bytes.Index(inputs[0].StrBytesGet(prgrm), inputs[1].StrBytesGet(prgrm))))
}

func OpStrLastIndex(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, int32(bytes.LastIndex(inputs[0].StrBytesGet(prgrm), inputs[1].StrBytesGet(prgrm))))
}

func OpStrTrimSpace(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].StrBytesSet(prgrm, bytes.TrimSpace(inputs[0].StrBytesGet(prgrm)))
}

func OpStrEqualByValue(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].BoolSet(prgrm, bytes.Equal(inputs[0].StrBytesGet(prgrm), inputs[1].StrBytesGet(prgrm)))
}

func OpStrEqualByPointer(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].BoolSet(prgrm, inputs[0].PtrGet(prgrm) == inputs[1].PtrGet(prgrm))
}

func OpStrCompareByValue(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, int32(bytes.Compare(inputs[0].StrBytesGet(prgrm), inputs[1].StrBytesGet(prgrm))))
}

func OpStrByteLength(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I32Set(prgrm, int32(len(inputs[0].StrBytesGet(prgrm))))
}

// OpStrBytesOfString returns a new []ui8 holding the bytes of a str, as
// []byte(s) in Go.
// GCREACH: allocates the output byte slice, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func OpStrBytesOfString(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	str := inputs[0].StrBytesGet(prgrm)
	count := types.IntToPtrCast(len(str))
	if count == 0 {
		outputs[0].PtrSet(prgrm, 0)
		return
	}
	if count > ast.SliceElementLimit(prgrm, 1) {
		ast.HeapOutOfMemory(prgrm, uint64(types.ObjectHeaderSize+constants.SliceHeaderSize)+uint64(count))
	}
	slice := ast.SliceResizeEx(prgrm, 0, count, 1)
	copy(ast.SliceElementsGet(prgrm, slice, 1), str)
	outputs[0].PtrSet(prgrm, slice)
}

// OpStrStringOfBytes returns a new str holding the bytes of a []ui8, as
// string(b) in Go.
func OpStrStringOfBytes(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].StrBytesSet(prgrm, ast.SliceElementsGet(prgrm, inputs[0].PtrGet(prgrm), 1))
}
