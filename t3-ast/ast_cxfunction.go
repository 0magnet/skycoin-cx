package ast

import (
	"errors"
	"fmt"

	"github.com/skycoin/skycoin-cx/t0-types"
	objmodel "github.com/skycoin/skycoin-cx/t1-cx-core-objmodel"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

// BuiltIn determines if opcode is not 0
// True if the function is native to CX, e.g. int32.add()
func (cxf CXFunction) BuiltIn() bool {
	return cxf.AtomicOPCode != 0
}

func (cxf CXFunction) Atomic() bool {
	return cxf.AtomicOPCode != 0
}

// FunctionMake allocates an empty function in the program's CXFunctionMng
// and returns its Gid and the pointer to fill it through. Parameters and
// contents are added later.
func FunctionMake(prgrm *CXProgram, name stringpool.StringGid, fileName SourceFileGid, fileLine int) (CXFunctionGid, *CXFunction) {
	gid, fn := prgrm.CXProgramFunctionAlloc()
	fn.Name = name
	fn.FileName = fileName
	fn.FileLine = fileLine
	// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
	// Now a heap literal, site S0135.
	fn.Inputs = &CXStruct{
		Name: prgrm.CXProgramNameIntern(prgrm.CXProgramNameText(name) + "_Input"),
	}
	fn.LocalVariables = []stringpool.StringGid{}
	return gid, fn
}

// ----------------------------------------------------------------
//                             `CXFunction` Getters

// ExpressionsGet is not used
func (fn *CXFunction) ExpressionsGet(prgrm *CXProgram) ([]CXExpression, error) {
	if fn.Expressions == nil {
		return nil, fmt.Errorf("function '%s' has no expressions", prgrm.CXProgramNameText(fn.Name))
	}

	return fn.Expressions, nil

}

// ExpressionByLabelGet
func (fn *CXFunction) ExpressionByLabelGet(prgrm *CXProgram, lbl string) (*CXExpression, error) {
	if fn.Expressions == nil {
		return nil, fmt.Errorf("function '%s' has no expressions", prgrm.CXProgramNameText(fn.Name))
	}

	for _, expr := range fn.Expressions {
		if prgrm.CXProgramNameText(expr.LabelGet(prgrm)) == lbl {
			return &expr, nil
		}
	}
	return nil, fmt.Errorf("expression '%s' not found in function '%s'", lbl, prgrm.CXProgramNameText(fn.Name))
}

// ExpressionByLineGet ...
func (fn *CXFunction) ExpressionByLineGet(prgrm *CXProgram, line int) (*CXExpression, error) {
	if fn.Expressions == nil {
		return nil, fmt.Errorf("function '%s' has no expressions", prgrm.CXProgramNameText(fn.Name))
	}

	if line > len(fn.Expressions) {
		return nil, fmt.Errorf("expression line number '%d' exceeds number of expressions in function '%s'", line, prgrm.CXProgramNameText(fn.Name))
	}

	return &fn.Expressions[line], nil
}

// ----------------------------------------------------------------
//                     `CXFunction` Member handling

// InputAdd adds param to the inputs of fn. It takes a managed param, which it
// releases when the signature does not name it or the input exists.
func (fn *CXFunction) InputAdd(prgrm *CXProgram, param *CXArgument) *CXFunction {
	fnInputs := fn.InputsGet(prgrm)
	for _, inputIdx := range fnInputs {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0136.
		input := prgrm.CXTypeSignatureFromArrayGet(inputIdx)
		if input.Name == param.Name {
			prgrm.CXArgBuilderRelease(param)
			return fn
		}
	}

	param.Package = fn.Package
	newFieldIdx := CXTypeSignatureOfCXArgBuilderAlloc(prgrm, param)

	if fn.Inputs == nil {
		// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
		// Now a heap literal, site S0137.
		fn.Inputs = &CXStruct{}
	}

	fn.Inputs.FieldTypeSignatureAdd(prgrm, newFieldIdx)

	return fn
}

func (fn *CXFunction) InputsGet(prgrm *CXProgram) []CXTypeSignatureIndex {
	if fn == nil || fn.Inputs == nil {
		return []CXTypeSignatureIndex{}
	}

	return fn.Inputs.Fields
}

// RemoveInput ...
// func (fn *CXFunction) RemoveInput(prgrm *CXProgram, inpName string) {
// 	if len(fn.Inputs) > 0 {
// 		lenInps := len(fn.Inputs)
// 		for i, inpIdx := range fn.Inputs {
// 			inp := prgrm.CXArgFromArrayGet(inpIdx)
// 			if inp.Name == inpName {
// 				if i == lenInps {
// 					fn.Inputs = fn.Inputs[:len(fn.Inputs)-1]
// 				} else {
// 					fn.Inputs = append(fn.Inputs[:i], fn.Inputs[i+1:]...)
// 				}
// 				break
// 			}
// 		}
// 	}
// }

// OutputAdd adds param to the outputs of fn like InputAdd.
func (fn *CXFunction) OutputAdd(prgrm *CXProgram, param *CXArgument) *CXFunction {
	fnOutputs := fn.OutputsGet(prgrm)
	for _, outputIdx := range fnOutputs {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0139.
		output := prgrm.CXTypeSignatureFromArrayGet(outputIdx)
		if output.Name == param.Name {
			prgrm.CXArgBuilderRelease(param)
			return fn
		}
	}

	param.Package = fn.Package

	// An array output, or a pointer to one, stays a CXArgument, the call,
	// return and assignment paths have no array signature case.
	var newFieldIdx CXTypeSignatureIndex
	if CXArgumentTypeArrayAtomic(param) || CXArgumentTypePointerArrayAtomic(param) {
		newFieldIdx = CXTypeSignatureOfCXArgDeprecateAlloc(prgrm, param)
	} else {
		newFieldIdx = CXTypeSignatureOfCXArgBuilderAlloc(prgrm, param)
	}

	if fn.Outputs == nil {
		// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
		// Now a heap literal, site S0140.
		fn.Outputs = &CXStruct{}
	}

	fn.Outputs.FieldTypeSignatureAdd(prgrm, newFieldIdx)

	return fn
}

func (fn *CXFunction) OutputsGet(prgrm *CXProgram) []CXTypeSignatureIndex {
	if fn == nil || fn.Outputs == nil {
		return []CXTypeSignatureIndex{}
	}

	return fn.Outputs.Fields
}

// RemoveOutput ...
// func (fn *CXFunction) RemoveOutput(prgrm *CXProgram, outName string) {
// 	if len(fn.Outputs) > 0 {
// 		lenOuts := len(fn.Outputs)
// 		for i, outIdx := range fn.Outputs {
// 			out := prgrm.CXArgFromArrayGet(outIdx)
// 			if out.Name == outName {
// 				if i == lenOuts {
// 					fn.Outputs = fn.Outputs[:len(fn.Outputs)-1]
// 				} else {
// 					fn.Outputs = append(fn.Outputs[:i], fn.Outputs[i+1:]...)
// 				}
// 				break
// 			}
// 		}
// 	}
// }

// ExpressionAdd ...
func (fn *CXFunction) ExpressionAdd(prgrm *CXProgram, expr *CXExpression) *CXFunction {
	if expr.Type == CXOperationTypeAtomicOperator {
		// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
		// Now an arena getter, site S0142.
		cxAtomicOp, err := prgrm.CXAtomicOpGet(expr.Index)
		if err != nil {
			panic(err)
		}

		cxAtomicOp.Package = fn.Package
		cxAtomicOp.Function = prgrm.CXProgramFunctionGid(fn)
	}

	fn.Expressions = append(fn.Expressions, *expr)
	fn.LineCount++
	return fn
}

func (fn *CXFunction) ExpressionByLineNumberAdd(prgrm *CXProgram, expr *CXExpression, line int) *CXFunction {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0143.
	cxAtomicOp, err := prgrm.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}
	cxAtomicOp.Package = fn.Package
	cxAtomicOp.Function = prgrm.CXProgramFunctionGid(fn)

	lenExprs := len(fn.Expressions)
	if lenExprs == line {
		fn.Expressions = append(fn.Expressions, *expr)
	} else {
		// OBJMODEL: state.CXExpressionMng.CXExpressionRelease(gid), phase 4 #346.
		// Now a dropped object, site S0144.
		fn.Expressions = append(fn.Expressions[:line+1], fn.Expressions[line:]...)
		// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
		// Now an object copy, site S0145.
		fn.Expressions[line] = *expr
	}

	fn.LineCount++
	return fn

}

