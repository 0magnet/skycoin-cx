package parsingcompletor

import (
	"github.com/skycoin/skycoin-cx/t4-actions"
	"github.com/skycoin/skycoin-cx/t4-opcodes"     // thin: only RegisterOpcodes; no depth gain
	"github.com/skycoin/skycoin-cx/t5-cxos"        // thin: only InitConstants; only for the InitCXCore helper, goes with the init move (A)
	"github.com/skycoin/skycoin-cx/t6-cxfx"        // thin: only InitConstants; only for the InitCXCore helper, goes with the init move (A)
	cxinit "github.com/skycoin/skycoin-cx/t7-init" // thin: only MakeProgram, RegisterPackages; only for the InitCXCore helper, goes with the init move (A), tier 8 to 5
)

var Initialized bool

// VIOLATION: init chaining. Init 4-9: the steps t11-cx Run calls, called from
// the playgrounds, the generator and tests. Not from t11-cx.
func InitCXCore() {
	if !Initialized {
		actions.InitConstants()
		cxos.InitConstants()
		cxfx.InitConstants()
		if actions.AST == nil {
			actions.AST = cxinit.MakeProgram()
			cxinit.RegisterPackages(actions.AST)
		}
		opcodes.RegisterOpcodes(actions.AST)

		Initialized = true
	}
}
