package ast

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

type CXTypeSignature struct {
	Index   CXTypeSignatureIndex
	Name    stringpool.StringGid
	Offset  valuetypes.Pointer
	Type    CXTypeSignatureType
	Package CXPackageIndex

	PassBy int  // pass by value or reference
	Deref  bool // true if it is dereferencing a pointer, array or slice

	// if type is complex, meta is complex id
	// if type is struct, meta is CXTypeSignature_Struct id
	// if type is array, meta is CXTypeSignature_Array id
	// if type is atomic, meta is the atomic type
	Meta int

	ArgDetails CXArgumentDebugGid // file name and line, shared with the CXArgument it was made from
}