// ExpressionRemove ...
func (fn *CXFunction) ExpressionRemove(line int) {
	if len(fn.Expressions) > 0 {
		lenExprs := len(fn.Expressions)
		if line >= lenExprs-1 || line < 0 {
			// OBJMODEL: state.CXExpressionMng.CXExpressionRelease(gid), phase 4 #346.
			// Now a dropped object, site S0146.
			fn.Expressions = fn.Expressions[:len(fn.Expressions)-1]
		} else {
			// OBJMODEL: state.CXExpressionMng.CXExpressionRelease(gid), phase 4 #346.
			// Now a dropped object, site S0147.
			fn.Expressions = append(fn.Expressions[:line], fn.Expressions[line+1:]...)
		}
	}
}

func (fn *CXFunction) LocalVariableNameAdd(name stringpool.StringGid) error {
	fn.LocalVariables = append(fn.LocalVariables, name)

	return nil
}

func (fn *CXFunction) LocalVariableDeclared(name stringpool.StringGid) bool {
	for _, localVar := range fn.LocalVariables {
		if name == localVar {
			return true
		}
	}

	return false
}

func (fn *CXFunction) LocalVariableFromArrayRemove(name stringpool.StringGid) error {
	for idx, localVar := range fn.LocalVariables {
		if name == localVar {
			fn.LocalVariables = append(fn.LocalVariables[:idx], fn.LocalVariables[idx+1:]...)
			return nil
		}
	}

	return nil
}

