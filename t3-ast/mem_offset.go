package ast

import (
	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
)

type DereferenceStruct struct {
	pointerTyped  bool
	baseOffset    types.Pointer
	sizeofElement types.Pointer
	derefPointer  bool
	idxCounter    int
	// sliceIndexed is set once a slice index is applied.
	sliceIndexed bool

	// OBJMODEL: no change, the struct lives on the stack of one FinalOffsetGet call.
	// Site S0366.
	arg         *CXArgument
	finalOffset types.Pointer
	fp          types.Pointer
	// fault, when set, records a bad index instead of a panic.
	fault *bool
}

// SizeGet ...
func ArgSizeGet(prgrm *CXProgram, arg *CXArgument) types.Pointer {
	if len(arg.Fields) > 0 {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0367.
		return ArgSizeGet(prgrm, prgrm.CXArgFromArrayGet(arg.Fields[len(arg.Fields)-1]))
	}

	derefCount := len(arg.DereferenceOperations)
	if derefCount > 0 {
		deref := arg.DereferenceOperations[derefCount-1]

		switch deref {
		case constants.DerefSlice, constants.DerefArray:
			if direct, behind := arg.DirectDimensions(); behind && len(arg.Indexes) <= direct {
				// e[1] of a [2]*[3]T is a pointer.
				return ValueSizeOfLengths(arg.Lengths[len(arg.Indexes):direct], types.PointerSize)
			}
			if len(arg.Indexes) < len(arg.Lengths) {
				// r[1] of a [2][3]i32 is a [3]i32.
				return ValueSizeOfLengths(arg.Lengths[len(arg.Indexes):], arg.Size)
			}
			return arg.Size
		case constants.DerefPointer:
			return arg.Type.Size()
		}
	}

	// The last specifier is the outermost, so [3]*i32 is an array of pointers.
	for i := len(arg.DeclarationSpecifiers) - 1; i >= 0; i-- {
		switch arg.DeclarationSpecifiers[i] {
		case constants.DeclarationArray:
			return arg.ArraySize()
		case constants.DeclarationPointer:
			return arg.Size
		case constants.DeclarationSlice:
			return types.PointerSize
		}
	}

	if arg.StructType != nil {
		return arg.StructType.StructSizeGet(prgrm)
	}

	return arg.Size
}

func NativeSizeGet(prgrm *CXProgram, arg *CXArgument) types.Pointer {
	derefCount := len(arg.DereferenceOperations)
	if derefCount > 0 {
		deref := arg.DereferenceOperations[derefCount-1]
		switch deref {
		case constants.DerefSlice, constants.DerefArray:
			return arg.Size
		case constants.DerefPointer:
			return arg.Type.Size()
		}
	}

	// The last specifier is the outermost, so [3]*i32 is an array of pointers.
	for i := len(arg.DeclarationSpecifiers) - 1; i >= 0; i-- {
		switch arg.DeclarationSpecifiers[i] {
		case constants.DeclarationArray:
			return arg.ArraySize()
		case constants.DeclarationPointer:
			return arg.Size
		case constants.DeclarationSlice:
			return types.PointerSize
		}
	}

	if arg.StructType != nil {
		return arg.StructType.StructSizeGet(prgrm)
	}

	return arg.Size
}

// DerefSizeGet ...
func DerefSizeGet(prgrm *CXProgram, arg *CXArgument, index int, derefPointer bool, derefArray bool) types.Pointer {
	if direct, behind := arg.DirectDimensions(); behind && index < direct {
		// The elements of the dimensions before a pointer, as in [2]*[3]T,
		// are pointers.
		if derefArray {
			return types.PointerSize
		}
		return ValueSizeOfLengths(arg.Lengths[index+1:direct], types.PointerSize)
	}
	if !derefArray && len(arg.Lengths) > 1 && ((index + 1) < len(arg.Lengths)) {
		if rest := arg.Lengths[index+1:]; rest[0] != 0 {
			// An element of [][N]T is an array.
			return ValueSizeOfLengths(rest, getDerefElementSize(prgrm, arg, derefPointer))
		}
		return types.PointerSize
	}
	return getDerefElementSize(prgrm, arg, derefPointer)
}

