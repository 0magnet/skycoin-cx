package ast

import (
	"errors"
	"fmt"

	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

type CXPackageIndex int

// CXPackageMethodKey names a method by the name of its receiver struct and
// its own name.
type CXPackageMethodKey struct {
	CXPackageMethodKeyStruct stringpool.StringGid
	CXPackageMethodKeyMethod stringpool.StringGid
}

// FunctionGet returns the function fnName of pkg. A function of an import
// is reached as pkg.name or through an import into scope, see #487.
func (pkg *CXPackage) FunctionGet(prgrm *CXProgram, fnName stringpool.StringGid) (*CXFunction, error) {
	if fnIdx, ok := pkg.Functions[fnName]; ok {
		fn := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx)

		return fn, nil
	}

	return nil, fmt.Errorf("function '%s' not found in package '%s'", prgrm.CXProgramNameText(fnName), prgrm.CXProgramNameText(pkg.Name))
}

// ImportGet ...
func (pkg *CXPackage) ImportGet(prgrm *CXProgram, impName stringpool.StringGid) (*CXPackage, error) {
	if _, ok := pkg.Imports[impName]; !ok {
		return nil, fmt.Errorf("package '%s' not imported", prgrm.CXProgramNameText(impName))
	}

	// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
	// Now an arena getter, site S0161.
	imp, err := prgrm.PackageFromArrayGet(pkg.Imports[impName])
	if err != nil {
		panic(err)
	}
	return imp, nil
}

// MethodGet returns the method methodName of the struct receiverName.
func (pkg *CXPackage) MethodGet(prgrm *CXProgram, receiverName stringpool.StringGid, methodName stringpool.StringGid) (CXFunctionGid, error) {
	if fnIdx, ok := pkg.Methods[CXPackageMethodKey{CXPackageMethodKeyStruct: receiverName, CXPackageMethodKeyMethod: methodName}]; ok {
		return fnIdx, nil
	}

	// Trying to find it among the built-ins.
	// Most likely a method from a core package.
	if native, found := prgrm.CXProgramNativeFind(&NativeNames, pkg.Name, receiverName, methodName); found {
		opFnIdx := prgrm.NativeFunctionInArrayAdd(NativeFunctionOf(prgrm, native.NativeNameEntryCode))
		return opFnIdx, nil
	}

	return CXFunctionGidNil, fmt.Errorf("method '%s.%s' not found in package '%s'", prgrm.CXProgramNameText(receiverName), prgrm.CXProgramNameText(methodName), prgrm.CXProgramNameText(pkg.Name))
}

// StructGet returns the struct strctName of pkg, not one of its imports.
func (pkg *CXPackage) StructGet(prgrm *CXProgram, strctName stringpool.StringGid) (*CXStruct, error) {
	if strctIdx, ok := pkg.Structs[strctName]; ok {
		// OBJMODEL: state.CXStructMng.CXStructGet(gid), phase 5 #337.
		// Now a pointer into an arena, site S0168.
		return &prgrm.CXStructs[strctIdx], nil
	}

	return nil, fmt.Errorf("struct '%s' not found in package '%s'", prgrm.CXProgramNameText(strctName), prgrm.CXProgramNameText(pkg.Name))

}

// GlobalGet ...
func (pkg *CXPackage) GlobalGet(prgrm *CXProgram, defName stringpool.StringGid) (*CXTypeSignature, error) {
	for _, fieldIdx := range pkg.Globals.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0171.
		field := prgrm.CXTypeSignatureFromArrayGet(fieldIdx)
		if field.Name == defName {
			return field, nil
		}
	}

	return nil, fmt.Errorf("global '%s' not found in package '%s'", prgrm.CXProgramNameText(defName), prgrm.CXProgramNameText(pkg.Name))
}

// CurrentFunctionGet ...
func (pkg *CXPackage) CurrentFunctionGet(prgrm *CXProgram) (*CXFunction, error) {
	if pkg.CurrentFunction == CXFunctionGidNil {
		return nil, errors.New("current function is nil")
	}

	return prgrm.CxIrState.CXFunctionMng.CXFunctionGet(pkg.CurrentFunction), nil
}

// ----------------------------------------------------------------
//                             `CXPackage` Selectors

