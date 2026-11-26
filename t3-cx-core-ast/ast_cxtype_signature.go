package ast

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	objmodel "github.com/skycoin/skycoin-cx/t1-cx-core-objmodel"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

// CXTypeSignatureType enum contains CXTypeSignature types.
type CXTypeSignatureType int

type CXTypeSignatureIndex int

const (
	TypeUnused CXTypeSignatureType = iota
	TypeAtomic
	TypePointerAtomic
	TypeArrayAtomic
	TypePointerArrayAtomic
	TypeSliceAtomic
	TypePointerSliceAtomic

	TypeStruct
	TypePointerStruct
	TypeArrayStruct
	TypePointerArrayStruct
	TypeSliceStruct
	TypeSlicePointerStruct

	TypeComplex
	TypePointerComplex
	TypeArrayComplex
	TypeArrayPointerComplex
	TypeSliceComplex
	TypeSlicePointerComplex

	// For CXArgument usage
	// To be deprecated
	TypeCXArgumentDeprecate
)

// ----------------------------------------------------------------
//                             `CXTypeSignature` Handlers

func (typeSignature *CXTypeSignature) SizeGet(prgrm *CXProgram, forSymbolsTableUpdate bool) valuetypes.Pointer {
	switch typeSignature.Type {
	case TypeAtomic:
		return valuetypes.Code(typeSignature.Meta).Size()
	case TypePointerAtomic:
		if typeSignature.Deref {
			return valuetypes.Code(typeSignature.Meta).Size()
		}

		return valuetypes.POINTER.Size()
	case TypeArrayAtomic:
		if forSymbolsTableUpdate {
			return typeSignature.ArraySizeGet(prgrm)
		}

		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0273.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		if len(arrDetails.Indexes) > 0 {
			return indexedValueSize(arrDetails)
		}

		return typeSignature.ArraySizeGet(prgrm)
	case TypePointerArrayAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0274.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		if typeSignature.Deref && len(arrDetails.Indexes) > 0 {
			return indexedValueSize(arrDetails)
		} else if typeSignature.Deref {
			return typeSignature.ArraySizeGet(prgrm)
		}

		return valuetypes.POINTER.Size()
	case TypeSliceAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0275.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)

		if len(sliceDetails.Indexes) > 0 {
			return indexedValueSize(sliceDetails)
		}
		if typeSignature.Deref {
			return valuetypes.Code(sliceDetails.Meta).Size()
		}

		return valuetypes.PointerSize
	case TypePointerSliceAtomic:

		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0276.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		if len(sliceDetails.Indexes) > 0 {
			return indexedValueSize(sliceDetails)
		}

		// *p without an index is the slice header pointer, like p itself.
		return valuetypes.POINTER.Size()
	case TypeStruct:
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0277.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)

		lenFlds := len(structDetails.Fields)
		if lenFlds > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S0278.
			return ArgSizeGet(prgrm, prgrm.CXArgGet(CXArgumentIndex(structDetails.Fields[lenFlds-1])))
		}
		return structDetails.StructType.StructSizeGet(prgrm)
	case TypePointerStruct:
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0279.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)

		lenFlds := len(structDetails.Fields)
		if lenFlds > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S0280.
			return ArgSizeGet(prgrm, prgrm.CXArgGet(CXArgumentIndex(structDetails.Fields[lenFlds-1])))
		}

		if typeSignature.Deref {
			return structDetails.StructType.StructSizeGet(prgrm)
		}

		return valuetypes.POINTER.Size()
	case TypeArrayStruct:
	case TypePointerArrayStruct:
	case TypeSliceStruct:
	case TypeSlicePointerStruct:

	case TypeComplex:
	case TypePointerComplex:
	case TypeArrayComplex:
	case TypeArrayPointerComplex:
	case TypeSliceComplex:
	case TypeSlicePointerComplex:

	case TypeCXArgumentDeprecate:
		argIdx := typeSignature.Meta
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0281.
		return ArgSizeGet(prgrm, prgrm.CXArgGet(CXArgumentIndex(argIdx)))
	}

	return 0
}

