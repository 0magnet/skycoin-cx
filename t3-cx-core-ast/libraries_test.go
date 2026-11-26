package ast

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// TestLibraryCallSandbox checks the D16 sandbox of the old loop. A library
// is disabled by default (D17), and an enabled call runs its function through
// the switch of its library. The disabled calls, which panic, are the
// test-library-disabled.cx cases of t1-cx-tools-programtest.
func TestLibraryCallSandbox(t *testing.T) {
	if opcodetable.LibraryPolicyEnabledWhenImported {
		t.Fatalf("libraries are enabled by default, decision D17 keeps them disabled")
	}
	prgrm := &CXProgram{}
	for library, enabled := range prgrm.ProgramLibrariesEnabled {
		if enabled {
			t.Errorf("library %d is enabled in a new program", library)
		}
	}
	saved := LibraryDispatches[opcodetable.LibraryNumberFile]
	defer func() { LibraryDispatches[opcodetable.LibraryNumberFile] = saved }()
	var ran []int
	LibraryDispatches[opcodetable.LibraryNumberFile] = func(prgrm *CXProgram, function int, inputs []CXValue, outputs []CXValue) bool {
		ran = append(ran, function)
		return true
	}
	prgrm.ProgramLibrariesEnabled[opcodetable.LibraryNumberFile] = true
	libraryCall(prgrm, &CXFunction{AtomicOPCode: opcodetable.OpLibraryCall, LibraryNumber: opcodetable.LibraryNumberFile, LibraryFunctionNumber: 3}, nil, nil)
	if len(ran) != 1 || ran[0] != 3 {
		t.Errorf("an enabled call ran functions %v, want [3]", ran)
	}
}
