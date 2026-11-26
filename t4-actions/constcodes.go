package actions

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// For the cxgo. These shouldn't be used in the runtime for performance reasons
var (
	// ConstNames finds a constant by the name Gids of its package and name.
	// VIOLATION: state outside State. Owner: t4-actions; moves to CXConstantMng in phase 1.
	ConstNames ast.NativeNameIndex
	// VIOLATION: state outside State. Owner: t4-actions; moves to CXConstantMng in phase 1.
	Constants = map[int]CXConstant{} //Todo: Only used once in actions/postfix.go
)

// constPointerSize is the code of cx.PointerSize, after the exit codes.
const constPointerSize = valuetypes.CXRuntimeNotImplemented + 1

// nolint typecheck
// ConstantsInit registers the cx.* constants in ConstNames and Constants.
func ConstantsInit() {
	// cx
	ConstI32Add(valuetypes.CXSuccess, "cx.SUCCESS", valuetypes.CXSuccess)
	ConstI32Add(valuetypes.CXCompilationError, "cx.COMPILATION_ERROR", valuetypes.CXCompilationError)
	ConstI32Add(valuetypes.CXPanic, "cx.PANIC", valuetypes.CXPanic)
	ConstI32Add(valuetypes.CXInternalError, "cx.INTERNAL_ERROR", valuetypes.CXInternalError)
	ConstI32Add(valuetypes.CXAssert, "cx.ASSERT", valuetypes.CXAssert)
	ConstI32Add(valuetypes.CXRuntimeError, "cx.RUNTIME_ERROR", valuetypes.CXRuntimeError)
	ConstI32Add(valuetypes.CXRuntimeStackOverflowError, "cx.RUNTIME_STACK_OVERFLOW_ERROR", valuetypes.CXRuntimeStackOverflowError)
	ConstI32Add(valuetypes.CXRuntimeHeapExhaustedError, "cx.RUNTIME_HEAP_EXHAUSTED_ERROR", valuetypes.CXRuntimeHeapExhaustedError)
	ConstI32Add(valuetypes.CXRuntimeInvalidArgument, "cx.RUNTIME_INVALID_ARGUMENT", valuetypes.CXRuntimeInvalidArgument)
	ConstI32Add(valuetypes.CXRuntimeSliceIndexOutOfRange, "cx.RUNTIME_SLICE_INDEX_OUT_OF_RANGE", valuetypes.CXRuntimeSliceIndexOutOfRange)
	ConstI32Add(valuetypes.CXRuntimeNotImplemented, "cx.RUNTIME_NOT_IMPLEMENTED", valuetypes.CXRuntimeNotImplemented)
	ConstI32Add(constPointerSize, "cx.POINTER_SIZE", valuetypes.PtrToI32Cast(valuetypes.PointerSize))
}

// ConstCodeAdd ...
func ConstCodeAdd(code int, name string, typeCode valuetypes.Code, value []byte) {
	ConstNames.NativeNameIndexAdd(code, name)
	Constants[code] = CXConstant{Type: typeCode, Value: value}
}

// ConstI32Add ...
func ConstI32Add(code int, name string, value int32) {
	var memory [4]byte
	valuetypes.I32Write(memory[:], 0, value)
	ConstCodeAdd(code, name, valuetypes.I32, memory[:])
}

// CXConstant ...
type CXConstant struct {
	// native constants. only used for pre-packaged constants (e.g. math package's PI)
	// these fields are used to feed PrimaryWrite
	Value []byte
	Type  valuetypes.Code
}