func (typeSignature *CXTypeSignature) ArrayLengthGet(prgrm *CXProgram) valuetypes.Pointer {
	switch typeSignature.Type {
	case TypeArrayAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0282.
		typeSignatureForArray := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		return valuetypes.TotalLength(typeSignatureForArray.Lengths)
	}

	return 0
}

func (typeSignature *CXTypeSignature) ArraySizeGet(prgrm *CXProgram) valuetypes.Pointer {
	switch typeSignature.Type {
	case TypeArrayAtomic, TypePointerArrayAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0283.
		typeSignatureForArray := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		arrType := valuetypes.Code(typeSignatureForArray.Meta)
		return valuetypes.TotalLength(typeSignatureForArray.Lengths) * arrType.Size()
	}

	return 0
}

func (typeSignature *CXTypeSignature) TypeGet(prgrm *CXProgram) valuetypes.Code {
	var atomicType valuetypes.Code
	switch typeSignature.Type {
	case TypeAtomic:
		atomicType = valuetypes.Code(typeSignature.Meta)
	case TypePointerAtomic:
		atomicType = valuetypes.Code(typeSignature.Meta)
	case TypeArrayAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0284.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		atomicType = valuetypes.Code(arrDetails.Meta)
	case TypePointerArrayAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0285.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		atomicType = valuetypes.Code(arrDetails.Meta)
	case TypeSliceAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0286.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		atomicType = valuetypes.Code(sliceDetails.Meta)
	case TypePointerSliceAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0287.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		atomicType = valuetypes.Code(sliceDetails.Meta)
	case TypeStruct:
		atomicType = valuetypes.STRUCT

		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0288.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)
		fldLen := len(structDetails.Fields)
		if fldLen > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S0289.
			fld := prgrm.CXArgFromArrayGet(structDetails.Fields[fldLen-1])
			fld = fld.AssignmentElementGet(prgrm)

			atomicType = fld.Type
			if fld.Type == valuetypes.POINTER {
				atomicType = fld.PointerTargetType
			}
		}
	case TypePointerStruct:
		atomicType = valuetypes.STRUCT

		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0290.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)
		fldLen := len(structDetails.Fields)
		if fldLen > 0 {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S0291.
			fld := prgrm.CXArgFromArrayGet(structDetails.Fields[fldLen-1])
			fld = fld.AssignmentElementGet(prgrm)

			atomicType = fld.Type
			if fld.Type == valuetypes.POINTER {
				atomicType = fld.PointerTargetType
			}
		}
	default:
		panic("type is not known")
	}

	return atomicType
}

// typeNameOf returns what names the type of typeSignature, the name Gid of a
// struct when named is true and the type code otherwise.
func (typeSignature *CXTypeSignature) typeNameOf(prgrm *CXProgram) (code valuetypes.Code, name stringpool.StringGid, named bool) {
	switch typeSignature.Type {
	case TypeAtomic, TypePointerAtomic:
		return valuetypes.Code(typeSignature.Meta), stringpool.StringGidNil, false
	case TypeArrayAtomic, TypePointerArrayAtomic, TypeSliceAtomic, TypePointerSliceAtomic:
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, sites S0292 S0293 S0294.
		arrDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		return valuetypes.Code(arrDetails.Meta), stringpool.StringGidNil, false
	case TypeStruct, TypePointerStruct:
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, sites S0295 S0297.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)
		fldsLen := len(structDetails.Fields)
		if fldsLen == 0 {
			return valuetypes.UNDEFINED, structDetails.StructType.Name, true
		}
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, sites S0296 S0298.
		elt := prgrm.CXArgFromArrayGet(structDetails.Fields[fldsLen-1]).AssignmentElementGet(prgrm)
		return argumentTypeNameOf(elt)
	}
	panic("type is not known")
}

// argumentTypeNameOf returns what names the type of elt, its struct or, for a
// pointer, the type it points to.
func argumentTypeNameOf(elt *CXArgument) (code valuetypes.Code, name stringpool.StringGid, named bool) {
	if elt.StructType != nil {
		return valuetypes.UNDEFINED, elt.StructType.Name, true
	}
	if elt.Type == valuetypes.POINTER {
		return elt.PointerTargetType, stringpool.StringGidNil, false
	}
	return elt.Type, stringpool.StringGidNil, false
}

