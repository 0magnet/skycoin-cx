package actions

import (
	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t3-ast"
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
const constPointerSize = types.CXRuntimeNotImplemented + 1

// nolint typecheck
// ConstantsInit registers the cx.* constants in ConstNames and Constants.
func ConstantsInit() {
	// cx
	ConstI32Add(types.CXSuccess, "cx.SUCCESS", types.CXSuccess)
	ConstI32Add(types.CXCompilationError, "cx.COMPILATION_ERROR", types.CXCompilationError)
	ConstI32Add(types.CXPanic, "cx.PANIC", types.CXPanic)
	ConstI32Add(types.CXInternalError, "cx.INTERNAL_ERROR", types.CXInternalError)
	ConstI32Add(types.CXAssert, "cx.ASSERT", types.CXAssert)
	ConstI32Add(types.CXRuntimeError, "cx.RUNTIME_ERROR", types.CXRuntimeError)
	ConstI32Add(types.CXRuntimeStackOverflowError, "cx.RUNTIME_STACK_OVERFLOW_ERROR", types.CXRuntimeStackOverflowError)
	ConstI32Add(types.CXRuntimeHeapExhaustedError, "cx.RUNTIME_HEAP_EXHAUSTED_ERROR", types.CXRuntimeHeapExhaustedError)
	ConstI32Add(types.CXRuntimeInvalidArgument, "cx.RUNTIME_INVALID_ARGUMENT", types.CXRuntimeInvalidArgument)
	ConstI32Add(types.CXRuntimeSliceIndexOutOfRange, "cx.RUNTIME_SLICE_INDEX_OUT_OF_RANGE", types.CXRuntimeSliceIndexOutOfRange)
	ConstI32Add(types.CXRuntimeNotImplemented, "cx.RUNTIME_NOT_IMPLEMENTED", types.CXRuntimeNotImplemented)
	ConstI32Add(constPointerSize, "cx.POINTER_SIZE", types.PtrToI32Cast(types.PointerSize))
}

// ConstCodeAdd ...
func ConstCodeAdd(code int, name string, typeCode types.Code, value []byte) {
	ConstNames.NativeNameIndexAdd(code, name)
	Constants[code] = CXConstant{Type: typeCode, Value: value}
}

// ConstI32Add ...
func ConstI32Add(code int, name string, value int32) {
	var memory [4]byte
	types.I32Write(memory[:], 0, value)
	ConstCodeAdd(code, name, types.I32, memory[:])
}

// CXConstant ...
type CXConstant struct {
	// native constants. only used for pre-packaged constants (e.g. math package's PI)
	// these fields are used to feed PrimaryWrite
	Value []byte
	Type  types.Code
}
