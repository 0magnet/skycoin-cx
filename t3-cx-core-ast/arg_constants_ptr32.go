//go:build !ptr64

package ast

import "github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"

// VIOLATION: init chaining. Init 2: see arg_constants.go.
// Pointer type, CXArg Constants
// VIOLATION: state outside State. Owner: t3-cx-core-ast; moves to CXArgumentMng in phase 1.
var ConstCXArgPtr = Param(valuetypes.I32) // TODO:PTR use right type when we have ptr alias in cx
