package actions

import (
	"math/bits"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-ast"
)

// DeclarationSpecifiers is called to build a type of a variable or parameter.
//
// It is called repeatedly while the type is parsed.
//
// Returns the new type build from `declSpec` and `opTyp`.
//
// Input arguments description:
//
//	declSpec:     The incoming type
//	arrayLengths: The lengths of the array if `opTyp` = cxcore.DeclarationArray
//	opTyp:        The type of modification to `declSpec` (array of, pointer to, ...)
func DeclarationSpecifiers(declSpec *ast.CXArgument, arrayLengths []types.Pointer, opTyp int) *ast.CXArgument {
	switch opTyp {
	case constants.DeclarationPointer:
		declSpec.DeclarationSpecifiers = append(declSpec.DeclarationSpecifiers, constants.DeclarationPointer)
		declSpec.Size = types.PointerSize

		if declSpec.Type == types.STR || declSpec.StructType != nil {
			declSpec.PointerTargetType = declSpec.Type
			declSpec.Type = types.POINTER
		}

		return declSpec
	case constants.DeclarationArray:
		for range arrayLengths {
			declSpec.DeclarationSpecifiers = append(declSpec.DeclarationSpecifiers, constants.DeclarationArray)
		}
		// Lengths are outermost first, as written: [2][3]i32 is {2, 3}.
		arg := declSpec
		arg.Lengths = append(append([]types.Pointer{}, arrayLengths...), arg.Lengths...)

		return arg
	case constants.DeclarationSlice:
		arg := declSpec
		arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, constants.DeclarationSlice)
		arg.Lengths = append([]types.Pointer{0}, arg.Lengths...)

		return arg
	case constants.DeclarationBasic:
		arg := declSpec

		return arg
	case constants.DeclarationFunc:
		// Creating this case if additional operations are needed in the
		// future.
		return declSpec
	}

	return nil
}

// DeclarationSpecifiersBasic() returns a type specifier created from one of the builtin types.
func DeclarationSpecifiersBasic(prgrm *ast.CXProgram, typeCode types.Code) *ast.CXArgument {
	_, arg := ast.ArgumentMake(prgrm, stringpool.StringGidNil, CurrentFile, LineNo)
	arg.TypeSet(typeCode)
	return DeclarationSpecifiers(arg, []types.Pointer{0}, constants.DeclarationBasic)
}

// DeclarationSpecifiersStruct() declares a struct. The errors of a missing
// struct or package are at currentFile, lineNo and column, where the type
// name starts, or at the token the parser is at for a column of 0.
func DeclarationSpecifiersStruct(prgrm *ast.CXProgram, ident string, pkgName string,
	external bool, currentFile ast.SourceFileGid, lineNo int, column int) *ast.CXArgument {
	pkg, err := prgrm.CurrentPackageGet()
	if err != nil {
		panic(err)
	}
	if column == 0 {
		column = prgrm.CXProgramCompileErrorColumn(currentFile, lineNo)
	}

	identGid := prgrm.CXProgramNameIntern(ident)
	var strct *ast.CXStruct
	if external {
		// custom type in an imported package
		imp, err := pkg.ImportGet(prgrm, prgrm.CXProgramNameIntern(pkgName))
		if err != nil {
			prgrm.CXProgramCompileErrorAddPosition(ast.CompileErrorCodeMessage, currentFile, lineNo, column, ast.CompileErrorOperands{prgrm.CXProgramCompileErrorText(err.Error())})

			return declarationSpecifiersMissing(prgrm, currentFile, lineNo)
		}

		strct, err = prgrm.StructGet(identGid, imp.Name)
		if err != nil {
			// A package whose import failed has no types, and its import is the
			// error. Its uses report nothing more.
			if prgrm.CXProgramImportFailed(imp.Name) {
				return declarationSpecifiersMissing(prgrm, currentFile, lineNo)
			}
			prgrm.CXProgramCompileErrorAddPosition(ast.CompileErrorCodeMessage, currentFile, lineNo, column, ast.CompileErrorOperands{prgrm.CXProgramCompileErrorText(err.Error())})

			return declarationSpecifiersMissing(prgrm, currentFile, lineNo)
		}
	} else {
		// custom type in the current package, or one an import into scope brings in
		strct, err = prgrm.StructGet(identGid, pkg.Name)
		if err != nil {
			if scopedStruct := importScopedStruct(prgrm, pkg, currentFile, identGid); scopedStruct != nil {
				strct, err = scopedStruct, nil
			}
		}
		if err != nil {
			prgrm.CXProgramCompileErrorAddPosition(ast.CompileErrorCodeMessage, currentFile, lineNo, column, ast.CompileErrorOperands{prgrm.CXProgramCompileErrorText(err.Error())})

			return declarationSpecifiersMissing(prgrm, currentFile, lineNo)
		}
	}

	_, arg := ast.ArgumentMake(prgrm, stringpool.StringGidNil, currentFile, lineNo)
	arg.Type = types.STRUCT
	arg.StructType = strct
	arg.Size = strct.StructSizeGet(prgrm)
	arg.Package = ast.CXPackageIndex(pkg.Index)
	arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, constants.DeclarationStruct)

	return arg
}

