package programgenerator

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"strconv"
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t7-cx-runtime-libraries"    // thin: only MakeProgram, RegisterPackages; no depth gain
	"github.com/skycoin/skycoin-cx/t8-cx-compiler-fullparsing" // thin: only InitCXCore; no depth gain
	"github.com/skycoin/skycoin-cx/t9-cx-core-astapi"
)

// RandomExpressionsGenerate appends random expressions built from fns to
// inputFn. Each expression owns a stack slot named after its position in
// inputFn.Expressions, and the last one writes to inputFn's first output.
func RandomExpressionsGenerate(prgrm *ast.CXProgram, inputFn *ast.CXFunction, inputPkg *ast.CXPackage, fns []*ast.CXNativeFunction, numExprs int) {
	// Adding an expression can grow CXAtomicOps, so operators are looked up
	// by index after every addition.
	fnIdx := prgrm.CXProgramFunctionGid(inputFn)
	fnOutputs := inputFn.OutputsGet(prgrm)
	newExprs := numExprs - len(inputFn.Expressions)

	for i := 0; i < newExprs; i++ {
		last := i == newExprs-1
		op := randomFunctionGet(fns)
		// Last expression output must be the same as function output.
		if last && len(fnOutputs) > 0 {
			want := signatureType(prgrm, fnOutputs[0])
			for len(op.Outputs) == 0 || op.Outputs[0].Type != want {
				op = randomFunctionGet(fns)
			}
		}

		// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
		// Now a constructor call, site S0021.
		exprCXLine := ast.CXLineExpressionMake(prgrm, ast.SourceFileGidNil, -1, stringpool.StringGidNil)
		// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
		// Now a constructor call, site S0022.
		expr := ast.AtomicOperatorExpressionMake(prgrm, op)
		cxAtomicOp := getAtomicOp(prgrm, expr.Index)
		cxAtomicOp.Package = ast.CXPackageIndex(inputPkg.Index)
		cxAtomicOp.Function = fnIdx

		for _, inp := range op.Inputs {
			inpIdx := randomInputGet(prgrm, fnIdx, op.AtomicOPCode, inp.Type)
			getAtomicOp(prgrm, expr.Index).InputAdd(prgrm, inpIdx)
		}

		// if operator is jmp, add then and else lines
		if OpcodeOperatorJump(op.AtomicOPCode) {
			lineNumOptions := numExprs - len(prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx).Expressions)
			if lineNumOptions < 0 {
				lineNumOptions = (lineNumOptions * -1) - 2
			}
			randThenLineIndex := 0
			if lineNumOptions > 0 {
				randThenLineIndex = rand.Intn(lineNumOptions)
			}

			cxAtomicOp = getAtomicOp(prgrm, expr.Index)
			cxAtomicOp.ThenLines = 1
			cxAtomicOp.ElseLines = randThenLineIndex
		}

		// The expression is added before its outputs are chosen, so its
		// own slot is one of the options.
		fn := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx)
		fn.ExpressionAdd(prgrm, exprCXLine)
		fn.ExpressionAdd(prgrm, expr)

		cxAtomicOp = getAtomicOp(prgrm, expr.Index)
		if last && len(fnOutputs) > 0 {
			cxAtomicOp.OutputAdd(prgrm, fnOutputs[0])
		} else {
			for _, out := range op.Outputs {
				cxAtomicOp.OutputAdd(prgrm, getRandOut(prgrm, fnIdx, out.Type))
			}
		}
	}

	fn := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx)
	fn.Size = functionSizeCalculate(prgrm, fn)
	fn.LineCount = len(fn.Expressions)
}

func OpcodeOperatorJump(opCode int) bool {
	switch opCode {
	case opcodetable.OpJmp,
		opcodetable.OpAbsJmp,
		opcodetable.OpJmpEq,
		opcodetable.OpJmpUneq,
		opcodetable.OpJmpGt,
		opcodetable.OpJmpGteq,
		opcodetable.OpJmpLt,
		opcodetable.OpJmpLteq,
		opcodetable.OpJmpZero,
		opcodetable.OpJmpNotZero:
		return true
	default:
		return false
	}
}

func randomFunctionGet(fnSet []*ast.CXNativeFunction) *ast.CXNativeFunction {
	return fnSet[rand.Intn(len(fnSet))]
}

func getAtomicOp(prgrm *ast.CXProgram, index int) *ast.CXAtomicOperator {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0027.
	cxAtomicOp, err := prgrm.CXAtomicOpGet(index)
	if err != nil {
		panic(err)
	}
	return cxAtomicOp
}

// signatureType reads the atomic type of a type signature, deprecated CXArgument
// signatures included, which CXTypeSignature.TypeGet does not handle.
func signatureType(prgrm *ast.CXProgram, idx ast.CXTypeSignatureIndex) valuetypes.Code {
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0028.
	sig := prgrm.CXTypeSignatureFromArrayGet(idx)
	if sig.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0029.
		return prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(sig.Meta)).Type
	}
	return sig.TypeGet(prgrm)
}

