package actions

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// declarationSpecifierFound determines if an argument has certain declaration specifier.
func declarationSpecifierFound(arg *ast.CXArgument, spec int) bool {
	found := false
	for _, s := range arg.DeclarationSpecifiers {
		if s == spec {
			found = true
		}
	}
	return found
}

// derefOpFound determines if an argument has certain dereference operation
func derefOpFound(arg *ast.CXArgument, spec int) bool {
	found := false
	for _, s := range arg.DereferenceOperations {
		if s == spec {
			found = true
		}
	}
	return found
}

// PrimaryWrite writes those bytes to prgrm.Data and returns a new argument
// for them. The caller takes the argument or releases it. A str
// literal points to itself, storage being the zero value of a variable,
// whose pointer stays nil.
func PrimaryWrite(prgrm *ast.CXProgram, typeCode valuetypes.Code, byts []byte, storage bool) *ast.CXArgument {
	pkg, err := prgrm.CurrentPackageGet()
	if err != nil {
		panic(err)
	}

	_, arg := ast.ArgumentMake(prgrm, stringpool.StringGidNil, CurrentFile, LineNo)
	arg.TypeSet(typeCode)
	arg.Package = ast.CXPackageIndex(pkg.Index)

	size := valuetypes.IntToPtrCast(len(byts))

	arg.Size = typeCode.Size()
	arg.Offset = prgrm.Data.Size + prgrm.Data.StartsAt

	if arg.Type == valuetypes.STR {
		arg.Size = valuetypes.PointerSize
		if !storage {
			valuetypes.PtrWrite(byts, 0, arg.Offset)
		}
	}

	// A CX program allocates min(HeapSizeInit, MaxHeapSize) bytes
	// after the stack segment. These bytes are used to allocate the data segment
	// at compile time. If the data segment is bigger than min(HeapSizeInit, MaxHeapSize),
	// we'll start appending the bytes to prgrm.Memory.
	// After compilation, we calculate how many bytes we need to add to have a heap segment
	// equal to `minHeapSize()` that is allocated after the data segment.
	memSize := valuetypes.IntToPtrCast(len(prgrm.Memory))
	if (size + prgrm.Data.Size + prgrm.Data.StartsAt) > memSize {
		var i valuetypes.Pointer
		// First we need to fill the remaining free bytes in
		// the current `prgrm.Memory` slice.
		for i = valuetypes.Pointer(0); i < memSize-(prgrm.Data.Size+prgrm.Data.StartsAt); i++ {
			prgrm.Memory[prgrm.Data.Size+prgrm.Data.StartsAt+i] = byts[i]
		}
		// Then we append the bytes that didn't fit.
		prgrm.Memory = append(prgrm.Memory, byts[i:]...)
	} else {
		for i, byt := range byts {
			prgrm.Memory[prgrm.Data.Size+prgrm.Data.StartsAt+valuetypes.IntToPtrCast(i)] = byt
		}
	}
	prgrm.Data.Size += size
	prgrm.Heap.StartsAt = prgrm.Data.Size + prgrm.Data.StartsAt

	return arg
}

// PrimaryWrite writes those bytes to prgrm.Data and returns a []CXExpression.
func PrimaryExpressionsWrite(prgrm *ast.CXProgram, typeCode valuetypes.Code, byts []byte, sliceTyped bool) []ast.CXExpression {
	pkg, err := prgrm.CurrentPackageGet()
	if err != nil {
		panic(err)
	}

	arg := PrimaryWrite(prgrm, typeCode, byts, sliceTyped)

	// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
	// Now a constructor call, site S1055.
	expr := ast.AtomicOperatorExpressionMake(prgrm, nil)
	prgrm.CXAtomicOps[expr.Index].Package = ast.CXPackageIndex(pkg.Index)

	typeSigIdx := ast.CXTypeSignatureOfCXArgBuilderAlloc(prgrm, arg)
	prgrm.CXAtomicOps[expr.Index].OutputAdd(prgrm, typeSigIdx)

	// OBJMODEL: state.CXExpressionMng (manager storage) or []CXExpressionGid, phase 4 #346.
	// Now a slice of objects, site S1057.
	return []ast.CXExpression{*expr}
}

