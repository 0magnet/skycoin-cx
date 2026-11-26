package ast

import stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"

// ----------------------------------------------------------------
//                             `CXAtomicOperator` Getters

// OperatorNameGet returns the name of the operator of op, for compile
// diagnostics. The stack dump reads OperatorNameBytesGet.
func (op *CXAtomicOperator) OperatorNameGet(prgrm *CXProgram) string {
	opOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(op.Operator)
	if opOperator.BuiltIn() {
		return opOperator.BuiltInName()
	}
	return prgrm.CXProgramNameText(opOperator.Name)

}

// OperatorNameBytesGet returns the bytes of OperatorNameGet, viewed from the
// string pool or the built-in name tables (#412). The caller must not write
// to them.
func (op *CXAtomicOperator) OperatorNameBytesGet(prgrm *CXProgram) []byte {
	opOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(op.Operator)
	if opOperator.BuiltIn() {
		return opOperator.BuiltInNameBytes()
	}
	return prgrm.CXProgramNameBytes(opOperator.Name)
}

// ----------------------------------------------------------------
//                     `CXAtomicOperator` Member handling

// InputAdd ...
func (op *CXAtomicOperator) InputAdd(prgrm *CXProgram, typeSignatureIdx CXTypeSignatureIndex) *CXAtomicOperator {
	if op.Inputs == nil {
		// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
		// Now a heap literal, site S0090.
		op.Inputs = &CXStruct{Package: op.Package}
	}

	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0091.
	typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
	if typeSignature.Package == -1 {
		typeSignature.Package = op.Package
	}

	// Add Package if arg has no package
	if typeSignature.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0092.
		arg := prgrm.CXArgFromArrayGet(CXArgumentIndex(typeSignature.Meta))
		if arg.Package == -1 {
			arg.Package = op.Package
		}
	}

	op.Inputs.FieldCXAtomicOpsAdd(prgrm, typeSignatureIdx)

	return op
}

func (op *CXAtomicOperator) InputsGet(prgrm *CXProgram) []CXTypeSignatureIndex {
	if op == nil || op.Inputs == nil {
		return []CXTypeSignatureIndex{}
	}

	return op.Inputs.Fields
}

// RemoveInput ...
// func (op *CXAtomicOperator) RemoveInput() {
// 	if len(op.Inputs) > 0 {
// 		op.Inputs = op.Inputs[:len(op.Inputs)-1]
// 	}
// }

// OutputAdd ...
func (op *CXAtomicOperator) OutputAdd(prgrm *CXProgram, typeSignatureIdx CXTypeSignatureIndex) *CXAtomicOperator {
	if op.Outputs == nil {
		// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
		// Now a heap literal, site S0093.
		op.Outputs = &CXStruct{Package: op.Package}
	}

	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0094.
	typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
	if typeSignature.Package == -1 {
		typeSignature.Package = op.Package
	}

	// Add Package if arg has no package
	if typeSignature.Type == TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0095.
		arg := prgrm.CXArgFromArrayGet(CXArgumentIndex(typeSignature.Meta))
		if arg.Package == -1 {
			arg.Package = op.Package
		}
	}

	op.Outputs.FieldCXAtomicOpsAdd(prgrm, typeSignatureIdx)
	return op
}

func (op *CXAtomicOperator) OutputsGet(prgrm *CXProgram) []CXTypeSignatureIndex {
	if op == nil || op.Outputs == nil {
		return []CXTypeSignatureIndex{}
	}

	return op.Outputs.Fields
}

// RemoveOutput ...
// func (op *CXAtomicOperator) RemoveOutput() {
// 	if len(op.Outputs) > 0 {
// 		op.Outputs = op.Outputs[:len(op.Outputs)-1]
// 	}
// }

// LabelAdd ...
func (op *CXAtomicOperator) LabelAdd(lbl stringpool.StringGid) *CXAtomicOperator {
	op.Label = lbl
	return op
}
