package ast

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
)

// TestPrintValueInvalidAddress checks that values whose pointers or lengths
// lead outside program memory print printValueInvalid instead of panicking,
// as the call stack does for locals whose declarations never ran.
func TestPrintValueInvalidAddress(t *testing.T) {
	const memorySize = 256
	prgrm := &CXProgram{Memory: make([]byte, memorySize)}
	memory := prgrm.Memory

	// A str with an object of "ab" at 64, and slices at 128.
	valuetypes.ObjectPayloadWrite(memory, 64, valuetypes.ObjectHeaderSize+2, []byte("ab"))
	sliceObject := valuetypes.Pointer(128)
	valuetypes.PtrWrite(memory, sliceObject+valuetypes.ObjectHeaderSize+valuetypes.PointerSize, 1<<20)

	cases := []struct {
		name  string
		value valuetypes.Pointer
		print func(dst []byte) []byte
		want  string
	}{
		{"str nil", 0, func(dst []byte) []byte { return appendAtomic(dst, memory, valuetypes.STR, 0) }, ""},
		{"str", 64, func(dst []byte) []byte { return appendAtomic(dst, memory, valuetypes.STR, 0) }, "ab"},
		{"str past memory", 100000000, func(dst []byte) []byte { return appendAtomic(dst, memory, valuetypes.STR, 0) }, printValueInvalid},
		{"str zero size", 200, func(dst []byte) []byte { return appendAtomic(dst, memory, valuetypes.STR, 0) }, printValueInvalid},
		{"str header at end", memorySize - 1, func(dst []byte) []byte { return appendAtomic(dst, memory, valuetypes.STR, 0) }, printValueInvalid},
		{"i64 past memory", 0, func(dst []byte) []byte { return appendAtomic(dst, memory, valuetypes.I64, memorySize-4) }, printValueInvalid},
		{"pointer past memory", 0, func(dst []byte) []byte { return appendPointerAt(dst, memory, memorySize) }, printValueInvalid},
		{"slice past memory", 100000000, func(dst []byte) []byte {
			return appendValueAt(dst, prgrm, 0, valuetypes.I32, valuetypes.I32Size, nil, []valuetypes.Pointer{0}, false)
		}, printValueInvalid},
		{"slice too long", sliceObject, func(dst []byte) []byte {
			return appendValueAt(dst, prgrm, 0, valuetypes.I32, valuetypes.I32Size, nil, []valuetypes.Pointer{0}, false)
		}, printValueInvalid},
		{"array of str", 100000000, func(dst []byte) []byte {
			return appendValueAt(dst, prgrm, 0, valuetypes.STR, valuetypes.StrSize, nil, []valuetypes.Pointer{2}, false)
		}, "[" + printValueInvalid + ", ]"},
		{"array past memory", 0, func(dst []byte) []byte {
			return appendValueAt(dst, prgrm, memorySize-8, valuetypes.I32, valuetypes.I32Size, nil, []valuetypes.Pointer{4}, false)
		}, printValueInvalid},
	}
	dst := make([]byte, 0, 256)
	for _, tc := range cases {
		valuetypes.PtrWrite(memory, 0, tc.value)
		if got := string(tc.print(dst[:0])); got != tc.want {
			t.Errorf("%s: printed %q, want %q", tc.name, got, tc.want)
		}
		if allocs := testing.AllocsPerRun(10, func() { dst = tc.print(dst[:0]) }); allocs != 0 {
			t.Errorf("%s: %v allocations, want 0", tc.name, allocs)
		}
	}
}
