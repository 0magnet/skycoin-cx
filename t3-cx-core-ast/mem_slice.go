package ast

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// TODO: rewrite slice api.
// TODO: remove all offset > 0 tests.
// TODO: panic(types.CX_RUNTIME_*) instead of silent failure.

// SliceIndexValid ...
func SliceIndexValid(prgrm *CXProgram, offset valuetypes.Pointer, index valuetypes.Pointer, sizeofElement valuetypes.Pointer) bool {
	sliceLen := SliceLenGet(prgrm, offset)
	bytesLen := sliceLen * sizeofElement
	index -= valuetypes.ObjectHeaderSize + opcodetable.SliceHeaderSize + offset
	// An index before the data wraps to a large value and fails index < bytesLen.
	if index < bytesLen && (index%sizeofElement) == 0 {
		return true
	}
	return false
}

// SliceOffsetGet ...
func SliceOffsetGet(prgrm *CXProgram, fp valuetypes.Pointer, argTypeSig *CXTypeSignature) valuetypes.Pointer {
	var element *CXArgument
	if argTypeSig.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0386.
		arg := prgrm.CXArgFromArrayGet(CXArgumentIndex(argTypeSig.Meta))
		element = arg.AssignmentElementGet(prgrm)

		if element.SliceTyped() {
			return valuetypes.PtrRead(prgrm.Memory, FinalOffsetGet(prgrm, fp, nil, argTypeSig))
		}
	} else if argTypeSig.Type == TypeAtomic || argTypeSig.Type == TypePointerAtomic {
		return valuetypes.InvalidPointer
	} else if argTypeSig.Type == TypeSliceAtomic || argTypeSig.Type == TypePointerSliceAtomic {
		return valuetypes.PtrRead(prgrm.Memory, FinalOffsetGet(prgrm, fp, nil, argTypeSig))
	} else if argTypeSig.Type == TypeStruct || argTypeSig.Type == TypePointerStruct {
		return valuetypes.PtrRead(prgrm.Memory, FinalOffsetGet(prgrm, fp, nil, argTypeSig))
	} else {
		panic("type is not known")
	}

	return valuetypes.InvalidPointer
}

// SliceHeaderGet ...
func SliceHeaderGet(prgrm *CXProgram, offset valuetypes.Pointer) []byte {
	if offset > 0 && offset.Valid() {
		return prgrm.Memory[offset+valuetypes.ObjectHeaderSize : offset+valuetypes.ObjectHeaderSize+opcodetable.SliceHeaderSize]
	}
	return nil
}

// SliceLenGet ...
func SliceLenGet(prgrm *CXProgram, offset valuetypes.Pointer) valuetypes.Pointer {
	if offset > 0 && offset.Valid() {
		sliceHeader := SliceHeaderGet(prgrm, offset)
		return valuetypes.PtrRead(sliceHeader, valuetypes.PointerSize)
	}
	return 0
}

// SliceGet ...
func SliceGet(prgrm *CXProgram, offset valuetypes.Pointer, sizeofElement valuetypes.Pointer) []byte {
	if offset > 0 && offset.Valid() {
		sliceLen := SliceLenGet(prgrm, offset)
		if sliceLen > 0 && sliceLen.Valid() {
			elementsOffset := offset + valuetypes.ObjectHeaderSize + opcodetable.SliceHeaderSize - valuetypes.PointerSize
			elementsLen := valuetypes.PointerSize + sliceLen*sizeofElement
			return prgrm.Memory[elementsOffset : elementsOffset+elementsLen]
		}
	}
	return nil
}

// SliceElementsGet ...
func SliceElementsGet(prgrm *CXProgram, offset valuetypes.Pointer, sizeofElement valuetypes.Pointer) []byte {
	if slice := SliceGet(prgrm, offset, sizeofElement); slice != nil {
		return slice[valuetypes.PointerSize:]
	}
	return nil
}

