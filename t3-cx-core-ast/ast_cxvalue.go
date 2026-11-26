package ast

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
)

type CXValue struct {
	// OBJMODEL: CXTypeSignatureGid, phase 5 #332.
	// Now a pointer at rest, site S0333.
	TypeSignature *CXTypeSignature // TODO:PTR remove Arg/TypeSignature
	// OBJMODEL: CXExpressionGid, phase 5 #347.
	// Now a pointer at rest, site S0334.
	Expression   *CXExpression
	Type         valuetypes.Code
	Offset       valuetypes.Pointer
	Size         valuetypes.Pointer
	FramePointer valuetypes.Pointer // TODO:PTR remove FramePointer
}

func (value *CXValue) BoolGet(prgrm *CXProgram) bool {
	return valuetypes.BoolRead(prgrm.Memory, value.Offset)
}

func (value *CXValue) I8Get(prgrm *CXProgram) int8 {
	return valuetypes.I8Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) I16Get(prgrm *CXProgram) int16 {
	return valuetypes.I16Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) I32Get(prgrm *CXProgram) int32 {
	return valuetypes.I32Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) I64Get(prgrm *CXProgram) int64 {
	return valuetypes.I64Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI8Get(prgrm *CXProgram) uint8 {
	return valuetypes.UI8Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI16Get(prgrm *CXProgram) uint16 {
	return valuetypes.UI16Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI32Get(prgrm *CXProgram) uint32 {
	return valuetypes.UI32Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI64Get(prgrm *CXProgram) uint64 {
	return valuetypes.UI64Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) F32Get(prgrm *CXProgram) float32 {
	return valuetypes.F32Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) F64Get(prgrm *CXProgram) float64 {
	return valuetypes.F64Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) PtrGet(prgrm *CXProgram) valuetypes.Pointer {
	return valuetypes.PtrRead(prgrm.Memory, value.Offset)
}

func (value *CXValue) BytesGet(prgrm *CXProgram) []byte {
	return valuetypes.SliceByteGet(prgrm.Memory, value.Offset, value.Size)
}

// VIOLATION STRHOT-029: B, only the boundary opcodes use Go strings, #412.
// VIOLATION VMALLOC-003: a Go string copy of a str, for the boundary opcodes only, #407.
func (value *CXValue) StrGet(prgrm *CXProgram) string {
	return valuetypes.StrRead(prgrm.Memory, value.Offset)
}

// StrBytesGet returns the bytes of the str value, a view into prgrm.Memory
// without a copy. An allocation may give prgrm a new Memory, and the view then
// still holds the bytes it had.
func (value *CXValue) StrBytesGet(prgrm *CXProgram) []byte {
	return valuetypes.StrBytesRead(prgrm.Memory, value.Offset)
}

func (value *CXValue) SliceI8Get(prgrm *CXProgram) []int8 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 1); mem != nil {
		return valuetypes.SliceI8Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceI16Get(prgrm *CXProgram) []int16 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 2); mem != nil {
		return valuetypes.SliceI16Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceI32Get(prgrm *CXProgram) []int32 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 4); mem != nil {
		return valuetypes.SliceI32Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceI64Get(prgrm *CXProgram) []int64 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 8); mem != nil {
		return valuetypes.SliceI64Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI8Get(prgrm *CXProgram) []uint8 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 1); mem != nil {
		return valuetypes.SliceUI8Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI16Get(prgrm *CXProgram) []uint16 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 2); mem != nil {
		return valuetypes.SliceUI16Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI32Get(prgrm *CXProgram) []uint32 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 4); mem != nil {
		return valuetypes.SliceUI32Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI64Get(prgrm *CXProgram) []uint64 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 8); mem != nil {
		return valuetypes.SliceUI64Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceF32Get(prgrm *CXProgram) []float32 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 4); mem != nil {
		return valuetypes.SliceF32Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceF64Get(prgrm *CXProgram) []float64 {
	if mem := SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), 8); mem != nil {
		return valuetypes.SliceF64Read(mem, 0)
	}
	return nil
}

// SliceBytesGet returns the bytes of the elements of the slice value.
func (value *CXValue) SliceBytesGet(prgrm *CXProgram) []byte {
	eltSize := value.Size
	if value.TypeSignature != nil {
		eltSize = DerefSizeSliceGet(prgrm, value.TypeSignature)
	}
	return SliceElementsGet(prgrm, valuetypes.PtrRead(prgrm.Memory, value.Offset), eltSize)
}

func (value *CXValue) BoolSet(prgrm *CXProgram, newValue bool) {
	valuetypes.BoolWrite(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I8Set(prgrm *CXProgram, newValue int8) {
	valuetypes.I8Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I16Set(prgrm *CXProgram, newValue int16) {
	valuetypes.I16Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I32Set(prgrm *CXProgram, newValue int32) {
	valuetypes.I32Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I64Set(prgrm *CXProgram, newValue int64) {
	valuetypes.I64Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI8Set(prgrm *CXProgram, newValue uint8) {
	valuetypes.UI8Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI16Set(prgrm *CXProgram, newValue uint16) {
	valuetypes.UI16Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI32Set(prgrm *CXProgram, newValue uint32) {
	valuetypes.UI32Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI64Set(prgrm *CXProgram, newValue uint64) {
	valuetypes.UI64Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) F32Set(prgrm *CXProgram, newValue float32) {
	valuetypes.F32Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) F64Set(prgrm *CXProgram, newValue float64) {
	valuetypes.F64Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) PtrSet(prgrm *CXProgram, newValue valuetypes.Pointer) {
	valuetypes.PtrWrite(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) BytesSet(prgrm *CXProgram, newValue []byte) {
	valuetypes.SliceByteWrite(prgrm.Memory, value.Offset, newValue)
}

// GCREACH: allocates a str object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
// VIOLATION STRHOT-030: B, only the boundary opcodes use Go strings, #412.
func (value *CXValue) StrSet(prgrm *CXProgram, newValue string) {
	StrAllocateAt(prgrm, value.Offset, newValue)
}

// StrBytesSet writes a new str holding data, which may be a view into
// prgrm.Memory.
// GCREACH: allocates a str object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func (value *CXValue) StrBytesSet(prgrm *CXProgram, newValue []byte) {
	StrBytesAllocateAt(prgrm, value.Offset, newValue, nil)
}

// StrConcatSet writes a new str holding first followed by second, either of
// which may be a view into prgrm.Memory.
// GCREACH: allocates a str object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func (value *CXValue) StrConcatSet(prgrm *CXProgram, first []byte, second []byte) {
	StrBytesAllocateAt(prgrm, value.Offset, first, second)
}
