package opcodes

import (
	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
	// GCREACH: t3-ast owns the heap and the GC roots. This file allocates slices, which
	// adds to the GC pressure, see docs/GarbageCollector.md.
	"github.com/skycoin/skycoin-cx/t3-ast"
)

// TODO: Rework
func OpSliceLen(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var inp0 *ast.CXArgument
	var sliceLen types.Pointer
	if inputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1310.
		inp0 = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[0].TypeSignature.Meta))

		elt := inp0.AssignmentElementGet(prgrm)

		if elt.SliceTyped() {
			sliceOffset := types.PtrRead(prgrm.Memory, inputs[0].Offset)
			if sliceOffset > 0 && sliceOffset.Valid() {
				sliceLen = ast.SliceLenGet(prgrm, sliceOffset)
			}
			// TODO: Had to add elt.Lengths to avoid doing this for arrays, but not entirely sure why
		} else if (elt.PointerTargetType == types.STR || elt.Type == types.STR) && elt.Lengths == nil {
			sliceLen = types.StrSizeRead(prgrm.Memory, inputs[0].Offset)
		} else {
			sliceLen = elt.Lengths[len(elt.Indexes)]
		}
	} else if inputs[0].TypeSignature.Type == ast.TypeArrayAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1311.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[0].TypeSignature.Meta)
		sliceLen = arrDetails.Lengths[len(arrDetails.Indexes)]
	} else if inputs[0].TypeSignature.Type == ast.TypeSliceAtomic || inputs[0].TypeSignature.Type == ast.TypePointerSliceAtomic && inputs[0].TypeSignature.Deref {
		sliceOffset := types.PtrRead(prgrm.Memory, inputs[0].Offset)
		if sliceOffset > 0 && sliceOffset.Valid() {
			sliceLen = ast.SliceLenGet(prgrm, sliceOffset)
		}
	} else if inputs[0].TypeSignature.Type == ast.TypeStruct || inputs[0].TypeSignature.Type == ast.TypePointerStruct {
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S1312.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(inputs[0].TypeSignature.Meta)

		fldLen := len(structDetails.Fields)
		if fldLen > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S1313.
			fld := prgrm.CXArgFromArrayGet(structDetails.Fields[fldLen-1])
			fld = fld.AssignmentElementGet(prgrm)

			if fld.SliceTyped() {
				sliceOffset := types.PtrRead(prgrm.Memory, inputs[0].Offset)
				if sliceOffset > 0 && sliceOffset.Valid() {
					sliceLen = ast.SliceLenGet(prgrm, sliceOffset)
				}
				// TODO: Had to add elt.Lengths to avoid doing this for arrays, but not entirely sure why
			} else if (fld.PointerTargetType == types.STR || fld.Type == types.STR) && fld.Lengths == nil {
				sliceLen = types.StrSizeRead(prgrm.Memory, inputs[0].Offset)
			} else {
				sliceLen = fld.Lengths[len(fld.Indexes)]
			}
		} else {
			panic("struct has no fields")
		}
	} else {
		panic("type is not known\n\n")
	}

	outputs[0].I32Set(prgrm, types.PtrToI32Cast(sliceLen)) // TODO:PTR remove hardcode i32, should use ptr alias.
}