// ----------------------------------------------------------------
//                             `CXFunction` Selectors

// AtomicOperatorExpressionMake ...
// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
// Now a constructor, site S0148.
func AtomicOperatorExpressionMake(prgrm *CXProgram, op *CXNativeFunction) *CXExpression {
	var opIdx CXFunctionGid
	if op != nil {
		opIdx = prgrm.NativeFunctionInArrayAdd(op)
	}

	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorAlloc(), phase 4 #341.
	// Now a heap literal, site S0150 S0151.
	index := prgrm.CXAtomicOpAdd(&CXAtomicOperator{
		Operator: opIdx,
		// Function: -1,
	})

	// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
	// Now a heap literal, site S0152.
	return &CXExpression{
		Index: index,
		Type:  CXOperationTypeAtomicOperator,
	}
}

// CXLineExpressionMake ...
// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
// Now a constructor, site S0153.
func CXLineExpressionMake(prgrm *CXProgram, fileName SourceFileGid, lineNo int, lineStr stringpool.StringGid) *CXExpression {
	lineGid := prgrm.CxIrState.CXLineMng.CXLineAlloc()
	line := prgrm.CxIrState.CXLineMng.CXLineGet(lineGid)
	line.FileName = fileName
	line.LineNumber = lineNo
	line.LineColumn = prgrm.CXProgramCompileErrorColumn(fileName, lineNo)
	line.LineStr = lineStr
	index := int(objmodel.GidIndex(int64(lineGid)))

	// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
	// Now a heap literal, site S0156.
	return &CXExpression{
		Index: index,
		Type:  CXOperationTypeLine,
	}
}

// SelectExpression ...
func (fn *CXFunction) SelectExpression(line int) (*CXExpression, error) {
	// prgrmStep := &CXProgramStep{
	// 	Action: func(cxt *CXProgram) {
	// 		if mod, err := cxt.CurrentPackageGet(); err == nil {
	// 			if fn, err := mod.CurrentFunctionGet(); err == nil {
	// 				fn.SelectExpression(line)
	// 			}
	// 		}
	// 	},
	// }
	// saveProgramStep(prgrmStep, fn.Context)
	if len(fn.Expressions) == 0 {
		return nil, errors.New("There are no expressions in this function")
	}

	if line >= len(fn.Expressions) {
		line = len(fn.Expressions) - 1
	}

	if line < 0 {
		line = 0
	}

	expr := fn.Expressions[line]

	return &expr, nil
}

// CXFunctionEscape says where a function may keep the address an input holds.
type CXFunctionEscape uint8

const (
	// CXFunctionEscapeNone keeps it only for the call.
	CXFunctionEscapeNone CXFunctionEscape = iota
	// CXFunctionEscapeResult may return it in an output.
	CXFunctionEscapeResult
	// CXFunctionEscapeHeap may keep it after the call.
	CXFunctionEscapeHeap
)

// CXFunctionSlot is a local of a function that holds a pointer, slice or str,
// with its declared type. The GC walks it as a root by its Type.
type CXFunctionSlot struct {
	// CXFunctionSlotOffset is the offset of the local, flagged when it escaped.
	CXFunctionSlotOffset types.Pointer
	// CXFunctionSlotLengths are the lengths of the arrays and slices of the
	// local, outermost first, as CXArgument.Lengths.
	CXFunctionSlotLengths []types.Pointer
	// CXFunctionSlotType is the base type of the local, types.STR for a str.
	CXFunctionSlotType types.Code
	// CXFunctionSlotSpecifiers are the declaration specifiers of the local.
	CXFunctionSlotSpecifiers []int
	// CXFunctionSlotStruct is the struct the local is made of, or nil.
	CXFunctionSlotStruct *CXStruct
}