// TypeStringGet returns the type name of typeSignature, for the compiler.
func (typeSignature *CXTypeSignature) TypeStringGet(prgrm *CXProgram) string {
	code, name, named := typeSignature.typeNameOf(prgrm)
	if named {
		return prgrm.CXProgramNameText(name)
	}
	return code.Name()
}

// ValueTypeNameAppend appends the type name printf gives an extra value to
// dst, the struct name for a struct and the element type name otherwise,
// viewed from the string pool or the type table (#412).
func (typeSignature *CXTypeSignature) ValueTypeNameAppend(dst []byte, prgrm *CXProgram) []byte {
	code, name, named := typeSignature.valueTypeNameOf(prgrm)
	if named {
		return append(dst, prgrm.CXProgramNameBytes(name)...)
	}
	return append(dst, code.NameBytes()...)
}

// ValueTypeNameGet returns the name ValueTypeNameAppend appends, for the
// compiler.
func (typeSignature *CXTypeSignature) ValueTypeNameGet(prgrm *CXProgram) string {
	code, name, named := typeSignature.valueTypeNameOf(prgrm)
	if named {
		return prgrm.CXProgramNameText(name)
	}
	return code.Name()
}

func (typeSignature *CXTypeSignature) valueTypeNameOf(prgrm *CXProgram) (code valuetypes.Code, name stringpool.StringGid, named bool) {
	switch typeSignature.Type {
	case TypeCXArgumentDeprecate:
		element := prgrm.CXArgFromArrayGet(CXArgumentIndex(typeSignature.Meta)).AssignmentElementGet(prgrm)
		if element.StructType != nil {
			return valuetypes.UNDEFINED, element.StructType.Name, true
		}
		return element.Type, stringpool.StringGidNil, false
	case TypeAtomic, TypePointerAtomic, TypeArrayAtomic, TypePointerArrayAtomic,
		TypeSliceAtomic, TypePointerSliceAtomic, TypeStruct, TypePointerStruct:
		return typeSignature.typeNameOf(prgrm)
	}
	return valuetypes.UNDEFINED, stringpool.StringGidNil, false
}