// TODO: Rework
func OpSliceAppend(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var inp0, out0 *ast.CXArgument
	var inp0Type, inp1Type, out0Type types.Code
	var input0IsSlice, outputIsSlice bool
	if inputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1314.
		inp0 = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[0].TypeSignature.Meta))
		inp0Type = inp0.Type
		if inp0.Type == types.POINTER {
			inp0Type = inp0.PointerTargetType
		}

		eltInp0 := inp0.AssignmentElementGet(prgrm)
		input0IsSlice = eltInp0.SliceTyped()
	} else if inputs[0].TypeSignature.Type == ast.TypeSliceAtomic || inputs[0].TypeSignature.Type == ast.TypePointerSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1315.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[0].TypeSignature.Meta)
		inp0Type = types.Code(sliceDetails.Meta)

		input0IsSlice = true
	} else if inputs[0].TypeSignature.Type == ast.TypeStruct || inputs[0].TypeSignature.Type == ast.TypePointerStruct {
		inp0Type = types.STRUCT

		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S1316.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(inputs[0].TypeSignature.Meta)
		fldLen := len(structDetails.Fields)
		if fldLen > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S1317.
			fld := prgrm.CXArgFromArrayGet(structDetails.Fields[fldLen-1])
			fld = fld.AssignmentElementGet(prgrm)
			input0IsSlice = fld.SliceTyped()

			inp0Type = fld.Type
			if fld.Type == types.POINTER {
				inp0Type = fld.PointerTargetType
			}
		}
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	if inputs[1].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1318.
		inp1 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[1].TypeSignature.Meta))
		inp1Type = inp1.Type
		if inp1.Type == types.POINTER {
			inp1Type = inp1.PointerTargetType
		}
	} else {
		inp1Type = inputs[1].TypeSignature.TypeGet(prgrm)
	}

	if outputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1319.
		out0 = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(outputs[0].TypeSignature.Meta))

		out0 := out0.AssignmentElementGet(prgrm)

		out0Type = out0.Type
		if out0.Type == types.POINTER {
			out0Type = out0.PointerTargetType
		}

		outputIsSlice = out0.SliceTyped()

	} else if outputs[0].TypeSignature.Type == ast.TypeSliceAtomic || outputs[0].TypeSignature.Type == ast.TypePointerSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1320.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(outputs[0].TypeSignature.Meta)
		out0Type = types.Code(sliceDetails.Meta)
		outputIsSlice = true
	} else if outputs[0].TypeSignature.Type == ast.TypeStruct || outputs[0].TypeSignature.Type == ast.TypePointerStruct {
		out0Type = types.STRUCT

		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S1321.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(outputs[0].TypeSignature.Meta)
		fldLen := len(structDetails.Fields)
		if fldLen > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S1322.
			fld := prgrm.CXArgFromArrayGet(structDetails.Fields[fldLen-1])
			fld = fld.AssignmentElementGet(prgrm)
			outputIsSlice = fld.SliceTyped()

			out0Type = fld.Type
			if fld.Type == types.POINTER {
				out0Type = fld.PointerTargetType
			}
		}
	} else {
		panic("type is not known")
	}

	sliceInputs := inputs[1:]
	sliceInputsLen := types.IntToPtrCast(len(sliceInputs))

	if inp0Type != inp1Type || inp0Type != out0Type || !input0IsSlice || !outputIsSlice {
		panic(types.CXRuntimeInvalidArgument)
	}

	var inputSliceLen types.Pointer
	inputSliceOffset := types.PtrRead(prgrm.Memory, inputs[0].Offset)
	if inputSliceOffset != 0 && inputSliceOffset.Valid() {
		inputSliceLen = ast.SliceLenGet(prgrm, inputSliceOffset)
	}

	// Preparing slice in case more memory is needed for the new element.
	eltSize := ast.DerefSizeSliceGet(prgrm, inputs[0].TypeSignature)
	outputSliceOffset := ast.SliceAppendResize(prgrm, inputs[0].FramePointer, outputs[0].TypeSignature, inputs[0].TypeSignature, eltSize, sliceInputsLen)

	// We need to update the address of the output and input, as the final offsets
	// could be on the heap and they could have been moved by the GC.

	for i, input := range sliceInputs {
		var inp *ast.CXArgument
		var inpType types.Code
		if input.TypeSignature.Type == ast.TypeCXArgumentDeprecate {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S1323.
			inp = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(input.TypeSignature.Meta))
			inpType = inp.Type
			if inp.Type == types.POINTER {
				inpType = inp.PointerTargetType
			}
		} else {
			inpType = input.TypeSignature.TypeGet(prgrm)
		}

		if inp0Type != inpType {
			panic(types.CXRuntimeInvalidArgument)
		}
		if ast.CXTypeSignaturePassBy(prgrm, input.TypeSignature) == constants.PassByReference {
			// append(ps, &x) appends the address of x.
			var obj [types.PointerSize]byte
			types.PtrWrite(obj[:], 0, ast.PointerOfAddress(prgrm, input.Offset))
			ast.SliceAppendWrite(prgrm, outputSliceOffset, obj[:], inputSliceLen+types.IntToPtrCast(i))
		} else if inpType == types.STR && eltSize == types.PointerSize {
			// An array of str, as [][2]str appends, is copied whole below.
			var obj [types.PointerSize]byte
			types.PtrWrite(obj[:], 0, types.PtrRead(prgrm.Memory, input.Offset))
			ast.SliceAppendWrite(prgrm, outputSliceOffset, obj[:], inputSliceLen+types.IntToPtrCast(i))
		} else {
			obj := types.SliceByteGet(prgrm.Memory, input.Offset, eltSize)
			ast.SliceAppendWrite(prgrm, outputSliceOffset, obj, inputSliceLen+types.IntToPtrCast(i))
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
}