// getDerefElementSize returns the size of the innermost elements of arg.
func getDerefElementSize(prgrm *CXProgram, arg *CXArgument, derefPointer bool) types.Pointer {
	if arg.StructType != nil && !cxArgumentElementPointer(arg) {
		return arg.StructType.StructSizeGet(prgrm)
	}
	// &(*p)[i] on a *[N]T ends its specifiers with a pointer but has the size
	// of a T.
	if derefPointer && arg.StructType == nil {
		// p[i] on a *[N]*T reads a pointer.
		if cxArgumentElementPointer(arg) {
			return types.PointerSize
		}
		return arg.Type.Size()
	}
	return arg.Size
}

// DerefSizeSliceGet returns the element size of the slice typeSignature names.
func DerefSizeSliceGet(prgrm *CXProgram, typeSignature *CXTypeSignature) types.Pointer {
	if typeSignature.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0368.
		arg := prgrm.CXArgFromArrayGet(CXArgumentIndex(typeSignature.Meta))
		return derefSizeSliceOfArg(prgrm, arg.AssignmentElementGet(prgrm))
	} else if typeSignature.Type == TypeStruct || typeSignature.Type == TypePointerStruct {
		// A field access such as t.s, sized by its last field, not by the slice header.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)
		if fldLen := len(structDetails.Fields); fldLen > 0 {
			fld := prgrm.CXArgFromArrayGet(structDetails.Fields[fldLen-1])
			return derefSizeSliceOfArg(prgrm, fld.AssignmentElementGet(prgrm))
		}
	} else if typeSignature.Type == TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0369.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		if len(sliceDetails.Lengths) > 1 && (len(sliceDetails.Lengths)-len(sliceDetails.Indexes)) > 1 {
			return types.PointerSize
		}
		return types.Code(sliceDetails.Meta).Size()
	} else if typeSignature.Type == TypePointerSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0370.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		if len(sliceDetails.Lengths) > 1 && (len(sliceDetails.Lengths)-len(sliceDetails.Indexes)) > 1 {
			return types.PointerSize
		}

		return types.Code(sliceDetails.Meta).Size()
	}

	return typeSignature.SizeGet(prgrm, false)
}

// FinalOffsetGet returns the address of the value argTypeSig names in the
// frame at fp. A negative or out of range index panics with its CX error.
func FinalOffsetGet(prgrm *CXProgram, fp types.Pointer, oldArg *CXArgument, argTypeSig *CXTypeSignature) types.Pointer {
	return finalOffsetCompute(prgrm, fp, oldArg, argTypeSig, nil)
}

// FinalOffsetChecked is FinalOffsetGet that returns false where it would
// panic, so the call stack of a runtime error can print the other values.
func FinalOffsetChecked(prgrm *CXProgram, fp types.Pointer, argTypeSig *CXTypeSignature) (types.Pointer, bool) {
	var fault bool
	offset := finalOffsetCompute(prgrm, fp, nil, argTypeSig, &fault)
	return offset, !fault
}

// offsetIndexFromI32 is types.I32ToPtrCast for an index. With fault set a
// negative index sets *fault and counts as 0 instead of panicking.
func offsetIndexFromI32(value int32, fault *bool) types.Pointer {
	if value < 0 && fault != nil {
		*fault = true
		return 0
	}
	return types.I32ToPtrCast(value)
}

// offsetSliceIndexCheck panics with CXRuntimeSliceIndexOutOfRange when
// an index is out of range, or sets *fault when fault is set.
func offsetSliceIndexCheck(valid bool, fault *bool) {
	if valid {
		return
	}
	if fault != nil {
		*fault = true
		return
	}
	panic(types.CXRuntimeSliceIndexOutOfRange)
}

