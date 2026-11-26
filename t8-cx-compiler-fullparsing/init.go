package fullparsing

import (
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t4-cx-library-netopcodes"  // thin: only ConstantsInit; only for the InitCXCore helper, goes with the init move (A)
	"github.com/skycoin/skycoin-cx/t5-cx-library-fileopcodes" // thin: only ConstantsInit; only for the InitCXCore helper, goes with the init move (A)
	"github.com/skycoin/skycoin-cx/t6-cx-library-graphics"    // thin: only ConstantsInit; only for the InitCXCore helper, goes with the init move (A)
	"github.com/skycoin/skycoin-cx/t7-cx-runtime-libraries"   // thin: only MakeProgram, RegisterPackages; only for the InitCXCore helper, goes with the init move (A), tier 8 to 5
)

// VIOLATION: state outside State. Owner: t8-cx-compiler-fullparsing; moves to CXNativeFunctionMng in phase 2.
var Initialized bool

// VIOLATION: init chaining. Init 4-9: the steps t11-cx Run calls, called from
// the playgrounds, the generator and tests. Not from t11-cx.
func CXCoreInit() {
	if !Initialized {
		astbuild.ConstantsInit()
		fileopcodes.ConstantsInit(astbuild.ConstI32Add)
		netopcodes.ConstantsInit(astbuild.ConstI32Add)
		graphics.ConstantsInit()
		if astbuild.AST == nil {
			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1466 S1467.
			astbuild.AST = libraries.ProgramMake()
			libraries.PackagesRegister(astbuild.AST)
		}

		Initialized = true
	}
}