// TODO: Rework
func OpSliceResize(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var inpType, outType types.Code
	var inpIsSlice, outIsSlice bool
	if inputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1324.
		inp0 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[0].TypeSignature.Meta))
		eltInp0 := inp0.AssignmentElementGet(prgrm)

		inpType = inp0.Type
		inpIsSlice = eltInp0.SliceTyped()
	} else if inputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1325.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[0].TypeSignature.Meta)

		inpType = types.Code(sliceDetails.Meta)
		inpIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	if outputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1326.
		out0 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(outputs[0].TypeSignature.Meta))
		outType = out0.Type
		outIsSlice = out0.AssignmentElementGet(prgrm).SliceTyped()
	} else if outputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1327.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(outputs[0].TypeSignature.Meta)

		outType = types.Code(sliceDetails.Meta)
		outIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	fp := inputs[0].FramePointer

	if inpType != outType || !inpIsSlice || !outIsSlice {
		panic(types.CXRuntimeInvalidArgument)
	}

	outputSliceOffset := ast.SliceResize(prgrm, fp, outputs[0].TypeSignature, inputs[0].TypeSignature, types.I32ToPtrCast(inputs[1].I32Get(prgrm)), ast.DerefSizeSliceGet(prgrm, inputs[0].TypeSignature))

	outputs[0].PtrSet(prgrm, outputSliceOffset)
}

// TODO: Rework
func OpSliceInsertElement(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var inp0Type, inp2Type, out0Type types.Code
	var inpIsSlice, outIsSlice bool
	if inputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1328.
		inp0 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[0].TypeSignature.Meta))

		inp0Type = inp0.Type
		if inp0.Type == types.POINTER {
			inp0Type = inp0.PointerTargetType
		}

		inpIsSlice = inp0.SliceTyped()
	} else if inputs[0].TypeSignature.Type == ast.TypeAtomic || inputs[0].TypeSignature.Type == ast.TypePointerAtomic {
		inp0Type = types.Code(inputs[0].TypeSignature.Meta)

		inpIsSlice = false
	} else if inputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1329.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[0].TypeSignature.Meta)

		inp0Type = types.Code(sliceDetails.Meta)
		inpIsSlice = true
	} else {
		panic("type is not known")
	}

	if inputs[2].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1330.
		inp2 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[2].TypeSignature.Meta))

		inp2Type = inp2.Type
		if inp2.Type == types.POINTER {
			inp2Type = inp2.PointerTargetType
		}
	} else {
		inp2Type = inputs[2].TypeSignature.TypeGet(prgrm)
	}

	if outputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1331.
		out0 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(outputs[0].TypeSignature.Meta))

		out0Type = out0.Type
		if out0.Type == types.POINTER {
			out0Type = out0.PointerTargetType
		}

		outIsSlice = out0.SliceTyped()
	} else if outputs[0].TypeSignature.Type == ast.TypeAtomic || outputs[0].TypeSignature.Type == ast.TypePointerAtomic {
		out0Type = types.Code(outputs[0].TypeSignature.Meta)

		outIsSlice = false
	} else if outputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1332.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(outputs[0].TypeSignature.Meta)

		out0Type = types.Code(sliceDetails.Meta)
		outIsSlice = true
	} else {
		panic("type is not known")
	}

	fp := inputs[0].FramePointer

	if inp0Type != inp2Type || inp0Type != out0Type || !inpIsSlice || !outIsSlice {
		panic(types.CXRuntimeInvalidArgument)
	}

	index := types.I32ToPtrCast(inputs[1].I32Get(prgrm))

	var outputSliceOffset types.Pointer
	if inp2Type == types.STR {
		var obj [types.PointerSize]byte
		types.PtrWrite(obj[:], 0, types.PtrRead(prgrm.Memory, inputs[2].Offset))
		outputSliceOffset = ast.SliceInsert(prgrm, fp, outputs[0].TypeSignature, inputs[0].TypeSignature, index, obj[:])
	} else {
		obj := inputs[2].BytesGet(prgrm)
		outputSliceOffset = ast.SliceInsert(prgrm, fp, outputs[0].TypeSignature, inputs[0].TypeSignature, index, obj)
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
}