// declarationSpecifiersMissing stands in for a type that does not exist, after
// its error. The variables of the type are declared types.INVALID, so their
// uses report nothing more.
func declarationSpecifiersMissing(prgrm *ast.CXProgram, currentFile ast.SourceFileGid, lineNo int) *ast.CXArgument {
	placeholder := declarationSpecifiersPlaceholder(prgrm, currentFile, lineNo)
	placeholder.Type = types.INVALID
	return placeholder
}

// declarationSpecifiersPlaceholder stands in for a type that failed to resolve
// after its compile error is recorded, so parsing goes on without a nil.
func declarationSpecifiersPlaceholder(prgrm *ast.CXProgram, currentFile ast.SourceFileGid, lineNo int) *ast.CXArgument {
	_, arg := ast.ArgumentMake(prgrm, stringpool.StringGidNil, currentFile, lineNo)
	arg.TypeSet(types.UNDEFINED)
	arg.Size = 0
	return arg
}

// declarationSizeLimit caps a declaration, a frame and the data segment at the
// range of 32 bit pointers, so offsets never wrap in either build.
const declarationSizeLimit = uint64(types.MaxUint32)

// declarationArgSize returns the bytes a declaration of arg takes, and false
// when they pass declarationSizeLimit. It follows ast.ArgSizeGet.
func declarationArgSize(arg *ast.CXArgument) (types.Pointer, bool) {
	// The last specifier is the outermost, so [3]*i32 is an array of pointers.
	for i := len(arg.DeclarationSpecifiers) - 1; i >= 0; i-- {
		switch arg.DeclarationSpecifiers[i] {
		case constants.DeclarationArray:
			total, lengths := uint64(arg.Size), arg.Lengths
			if direct, behind := arg.DirectDimensions(); behind {
				total, lengths = uint64(types.PointerSize), lengths[:direct]
			} else if dims := arg.ArraySliceDimensions(); dims > 0 {
				total, lengths = uint64(types.PointerSize), lengths[:dims]
			}
			for _, length := range lengths {
				high, low := bits.Mul64(total, uint64(length))
				if high != 0 || low > declarationSizeLimit {
					return 0, false
				}
				total = low
			}
			return types.Pointer(total), true
		case constants.DeclarationPointer:
			return arg.Size, true
		case constants.DeclarationSlice:
			return types.PointerSize, true
		}
	}
	return arg.Size, uint64(arg.Size) <= declarationSizeLimit
}

// declarationSizeSumFits reports whether size bytes more after used stay
// within declarationSizeLimit.
func declarationSizeSumFits(used types.Pointer, size types.Pointer) bool {
	return uint64(used)+uint64(size) <= declarationSizeLimit
}
