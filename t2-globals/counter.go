package globals

import (
	"github.com/skycoin/skycoin-cx/t1-constants" // thin: only EndOfNamedOpcodes; core, keep
)

var (
	// Unused since opcode numbers are fixed in constants.OpcodeLifecycleTable.
	// Removing it drops t2-globals to tier 1, so it goes with that rename.
	OpCodeSystemCounter int = constants.EndOfNamedOpcodes
)