// finalOffsetCompute is FinalOffsetGet. With fault set it records a bad index
// in *fault instead of panicking.
func finalOffsetCompute(prgrm *CXProgram, fp types.Pointer, oldArg *CXArgument, argTypeSig *CXTypeSignature, fault *bool) types.Pointer {
	var finalOffset types.Pointer

	var arg *CXArgument
	if oldArg != nil {
		arg = oldArg
	} else if argTypeSig.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0372.
		arg = prgrm.CXArgGet(CXArgumentIndex(argTypeSig.Meta))
	} else if argTypeSig.Type == TypeAtomic {
		argTypeSigOffset := frameAddress(prgrm, fp, argTypeSig.Offset)

		return argTypeSigOffset
	} else if argTypeSig.Type == TypePointerAtomic {
		if OffsetEscaped(argTypeSig.Offset) {
			return finalOffsetEscapedPointer(prgrm, fp, argTypeSig)
		}
		finalOffset = argTypeSig.Offset

		if argTypeSig.Deref {
			finalOffset = types.PtrRead(prgrm.Memory, finalOffset)
		}

		//Todo: find way to eliminate this check
		if finalOffset < prgrm.Stack.Size {
			// Then it's in the stack, not in data or heap and we need to consider the frame pointer.
			finalOffset += fp
		}

		if finalOffset.Valid() && finalOffset >= prgrm.Heap.StartsAt {
			// then it's an object
			finalOffset += types.ObjectHeaderSize
		}

		return finalOffset
	} else if argTypeSig.Type == TypeArrayAtomic {
		argTypeSigOffset := frameAddress(prgrm, fp, argTypeSig.Offset)

		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0373.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(argTypeSig.Meta)
		indexesLen := len(arrDetails.Indexes)

		for i := 0; i < indexesLen; i++ {
			var subSize = types.Pointer(1)
			for _, len := range arrDetails.Lengths[i+1:] {
				subSize *= len
			}

			sizeToUse := types.Code(arrDetails.Meta).Size()
			sizeOfElement := subSize * sizeToUse
			// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
			// Now an arena getter, site S0374.
			argTypeSigOffset += offsetIndexFromI32(types.I32Read(prgrm.Memory, finalOffsetCompute(prgrm, fp, nil, prgrm.CXTypeSignatureFromArrayGet(arrDetails.Indexes[i]), fault)), fault) * sizeOfElement
		}

		return argTypeSigOffset
	} else if argTypeSig.Type == TypePointerArrayAtomic {
		finalOffset = frameAddress(prgrm, fp, argTypeSig.Offset)

		if argTypeSig.Deref {
			finalOffset = types.PtrRead(prgrm.Memory, finalOffset)

			// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
			// Now an arena getter, site S0375.
			arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(argTypeSig.Meta)
			indexesLen := len(arrDetails.Indexes)

			for i := 0; i < indexesLen; i++ {
				var subSize = types.Pointer(1)
				for _, len := range arrDetails.Lengths[i+1:] {
					subSize *= len
				}

				sizeToUse := types.Code(arrDetails.Meta).Size()
				sizeOfElement := subSize * sizeToUse
				// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
				// Now an arena getter, site S0376.
				finalOffset += offsetIndexFromI32(types.I32Read(prgrm.Memory, finalOffsetCompute(prgrm, fp, nil, prgrm.CXTypeSignatureFromArrayGet(arrDetails.Indexes[i]), fault)), fault) * sizeOfElement
			}
		}

		if argTypeSig.Deref && finalOffset.Valid() && finalOffset >= prgrm.Heap.StartsAt {
			// then it's an object
			finalOffset += types.ObjectHeaderSize
		}

		return finalOffset
	} else if argTypeSig.Type == TypeSliceAtomic {
		argTypeSigOffset := frameAddress(prgrm, fp, argTypeSig.Offset)

		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0377.
		return finalOffsetSliceIndexes(prgrm, fp, argTypeSigOffset, prgrm.CXTypeSignatureArrayFromArrayGet(argTypeSig.Meta), fault)
	} else if argTypeSig.Type == TypePointerSliceAtomic {
		argTypeSigOffset := frameAddress(prgrm, fp, argTypeSig.Offset)

		if argTypeSig.Deref {
			argTypeSigOffset = types.PtrRead(prgrm.Memory, argTypeSigOffset)
			if argTypeSigOffset.Valid() && argTypeSigOffset >= prgrm.Heap.StartsAt {
				// p points to a slice variable in a heap object.
				argTypeSigOffset += types.ObjectHeaderSize
			}

			// (*p)[i] indexes the slice that p points to.
			sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(argTypeSig.Meta)
			if len(sliceDetails.Indexes) > 0 {
				return finalOffsetSliceIndexes(prgrm, fp, argTypeSigOffset, sliceDetails, fault)
			}
		}

		return argTypeSigOffset
	} else if argTypeSig.Type == TypeStruct {
		argTypeSigOffset := frameAddress(prgrm, fp, argTypeSig.Offset)

		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0379.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(argTypeSig.Meta)
		for _, fldIdx := range structDetails.Fields {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S0380.
			fld := prgrm.CXArgFromArrayGet(fldIdx)
			argTypeSigOffset += fld.Offset
			argTypeSigOffset = calculateDereferencesCompute(prgrm, fld, argTypeSigOffset, fp, fault)
		}

		return argTypeSigOffset
	} else if argTypeSig.Type == TypePointerStruct {
		argTypeSigOffset := frameAddress(prgrm, fp, argTypeSig.Offset)

		if argTypeSig.Deref {
			argTypeSigOffset = types.PtrRead(prgrm.Memory, argTypeSigOffset)
			if argTypeSigOffset.Valid() && argTypeSigOffset >= prgrm.Heap.StartsAt {
				// then it's an object
				argTypeSigOffset += types.ObjectHeaderSize
			}
		}

		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0381.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(argTypeSig.Meta)
		for _, fldIdx := range structDetails.Fields {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S0382.
			fld := prgrm.CXArgFromArrayGet(fldIdx)
			argTypeSigOffset += fld.Offset
			argTypeSigOffset = calculateDereferencesCompute(prgrm, fld, argTypeSigOffset, fp, fault)
		}

		return argTypeSigOffset
	}

	if arg == nil {
		// A signature of another type has no offset or dereferences of its own.
		if prgrm.Stack.Size > 0 {
			return fp
		}
		return 0
	}

	finalOffset = frameAddress(prgrm, fp, arg.Offset)

	// elt = arg
	//TODO: Eliminate all op codes with more than one return type
	//TODO: Eliminate this loop
	//Q: How can DereferencesCalculate change offset?
	//Why is finalOffset fed in as a pointer?

	finalOffset = calculateDereferencesCompute(prgrm, arg, finalOffset, fp, fault)
	for _, fldIdx := range arg.Fields {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0383.
		fld := prgrm.CXArgFromArrayGet(fldIdx)
		// elt = fld
		finalOffset += fld.Offset
		finalOffset = calculateDereferencesCompute(prgrm, fld, finalOffset, fp, fault)
	}

	return finalOffset
}

