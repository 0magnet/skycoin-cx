//go:build cxcorelibraryprocess

package main

import (
	"github.com/skycoin/skycoin-cx/t5-cx-library-fileopcodes"
)

// The process library is linked into this core, see coreLibraryRegisters.
func init() {
	coreLibraryRegisters = append(coreLibraryRegisters, fileopcodes.LibraryProcessRegister)
}
