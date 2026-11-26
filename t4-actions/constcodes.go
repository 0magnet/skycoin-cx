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

// constPointerSize is the code of cx.POINTER_SIZE, after the exit codes.
const constPointerSize = types.CX_RUNTIME_NOT_IMPLEMENTED + 1

// nolint typecheck
// ConstantsInit registers the cx.* constants in ConstNames and Constants.
func ConstantsInit() {
	// cx
	ConstI32Add(types.CX_SUCCESS, "cx.SUCCESS", types.CX_SUCCESS)
	ConstI32Add(types.CX_COMPILATION_ERROR, "cx.COMPILATION_ERROR", types.CX_COMPILATION_ERROR)
	ConstI32Add(types.CX_PANIC, "cx.PANIC", types.CX_PANIC)
	ConstI32Add(types.CX_INTERNAL_ERROR, "cx.INTERNAL_ERROR", types.CX_INTERNAL_ERROR)
	ConstI32Add(types.CX_ASSERT, "cx.ASSERT", types.CX_ASSERT)
	ConstI32Add(types.CX_RUNTIME_ERROR, "cx.RUNTIME_ERROR", types.CX_RUNTIME_ERROR)
	ConstI32Add(types.CX_RUNTIME_STACK_OVERFLOW_ERROR, "cx.RUNTIME_STACK_OVERFLOW_ERROR", types.CX_RUNTIME_STACK_OVERFLOW_ERROR)
	ConstI32Add(types.CX_RUNTIME_HEAP_EXHAUSTED_ERROR, "cx.RUNTIME_HEAP_EXHAUSTED_ERROR", types.CX_RUNTIME_HEAP_EXHAUSTED_ERROR)
	ConstI32Add(types.CX_RUNTIME_INVALID_ARGUMENT, "cx.RUNTIME_INVALID_ARGUMENT", types.CX_RUNTIME_INVALID_ARGUMENT)
	ConstI32Add(types.CX_RUNTIME_SLICE_INDEX_OUT_OF_RANGE, "cx.RUNTIME_SLICE_INDEX_OUT_OF_RANGE", types.CX_RUNTIME_SLICE_INDEX_OUT_OF_RANGE)
	ConstI32Add(types.CX_RUNTIME_NOT_IMPLEMENTED, "cx.RUNTIME_NOT_IMPLEMENTED", types.CX_RUNTIME_NOT_IMPLEMENTED)
	ConstI32Add(constPointerSize, "cx.POINTER_SIZE", types.Cast_ptr_to_i32(types.POINTER_SIZE))
}

// ConstCodeAdd ...
func ConstCodeAdd(code int, name string, typeCode types.Code, value []byte) {
	ConstNames.NativeNameIndexAdd(code, name)
	Constants[code] = CXConstant{Type: typeCode, Value: value}
}

// ConstI32Add ...
func ConstI32Add(code int, name string, value int32) {
	var memory [4]byte
	types.Write_i32(memory[:], 0, value)
	ConstCodeAdd(code, name, types.I32, memory[:])
}

// CXConstant ...
type CXConstant struct {
	// native constants. only used for pre-packaged constants (e.g. math package's PI)
	// these fields are used to feed PrimaryWrite
	Value []byte
	Type  types.Code
}