// finalOffsetSliceIndexes applies the indexes of sliceDetails to the slice
// variable at offset and returns the address of the element.
func finalOffsetSliceIndexes(prgrm *CXProgram, fp types.Pointer, offset types.Pointer, sliceDetails *CXTypeSignature_Array, fault *bool) types.Pointer {
	eltSize := types.Code(sliceDetails.Meta).Size()
	for i := 0; i < len(sliceDetails.Indexes); i++ {
		// The elements of an outer slice are pointers to the inner slices.
		sizeToUse := eltSize
		if i+1 < len(sliceDetails.Lengths) {
			sizeToUse = types.PointerSize
		}

		offset = types.PtrRead(prgrm.Memory, offset)
		baseOffset := offset

		offset += types.ObjectHeaderSize
		offset += constants.SliceHeaderSize

		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0378.
		indexOffset := finalOffsetCompute(prgrm, fp, nil, prgrm.CXTypeSignatureFromArrayGet(sliceDetails.Indexes[i]), fault)
		indexValue := types.I32Read(prgrm.Memory, indexOffset)

		offset += offsetIndexFromI32(indexValue, fault) * sizeToUse // TODO:PTR Use ptr/PtrRead, array/slice indexing only works with i32.

		offsetSliceIndexCheck(SliceIndexValid(prgrm, baseOffset, offset, sizeToUse), fault)
	}
	return offset
}

