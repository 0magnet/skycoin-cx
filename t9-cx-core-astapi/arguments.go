package astapi

import (
	"errors"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes" // thin: only Code; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild" // thin: only AST, WritePrimary; no depth gain
)

// ExpressionNativeInputAdd adds native input to an expression
// in a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageName - the name of the package where the function is located.
// functionName - the name of the function where the expression is located.
// inputName - the name of input to be added.
// inputType - the type of the input to be added.
// lineNumber - the line number of the expression in the function.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z i16 = add(x i16)
//
// We use ExpressionNativeInputAdd(cxprogram, "main", "TestFunction", "y", cxtypes.I16, 0).
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z i16 = add(x i16, y i16)
//
// Note the y i16 added as an input to the expression in line 0.
func ExpressionNativeInputAdd(cxprogram *ast.CXProgram, packageName, functionName, inputName string, inputType valuetypes.Code, lineNumber int) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}

	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}

	expr, err := fn.ExpressionByLineGet(cxprogram, lineNumber)
	if err != nil {
		return err
	}

	_, arg := ast.FieldMake(cxprogram, cxprogram.CXProgramNameIntern(inputName), inputType, ast.SourceFileGidNil, -1)
	arg.TypeSet(valuetypes.Code(inputType))
	arg.Package = ast.CXPackageIndex(pkg.Index)
	argIdx := ast.CXArgumentIndex(arg.Index)
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S1563.
	cxAtomicOp, err := cxprogram.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}

	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S1564.
	typeSigIdx := ast.CXTypeSignatureOfCXArgBuilderAlloc(cxprogram, cxprogram.CXArgFromArrayGet(ast.CXArgumentIndex(argIdx)))
	cxAtomicOp.InputAdd(cxprogram, typeSigIdx)

	return nil
}

// RemoveInputFromExpression removes an input from an
// expression in a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// functionName - the name of the function where the expression is located.
// lineNumber - the line number of the expression in the function.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction (inputOne i32) (outputOne i16)
//             0.- Expression: z i16 = add(x i16, y i16)
//
// We use RemoveInputFromExpression(cxprogram, "TestFunction", 0).
// The Result will be:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction (inputOne i32) (outputOne i16)
//             0.- Expression: z i16 = add(x i16)
//
// Note the y i16 removed from the expression in line 0.
// func RemoveInputFromExpression(cxprogram *cxast.CXProgram, functionName string, lineNumber int) error {
// 	fn, err := FunctionFind(cxprogram, functionName)
// 	if err != nil {
// 		return err
// 	}

// 	expr, err := fn.ExpressionByLineGet(lineNumber)
// 	if err != nil {
// 		return err
// 	}

// 	cxAtomicOp, err := cxprogram.CXAtomicOpGet(expr.Index)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// cxAtomicOp.RemoveInput()
// 	return nil
// }

// ExpressionNativeOutputAdd adds native output to
// an expression in a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageName - the name of the package where the function is located.
// functionName - the name of the function where the expression is located.
// outputName - the name of output to be added.
// outputType - the type of the output to be added.
// lineNumber - the line number of the expression in the function.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: add(x i16, y i16)
//
// We use ExpressionNativeOutputAdd(cxprogram, "main", "TestFunction", "z", cxtypes.I16, 0).
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z i16 = add(x i16, y i16)
//
// Note the z i16 added as an output of the expression in line 0.
func ExpressionNativeOutputAdd(cxprogram *ast.CXProgram, packageName, functionName, outputName string, outputType valuetypes.Code, lineNumber int) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}

	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}

	expr, err := fn.ExpressionByLineGet(cxprogram, lineNumber)
	if err != nil {
		return err
	}

	_, arg := ast.FieldMake(cxprogram, cxprogram.CXProgramNameIntern(outputName), outputType, ast.SourceFileGidNil, -1)
	arg.TypeSet(valuetypes.Code(outputType))
	arg.Package = ast.CXPackageIndex(pkg.Index)
	argIdx := ast.CXArgumentIndex(arg.Index)
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S1568.
	cxAtomicOp, err := cxprogram.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}

	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S1569.
	typeSigIdx := ast.CXTypeSignatureOfCXArgBuilderAlloc(cxprogram, cxprogram.CXArgFromArrayGet(ast.CXArgumentIndex(argIdx)))
	cxAtomicOp.OutputAdd(cxprogram, typeSigIdx)

	return nil
}

