package ast

import "github.com/skycoin/skycoin-cx/t0-types"

// VIOLATION: init chaining. Init 2: the shared ConstCxArg_* templates here and
// in arg_constants_ptr*.go, called from Go package init. Library registration
// reads them.

// Integer Types, CxArg Constants
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgI8 = Param(types.I8)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgI16 = Param(types.I16)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgI32 = Param(types.I32)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgI64 = Param(types.I64)

// Unsigned Integer Types, CxArg Constants
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgUI8 = Param(types.UI8)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgUI16 = Param(types.UI16)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgUI32 = Param(types.UI32)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgUI64 = Param(types.UI64)

// Floating point, f32, f64, CxArg Constants
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgF32 = Param(types.F32)

// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgF64 = Param(types.F64)

// ConstCXArgBool Default bool parameter
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgBool = Param(types.BOOL)

// ConstCXArgStr Default str parameter
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgStr = Param(types.STR)

// ConstCXArgUndType Default und parameter
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgUndType = Param(types.UNDEFINED)

// ConstCXArgFunc is a function reference parameter. The compiler resolves
// the function named by the argument and passes its CXFunctionGid as an i64
// literal, so the opcode never looks a name up, see #432.
// VIOLATION: state outside State. Owner: t3-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgFunc = Param(types.FUNC)
