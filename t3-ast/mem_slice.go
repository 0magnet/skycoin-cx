package ast

import (
	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
)

// TODO: rewrite slice api.
// TODO: remove all offset > 0 tests.
// TODO: panic(types.CX_RUNTIME_*) instead of silent failure.

// SliceIndexValid ...
func SliceIndexValid(prgrm *CXProgram, offset types.Pointer, index types.Pointer, sizeofElement types.Pointer) bool {
	sliceLen := SliceLenGet(prgrm, offset)
	bytesLen := sliceLen * sizeofElement
	index -= types.ObjectHeaderSize + constants.SliceHeaderSize + offset
	// An index before the data wraps to a large value and fails index < bytesLen.
	if index < bytesLen && (index%sizeofElement) == 0 {
		return true
	}
	return false
}

// SliceOffsetGet ...
func SliceOffsetGet(prgrm *CXProgram, fp types.Pointer, argTypeSig *CXTypeSignature) types.Pointer {
	var element *CXArgument
	if argTypeSig.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0386.
		arg := prgrm.CXArgFromArrayGet(CXArgumentIndex(argTypeSig.Meta))
		element = arg.AssignmentElementGet(prgrm)

		if element.SliceTyped() {
			return types.PtrRead(prgrm.Memory, FinalOffsetGet(prgrm, fp, nil, argTypeSig))
		}
	} else if argTypeSig.Type == TypeAtomic || argTypeSig.Type == TypePointerAtomic {
		return types.InvalidPointer
	} else if argTypeSig.Type == TypeSliceAtomic || argTypeSig.Type == TypePointerSliceAtomic {
		return types.PtrRead(prgrm.Memory, FinalOffsetGet(prgrm, fp, nil, argTypeSig))
	} else if argTypeSig.Type == TypeStruct || argTypeSig.Type == TypePointerStruct {
		return types.PtrRead(prgrm.Memory, FinalOffsetGet(prgrm, fp, nil, argTypeSig))
	} else {
		panic("type is not known")
	}

	return types.InvalidPointer
}

// SliceHeaderGet ...
func SliceHeaderGet(prgrm *CXProgram, offset types.Pointer) []byte {
	if offset > 0 && offset.Valid() {
		return prgrm.Memory[offset+types.ObjectHeaderSize : offset+types.ObjectHeaderSize+constants.SliceHeaderSize]
	}
	return nil
}

// SliceLenGet ...
func SliceLenGet(prgrm *CXProgram, offset types.Pointer) types.Pointer {
	if offset > 0 && offset.Valid() {
		sliceHeader := SliceHeaderGet(prgrm, offset)
		return types.PtrRead(sliceHeader, types.PointerSize)
	}
	return 0
}

// SliceGet ...
func SliceGet(prgrm *CXProgram, offset types.Pointer, sizeofElement types.Pointer) []byte {
	if offset > 0 && offset.Valid() {
		sliceLen := SliceLenGet(prgrm, offset)
		if sliceLen > 0 && sliceLen.Valid() {
			elementsOffset := offset + types.ObjectHeaderSize + constants.SliceHeaderSize - types.PointerSize
			elementsLen := types.PointerSize + sliceLen*sizeofElement
			return prgrm.Memory[elementsOffset : elementsOffset+elementsLen]
		}
	}
	return nil
}

// SliceElementsGet ...
func SliceElementsGet(prgrm *CXProgram, offset types.Pointer, sizeofElement types.Pointer) []byte {
	if slice := SliceGet(prgrm, offset, sizeofElement); slice != nil {
		return slice[types.PointerSize:]
	}
	return nil
}

// SliceResizeEx does the logic required by `SliceResize`. It is separated because some other functions might have access to the offsets of the slices, but not the `CXArgument`s.
// GCREACH: allocates a larger slice object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func SliceResizeEx(prgrm *CXProgram, outputSliceOffset types.Pointer, count types.Pointer, sizeofElement types.Pointer) types.Pointer {
	var outputSliceHeader []byte
	var outputSliceCap types.Pointer

	if outputSliceOffset > 0 && outputSliceOffset.Valid() {
		outputSliceHeader = SliceHeaderGet(prgrm, outputSliceOffset)
		outputSliceCap = types.PtrRead(outputSliceHeader, 0)
	}

	var newLen = count
	var newCap = outputSliceCap
	if newLen > newCap {
		newCap = sliceCapacityGrow(prgrm, newCap, newLen, sizeofElement)
		var outputObjectSize = types.ObjectHeaderSize + constants.SliceHeaderSize + newCap*sizeofElement
		outputSliceOffset = SequenceAllocate(prgrm, outputObjectSize)
		types.PtrWrite(types.ObjectHeaderGet(prgrm.Memory, outputSliceOffset), types.ObjectGCHeaderSize, outputObjectSize)

		outputSliceHeader = SliceHeaderGet(prgrm, outputSliceOffset)
		types.PtrWrite(outputSliceHeader, 0, newCap)
	}

	if outputSliceHeader != nil {
		types.PtrWrite(outputSliceHeader, types.PointerSize, newLen)
	}
	return outputSliceOffset
}

