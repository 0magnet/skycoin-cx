package opcodes

import (
	"io"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	// GCREACH: t3-cx-core-ast owns the heap and the GC roots. This file allocates strs, which
	// adds to the GC pressure, see docs/GarbageCollector.md.
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// OpReadStdin returns the next line of the program's input without its line
// ending. At the end of input it returns "", see LanguageGuide.md.
// VIOLATION VMALLOC-018: reads a line into Go memory, #407.
func OpReadStdin(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var line []byte
	if prgrm.ProgramStdin != nil {
		var err error
		line, err = prgrm.ProgramStdin.ReadBytes('\n')
		if err != nil && err != io.EOF {
			panic(valuetypes.CXInternalError)
		}
	}

	kept := line[:0]
	for _, b := range line {
		if b != '\n' && b != '\r' {
			kept = append(kept, b)
		}
	}
	outputs[0].StrBytesSet(prgrm, kept)
}

// OpOsOutputFlush writes the buffered output. The old loop writes every
// print as it runs, so it has nothing to write (decision D142).
func OpOsOutputFlush(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
}