// SliceResizeEx does the logic required by `SliceResize`. It is separated because some other functions might have access to the offsets of the slices, but not the `CXArgument`s.
// GCREACH: allocates a larger slice object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func SliceResizeEx(prgrm *CXProgram, outputSliceOffset valuetypes.Pointer, count valuetypes.Pointer, sizeofElement valuetypes.Pointer) valuetypes.Pointer {
	var outputSliceHeader []byte
	var outputSliceCap valuetypes.Pointer

	if outputSliceOffset > 0 && outputSliceOffset.Valid() {
		outputSliceHeader = SliceHeaderGet(prgrm, outputSliceOffset)
		outputSliceCap = valuetypes.PtrRead(outputSliceHeader, 0)
	}

	var newLen = count
	var newCap = outputSliceCap
	if newLen > newCap {
		newCap = sliceCapacityGrow(prgrm, newCap, newLen, sizeofElement)
		var outputObjectSize = valuetypes.ObjectHeaderSize + opcodetable.SliceHeaderSize + newCap*sizeofElement
		outputSliceOffset = SequenceAllocate(prgrm, outputObjectSize)
		valuetypes.PtrWrite(valuetypes.ObjectHeaderGet(prgrm.Memory, outputSliceOffset), valuetypes.ObjectGCHeaderSize, outputObjectSize)

		outputSliceHeader = SliceHeaderGet(prgrm, outputSliceOffset)
		valuetypes.PtrWrite(outputSliceHeader, 0, newCap)
	}

	if outputSliceHeader != nil {
		valuetypes.PtrWrite(outputSliceHeader, valuetypes.PointerSize, newLen)
	}
	return outputSliceOffset
}

// SliceElementLimit is the most elements of sizeofElement bytes that one
// slice object can hold within the heap cap of prgrm.
func SliceElementLimit(prgrm *CXProgram, sizeofElement valuetypes.Pointer) valuetypes.Pointer {
	var limit valuetypes.Pointer
	if heapCap := HeapCap(prgrm); heapCap > valuetypes.ObjectHeaderSize+opcodetable.SliceHeaderSize {
		limit = heapCap - valuetypes.ObjectHeaderSize - opcodetable.SliceHeaderSize
	}
	if sizeofElement > 1 {
		limit /= sizeofElement
	}
	return limit
}

// sliceCapacityGrow doubles capacity up to at least newLen elements. A slice too
// large for the heap is a heap exhausted error, before its size can wrap.
func sliceCapacityGrow(prgrm *CXProgram, capacity valuetypes.Pointer, newLen valuetypes.Pointer, sizeofElement valuetypes.Pointer) valuetypes.Pointer {
	limit := SliceElementLimit(prgrm, sizeofElement)
	if newLen > limit {
		HeapOutOfMemory(prgrm, uint64(valuetypes.ObjectHeaderSize+opcodetable.SliceHeaderSize)+uint64(newLen)*uint64(sizeofElement))
	}
	grown := limit
	if capacity <= limit/2 {
		grown = capacity * 2
	}
	if grown < newLen {
		grown = newLen
	}
	return grown
}

// SliceResize ...
// GCREACH: allocates a larger slice object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func SliceResize(prgrm *CXProgram, fp valuetypes.Pointer, out *CXTypeSignature, inp *CXTypeSignature, count valuetypes.Pointer, sizeofElement valuetypes.Pointer) valuetypes.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)
	outputSliceOffset := SliceResizeEx(prgrm, inputSliceOffset, count, sizeofElement)
	// SliceResizeEx may collect, which moves the input slice.
	inputSliceOffset = SliceOffsetGet(prgrm, fp, inp)

	if outputSliceOffset != inputSliceOffset {
		inputSliceLen := SliceLenGet(prgrm, inputSliceOffset)
		if inputSliceLen > 0 {
			SliceCopy(prgrm, fp, outputSliceOffset, inp, inputSliceLen, sizeofElement)
		}
	}

	return outputSliceOffset
}

// SliceCopyEx does the logic required by `SliceCopy`. It is separated because some other functions might have access to the offsets of the slices, but not the `CXArgument`s.
func SliceCopyEx(prgrm *CXProgram, outputSliceOffset valuetypes.Pointer, inputSliceOffset valuetypes.Pointer, count valuetypes.Pointer, sizeofElement valuetypes.Pointer) {
	var inputSliceLen valuetypes.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	if outputSliceOffset > 0 && outputSliceOffset.Valid() {
		outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
		if (outputSliceOffset != inputSliceOffset) && inputSliceLen > 0 {
			copy(outputSliceElements, SliceElementsGet(prgrm, inputSliceOffset, sizeofElement))
		}
	}
}

// SliceCopy copies the contents from the slice located at `inputSliceOffset` to the slice located at `outputSliceOffset`.
func SliceCopy(prgrm *CXProgram, fp valuetypes.Pointer, outputSliceOffset valuetypes.Pointer, inp *CXTypeSignature, count valuetypes.Pointer, sizeofElement valuetypes.Pointer) {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)
	SliceCopyEx(prgrm, outputSliceOffset, inputSliceOffset, count, sizeofElement)
}