// SelectFunction ...
func (pkg *CXPackage) SelectFunction(prgrm *CXProgram, name stringpool.StringGid) (*CXFunction, error) {
	if _, ok := pkg.Functions[name]; !ok {
		return nil, fmt.Errorf("function '%s' does not exist", prgrm.CXProgramNameText(name))
	}

	idx := pkg.Functions[name]
	pkg.CurrentFunction = idx

	return prgrm.CxIrState.CXFunctionMng.CXFunctionGet(idx), nil
}

// PackageMake creates a new empty CXPackage.
//
// It can be filled in later with imports, structs, globals and functions.
// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
// Now a constructor, site S0174.
func PackageMake(name stringpool.StringGid) *CXPackage {
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now a heap literal, site S0175.
	return &CXPackage{
		Name: name,
		// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
		// Now a heap literal, site S0176.
		Globals:         &CXStruct{},
		Imports:         make(map[stringpool.StringGid]CXPackageIndex, 0),
		Structs:         make(map[stringpool.StringGid]CXStructIndex, 0),
		Functions:       make(map[stringpool.StringGid]CXFunctionGid, 0),
		Methods:         make(map[CXPackageMethodKey]CXFunctionGid, 0),
		CurrentFunction: CXFunctionGidNil,
	}
}

// ----------------------------------------------------------------
//                             `CXPackage` Getters

// ----------------------------------------------------------------
//                     `CXPackage` Member handling

// ImportAdd ...
func (pkg *CXPackage) ImportAdd(prgrm *CXProgram, imp *CXPackage) *CXPackage {
	if _, ok := pkg.Imports[imp.Name]; !ok {
		// impIdx := prgrm.PackageInArrayAdd(imp)
		pkg.Imports[imp.Name] = CXPackageIndex(imp.Index)
	}

	return pkg
}

// ImportRemove ...
func (pkg *CXPackage) ImportRemove(impName stringpool.StringGid) {
	if _, ok := pkg.Imports[impName]; ok {
		// OBJMODEL: state.CXPackageMng.CXPackageRelease(gid), phase 4 #361.
		// Now a dropped object, site S0177.
		delete(pkg.Imports, impName)
	}
}

// FunctionAdd adds fn, allocated by FunctionMake, to pkg. A second function of the same name is an
// error and stays out of pkg.Functions, so the first keeps the name as in Go.
func (pkg *CXPackage) FunctionAdd(prgrm *CXProgram, fn *CXFunction) (*CXPackage, CXFunctionGid) {
	return pkg, pkg.addFunction(prgrm, fn, CompileErrorCodeRedeclaredBlock)
}

func (pkg *CXPackage) addFunction(prgrm *CXProgram, fn *CXFunction, code CompileErrorCode) CXFunctionGid {
	fn.Package = CXPackageIndex(pkg.Index)

	fnIdx := prgrm.CXProgramFunctionGid(fn)
	if firstIdx, ok := pkg.Functions[fn.Name]; ok {
		first := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(firstIdx)
		column := fn.FileColumn
		if column == 0 {
			column = prgrm.CXProgramCompileErrorColumn(fn.FileName, fn.FileLine)
		}
		prgrm.CXProgramCompileErrorAddPosition(code, fn.FileName, fn.FileLine, column, CompileErrorOperands{int64(fn.Name), int64(first.FileName), int64(first.FileLine), int64(first.FileColumn)})
	} else {
		pkg.Functions[fn.Name] = fnIdx
	}
	pkg.CurrentFunction = fnIdx

	return fnIdx
}

// MethodAdd adds fn, the method methodName of the struct receiverName. The
// name of fn is "receiverName.methodName", see CXProgramMethodNameIntern.
func (pkg *CXPackage) MethodAdd(prgrm *CXProgram, fn *CXFunction, receiverName stringpool.StringGid, methodName stringpool.StringGid) (*CXPackage, CXFunctionGid) {
	key := CXPackageMethodKey{CXPackageMethodKeyStruct: receiverName, CXPackageMethodKeyMethod: methodName}
	_, declared := pkg.Methods[key]
	fnIdx := pkg.addFunction(prgrm, fn, CompileErrorCodeMethodRedeclared)
	if !declared {
		pkg.Methods[key] = fnIdx
	}
	return pkg, fnIdx
}

