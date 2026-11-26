package ast

import (
	"github.com/skycoin/skycoin-cx/t0-types"
)

type CXValue struct {
	// OBJMODEL: CXTypeSignatureGid, phase 5 #332.
	// Now a pointer at rest, site S0333.
	TypeSignature *CXTypeSignature // TODO:PTR remove Arg/TypeSignature
	// OBJMODEL: CXExpressionGid, phase 5 #347.
	// Now a pointer at rest, site S0334.
	Expression   *CXExpression
	Type         types.Code
	Offset       types.Pointer
	Size         types.Pointer
	FramePointer types.Pointer // TODO:PTR remove FramePointer
}

func (value *CXValue) BoolGet(prgrm *CXProgram) bool {
	return types.BoolRead(prgrm.Memory, value.Offset)
}

func (value *CXValue) I8Get(prgrm *CXProgram) int8 {
	return types.I8Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) I16Get(prgrm *CXProgram) int16 {
	return types.I16Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) I32Get(prgrm *CXProgram) int32 {
	return types.I32Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) I64Get(prgrm *CXProgram) int64 {
	return types.I64Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI8Get(prgrm *CXProgram) uint8 {
	return types.UI8Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI16Get(prgrm *CXProgram) uint16 {
	return types.UI16Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI32Get(prgrm *CXProgram) uint32 {
	return types.UI32Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) UI64Get(prgrm *CXProgram) uint64 {
	return types.UI64Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) F32Get(prgrm *CXProgram) float32 {
	return types.F32Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) F64Get(prgrm *CXProgram) float64 {
	return types.F64Read(prgrm.Memory, value.Offset)
}

func (value *CXValue) PtrGet(prgrm *CXProgram) types.Pointer {
	return types.PtrRead(prgrm.Memory, value.Offset)
}

func (value *CXValue) BytesGet(prgrm *CXProgram) []byte {
	return types.SliceByteGet(prgrm.Memory, value.Offset, value.Size)
}

// VIOLATION STRHOT-029: B, only the boundary opcodes use Go strings, #412.
// VIOLATION VMALLOC-003: a Go string copy of a str, for the boundary opcodes only, #407.
func (value *CXValue) StrGet(prgrm *CXProgram) string {
	return types.StrRead(prgrm.Memory, value.Offset)
}

// StrBytesGet returns the bytes of the str value, a view into prgrm.Memory
// without a copy. An allocation may give prgrm a new Memory, and the view then
// still holds the bytes it had.
func (value *CXValue) StrBytesGet(prgrm *CXProgram) []byte {
	return types.StrBytesRead(prgrm.Memory, value.Offset)
}

func (value *CXValue) SliceI8Get(prgrm *CXProgram) []int8 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 1); mem != nil {
		return types.SliceI8Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceI16Get(prgrm *CXProgram) []int16 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 2); mem != nil {
		return types.SliceI16Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceI32Get(prgrm *CXProgram) []int32 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 4); mem != nil {
		return types.SliceI32Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceI64Get(prgrm *CXProgram) []int64 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 8); mem != nil {
		return types.SliceI64Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI8Get(prgrm *CXProgram) []uint8 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 1); mem != nil {
		return types.SliceUI8Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI16Get(prgrm *CXProgram) []uint16 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 2); mem != nil {
		return types.SliceUI16Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI32Get(prgrm *CXProgram) []uint32 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 4); mem != nil {
		return types.SliceUI32Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceUI64Get(prgrm *CXProgram) []uint64 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 8); mem != nil {
		return types.SliceUI64Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceF32Get(prgrm *CXProgram) []float32 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 4); mem != nil {
		return types.SliceF32Read(mem, 0)
	}
	return nil
}

func (value *CXValue) SliceF64Get(prgrm *CXProgram) []float64 {
	if mem := SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), 8); mem != nil {
		return types.SliceF64Read(mem, 0)
	}
	return nil
}

// SliceBytesGet returns the bytes of the elements of the slice value.
func (value *CXValue) SliceBytesGet(prgrm *CXProgram) []byte {
	eltSize := value.Size
	if value.TypeSignature != nil {
		eltSize = DerefSizeSliceGet(prgrm, value.TypeSignature)
	}
	return SliceElementsGet(prgrm, types.PtrRead(prgrm.Memory, value.Offset), eltSize)
}

func (value *CXValue) BoolSet(prgrm *CXProgram, newValue bool) {
	types.BoolWrite(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I8Set(prgrm *CXProgram, newValue int8) {
	types.I8Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I16Set(prgrm *CXProgram, newValue int16) {
	types.I16Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I32Set(prgrm *CXProgram, newValue int32) {
	types.I32Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) I64Set(prgrm *CXProgram, newValue int64) {
	types.I64Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI8Set(prgrm *CXProgram, newValue uint8) {
	types.UI8Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI16Set(prgrm *CXProgram, newValue uint16) {
	types.UI16Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI32Set(prgrm *CXProgram, newValue uint32) {
	types.UI32Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) UI64Set(prgrm *CXProgram, newValue uint64) {
	types.UI64Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) F32Set(prgrm *CXProgram, newValue float32) {
	types.F32Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) F64Set(prgrm *CXProgram, newValue float64) {
	types.F64Write(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) PtrSet(prgrm *CXProgram, newValue types.Pointer) {
	types.PtrWrite(prgrm.Memory, value.Offset, newValue)
}

func (value *CXValue) BytesSet(prgrm *CXProgram, newValue []byte) {
	types.SliceByteWrite(prgrm.Memory, value.Offset, newValue)
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