func DereferenceSliceCalculate(prgrm *CXProgram, drfsStruct *DereferenceStruct) {
	if drfsStruct.pointerTyped && drfsStruct.finalOffset.Valid() && drfsStruct.finalOffset >= prgrm.Heap.StartsAt {
		// A slice behind a pointer into a heap object, as in (*s[i])[j] on a
		// []*[]T, see PointerOfAddress.
		drfsStruct.finalOffset += types.ObjectHeaderSize
	}
	drfsStruct.sliceIndexed = true
	drfsStruct.pointerTyped = false
	drfsStruct.finalOffset = types.PtrRead(prgrm.Memory, drfsStruct.finalOffset)
	drfsStruct.baseOffset = drfsStruct.finalOffset

	drfsStruct.finalOffset += types.ObjectHeaderSize
	drfsStruct.finalOffset += constants.SliceHeaderSize

	//TODO: delete
	sizeToUse := DerefSizeGet(prgrm, drfsStruct.arg, drfsStruct.idxCounter, drfsStruct.derefPointer, false) //TODO: is always arg.Size unless arg.StructType != nil
	drfsStruct.derefPointer = false

	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0384.
	indexOffset := finalOffsetCompute(prgrm, drfsStruct.fp, nil, prgrm.CXTypeSignatureFromArrayGet(drfsStruct.arg.Indexes[drfsStruct.idxCounter]), drfsStruct.fault)
	indexValue := types.I32Read(prgrm.Memory, indexOffset)

	drfsStruct.finalOffset += offsetIndexFromI32(indexValue, drfsStruct.fault) * sizeToUse // TODO:PTR Use ptr/PtrRead, array/slice indexing only works with i32.

	offsetSliceIndexCheck(SliceIndexValid(prgrm, drfsStruct.baseOffset, drfsStruct.finalOffset, sizeToUse), drfsStruct.fault)

	drfsStruct.idxCounter++
}

func DereferenceArrayCalculate(prgrm *CXProgram, drfsStruct *DereferenceStruct) {
	if drfsStruct.fault != nil && drfsStruct.idxCounter >= len(drfsStruct.arg.Lengths) {
		*drfsStruct.fault = true
		return
	}
	var subSize = types.Pointer(1)
	elementSlice := false
	lengths := drfsStruct.arg.Lengths[drfsStruct.idxCounter+1:]
	if direct, behind := drfsStruct.arg.DirectDimensions(); behind && drfsStruct.idxCounter < direct {
		lengths = drfsStruct.arg.Lengths[drfsStruct.idxCounter+1 : direct]
	}
	for _, len := range lengths {
		if len == 0 {
			// The elements of [N][]T are slices.
			elementSlice = true
			break
		}
		subSize *= len
	}

	//TODO: Delete
	sizeToUse := DerefSizeGet(prgrm, drfsStruct.arg, drfsStruct.idxCounter, drfsStruct.derefPointer, true) //TODO: is always arg.Size unless arg.StructType != nil
	if elementSlice {
		sizeToUse = types.PointerSize
	}
	drfsStruct.derefPointer = false
	drfsStruct.baseOffset = drfsStruct.finalOffset
	drfsStruct.sizeofElement = subSize * sizeToUse

	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0385.
	drfsStruct.finalOffset += offsetIndexFromI32(types.I32Read(prgrm.Memory, finalOffsetCompute(prgrm, drfsStruct.fp, nil, prgrm.CXTypeSignatureFromArrayGet(drfsStruct.arg.Indexes[drfsStruct.idxCounter]), drfsStruct.fault)), drfsStruct.fault) * drfsStruct.sizeofElement // TODO:PTR Use PtrRead
	drfsStruct.idxCounter++
}

