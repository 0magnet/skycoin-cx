package ast

import "github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"

// CXCall ...
type CXCall struct {
	Operator     CXFunctionGid      // What CX function will be called when running this CXCall in the runtime
	Line         int                // What line in the CX function is currently being executed
	FramePointer valuetypes.Pointer // Where in the stack is this function call's local variables stored
}