func (typeSignature *CXTypeSignature) CXArgFormatGet(prgrm *CXProgram) *CXArgument {
	// MISUSE: builds an IR argument for the garbage collector, which needs only
	// the offset, types and declaration specifiers of a field, see #423.
	// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
	// Now a heap literal, site S0299.
	var arg *CXArgument = &CXArgument{}
	if typeSignature.Type == TypeAtomic {
		arg.Type = valuetypes.Code(typeSignature.Meta)
		arg.StructType = nil
		arg.Size = typeSignature.SizeGet(prgrm, false)

		// TODO: this should not be needed.
		if len(arg.DeclarationSpecifiers) > 0 {
			arg.DeclarationSpecifiers = append([]int{opcodetable.DeclarationBasic}, arg.DeclarationSpecifiers[1:]...)
		} else {
			arg.DeclarationSpecifiers = []int{opcodetable.DeclarationBasic}
		}

	} else if typeSignature.Type == TypePointerAtomic {
		arg.Type = valuetypes.POINTER
		arg.PointerTargetType = valuetypes.Code(typeSignature.Meta)
		arg.StructType = nil
		arg.Size = valuetypes.POINTER.Size()

		// TODO: this should not be needed.
		if len(arg.DeclarationSpecifiers) > 0 {
			arg.DeclarationSpecifiers = append([]int{opcodetable.DeclarationBasic, opcodetable.DeclarationPointer}, arg.DeclarationSpecifiers[1:]...)
		} else {
			arg.DeclarationSpecifiers = []int{opcodetable.DeclarationPointer}
		}
	} else if typeSignature.Type == TypeArrayAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0300.
		typeSignatureArray := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		arg.Type = valuetypes.Code(typeSignatureArray.Meta)
		arg.StructType = nil
		arg.Size = typeSignature.SizeGet(prgrm, false)
		arg.Lengths = typeSignatureArray.Lengths
		arg.Indexes = typeSignatureArray.Indexes

		// TODO: this should not be needed.
		if len(arg.DeclarationSpecifiers) > 0 {
			arg.DeclarationSpecifiers = append([]int{opcodetable.DeclarationBasic, opcodetable.DeclarationArray}, arg.DeclarationSpecifiers[1:]...)
		} else {
			arg.DeclarationSpecifiers = []int{opcodetable.DeclarationBasic}
		}

	} else if typeSignature.Type == TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S0301.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(typeSignature.Meta)
		arg.Type = valuetypes.Code(sliceDetails.Meta)
		arg.StructType = nil
		arg.Size = valuetypes.Code(sliceDetails.Meta).Size()
		arg.Lengths = sliceDetails.Lengths

		// TODO: this should not be needed.
		if len(arg.DeclarationSpecifiers) > 0 {
			arg.DeclarationSpecifiers = append([]int{opcodetable.DeclarationBasic, opcodetable.DeclarationSlice}, arg.DeclarationSpecifiers[1:]...)
		} else {
			arg.DeclarationSpecifiers = []int{opcodetable.DeclarationBasic}
		}

		arg.DereferenceOperations = append([]int{opcodetable.DerefPointer}, arg.DereferenceOperations...)

	} else if typeSignature.Type == TypeStruct {
		arg.Type = valuetypes.STRUCT
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0302.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)
		arg.StructType = structDetails.StructType
		// TODO: this should not be needed.
		if len(arg.DeclarationSpecifiers) > 0 {
			arg.DeclarationSpecifiers = append([]int{opcodetable.DeclarationStruct}, arg.DeclarationSpecifiers[1:]...)
		} else {
			arg.DeclarationSpecifiers = []int{opcodetable.DeclarationStruct}
		}

	} else if typeSignature.Type == TypePointerStruct {
		arg.Type = valuetypes.POINTER
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0303.
		structDetails := prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta)
		arg.StructType = structDetails.StructType
		// TODO: this should not be needed.
		if len(arg.DeclarationSpecifiers) > 0 {
			arg.DeclarationSpecifiers = append([]int{opcodetable.DeclarationStruct, opcodetable.DeclarationPointer}, arg.DeclarationSpecifiers[1:]...)
		} else {
			arg.DeclarationSpecifiers = []int{opcodetable.DeclarationStruct, opcodetable.DeclarationPointer}
		}

	}
	arg.Offset = typeSignature.Offset

	return arg
}

// InnerReference determines if it is an inner reference
// for example: &slice[0] or &struct.field
func (typeSignature *CXTypeSignature) InnerReference(prgrm *CXProgram) bool {
	switch typeSignature.Type {
	case TypeCXArgumentDeprecate:
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0304.
		arg := prgrm.CXArgFromArrayGet(CXArgumentIndex(typeSignature.Meta))

		if arg.PassBy == opcodetable.PassByReference && len(arg.Indexes) > 0 {
			for _, decl := range arg.DeclarationSpecifiers {
				if decl == opcodetable.DeclarationPointer {
					return true
				}
			}

		}
	}

	return false
}

