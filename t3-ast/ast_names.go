package ast

import stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"

// CXProgramNameIntern interns a name in the string pool of the program, see
// #413. It is called while compiling only.
func (prgrm *CXProgram) CXProgramNameIntern(text string) stringpool.StringGid {
	return prgrm.CxIrState.StringPool.StringPoolIntern([]byte(text))
}

// CXProgramNameBytes returns the bytes of a name. Only boundary code calls it,
// diagnostics, printing, the serializer and the REPL. The caller must not
// write to them.
func (prgrm *CXProgram) CXProgramNameBytes(gid stringpool.StringGid) []byte {
	return prgrm.CxIrState.StringPool.StringPoolBytes(gid)
}

// CXProgramNameText returns a name as a Go string that views the pool, for
// boundary code that still takes strings. It does not allocate.
func (prgrm *CXProgram) CXProgramNameText(gid stringpool.StringGid) string {
	return prgrm.CxIrState.StringPool.StringPoolText(gid)
}