func DereferencePointerCalculate(prgrm *CXProgram, drfsStruct *DereferenceStruct) {
	if drfsStruct.pointerTyped && drfsStruct.finalOffset.Valid() && drfsStruct.finalOffset >= prgrm.Heap.StartsAt {
		// A pointer read from a heap object, as in **pp.
		drfsStruct.finalOffset += types.ObjectHeaderSize
	}
	drfsStruct.pointerTyped = true
	drfsStruct.finalOffset = types.PtrRead(prgrm.Memory, drfsStruct.finalOffset)
	drfsStruct.derefPointer = true
}

// DereferencesCalculate applies the index and pointer dereferences of arg to
// finalOffset.
func DereferencesCalculate(prgrm *CXProgram, arg *CXArgument, finalOffset types.Pointer, fp types.Pointer) types.Pointer {
	return calculateDereferencesCompute(prgrm, arg, finalOffset, fp, nil)
}

// calculateDereferencesCompute is DereferencesCalculate, see finalOffsetCompute
// for fault.
func calculateDereferencesCompute(prgrm *CXProgram, arg *CXArgument, finalOffset types.Pointer, fp types.Pointer, fault *bool) types.Pointer {
	drfsStruct := &DereferenceStruct{
		idxCounter:  0,
		arg:         arg,
		finalOffset: finalOffset,
		fp:          fp,
		fault:       fault,
	}

	for _, op := range drfsStruct.arg.DereferenceOperations {
		if len(drfsStruct.arg.Indexes) == 0 && op != constants.DerefPointer {
			continue
		}

		switch op {
		case constants.DerefSlice:
			DereferenceSliceCalculate(prgrm, drfsStruct)
		case constants.DerefArray:
			DereferenceArrayCalculate(prgrm, drfsStruct)
		case constants.DerefPointer:
			DereferencePointerCalculate(prgrm, drfsStruct)
		}
	}

	if drfsStruct.finalOffset.Valid() && drfsStruct.finalOffset >= prgrm.Heap.StartsAt && drfsStruct.pointerTyped {
		// then it's an object
		drfsStruct.finalOffset += types.ObjectHeaderSize
		// An element of a slice, as ps[i] is, is not the slice. The Lengths of
		// an arg with fields, as in gp.s, are those of its last field.
		if drfsStruct.arg.SliceTyped() && !drfsStruct.sliceIndexed && len(drfsStruct.arg.Fields) == 0 {
			drfsStruct.finalOffset += constants.SliceHeaderSize
			offsetSliceIndexCheck(SliceIndexValid(prgrm, drfsStruct.baseOffset, drfsStruct.finalOffset, drfsStruct.sizeofElement), drfsStruct.fault)
		}
	}

	return drfsStruct.finalOffset
}

// derefSizeSliceOfArg returns the element size of the slice arg.
func derefSizeSliceOfArg(prgrm *CXProgram, arg *CXArgument) types.Pointer {
	if len(arg.Lengths) > 1 && (len(arg.Lengths)-len(arg.Indexes)) > 1 {
		if rest := arg.Lengths[len(arg.Indexes)+1:]; rest[0] != 0 {
			return ValueSizeOfLengths(rest, derefSizeSliceElement(prgrm, arg))
		}
		return types.PointerSize
	}
	return derefSizeSliceElement(prgrm, arg)
}

// derefSizeSliceElement returns the size of the innermost elements of arg.
func derefSizeSliceElement(prgrm *CXProgram, arg *CXArgument) types.Pointer {
	if arg.StructType != nil && !cxArgumentElementPointer(arg) {
		return arg.StructType.StructSizeGet(prgrm)
	}
	return arg.Size
}

// cxArgumentElementPointer reports whether the elements of the array or slice arg are
// pointers, as in [2]*T, where a pointer comes before the innermost dimension.
func cxArgumentElementPointer(arg *CXArgument) bool {
	for _, spec := range arg.DeclarationSpecifiers {
		switch spec {
		case constants.DeclarationPointer:
			return true
		case constants.DeclarationArray, constants.DeclarationSlice:
			return false
		}
	}
	return false
}
