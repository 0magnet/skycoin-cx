//go:build !nocipher

package cipher

import (
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// PackageRegister adds the cipher package to prgrm and makes the hash
// states. The functions of sys/cipher are the opcodes of
// constants.OpcodeGroupCipher, which CipherOpcodeDispatch runs, see
// docs/CipherLibrary.md.
func PackageRegister(prgrm *ast.CXProgram) {
	cipherHashTablePrepare()

	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now a constructor call, site S1359.
	cipherPkg := ast.PackageMake(prgrm.CXProgramNameIntern("cipher"))
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now an arena insert, site S1362.
	prgrm.PackageAdd(cipherPkg)
}
