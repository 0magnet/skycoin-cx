package environment

import (
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable" // thin: only EndOfNamedOpcodes; core, keep
)

var (
	// Unused since opcode numbers are fixed in constants.OpcodeLifecycleTable.
	// Removing it drops t2-cx-core-environment to tier 1, so it goes with that rename.
	OpCodeSystemCounter int = opcodetable.EndOfNamedOpcodes
)