// RemoveOutputFromExpression removes an output
// from an expression in a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// functionName - the name of the function where the expression is located.
// lineNumber - the line number of the expression in the function.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction (inputOne i32) (outputOne i16)
//             0.- Expression: z i16 = add(x i16, y i16)
//
// We use RemoveOutputFromExpression(cxprogram, "TestFunction", 0).
// The Result will be:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction (inputOne i32) (outputOne i16)
//             0.- Expression: add(x i16, y i16)
//
// Note the z i16 removed as an output from the expression in line 0.
// func RemoveOutputFromExpression(cxprogram *cxast.CXProgram, functionName string, lineNumber int) error {
// 	fn, err := FunctionFind(cxprogram, functionName)
// 	if err != nil {
// 		return err
// 	}

// 	expr, err := fn.ExpressionByLineGet(lineNumber)
// 	if err != nil {
// 		return err
// 	}

// 	cxAtomicOp, err := cxprogram.CXAtomicOpGet(expr.Index)
// 	if err != nil {
// 		panic(err)
// 	}

// 	cxAtomicOp.RemoveOutput()
// 	return nil
// }

// ExpressionInputPointerMake makes an input of an
// expression a pointer.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// functionName - the name of the function where the expression is located.
// lineNumber - the line number of the expression in the function.
// expressionNumber
// inputNumber - the input number of the expression to be made as a pointer.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z i16 = add(x i16, y i16)
//
// We use ExpressionInputPointerMake(cxprogram, "TestFunction", 0, 1).
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z i16 = add(x i16, y *i16)
//
// Note that y *i16 became a pointer.
func ExpressionInputPointerMake(cxprogram *ast.CXProgram, functionName string, lineNumber, inputNumber int) error {
	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}

	expr, err := fn.ExpressionByLineGet(cxprogram, lineNumber)
	if err != nil {
		return err
	}

	if expr.Type == ast.CXOperationTypeLine {
		return errors.New("Expression is a CXLine")
	}
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S1571.
	cxAtomicOp, err := cxprogram.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}

	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S1572.
	cxAtomicOpInputTypeSig := cxprogram.CXTypeSignatureFromArrayGet(cxAtomicOp.InputsGet(cxprogram)[inputNumber])
	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S1573.
	ast.PointerMake(cxprogram.CXArgFromArrayGet(ast.CXArgumentIndex(cxAtomicOpInputTypeSig.Meta)))
	return nil
}

// ExpressionOutputPointerMake makes an output
// of an expression a pointer.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// functionName - the name of the function where the expression is located.
// lineNumber - the line number of the expression in the function.
// expressionNumber
// outputNumber - the output number of the expression to be made as a pointer.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z i16 = add(x i16, y i16)
//
// We use ExpressionOutputPointerMake(cxprogram, "TestFunction", 0, 0).
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	        0.- Expression: z *i16 = add(x i16, y i16)
//
// Note that z *i16 became a pointer.
func ExpressionOutputPointerMake(cxprogram *ast.CXProgram, functionName string, lineNumber, outputNumber int) error {
	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}

	expr, err := fn.ExpressionByLineGet(cxprogram, lineNumber)
	if err != nil {
		return err
	}

	if expr.Type == ast.CXOperationTypeLine {
		return errors.New("Expression is a CXLine")
	}
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S1574.
	cxAtomicOp, err := cxprogram.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}

	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S1575.
	cxAtomicOpOutputTypeSig := cxprogram.CXTypeSignatureFromArrayGet(cxAtomicOp.OutputsGet(cxprogram)[outputNumber])
	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S1576.
	ast.PointerMake(cxprogram.CXArgFromArrayGet(ast.CXArgumentIndex(cxAtomicOpOutputTypeSig.Meta)))
	return nil
}