func signatureSize(prgrm *ast.CXProgram, idx ast.CXTypeSignatureIndex) valuetypes.Pointer {
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0030.
	return prgrm.CXTypeSignatureFromArrayGet(idx).SizeGet(prgrm, false)
}

// opOutput returns the first output of the operator behind expression exprPos
// of fn. Only one output per operator is considered, multiple returns in CX
// are buggy anyway.
func opOutput(prgrm *ast.CXProgram, fn *ast.CXFunction, exprPos int) (ast.CXTypeSignatureIndex, bool) {
	expr := fn.Expressions[exprPos]
	if expr.Type != ast.CXOperationTypeAtomicOperator {
		return 0, false
	}
	operator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(getAtomicOp(prgrm, expr.Index).Operator)
	outputs := operator.OutputsGet(prgrm)
	if len(outputs) == 0 {
		return 0, false
	}
	return outputs[0], true
}

// slotOffset is the stack offset of the slot owned by expression exprPos,
// after fn's inputs, outputs and the slots of the expressions before it.
func slotOffset(prgrm *ast.CXProgram, fn *ast.CXFunction, exprPos int) (offset valuetypes.Pointer) {
	for _, idx := range fn.InputsGet(prgrm) {
		offset += signatureSize(prgrm, idx)
	}
	for _, idx := range fn.OutputsGet(prgrm) {
		offset += signatureSize(prgrm, idx)
	}
	for p := 0; p < exprPos; p++ {
		if out, ok := opOutput(prgrm, fn, p); ok {
			offset += signatureSize(prgrm, out)
		}
	}
	return offset
}

func functionSizeCalculate(prgrm *ast.CXProgram, fn *ast.CXFunction) valuetypes.Pointer {
	return slotOffset(prgrm, fn, len(fn.Expressions))
}

// newSlot adds an atomic type signature for the slot of expression exprPos.
func newSlot(prgrm *ast.CXProgram, fn *ast.CXFunction, exprPos int, typ valuetypes.Code) ast.CXTypeSignatureIndex {
	_, arg := ast.FieldMake(prgrm, prgrm.CXProgramNameIntern(strconv.Itoa(exprPos)), typ, ast.SourceFileGidNil, -1)
	arg.TypeSet(typ)
	arg.Package = fn.Package
	arg.Offset = slotOffset(prgrm, fn, exprPos)
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureAlloc(), phase 4 #331.
	// Now an arena insert, site S0033.
	return prgrm.CXTypeSignatureInArrayAdd(ast.CXTypeSignatureRepresentationOfCXArgGet(prgrm, arg))
}

// signatureCopy adds a copy of the type signature at idx, so each use of a
// variable has its own signature.
func signatureCopy(prgrm *ast.CXProgram, idx ast.CXTypeSignatureIndex) ast.CXTypeSignatureIndex {
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureAlloc(), phase 4 #331.
	// Also state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332. Site S0034 S0035.
	sig := *prgrm.CXTypeSignatureFromArrayGet(idx)
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureAlloc(), phase 4 #331.
	// Now an arena insert, site S0036.
	return prgrm.CXTypeSignatureInArrayAdd(&sig)
}

// findArgOptions lists the inputs of fn and the outputs already written by
// its expressions that have type argTypeToFind.
func findArgOptions(prgrm *ast.CXProgram, fn *ast.CXFunction, argTypeToFind valuetypes.Code) []ast.CXTypeSignatureIndex {
	var options []ast.CXTypeSignatureIndex
	for _, idx := range fn.InputsGet(prgrm) {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0037.
		if signatureType(prgrm, idx) == argTypeToFind && prgrm.CXTypeSignatureFromArrayGet(idx).Name != stringpool.StringGidNil {
			options = append(options, idx)
		}
	}

	for _, expr := range fn.Expressions {
		if expr.Type != ast.CXOperationTypeAtomicOperator {
			continue
		}
		outputs := getAtomicOp(prgrm, expr.Index).OutputsGet(prgrm)
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0038.
		if len(outputs) > 0 && signatureType(prgrm, outputs[0]) == argTypeToFind && prgrm.CXTypeSignatureFromArrayGet(outputs[0]).Name != stringpool.StringGidNil {
			options = append(options, outputs[0])
		}
	}
	return options
}

func randomInputGet(prgrm *ast.CXProgram, fnIdx ast.CXFunctionGid, opCode int, argType valuetypes.Code) ast.CXTypeSignatureIndex {
	options := findArgOptions(prgrm, prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx), argType)

	// if no options available or if operator is jump, add new i32_LT expression.
	if len(options) == 0 || opCode == opcodetable.OpJmp {
		// TODO: improve process when there's OpJmp
		return addNewExpression(prgrm, fnIdx, nativeCode("i32.lt"))
	}

	return signatureCopy(prgrm, options[rand.Intn(len(options))])
}

