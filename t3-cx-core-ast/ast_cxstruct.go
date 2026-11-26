package ast

import (
	"fmt"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

type CXStructIndex int

// ----------------------------------------------------------------
//                             `CXStruct` Getters

// FieldGet returns the signature of the field name of strct. Its
// StructTypeGet and OffsetGet give what a caller needs of the field.
func (strct *CXStruct) FieldGet(prgrm *CXProgram, name stringpool.StringGid) (*CXTypeSignature, error) {
	for _, typeSignatureIdx := range strct.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0254.
		typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
		if typeSignature.Name == name {
			return typeSignature, nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found in struct '%s'", prgrm.CXProgramNameText(name), prgrm.CXProgramNameText(strct.Name))
}

// ----------------------------------------------------------------
//                     `CXStruct` Member handling

// StructMake ...
// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
// Now a constructor, site S0261.
func StructMake(name stringpool.StringGid) *CXStruct {
	// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
	// Now a heap literal, site S0262.
	return &CXStruct{
		Name: name,
	}
}

// FieldAdd adds the field cxArgument to strct. It takes a managed cxArgument,
// which it releases when the signature does not name it.
func (strct *CXStruct) FieldAdd(prgrm *CXProgram, cxArgument *CXArgument) *CXStruct {
	// Check if field already exist
	for _, typeSignatureIdx := range strct.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0263.
		typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
		if typeSignature.Name == cxArgument.Name {
			prgrm.CXProgramCompileErrorAddDebug(CompileErrorCodeStructFieldDuplicate, cxArgument.ArgDetails, CompileErrorOperands{int64(cxArgument.Name)})
			prgrm.CXArgBuilderRelease(cxArgument)
			return strct
		}
	}

	typeSignatureIdx := CXTypeSignatureOfCXArgForStructsAlloc(prgrm, cxArgument)
	prgrm.CXArgUnlessNamedRelease(cxArgument, typeSignatureIdx)
	strct.Fields = append(strct.Fields, typeSignatureIdx)

	// TODO: Found out the effect and completely remove this.
	// Dont remove this yet as removing this gives another error on cxfx
	// even though all cx tests passed already.
	// numFlds := len(strct.Fields)
	// if numFlds > 2 {
	// 	// Pre-compiling the offset of the field.
	// 	lastTypeSignature := strct.Fields[numFlds-2]
	// 	currentTypeSignature := strct.Fields[numFlds-1]
	// 	if lastTypeSignature.Type == TypeCXArgumentDeprecate && currentTypeSignature.Type == TypeCXArgumentDeprecate {
	// 		lastFldIdx := lastTypeSignature.Meta
	// 		currFldIdx := currentTypeSignature.Meta
	// 		prgrm.CXArgGet(CXArgumentIndex(currFldIdx)).Offset = prgrm.CXArgGet(CXArgumentIndex(lastFldIdx)).Offset + prgrm.CXArgGet(CXArgumentIndex(lastFldIdx)).TotalSize
	// 	} else if lastTypeSignature.Type == TypeCXArgumentDeprecate && currentTypeSignature.Type != TypeCXArgumentDeprecate {
	// 		lastFldIdx := lastTypeSignature.Meta
	// 		currentTypeSignature.Offset = prgrm.CXArgGet(CXArgumentIndex(lastFldIdx)).Offset + prgrm.CXArgGet(CXArgumentIndex(lastFldIdx)).TotalSize
	// 	} else if lastTypeSignature.Type != TypeCXArgumentDeprecate && currentTypeSignature.Type == TypeCXArgumentDeprecate {
	// 		currFldIdx := currentTypeSignature.Meta
	// 		prgrm.CXArgGet(CXArgumentIndex(currFldIdx)).Offset = lastTypeSignature.Offset + lastTypeSignature.SizeGet(prgrm, false)
	// 	} else if lastTypeSignature.Type != TypeCXArgumentDeprecate && currentTypeSignature.Type != TypeCXArgumentDeprecate {
	// 		currentTypeSignature.Offset = lastTypeSignature.Offset + lastTypeSignature.SizeGet(prgrm, false)
	// 	}
	// }

	return strct
}

// FieldAdd ...
func (strct *CXStruct) FieldTypeSignatureAdd(prgrm *CXProgram, fieldIdx CXTypeSignatureIndex) *CXStruct {
	// Check if field already exist
	for _, typeSignatureIdx := range strct.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0265.
		typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0266.
		field := prgrm.CXTypeSignatureFromArrayGet(fieldIdx)
		if typeSignature.Name == field.Name {
			prgrm.CXProgramCompileErrorAddDebug(CompileErrorCodeStructFieldDuplicate, field.ArgDetails, CompileErrorOperands{int64(field.Name)})
			return strct
		}
	}

	strct.Fields = append(strct.Fields, fieldIdx)

	return strct
}

func (strct *CXStruct) FieldCXAtomicOpsAdd(prgrm *CXProgram, fieldIdx CXTypeSignatureIndex) *CXStruct {
	strct.Fields = append(strct.Fields, fieldIdx)

	return strct
}

func (strct *CXStruct) FieldGlobalsCXAtomicOpsAdd(prgrm *CXProgram, cxArgIdx CXArgumentIndex) *CXStruct {
	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S0267.
	cxArgument := prgrm.CXArgFromArrayGet(cxArgIdx)

	// Check if field already exist
	for _, typeSignatureIdx := range strct.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0268.
		typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
		if typeSignature.Name == cxArgument.Name {
			prgrm.CXProgramCompileErrorAddDebug(CompileErrorCodeStructFieldDuplicate, cxArgument.ArgDetails, CompileErrorOperands{int64(cxArgument.Name)})
			return strct
		}
	}

	newCXTypeSignatureIdx := CXTypeSignatureOfCXArgAlloc(prgrm, cxArgument)
	prgrm.CXArgUnlessNamedRelease(cxArgument, newCXTypeSignatureIdx)
	strct.Fields = append(strct.Fields, newCXTypeSignatureIdx)

	return strct
}

// RemoveField ...
// func (strct *CXStruct) RemoveField(prgrm *CXProgram, fldName string) {
// 	if len(strct.Fields) > 0 {
// 		lenFlds := len(strct.Fields)
// 		for i, fldIdx := range strct.Fields {
// 			if prgrm.CXArgGet(CXArgumentIndex(fldIdx)).Name == fldName {
// 				if i == lenFlds-1 {
// 					strct.Fields = strct.Fields[:len(strct.Fields)-1]
// 				} else {
// 					strct.Fields = append(strct.Fields[:i], strct.Fields[i+1:]...)
// 				}
// 				break
// 			}
// 		}
// 	}
// }

func (strct *CXStruct) StructSizeGet(prgrm *CXProgram) valuetypes.Pointer {
	var structSize valuetypes.Pointer
	for _, typeSignatureIdx := range strct.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0270.
		typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
		structSize += typeSignature.SizeGet(prgrm, false)
	}

	return structSize
}

// StructFieldOffset returns the offset of field index of strct. The
// signature of a field holds its size in Offset, so the offset is a sum.
func StructFieldOffset(prgrm *CXProgram, strct *CXStruct, index int) valuetypes.Pointer {
	var offset valuetypes.Pointer
	for _, fieldIdx := range strct.Fields[:index] {
		offset += prgrm.CXTypeSignatureFromArrayGet(fieldIdx).Offset
	}
	return offset
}