// SliceAppendResize prepares a slice to be able to store a new object of length `sizeofElement`. It checks if the slice needs to be relocated in memory, and if it is needed it relocates it and a new `outputSliceOffset` is calculated for the new slice.
func SliceAppendResize(prgrm *CXProgram, fp valuetypes.Pointer, out *CXTypeSignature, inp *CXTypeSignature, sizeofElement valuetypes.Pointer, appendLen valuetypes.Pointer) valuetypes.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)
	var inputSliceLen valuetypes.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	outputSliceOffset := SliceResize(prgrm, fp, out, inp, inputSliceLen+appendLen, sizeofElement)
	return outputSliceOffset
}

// SliceAppendWrite writes `object` to a slice that is guaranteed to be able to hold `object`, i.e. it had to be checked by `SliceAppendResize` first in case it needed to be resized.
func SliceAppendWrite(prgrm *CXProgram, outputSliceOffset valuetypes.Pointer, object []byte, index valuetypes.Pointer) {
	sizeofElement := valuetypes.IntToPtrCast(len(object))
	outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
	copy(outputSliceElements[index*sizeofElement:], object)
}

// SliceAppendWriteByte writes `object` to a slice that is guaranteed to be able to hold `object`, i.e. it had to be checked by `SliceAppendResize` first in case it needed to be resized.
func SliceAppendWriteByte(prgrm *CXProgram, outputSliceOffset valuetypes.Pointer, object []byte, index valuetypes.Pointer) {
	outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, 1)
	copy(outputSliceElements[index:], object)
}

// SliceInsert ...
func SliceInsert(prgrm *CXProgram, fp valuetypes.Pointer, out *CXTypeSignature, inp *CXTypeSignature, index valuetypes.Pointer, object []byte) valuetypes.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)

	var inputSliceLen valuetypes.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	if index > inputSliceLen {
		panic(valuetypes.CXRuntimeSliceIndexOutOfRange)
	}

	var newLen = inputSliceLen + 1
	sizeofElement := valuetypes.IntToPtrCast(len(object))
	outputSliceOffset := SliceResize(prgrm, fp, out, inp, newLen, sizeofElement)
	outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
	copy(outputSliceElements[(index+1)*sizeofElement:], outputSliceElements[index*sizeofElement:])
	copy(outputSliceElements[index*sizeofElement:], object)
	return outputSliceOffset
}

// SliceRemove removes the element at index in place and returns the input
// slice object, which the result shares with the input as append does.
func SliceRemove(prgrm *CXProgram, fp valuetypes.Pointer, inp *CXTypeSignature, index valuetypes.Pointer, sizeofElement valuetypes.Pointer) valuetypes.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)

	var inputSliceLen valuetypes.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	if index >= inputSliceLen {
		panic(valuetypes.CXRuntimeSliceIndexOutOfRange)
	}

	sliceElements := SliceElementsGet(prgrm, inputSliceOffset, sizeofElement)
	copy(sliceElements[index*sizeofElement:], sliceElements[(index+1)*sizeofElement:])
	valuetypes.PtrWrite(SliceHeaderGet(prgrm, inputSliceOffset), valuetypes.PointerSize, inputSliceLen-1)
	return inputSliceOffset
}

// SliceWrite is used to create slices in the backend, i.e. not by calling `append`
// in a CX program, but rather by the CX code itself. This function is used by
// serialization and to store OS input arguments.
// GCREACH: allocates a larger slice object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func SliceWrite(prgrm *CXProgram, off valuetypes.Pointer, inp []byte) valuetypes.Pointer {
	// TODO: Check all these parses from/to int32/int.
	var inputSliceLen valuetypes.Pointer
	if off != 0 && off.Valid() {
		inputSliceLen = SliceLenGet(prgrm, off)
	}

	inpLen := valuetypes.IntToPtrCast(len(inp))
	// We first check if a resize is needed. If a resize occurred
	// the address of the new slice will be stored in `newOff` and will
	// be different to `off`.
	newOff := SliceResizeEx(prgrm, off, inputSliceLen+1, inpLen)

	// Copy the data from the old slice at `off` to `newOff`.
	SliceCopyEx(prgrm, newOff, off, inputSliceLen, inpLen)

	// Write the new slice element `inp` to the slice located at `newOff`.
	SliceAppendWrite(prgrm, newOff, inp, inputSliceLen)
	return newOff

}
