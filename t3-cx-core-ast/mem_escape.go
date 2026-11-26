package ast

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// An escaped local has its frame slot offset with EscapeOffsetFlag set, and
// the slot holds its heap object. Every use copies the offset and reads through it.

// OffsetEscaped reports whether offset is the offset of an escaped local.
func OffsetEscaped(offset valuetypes.Pointer) bool {
	return offset&opcodetable.EscapeOffsetFlag != 0
}

// OffsetEscapedSlot returns the frame offset of the slot of an escaped local.
func OffsetEscapedSlot(offset valuetypes.Pointer) valuetypes.Pointer {
	return offset &^ opcodetable.EscapeOffsetFlag
}

// frameAddress returns the address of the variable at offset in the frame at
// fp. An offset outside the stack is an address already.
func frameAddress(prgrm *CXProgram, fp valuetypes.Pointer, offset valuetypes.Pointer) valuetypes.Pointer {
	if OffsetEscaped(offset) {
		return valuetypes.PtrRead(prgrm.Memory, fp+OffsetEscapedSlot(offset)) + valuetypes.ObjectHeaderSize
	}
	if offset < prgrm.Stack.Size {
		return offset + fp
	}
	return offset
}

// FrameAddress is frameAddress for the GC and the error printer.
func FrameAddress(prgrm *CXProgram, fp valuetypes.Pointer, offset valuetypes.Pointer) valuetypes.Pointer {
	return frameAddress(prgrm, fp, offset)
}

// finalOffsetEscapedPointer is FinalOffsetGet of an escaped *T variable.
func finalOffsetEscapedPointer(prgrm *CXProgram, fp valuetypes.Pointer, argTypeSig *CXTypeSignature) valuetypes.Pointer {
	finalOffset := frameAddress(prgrm, fp, argTypeSig.Offset)
	if !argTypeSig.Deref {
		return finalOffset
	}
	finalOffset = valuetypes.PtrRead(prgrm.Memory, finalOffset)
	if finalOffset.Valid() && finalOffset >= prgrm.Heap.StartsAt {
		finalOffset += valuetypes.ObjectHeaderSize
	}
	return finalOffset
}

// EscapedAllocate gives the escaped local at offset in the frame at fp a new
// zeroed heap object of size bytes, as its declaration runs.
// GCREACH: allocates the object of an escaped local, which adds to
// the GC pressure, see docs/GarbageCollector.md.
func EscapedAllocate(prgrm *CXProgram, fp valuetypes.Pointer, offset valuetypes.Pointer, size valuetypes.Pointer) {
	objectSize := valuetypes.ObjectHeaderSize + size
	object := SequenceAllocate(prgrm, objectSize)
	header := valuetypes.ObjectHeaderGet(prgrm.Memory, object)
	for c := range header {
		header[c] = 0
	}
	valuetypes.ObjectSizeWrite(prgrm.Memory, object, objectSize)
	payload := valuetypes.ObjectPayloadGet(prgrm.Memory, object, size)
	for c := range payload {
		payload[c] = 0
	}
	valuetypes.PtrWrite(prgrm.Memory, fp+OffsetEscapedSlot(offset), object)
}

// PointerOfAddress returns the pointer to the value at address. A pointer to
// the heap is the address less the object header, which a dereference adds.
func PointerOfAddress(prgrm *CXProgram, address valuetypes.Pointer) valuetypes.Pointer {
	if address.Valid() && address >= prgrm.Heap.StartsAt+valuetypes.ObjectHeaderSize {
		return address - valuetypes.ObjectHeaderSize
	}
	return address
}

// escapedSize is the size of the heap object of an escaped local, the size
// its frame slot would have had.
func escapedSize(prgrm *CXProgram, sig *CXTypeSignature) valuetypes.Pointer {
	if sig.Type == TypeCXArgumentDeprecate {
		return ArgSizeGet(prgrm, prgrm.CXArgGet(CXArgumentIndex(sig.Meta)))
	}
	return sig.SizeGet(prgrm, true)
}