// PrimaryIdentifier creates an identifier expression with an output name of 'ident'.
func PrimaryIdentifier(prgrm *ast.CXProgram, ident string) []ast.CXExpression {
	identGid := prgrm.CXProgramNameIntern(ident)
	if exprs := importScopedIdentifier(prgrm, identGid); exprs != nil {
		return exprs
	}
	return primaryIdentifierExpressions(prgrm, identGid)
}

// primaryIdentifierExpressions is PrimaryIdentifier for an interned name.
func primaryIdentifierExpressions(prgrm *ast.CXProgram, identGid stringpool.StringGid) []ast.CXExpression {
	pkg, err := prgrm.CurrentPackageGet()
	if err != nil {
		panic(err)
	}

	_, arg := ast.ArgumentMake(prgrm, identGid, CurrentFile, LineNo) // fix: line numbers in errors sometimes report +1 or -1. Issue #195
	arg.TypeSet(valuetypes.IDENTIFIER)
	arg.Package = ast.CXPackageIndex(pkg.Index)

	// OBJMODEL: state.CXExpressionMng.CXExpressionAlloc(), phase 4 #346.
	// Now a constructor call, site S1080.
	expr := ast.AtomicOperatorExpressionMake(prgrm, nil)
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S1081.
	expression, err := prgrm.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}

	typeSigIdx := ast.CXTypeSignatureOfCXArgBuilderAlloc(prgrm, arg)
	expression.OutputAdd(prgrm, typeSigIdx)
	expression.Package = ast.CXPackageIndex(pkg.Index)
	// OBJMODEL: state.CXExpressionMng (manager storage) or []CXExpressionGid, phase 4 #346.
	// Now a slice of objects, site S1083.
	return []ast.CXExpression{*expr}
}

// ArgsAllBasicTypes checks if all the input arguments in an expressions are of basic type.
func ArgsAllBasicTypes(prgrm *ast.CXProgram, expr *ast.CXExpression) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S1084.
	expression, err := prgrm.CXAtomicOpGet(expr.Index)
	if err != nil {
		panic(err)
	}

	for _, inputIdx := range expression.InputsGet(prgrm) {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S1085.
		input := prgrm.CXTypeSignatureFromArrayGet(inputIdx)

		var inpType valuetypes.Code

		var inp *ast.CXArgument
		if input.Type == ast.TypeCXArgumentDeprecate {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S1087.
			inp = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(input.Meta))

			inpType = inp.Type
			if inp.Type == valuetypes.POINTER {
				inpType = inp.PointerTargetType
			}
		} else {
			inpType = input.TypeGet(prgrm)
		}

		// TODO: Check why STR is considered as basic type.
		if !inpType.Primitive() && inpType != valuetypes.STR {
			return false
		}
	}
	return true
}

// TokenPosition is a position of the parser, see TokenPositionPush.
type TokenPosition struct {
	TokenPositionLine   int
	TokenPositionColumn int
}

// TokenPositionPush makes line and column, where the rule being reduced
// starts, the column of the records made until TokenPositionPop, as the
// parser may have read the token after the rule already.
func TokenPositionPush(line int, column int) TokenPosition {
	list := &AST.CxIrState.CompileErrorList
	position := TokenPosition{list.CompileErrorListTokenLine, list.CompileErrorListTokenColumn}
	AST.CXProgramCompileErrorTokenSet(CurrentFile, line, column)
	return position
}

// TokenPositionPop restores the position TokenPositionPush returned.
func TokenPositionPop(position TokenPosition) {
	AST.CXProgramCompileErrorTokenSet(CurrentFile, position.TokenPositionLine, position.TokenPositionColumn)
}