// FunctionDeclarationAt returns the function declared at file, line and
// column with the name and package of fnIdx. It is fnIdx unless the name was
// declared twice, or for a line of 0.
func (prgrm *CXProgram) FunctionDeclarationAt(fnIdx CXFunctionGid, file SourceFileGid, line int, column int) CXFunctionGid {
	fn := prgrm.CxIrState.CXFunctionMng.CXFunctionGet(fnIdx)
	if line == 0 || fn.FileName == file && fn.FileLine == line && fn.FileColumn == column {
		return fnIdx
	}
	for index := 0; index < prgrm.CXProgramFunctionSlotCount(); index++ {
		other := prgrm.CXProgramFunctionSlotLive(index)
		if other != nil && other.Name == fn.Name && other.Package == fn.Package && other.FileName == file && other.FileLine == line && other.FileColumn == column {
			return prgrm.CXProgramFunctionSlotGid(uint32(index))
		}
	}
	return fnIdx
}

// FunctionRemove removes the function fnName from pkg and releases it.
func (pkg *CXPackage) FunctionRemove(prgrm *CXProgram, fnName stringpool.StringGid) {
	fnIdx, ok := pkg.Functions[fnName]
	if !ok {
		return
	}
	for key, methodIdx := range pkg.Methods {
		if methodIdx == fnIdx {
			delete(pkg.Methods, key)
		}
	}

	delete(pkg.Functions, fnName)
	if pkg.CurrentFunction == fnIdx {
		pkg.CurrentFunction = CXFunctionGidNil
	}
	prgrm.CXProgramFunctionRelease(fnIdx)
}

// StructAdd ...
func (pkg *CXPackage) StructAdd(prgrm *CXProgram, strct *CXStruct) *CXPackage {
	if _, ok := pkg.Structs[strct.Name]; ok {
		return pkg
	}
	strct.Package = CXPackageIndex(pkg.Index)
	// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
	// Now an arena insert, site S0180.
	strctIdx := prgrm.StructInArrayAdd(strct)
	pkg.Structs[strct.Name] = strctIdx

	return pkg
}

// StructRemove ...
func (pkg *CXPackage) StructRemove(strctName stringpool.StringGid) {
	if _, ok := pkg.Structs[strctName]; !ok {
		return
	}
	// OBJMODEL: state.CXStructMng.CXStructRelease(gid), phase 4 #336.
	// Now a dropped object, site S0181.
	delete(pkg.Structs, strctName)
}

// GlobalAdd ...
func (pkg *CXPackage) GlobalAdd(prgrm *CXProgram, defIdx CXArgumentIndex) *CXPackage {
	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S0182.
	defArg := prgrm.CXArgFromArrayGet(defIdx)
	// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
	// Now an arena getter, site S1631.
	prgrm.CXArgGet(CXArgumentIndex(defIdx)).Package = CXPackageIndex(pkg.Index)

	for _, fieldIdx := range pkg.Globals.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0183.
		field := prgrm.CXTypeSignatureFromArrayGet(fieldIdx)
		if field.Name == defArg.Name && field.Type == TypeCXArgumentDeprecate {
			return pkg
		}
	}

	pkg.Globals.FieldGlobalsCXAtomicOpsAdd(prgrm, defIdx)

	return pkg
}

func (pkg *CXPackage) GlobalTypeSignatureAdd(prgrm *CXProgram, typeSignatureIdx CXTypeSignatureIndex) *CXPackage {
	// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
	// Now an arena getter, site S0184.
	typeSignature := prgrm.CXTypeSignatureFromArrayGet(typeSignatureIdx)
	typeSignature.Package = CXPackageIndex(pkg.Index)
	for _, fieldIdx := range pkg.Globals.Fields {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S0185.
		field := prgrm.CXTypeSignatureFromArrayGet(fieldIdx)
		if field.Name == typeSignature.Name {
			return pkg
		}
	}

	pkg.Globals.FieldTypeSignatureAdd(prgrm, typeSignatureIdx)

	return pkg
}

// RemoveGlobal ...
// func (pkg *CXPackage) RemoveGlobal(prgrm *CXProgram, defName string) {
// 	lenGlobals := len(pkg.Globals)
// 	for i, defIdx := range pkg.Globals {
// 		def := prgrm.CXArgGet(defIdx)
// 		if def.Name == defName {
// 			if i == lenGlobals-1 {
// 				pkg.Globals = pkg.Globals[:len(pkg.Globals)-1]
// 			} else {
// 				pkg.Globals = append(pkg.Globals[:i], pkg.Globals[i+1:]...)
// 			}
// 			break
// 		}
// 	}
// }
