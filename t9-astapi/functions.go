package astapi

import (
	"github.com/skycoin/skycoin-cx/t0-types" // thin: only Code; no depth gain
	cxast "github.com/skycoin/skycoin-cx/t3-ast"
)

// PackageEmptyFunctionAdd adds an empty function to a package in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageName - the name of the package where the
//
//	function to be added will be located.
//
// functionName - the name of the function to be added.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
// We use PackageEmptyFunctionAdd(cxprogram, "main", "TestFunction").
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction()()
//
// Note the new function added to the package.
func PackageEmptyFunctionAdd(cxprogram *cxast.CXProgram, packageName, functionName string) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}

	_, fn := cxast.FunctionMake(cxprogram, cxprogram.CXProgramNameIntern(functionName), cxast.SourceFileGidNil, -1)

	pkg.FunctionAdd(cxprogram, fn)

	return nil
}

// PackageFunctionRemove removes a function from a package in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageName - the name of the package where the
//
//	function to be removed is located.
//
// functionName - the name of the function to be removed.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction()()
//
// We use PackageFunctionRemove(cxprogram, "main", "TestFunction").
// The Result will be:
// 0.- Package: main
//
// Note the TestFunction was removed from the package.
func PackageFunctionRemove(cxprogram *cxast.CXProgram, packageName, functionName string) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}
	if functionNameGid, found := cxprogram.CXProgramNameFind(functionName); found {
		pkg.FunctionRemove(cxprogram, functionNameGid)
	}

	return nil
}

// FunctionNativeInputAdd adds a native input to a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageName - the name of the package where the
//
//	function is located.
//
// functionName - the name of the function where the
//
//	input is to be added.
//
// inputName - the name of the input to be added.
// inputType - the type of the input to be added.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction()()
//
// We use FunctionNativeInputAdd(cxprogram, "main", "TestFunction", "inputOne", cxtypes.I32).
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction(inputOne i32) ()
//
// Note the new inputOne i32 input to the TestFunction.
func FunctionNativeInputAdd(cxprogram *cxast.CXProgram, packageName, functionName, inputName string, inputType types.Code) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}

	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}

	_, arg := cxast.FieldMake(cxprogram, cxprogram.CXProgramNameIntern(inputName), inputType, cxast.SourceFileGidNil, -1)
	arg.TypeSet(types.Code(inputType))
	arg.Package = cxast.CXPackageIndex(pkg.Index)
	fn.InputAdd(cxprogram, arg)

	return nil
}

// RemoveFunctionInput removes an input from a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// functionName - the name of the function where the
// 				  input is to be removed.
// inputName - the name of the input to be removed.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction(inputOne i32) ()
//
// We use RemoveFunctionInput(cxprogram, "TestFunction", "inputOne").
// The Result will be:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction() ()
//
// Note the inputOne i32 input was removed from the TestFunction.
// func RemoveFunctionInput(cxprogram *cxast.CXProgram, functionName, inputName string) error {
// 	fn, err := FunctionFind(cxprogram, functionName)
// 	if err != nil {
// 		return err
// 	}

// 	fn.RemoveInput(cxprogram, inputName)
// 	return nil
// }

// FunctionNativeOutputAdd adds a native output to a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// packageName - the name of the package where the
//
//	function is located.
//
// functionName - the name of the function where the
//
//	output is to be added.
//
// outputName - the name of the output to be added.
// outputType - the type of the output to be added.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction()()
//
// We use FunctionNativeOutputAdd(cxprogram, "main", "TestFunction", "outputOne", cxtypes.I32).
// The Result will be:
// 0.- Package: main
//
//	Functions
//	    0.- Function: TestFunction() (outputOne i32)
//
// Note the new outputOne i32 output of the TestFunction.
func FunctionNativeOutputAdd(cxprogram *cxast.CXProgram, packageName, functionName, outputName string, outputType types.Code) error {
	pkg, err := PackageFind(cxprogram, packageName)
	if err != nil {
		return err
	}

	fn, err := FunctionFind(cxprogram, functionName)
	if err != nil {
		return err
	}
	_, arg := cxast.FieldMake(cxprogram, cxprogram.CXProgramNameIntern(outputName), outputType, cxast.SourceFileGidNil, -1)
	arg.TypeSet(types.Code(outputType))
	arg.Package = cxast.CXPackageIndex(pkg.Index)
	fn.OutputAdd(cxprogram, arg)

	return nil
}

// RemoveFunctionOutput removes an output from a function in cx program.
//
// The inputs are
// cxprogram - in the form of cxast.CXProgram.
// functionName - the name of the function where the
// 				  output is to be removed.
// outputName - the name of the output to be removed.
//
// Example:
// We have this CX Program:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction() (outputOne i32)
//
// We use RemoveFunctionOutput(cxprogram, "TestFunction", "outputOne").
// The Result will be:
// 0.- Package: main
//     Functions
//         0.- Function: TestFunction() ()
//
// Note the outputOne i32 output was removed from the TestFunction.
// func RemoveFunctionOutput(cxprogram *cxast.CXProgram, functionName, outputName string) error {
// 	fn, err := FunctionFind(cxprogram, functionName)
// 	if err != nil {
// 		return err
// 	}

// 	fn.RemoveOutput(cxprogram, outputName)
// 	return nil
// }