// TODO: Rework
func OpSliceRemoveElement(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var inpType, outType types.Code
	var inpIsSlice, outIsSlice bool

	if inputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1333.
		inp0 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[0].TypeSignature.Meta))

		inpType = inp0.Type
		inpIsSlice = inp0.AssignmentElementGet(prgrm).SliceTyped()
	} else if inputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1334.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[0].TypeSignature.Meta)

		inpType = types.Code(sliceDetails.Meta)
		inpIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	if outputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1335.
		out0 := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(outputs[0].TypeSignature.Meta))

		outType = out0.Type
		outIsSlice = out0.AssignmentElementGet(prgrm).SliceTyped()
	} else if outputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1336.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(outputs[0].TypeSignature.Meta)

		outType = types.Code(sliceDetails.Meta)
		outIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	fp := inputs[0].FramePointer

	if inpType != outType || !inpIsSlice || !outIsSlice {
		panic(types.CXRuntimeInvalidArgument)
	}

	outputSliceOffset := ast.SliceRemove(prgrm, fp, inputs[0].TypeSignature, types.I32ToPtrCast(inputs[1].I32Get(prgrm)), ast.DerefSizeSliceGet(prgrm, inputs[0].TypeSignature))

	outputs[0].PtrSet(prgrm, outputSliceOffset)
}

func OpSliceCopy(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var dstType types.Code
	var dstSize types.Pointer
	var dstIsSlice bool
	if inputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1337.
		dstInput := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[0].TypeSignature.Meta))
		dstElem := dstInput.AssignmentElementGet(prgrm)

		dstType = dstInput.Type
		dstSize = dstElem.Size
		dstIsSlice = dstElem.SliceTyped()
	} else if inputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1338.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[0].TypeSignature.Meta)

		dstType = types.Code(sliceDetails.Meta)
		dstSize = types.Code(sliceDetails.Meta).Size()
		dstIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	var srcType types.Code
	var srcSize types.Pointer
	var srcIsSlice bool
	if inputs[1].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1339.
		srcInput := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[1].TypeSignature.Meta))
		srcElem := srcInput.AssignmentElementGet(prgrm)

		srcType = srcInput.Type
		srcSize = srcElem.Size
		srcIsSlice = srcElem.SliceTyped()
	} else if inputs[1].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1340.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[1].TypeSignature.Meta)

		srcType = types.Code(sliceDetails.Meta)
		srcSize = types.Code(sliceDetails.Meta).Size()
		srcIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	fp := inputs[0].FramePointer

	dstOffset := ast.SliceOffsetGet(prgrm, fp, inputs[0].TypeSignature)
	srcOffset := ast.SliceOffsetGet(prgrm, fp, inputs[1].TypeSignature)

	if dstType != srcType || !dstIsSlice || !srcIsSlice || dstSize != srcSize {
		panic(types.CXRuntimeInvalidArgument)
	}

	dst := ast.SliceElementsGet(prgrm, dstOffset, dstSize)
	src := ast.SliceElementsGet(prgrm, srcOffset, srcSize)

	var count types.Pointer
	if dstType == srcType {
		count = types.IntToPtrCast(copy(dst, src))
		if count%dstSize != 0 {
			panic(types.CXRuntimeError)
		}
	} else {
		panic(types.CXRuntimeInvalidArgument)
	}

	outputs[0].I32Set(prgrm, types.PtrToI32Cast(count/dstSize)) // TODO:PTR use ptr instead of i32
}