// fillCXTypeSignatureOfCXArgForStructs fills newCXTypeSignature from the struct
// field cxArgument. It returns false when the field has no signature type of
// its own, and the caller makes the signature name the argument.
// Added _ForStructs for now so it will be separate from others for smooth transition.
// TODO: Remove _ForStructs
func fillCXTypeSignatureOfCXArgForStructs(prgrm *CXProgram, newCXTypeSignature *CXTypeSignature, cxArgument *CXArgument) bool {
	newCXTypeSignature.Name = cxArgument.Name
	newCXTypeSignature.Package = cxArgument.Package
	newCXTypeSignature.PassBy = cxArgument.PassBy
	newCXTypeSignature.ArgDetails = cxArgument.ArgDetails

	fieldType := cxArgument.Type
	if CXArgumentTypeAtomic(cxArgument) {
		// If atomic type. i.e. i8, i16, i32, f32, etc.

		newCXTypeSignature.Type = TypeAtomic
		newCXTypeSignature.Meta = int(fieldType)

	} else if CXArgumentTypePointerAtomic(cxArgument) {
		// If pointer atomic, i.e. *i32, *f32, etc.

		newCXTypeSignature.Type = TypePointerAtomic
		newCXTypeSignature.Offset = cxArgument.Offset

		newCXTypeSignature.Meta = int(cxArgument.Type)
		if cxArgument.Type == valuetypes.STR || cxArgument.StructType != nil {
			newCXTypeSignature.Meta = int(cxArgument.PointerTargetType)
		}

	} else if CXArgumentTypeArrayAtomic(cxArgument) {
		// If simple array atomic type, i.e. [5]i32, [2]f64, etc.

		newCXTypeSignature.Type = TypeArrayAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(fieldType)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))

	} else if CXArgumentTypePointerArrayAtomic(cxArgument) {
		// If pointer array atomic type, i.e. *[5]i32, *[2]f64, etc.

		newCXTypeSignature.Type = TypePointerArrayAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(fieldType)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))

	} else if CXArgumentTypeSliceAtomic(cxArgument) {
		// If slice atomic type, i.e. []i32, []f64, etc.

		newCXTypeSignature.Type = TypeSliceAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(cxArgument.Type)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))
	} else if CXArgumentTypePointerSliceAtomic(cxArgument) {
		// If pointer slice atomic type, i.e. *[]i32, *[]f64, etc.

		newCXTypeSignature.Type = TypePointerSliceAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(cxArgument.Type)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))
	} else if CXArgumentTypeStruct(cxArgument) {
		// If type is struct
		newCXTypeSignature.Type = TypeStruct

		structGid := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructAlloc()
		typeSignatureForStruct := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructGet(structGid)
		typeSignatureForStruct.Fields = cxArgument.Fields
		typeSignatureForStruct.StructType = cxArgument.StructType

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(structGid)))
	} else if CXArgumentTypePointerStruct(cxArgument) {
		// If type is struct
		newCXTypeSignature.Type = TypePointerStruct

		structGid := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructAlloc()
		typeSignatureForStruct := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructGet(structGid)
		typeSignatureForStruct.Fields = cxArgument.Fields
		typeSignatureForStruct.StructType = cxArgument.StructType

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(structGid)))
	} else {
		return false
	}
	return true
}

// fillCXTypeSignatureOfCXArg fills newCXTypeSignature from cxArgument. It
// returns false when the argument has no signature type of its own, and the
// caller makes the signature name the argument.
func fillCXTypeSignatureOfCXArg(prgrm *CXProgram, newCXTypeSignature *CXTypeSignature, cxArgument *CXArgument) bool {
	newCXTypeSignature.Name = cxArgument.Name
	newCXTypeSignature.Package = cxArgument.Package
	newCXTypeSignature.PassBy = cxArgument.PassBy
	newCXTypeSignature.ArgDetails = cxArgument.ArgDetails

	if CXArgumentTypeAtomic(cxArgument) {
		// If atomic type. i.e. i8, i16, i32, f32, etc.
		newCXTypeSignature.Type = TypeAtomic
		newCXTypeSignature.Meta = int(cxArgument.Type)
		newCXTypeSignature.Offset = cxArgument.Offset
	} else if CXArgumentTypePointerAtomic(cxArgument) {
		// If pointer atomic, i.e. *i32, *f32, etc.
		newCXTypeSignature.Type = TypePointerAtomic
		newCXTypeSignature.Offset = cxArgument.Offset

		newCXTypeSignature.Meta = int(cxArgument.Type)
		if cxArgument.Type == valuetypes.STR || cxArgument.StructType != nil {
			newCXTypeSignature.Meta = int(cxArgument.PointerTargetType)
		}
	} else if CXArgumentTypeArrayAtomic(cxArgument) {
		// If simple array atomic type, i.e. [5]i32, [2]f64, etc.
		newCXTypeSignature.Type = TypeArrayAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(cxArgument.Type)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))
	} else if CXArgumentTypePointerArrayAtomic(cxArgument) {
		// If pointer array atomic type, i.e. *[5]i32, *[2]f64, etc.

		newCXTypeSignature.Type = TypePointerArrayAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(cxArgument.Type)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))
	} else if CXArgumentTypeSliceAtomic(cxArgument) {
		// If slice atomic type, i.e. []i32, []f64, etc.

		newCXTypeSignature.Type = TypeSliceAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(cxArgument.Type)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))
	} else if CXArgumentTypePointerSliceAtomic(cxArgument) {
		// If pointer slice atomic type, i.e. *[]i32, *[]f64, etc.

		newCXTypeSignature.Type = TypePointerSliceAtomic

		arrayGid := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayAlloc()
		typeSignatureForArray := prgrm.CxIrState.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(arrayGid)
		typeSignatureForArray.Meta = int(cxArgument.Type)
		typeSignatureForArray.Lengths = cxArgument.Lengths
		typeSignatureForArray.Indexes = cxArgument.Indexes

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(arrayGid)))
	} else if CXArgumentTypeStruct(cxArgument) {
		// If type is struct
		newCXTypeSignature.Type = TypeStruct

		structGid := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructAlloc()
		typeSignatureForStruct := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructGet(structGid)
		typeSignatureForStruct.Fields = cxArgument.Fields
		typeSignatureForStruct.StructType = cxArgument.StructType

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(structGid)))
	} else if CXArgumentTypePointerStruct(cxArgument) {
		// If type is struct
		newCXTypeSignature.Type = TypePointerStruct

		structGid := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructAlloc()
		typeSignatureForStruct := prgrm.CxIrState.CXTypeSignature_StructMng.CXTypeSignature_StructGet(structGid)
		typeSignatureForStruct.Fields = cxArgument.Fields
		typeSignatureForStruct.StructType = cxArgument.StructType

		newCXTypeSignature.Meta = int(objmodel.GidIndex(int64(structGid)))
	} else {
		return false
	}
	return true
}

