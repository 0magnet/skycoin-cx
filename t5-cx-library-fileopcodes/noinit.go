//go:build nocxos

package fileopcodes

import (
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

func PackageRegister(prgrm *ast.CXProgram) {

}

func ConstantsInit(constantAdd func(code int, name string, value int32)) {}

// LibraryProcessRegister registers nothing in a cx built with nocxos.
func LibraryProcessRegister(prgrm *ast.CXProgram) {}
