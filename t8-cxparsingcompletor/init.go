package parsingcompletor

import (
	"github.com/skycoin/skycoin-cx/t4-actions"
	"github.com/skycoin/skycoin-cx/t4-cxnet"       // thin: only ConstantsInit; only for the InitCXCore helper, goes with the init move (A)
	"github.com/skycoin/skycoin-cx/t5-cxos"        // thin: only ConstantsInit; only for the InitCXCore helper, goes with the init move (A)
	"github.com/skycoin/skycoin-cx/t6-cxfx"        // thin: only ConstantsInit; only for the InitCXCore helper, goes with the init move (A)
	cxinit "github.com/skycoin/skycoin-cx/t7-init" // thin: only MakeProgram, RegisterPackages; only for the InitCXCore helper, goes with the init move (A), tier 8 to 5
)

// VIOLATION: state outside State. Owner: t8-cxparsingcompletor; moves to CXNativeFunctionMng in phase 2.
var Initialized bool

// VIOLATION: init chaining. Init 4-9: the steps t11-cx Run calls, called from
// the playgrounds, the generator and tests. Not from t11-cx.
func CXCoreInit() {
	if !Initialized {
		actions.ConstantsInit()
		cxos.ConstantsInit(actions.ConstI32Add)
		cxnet.ConstantsInit(actions.ConstI32Add)
		cxfx.ConstantsInit()
		if actions.AST == nil {
			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1466 S1467.
			actions.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(actions.AST)
		}

		Initialized = true
	}
}