// CXTypeSignatureOfCXArgForStructsAlloc allocates the signature of the struct
// field cxArgument in the program's CXTypeSignatureMng and returns its slot.
// A field with no signature type of its own is named by the signature, and a
// heap field is copied in first.
func CXTypeSignatureOfCXArgForStructsAlloc(prgrm *CXProgram, cxArgument *CXArgument) CXTypeSignatureIndex {
	_, newCXTypeSignature := prgrm.CXTypeSignatureAlloc()
	if !fillCXTypeSignatureOfCXArgForStructs(prgrm, newCXTypeSignature, cxArgument) {
		// All are TypeCXArgumentDeprecate for now. The CXArg slot is in Meta.
		newCXTypeSignature.Type = TypeCXArgumentDeprecate
		newCXTypeSignature.Meta = int(prgrm.cxArgumentSlot(cxArgument, true))
	}

	newCXTypeSignature.Offset = newCXTypeSignature.SizeGet(prgrm, false)

	return newCXTypeSignature.Index
}

// RetypeCXTypeSignatureOfCXArg gives typeSig, which names cxArgument, the
// signature type of cxArgument once its type is known, as x := v learns it.
func RetypeCXTypeSignatureOfCXArg(prgrm *CXProgram, typeSig *CXTypeSignature, cxArgument *CXArgument) {
	if fillCXTypeSignatureOfCXArg(prgrm, typeSig, cxArgument) {
		typeSig.Offset = cxArgument.Offset
	}
}

// CXTypeSignatureOfCXArgAlloc allocates the signature of cxArgument in the
// program's CXTypeSignatureMng and returns its slot. An argument with no
// signature type of its own is named by the signature, and a heap argument
// that has no slot yet is copied in first.
func CXTypeSignatureOfCXArgAlloc(prgrm *CXProgram, cxArgument *CXArgument) CXTypeSignatureIndex {
	_, newCXTypeSignature := prgrm.CXTypeSignatureAlloc()
	if !fillCXTypeSignatureOfCXArg(prgrm, newCXTypeSignature, cxArgument) {
		// TypeCXArgumentDeprecate, the CXArg slot is in Meta.
		newCXTypeSignature.Type = TypeCXArgumentDeprecate
		newCXTypeSignature.Meta = int(prgrm.cxArgumentSlot(cxArgument, false))
		newCXTypeSignature.Offset = cxArgument.Offset
	}

	return newCXTypeSignature.Index
}