// SliceElementLimit is the most elements of sizeofElement bytes that one
// slice object can hold within the heap cap of prgrm.
func SliceElementLimit(prgrm *CXProgram, sizeofElement types.Pointer) types.Pointer {
	var limit types.Pointer
	if heapCap := HeapCap(prgrm); heapCap > types.ObjectHeaderSize+constants.SliceHeaderSize {
		limit = heapCap - types.ObjectHeaderSize - constants.SliceHeaderSize
	}
	if sizeofElement > 1 {
		limit /= sizeofElement
	}
	return limit
}

// sliceCapacityGrow doubles capacity up to at least newLen elements. A slice too
// large for the heap is a heap exhausted error, before its size can wrap.
func sliceCapacityGrow(prgrm *CXProgram, capacity types.Pointer, newLen types.Pointer, sizeofElement types.Pointer) types.Pointer {
	limit := SliceElementLimit(prgrm, sizeofElement)
	if newLen > limit {
		HeapOutOfMemory(prgrm, uint64(types.ObjectHeaderSize+constants.SliceHeaderSize)+uint64(newLen)*uint64(sizeofElement))
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
func SliceResize(prgrm *CXProgram, fp types.Pointer, out *CXTypeSignature, inp *CXTypeSignature, count types.Pointer, sizeofElement types.Pointer) types.Pointer {
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
func SliceCopyEx(prgrm *CXProgram, outputSliceOffset types.Pointer, inputSliceOffset types.Pointer, count types.Pointer, sizeofElement types.Pointer) {
	var inputSliceLen types.Pointer
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
func SliceCopy(prgrm *CXProgram, fp types.Pointer, outputSliceOffset types.Pointer, inp *CXTypeSignature, count types.Pointer, sizeofElement types.Pointer) {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)
	SliceCopyEx(prgrm, outputSliceOffset, inputSliceOffset, count, sizeofElement)
}

// SliceAppendResize prepares a slice to be able to store a new object of length `sizeofElement`. It checks if the slice needs to be relocated in memory, and if it is needed it relocates it and a new `outputSliceOffset` is calculated for the new slice.
func SliceAppendResize(prgrm *CXProgram, fp types.Pointer, out *CXTypeSignature, inp *CXTypeSignature, sizeofElement types.Pointer, appendLen types.Pointer) types.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)
	var inputSliceLen types.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	outputSliceOffset := SliceResize(prgrm, fp, out, inp, inputSliceLen+appendLen, sizeofElement)
	return outputSliceOffset
}

// SliceAppendWrite writes `object` to a slice that is guaranteed to be able to hold `object`, i.e. it had to be checked by `SliceAppendResize` first in case it needed to be resized.
func SliceAppendWrite(prgrm *CXProgram, outputSliceOffset types.Pointer, object []byte, index types.Pointer) {
	sizeofElement := types.IntToPtrCast(len(object))
	outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
	copy(outputSliceElements[index*sizeofElement:], object)
}

// SliceAppendWriteByte writes `object` to a slice that is guaranteed to be able to hold `object`, i.e. it had to be checked by `SliceAppendResize` first in case it needed to be resized.
func SliceAppendWriteByte(prgrm *CXProgram, outputSliceOffset types.Pointer, object []byte, index types.Pointer) {
	outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, 1)
	copy(outputSliceElements[index:], object)
}

// SliceInsert ...
func SliceInsert(prgrm *CXProgram, fp types.Pointer, out *CXTypeSignature, inp *CXTypeSignature, index types.Pointer, object []byte) types.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)

	var inputSliceLen types.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	if index > inputSliceLen {
		panic(types.CXRuntimeSliceIndexOutOfRange)
	}

	var newLen = inputSliceLen + 1
	sizeofElement := types.IntToPtrCast(len(object))
	outputSliceOffset := SliceResize(prgrm, fp, out, inp, newLen, sizeofElement)
	outputSliceElements := SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
	copy(outputSliceElements[(index+1)*sizeofElement:], outputSliceElements[index*sizeofElement:])
	copy(outputSliceElements[index*sizeofElement:], object)
	return outputSliceOffset
}

// SliceRemove removes the element at index in place and returns the input
// slice object, which the result shares with the input as append does.
func SliceRemove(prgrm *CXProgram, fp types.Pointer, inp *CXTypeSignature, index types.Pointer, sizeofElement types.Pointer) types.Pointer {
	inputSliceOffset := SliceOffsetGet(prgrm, fp, inp)

	var inputSliceLen types.Pointer
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = SliceLenGet(prgrm, inputSliceOffset)
	}

	if index >= inputSliceLen {
		panic(types.CXRuntimeSliceIndexOutOfRange)
	}

	sliceElements := SliceElementsGet(prgrm, inputSliceOffset, sizeofElement)
	copy(sliceElements[index*sizeofElement:], sliceElements[(index+1)*sizeofElement:])
	types.PtrWrite(SliceHeaderGet(prgrm, inputSliceOffset), types.PointerSize, inputSliceLen-1)
	return inputSliceOffset
}

// SliceWrite is used to create slices in the backend, i.e. not by calling `append`
// in a CX program, but rather by the CX code itself. This function is used by
// serialization and to store OS input arguments.
// GCREACH: allocates a larger slice object, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func SliceWrite(prgrm *CXProgram, off types.Pointer, inp []byte) types.Pointer {
	// TODO: Check all these parses from/to int32/int.
	var inputSliceLen types.Pointer
	if off != 0 && off.Valid() {
		inputSliceLen = SliceLenGet(prgrm, off)
	}

	inpLen := types.IntToPtrCast(len(inp))
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