// addNewExpression appends an expression of type expressionType to fn, with
// random inputs, and returns a signature reading its output.
func addNewExpression(prgrm *ast.CXProgram, fnIdx ast.CXFunctionGid, expressionType int) ast.CXTypeSignatureIndex {
	native := ast.NativeFunctionOf(prgrm, expressionType)
	// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
	// Now a constructor call, site S0040.
	newExprCXLine := ast.CXLineExpressionMake(prgrm, ast.SourceFileGidNil, -1, stringpool.StringGidNil)
	// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
	// Now a constructor call, site S0041.
	newExpr := ast.AtomicOperatorExpressionMake(prgrm, native)

	for _, inp := range native.Inputs {
		options := findArgOptions(prgrm, prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx), inp.Type)
		inpIdx := signatureCopy(prgrm, options[rand.Intn(len(options))])
		getAtomicOp(prgrm, newExpr.Index).InputAdd(prgrm, inpIdx)
	}

	fn := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx)
	fn.ExpressionAdd(prgrm, newExprCXLine)
	fn.ExpressionAdd(prgrm, newExpr)

	exprPos := len(fn.Expressions) - 1
	outIdx := newSlot(prgrm, fn, exprPos, native.Outputs[0].Type)
	getAtomicOp(prgrm, newExpr.Index).OutputAdd(prgrm, outIdx)

	return signatureCopy(prgrm, outIdx)
}

// getRandOut picks the slot of a random expression of fn whose operator
// returns outType, so expressions may overwrite each other's results.
func getRandOut(prgrm *ast.CXProgram, fnIdx ast.CXFunctionGid, outType valuetypes.Code) ast.CXTypeSignatureIndex {
	fn := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx)

	var options []int
	for p := range fn.Expressions {
		if out, ok := opOutput(prgrm, fn, p); ok && signatureType(prgrm, out) == outType {
			options = append(options, p)
		}
	}

	return newSlot(prgrm, fn, options[rand.Intn(len(options))], outType)
}

func FunctionSetGet(prgrm *ast.CXProgram, fnNames []string) (fns []*ast.CXNativeFunction) {
	for _, fnName := range fnNames {
		fn := ast.NativeFunctionOf(prgrm, nativeCode(fnName))
		if fn == nil {
			panic("standard library function not found.")
		}

		fns = append(fns, fn)
	}
	return fns
}

// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
// Now a constructor, site S0045.
func SampleProgramGenerate(t *testing.T, withLiteral bool) *ast.CXProgram {
	var cxProgram *ast.CXProgram

	// Needed for FunctionNativeExpressionAdd
	// because of dependency on cxast.NativeNames
	fullparsing.CXCoreInit()
	// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
	// Now a constructor call, site S0046.
	cxProgram = libraries.ProgramMake()
	libraries.PackagesRegister(cxProgram)

	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now an arena insert, site S0047.
	err := astapi.EmptyPackageAdd(cxProgram, "main")
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}

	err = astapi.PackageEmptyFunctionAdd(cxProgram, "main", "TestFunction")
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}

	err = astapi.FunctionNativeInputAdd(cxProgram, "main", "TestFunction", "inputOne", valuetypes.Code(valuetypes.I32))
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}

	err = astapi.FunctionNativeOutputAdd(cxProgram, "main", "TestFunction", "outputOne", valuetypes.Code(valuetypes.I32))
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}
	functionSetNames := []string{"i32.add", "i32.mul", "i32.sub", "i32.eq", "i32.uneq", "i32.gt", "i32.gteq", "i32.lt", "i32.lteq", "bool.not", "bool.or", "bool.and", "bool.uneq", "bool.eq", "i32.neg", "i32.abs", "i32.bitand", "i32.bitor", "i32.bitxor", "i32.bitclear", "i32.bitshl", "i32.bitshr", "i32.max", "i32.min", "i32.rand"}
	fns := FunctionSetGet(cxProgram, functionSetNames)

	fn, _ := cxProgram.FunctionGet(cxProgram.CXProgramNameIntern("TestFunction"), cxProgram.CxIrState.CxIrStateNameMain)
	pkg, _ := cxProgram.PackageGet(cxProgram.CxIrState.CxIrStateNameMain)
	RandomExpressionsGenerate(cxProgram, fn, pkg, fns, 30)

	if withLiteral {
		buf := new(bytes.Buffer)
		var num int32 = 5
		binary.Write(buf, binary.LittleEndian, num)
		err = astapi.ExpressionLiteralInputAdd(cxProgram, "main", "TestFunction", buf.Bytes(), valuetypes.I32, 2)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}

	}

	cxProgram.ProgramPrint()
	return cxProgram
}

// nativeCode returns the code of the built-in name, or 0 when none is
// linked.
func nativeCode(name string) int {
	entry, _ := ast.NativeNames.NativeNameIndexText(name)
	return entry.NativeNameEntryCode
}