// CXTypeSignatureRepresentationOfCXArgGet builds the signature of cxArgument
// on the heap, for the callers that still copy it in with
// CXTypeSignatureInArrayAdd, the natives, serialization and tests.
func CXTypeSignatureRepresentationOfCXArgGet(prgrm *CXProgram, cxArgument *CXArgument) *CXTypeSignature {
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureAlloc(), phase 4 #331.
	// Now a value literal, site S0319.
	newCXTypeSignature := CXTypeSignature{}
	if !fillCXTypeSignatureOfCXArg(prgrm, &newCXTypeSignature, cxArgument) {
		// TypeCXArgumentDeprecate
		// FieldIdx or the CXArg ID is in Meta field.
		newCXTypeSignature.Type = TypeCXArgumentDeprecate
		newCXTypeSignature.Meta = int(prgrm.cxArgumentSlot(cxArgument, false))
		newCXTypeSignature.Offset = cxArgument.Offset
	}

	return &newCXTypeSignature
}

// CXTypeSignatureOfCXArgBuilderAlloc is CXTypeSignatureOfCXArgAlloc for a
// builder that nothing else holds. The signature takes it, and when the
// signature does not name it the builder is released.
func CXTypeSignatureOfCXArgBuilderAlloc(prgrm *CXProgram, cxArgument *CXArgument) CXTypeSignatureIndex {
	typeSigIdx := CXTypeSignatureOfCXArgAlloc(prgrm, cxArgument)
	prgrm.CXArgUnlessNamedRelease(cxArgument, typeSigIdx)
	return typeSigIdx
}

// CXTypeSignatureOfCXArgDeprecateAlloc allocates a TypeCXArgumentDeprecate
// signature that names cxArgument, whatever its type.
func CXTypeSignatureOfCXArgDeprecateAlloc(prgrm *CXProgram, cxArgument *CXArgument) CXTypeSignatureIndex {
	_, typeSig := prgrm.CXTypeSignatureAlloc()
	typeSig.Name = cxArgument.Name
	typeSig.Package = cxArgument.Package
	typeSig.PassBy = cxArgument.PassBy
	typeSig.ArgDetails = cxArgument.ArgDetails
	typeSig.Type = TypeCXArgumentDeprecate
	typeSig.Meta = int(prgrm.cxArgumentSlot(cxArgument, false))
	typeSig.Offset = cxArgument.Offset
	return typeSig.Index
}

// StructTypeGet returns the struct type of the value typeSignature names,
// or nil when it is not a struct or a pointer to one.
func (typeSignature *CXTypeSignature) StructTypeGet(prgrm *CXProgram) *CXStruct {
	switch typeSignature.Type {
	case TypeCXArgumentDeprecate:
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1784.
		return prgrm.CXArgGet(CXArgumentIndex(typeSignature.Meta)).StructType
	case TypeStruct, TypePointerStruct:
		// OBJMODEL: state.CXTypeSignature_StructMng.CXTypeSignature_StructGet(gid), phase 5 #322.
		// Now an arena getter, site S0255.
		return prgrm.CXTypeSignatureStructFromArrayGet(typeSignature.Meta).StructType
	}
	return nil
}

// OffsetGet returns the offset of the value typeSignature names. A signature
// that names an argument has the offset of the argument.
func (typeSignature *CXTypeSignature) OffsetGet(prgrm *CXProgram) valuetypes.Pointer {
	if typeSignature.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0260.
		return prgrm.CXArgGet(CXArgumentIndex(typeSignature.Meta)).Offset
	}
	return typeSignature.Offset
}

// CXTypeSignaturePassBy is the PassBy of sig, read from its argument when
// sig is a TypeCXArgumentDeprecate.
func CXTypeSignaturePassBy(prgrm *CXProgram, sig *CXTypeSignature) int {
	if sig.Type == TypeCXArgumentDeprecate {
		return prgrm.CXArgFromArrayGet(CXArgumentIndex(sig.Meta)).PassBy
	}
	return sig.PassBy
}

// indexedValueSize returns the size of the value an indexed array or slice
// reads. r[1] of a [2][3]i32 is a [3]i32 and m[0] of a [][]i32 a slice.
func indexedValueSize(details *CXTypeSignature_Array) valuetypes.Pointer {
	if len(details.Indexes) >= len(details.Lengths) {
		return valuetypes.Code(details.Meta).Size()
	}
	return ValueSizeOfLengths(details.Lengths[len(details.Indexes):], valuetypes.Code(details.Meta).Size())
}