// FunctionAccessibleArgsByTypeGet gets all accessible
// arguments in cx program for a function by specified
// argument type.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageLocationName - the name of the package where the function is located.
// functionName - the name of the function where the argument will be used.
// argType - the type of argument we are looking for.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Globals
//	        0.- Global: testGlobal
//	Functions
//	        0.- Function: TestFunction (inputOne i32) (outputOne i16)
//	                0.- Expression: add(x i16)
//
// We use FunctionAccessibleArgsByTypeGet(cxprogram, "main", "TestFunction", cxtypes.I16).
// We then get a result of
// 1. testGlobal i16 arg in the form of cxast.CXTypeSignature.
// 2. x i16 arg in the form of cxast.CXTypeSignature.
func FunctionAccessibleArgsByTypeGet(cxprogram *ast.CXProgram, packageLocationName, functionName string, argType valuetypes.Code) ([]*ast.CXTypeSignature, error) {
	var argsList []*ast.CXTypeSignature

	// Get all globals
	pkg, err := PackageFind(cxprogram, packageLocationName)
	if err != nil {
		return nil, err
	}

	for _, glblFldIdx := range pkg.Globals.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S1577.
		argsList = appendArgsByType(cxprogram, argsList, cxprogram.CXTypeSignatureFromArrayGet(glblFldIdx), argType)
	}

	for _, impIdx := range pkg.Imports {
		// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
		// Now an arena getter, site S1578.
		imp, err := cxprogram.PackageFromArrayGet(impIdx)
		if err != nil {
			panic(err)
		}
		for _, glblFldIdx := range imp.Globals.Fields {
			// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
			// Now an arena getter, site S1579.
			argsList = appendArgsByType(cxprogram, argsList, cxprogram.CXTypeSignatureFromArrayGet(glblFldIdx), argType)
		}
	}

	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return nil, err
	}

	// Get all args from expression inputs
	for _, expr := range fn.Expressions {
		cxAtomicOp, _, err := cxprogram.OperationGet(&expr)
		if err != nil {
			panic(err)
		}
		for _, inputArgIdx := range cxAtomicOp.InputsGet(cxprogram) {
			// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
			// Now an arena getter, site S1580.
			argsList = appendArgsByType(cxprogram, argsList, cxprogram.CXTypeSignatureFromArrayGet(inputArgIdx), argType)
		}
	}

	return argsList, nil
}

// appendArgsByType appends typeSig, or its struct fields, to argsList if
// they are of argType. typeSig may be atomic or a deprecated CXArgument.
func appendArgsByType(cxprogram *ast.CXProgram, argsList []*ast.CXTypeSignature, typeSig *ast.CXTypeSignature, argType valuetypes.Code) []*ast.CXTypeSignature {
	switch typeSig.Type {
	case ast.TypeAtomic, ast.TypePointerAtomic:
		if valuetypes.Code(typeSig.Meta) == argType {
			argsList = append(argsList, typeSig)
		}
	case ast.TypeCXArgumentDeprecate:
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1581.
		arg := cxprogram.CXArgFromArrayGet(ast.CXArgumentIndex(typeSig.Meta))
		if arg.StructTyped() {
			for _, fieldIdx := range arg.StructType.Fields {
				// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
				// Now an arena getter, site S1582.
				argsList = appendArgsByType(cxprogram, argsList, cxprogram.CXTypeSignatureFromArrayGet(fieldIdx), argType)
			}
		} else if arg.Type == argType {
			argsList = append(argsList, ast.CXTypeSignatureRepresentationOfCXArgGet(cxprogram, arg))
		}
	}
	return argsList
}

func ExpressionLiteralInputAdd(cxprogram *ast.CXProgram, packageName, functionName string, bytes []byte, argType valuetypes.Code, lineNumber int) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}
	cxprogram.CurrentPackage = ast.CXPackageIndex(pkg.Index)
	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}

	expr, err := fn.ExpressionByLineGet(cxprogram, lineNumber)
	if err != nil {
		return err
	}

	astbuild.AST = cxprogram
	litArg := astbuild.PrimaryWrite(cxprogram, argType, bytes, false)

	cxAtomicOp2, _, err := cxprogram.OperationGet(expr)
	if err != nil {
		panic(err)
	}

	litArg.Package = ast.CXPackageIndex(pkg.Index)

	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S1583.
	typeSigIdx := ast.CXTypeSignatureOfCXArgBuilderAlloc(cxprogram, cxprogram.CXArgFromArrayGet(ast.CXArgumentIndex(litArg.Index)))
	cxAtomicOp2.InputAdd(cxprogram, typeSigIdx)

	return nil
}
