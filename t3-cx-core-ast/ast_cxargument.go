package ast

import (
	"slices"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t2-cx-core-environment" // thin: only DebugGolangStackTrace, FoundCompileErrors, HeapOffset; core globals, keep
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

type CXArgumentIndex int

func (arg CXArgument) PointerTyped() bool {
	return arg.Type == valuetypes.POINTER
}

func (arg CXArgument) StringTyped() bool {
	return arg.PointerTargetType == valuetypes.STR || arg.Type == valuetypes.STR
}

// CXArgumentStringSlicePointer reports whether arg is a *[]str, which takes
// an address as a *[]i32 does instead of the string path of a *str.
func CXArgumentStringSlicePointer(arg *CXArgument) bool {
	return arg.PointerTyped() && arg.PointerTargetType == valuetypes.STR && arg.SliceTyped()
}

// CXArgumentTakesAddress reports whether assigning input to the output element
// out writes the address of input rather than copying its bytes.
func CXArgumentTakesAddress(prgrm *CXProgram, out *CXArgument, input *CXTypeSignature) bool {
	if out.PassBy == opcodetable.PassByReference {
		return true
	}
	// A pointer that keeps the type of its target, as a *[]*i32 does, takes &v.
	if specs := out.DeclarationSpecifiers; len(specs) > 0 && specs[len(specs)-1] == opcodetable.DeclarationPointer && len(out.DereferenceOperations) == 0 && CXTypeSignaturePassBy(prgrm, input) == opcodetable.PassByReference {
		return true
	}
	return out.PointerTyped() && out.PointerTargetType != valuetypes.STR && !CXTypeSignaturePointerValue(prgrm, input)
}

// CXTypeSignatureStructTakesAddress reports whether assigning input to the
// struct or pointer struct output out writes the address of input.
func CXTypeSignatureStructTakesAddress(prgrm *CXProgram, out *CXTypeSignature, input *CXTypeSignature) bool {
	if out.Type == TypeStruct && out.PassBy == opcodetable.PassByReference {
		return true
	}
	// A field such as b.p is the pointer when it is last.
	pointer := out.Type == TypePointerStruct && !out.Deref
	if fields := prgrm.CXTypeSignatureStructFromArrayGet(out.Meta).Fields; len(fields) > 0 {
		pointer = cxArgumentPointerField(prgrm.CXArgFromArrayGet(fields[len(fields)-1]))
	}
	return pointer && !CXTypeSignaturePointerValue(prgrm, input)
}

// CXTypeSignaturePointerValue reports whether sig reads a pointer by value, as s
// or pa[0] do, so assigning it copies the pointer. &x and a struct literal do not.
func CXTypeSignaturePointerValue(prgrm *CXProgram, sig *CXTypeSignature) bool {
	if CXTypeSignaturePassBy(prgrm, sig) == opcodetable.PassByReference {
		return false
	}
	switch sig.Type {
	case TypeStruct, TypePointerStruct:
		// A field access such as b.p is decided by its last field. Deref of
		// rb.p reads rb, not the field.
		if fields := prgrm.CXTypeSignatureStructFromArrayGet(sig.Meta).Fields; len(fields) > 0 {
			return cxArgumentPointerField(prgrm.CXArgFromArrayGet(fields[len(fields)-1]).AssignmentElementGet(prgrm))
		}
		return sig.Type == TypePointerStruct && !sig.Deref
	}
	if sig.Deref {
		return false
	}
	switch sig.Type {
	case TypePointerAtomic, TypePointerArrayAtomic, TypePointerSliceAtomic:
		return true
	case TypeCXArgumentDeprecate:
		elt := prgrm.CXArgFromArrayGet(CXArgumentIndex(sig.Meta)).AssignmentElementGet(prgrm)
		if elt.PointerTyped() {
			return true
		}
		// A *i32 keeps the type i32 and ends its specifiers with the pointer. A
		// &T{} literal ends them so too but is a struct.
		specs := elt.DeclarationSpecifiers
		// An element such as (*v)[0] of a *[]*i32 reads the pointer too.
		ops := elt.DereferenceOperations
		element := len(ops) > 0 && (ops[len(ops)-1] == opcodetable.DerefArray || ops[len(ops)-1] == opcodetable.DerefSlice)
		return elt.StructType == nil && (len(ops) == 0 || element) && len(specs) > 0 && specs[len(specs)-1] == opcodetable.DeclarationPointer
	}
	return false
}

func (arg CXArgument) StructTyped() bool {
	return arg.StructType != nil
}

func (arg CXArgument) SliceTyped() bool {
	if arg.ArraySliceMixed() {
		// An element of [][N]T, as a[i] is, is an array, and of [N][]T a slice.
		return len(arg.Indexes) < len(arg.Lengths) && arg.Lengths[len(arg.Indexes)] == 0
	}
	if len(arg.Lengths) > 0 {
		for _, val := range arg.Lengths {
			if val != 0 {
				return false
			}
		}

		return true
	}

	return false
}

/*
	FileName              string
- filename and line number
- can be moved to CX AST annotations (comments to be skipped or map)

	FileLine
*/

/*

Note: PAssBy is not used too many place
Note: Low priority for deprecation
- isnt this same as "pointer"
*/

// ----------------------------------------------------------------
//                             `CXArgument` Getters

// AssignmentElementGet ...
func (arg *CXArgument) AssignmentElementGet(prgrm *CXProgram) *CXArgument {
	if len(arg.Fields) > 0 {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0068.
		return prgrm.CXArgFromArrayGet(arg.Fields[len(arg.Fields)-1]).AssignmentElementGet(prgrm)
	}
	return arg

}

// TypeGet ...
func (arg *CXArgument) TypeGet(prgrm *CXProgram) valuetypes.Code {
	fieldCount := len(arg.Fields)
	if fieldCount > 0 {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S0069.
		return prgrm.CXArgFromArrayGet(arg.Fields[fieldCount-1]).TypeGet(prgrm)
	}

	if arg.Type == valuetypes.POINTER {
		return arg.PointerTargetType
	}
	return arg.Type
}

func (arg *CXArgument) SizeGet() valuetypes.Pointer {
	if arg.Type != valuetypes.I8 {
		return arg.Type.Size()
	}

	return arg.Size
}

// ----------------------------------------------------------------
//                     `CXArgument` Member handling

// PackageSet sets CX package `pkg` of CX argument `arg`.
func (arg *CXArgument) PackageSet(pkg *CXPackage) *CXArgument {
	arg.Package = CXPackageIndex(pkg.Index)
	return arg
}

// TypeSet ...
func (arg *CXArgument) TypeSet(typeCode valuetypes.Code) *CXArgument {
	arg.Type = typeCode
	size := typeCode.Size()
	arg.Size = size
	arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, opcodetable.DeclarationBasic)
	return arg
}

// PointerMake takes an already defined `CXArgument` and turns it into a pointer.
// Only used once, deprecate
// TODO: only used by HTTP, create a better module system
func PointerMake(arg *CXArgument) *CXArgument {
	arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, opcodetable.DeclarationPointer)
	arg.PointerTargetType = arg.Type
	arg.Type = valuetypes.POINTER
	arg.Size = valuetypes.PointerSize

	return arg
}

// StructParameterMake helper for creating a struct parameter. It creates a
// `CXArgument` named `argName`, that represents a structure instane of
// `strctName`, from package `pkgName`.
// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
// Now a constructor, site S0072.
func StructParameterMake(prgrm *CXProgram, pkgName, strctName, argName string) *CXArgument {
	pkg, err := prgrm.PackageGet(prgrm.CXProgramNameIntern(pkgName))
	if err != nil {
		panic(err)
	}

	strct, err := pkg.StructGet(prgrm, prgrm.CXProgramNameIntern(strctName))
	if err != nil {
		panic(err)
	}

	// Natives are shared by every program, so this argument belongs to none.
	// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
	// Now a constructor call, site S0073.
	arg := nativeArgument(prgrm.CXProgramNameIntern(argName)).TypeSet(valuetypes.STRUCT)
	// arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, constants.DeclarationStruct)
	arg.Size = strct.StructSizeGet(prgrm)
	arg.StructType = strct

	return arg
}

// Slice Helper function for creating parameters for standard library operators.
// The current standard library only uses basic types and slices. If more options are needed, modify this function
// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
// Now a constructor, site S0074.
func Slice(typeCode valuetypes.Code) *CXArgument {
	// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
	// Now a constructor call, site S0075.
	arg := Param(typeCode)
	arg.DeclarationSpecifiers = append(arg.DeclarationSpecifiers, opcodetable.DeclarationSlice)
	// A typed []T is a slice as a declared one is. A []und takes its
	// dimensions from its operand.
	if typeCode != valuetypes.UNDEFINED {
		arg.Lengths = []valuetypes.Pointer{0}
	}
	return arg
}

// Func Helper function for creating function parameters for standard library operators.
// The current standard library only uses basic types and slices. If more options are needed, modify this function
// func Func(pkg *CXPackage, inputs []*CXArgument, outputs []*CXArgument) *CXArgument {
// 	arg := Param(types.FUNC)
// 	arg.Package = CXPackageIndex(pkg.Index)
// 	arg.Inputs = inputs
// 	arg.Outputs = outputs
// 	return arg
// }

// Param ...
// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
// Now a constructor, site S0076.
func Param(typeCode valuetypes.Code) *CXArgument {
	// Natives are shared by every program, so this argument belongs to none.
	// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
	// Now a constructor call, site S0077.
	arg := nativeArgument(stringpool.StringGidNil).TypeSet(typeCode)
	// arg.IsLocalDeclaration = true
	return arg
}

// nativeArgument makes an argument for a native. The natives are shared by
// every program, so it lives on the heap, belongs to none and has no debug
// record.
// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
// Now a constructor, site S0078.
func nativeArgument(name stringpool.StringGid) *CXArgument {
	// OBJMODEL: state.CXArgumentMng.CXArgumentAlloc(), phase 4 #326.
	// Now a heap literal, site S0079.
	return &CXArgument{
		Name:    name,
		Package: -1,
	}
}

// ArgumentMake allocates an argument named name in the CXArgumentMng of
// prgrm, with a debug record for fileName and fileLine, and returns its Gid
// and the pointer to fill it through.
func ArgumentMake(prgrm *CXProgram, name stringpool.StringGid, fileName SourceFileGid, fileLine int) (CXArgumentGid, *CXArgument) {
	argDetailsGid := argumentDebugAlloc(prgrm, fileName, fileLine)
	gid, arg := prgrm.CXArgumentAlloc()
	arg.Name = name
	arg.Package = -1
	arg.ArgDetails = argDetailsGid
	return gid, arg
}

// FieldMake allocates an argument of type typeCode like ArgumentMake.
func FieldMake(prgrm *CXProgram, name stringpool.StringGid, typeCode valuetypes.Code, fileName SourceFileGid, fileLine int) (CXArgumentGid, *CXArgument) {
	gid, arg := ArgumentMake(prgrm, name, fileName, fileLine)
	arg.Type = typeCode
	return gid, arg
}

// GlobalMake allocates a global of type typeCode like ArgumentMake and gives
// it the next offset in the heap segment.
func GlobalMake(prgrm *CXProgram, name stringpool.StringGid, typeCode valuetypes.Code, fileName SourceFileGid, fileLine int) (CXArgumentGid, *CXArgument) {
	gid, global := FieldMake(prgrm, name, typeCode, fileName, fileLine)
	global.Size = typeCode.Size()
	global.Offset = environment.HeapOffset
	environment.HeapOffset += global.Size
	return gid, global
}

// ------------------------------------------------------------------------------------------
//           Special functions to determine its type (atomic, array etomic, etc)

func CXArgumentTypeAtomic(arg *CXArgument) bool {
	// TODO: implement including types.IDENTIFIER
	// return (arg.Type.Primitive() || arg.Type == types.IDENTIFIER) && !arg.SliceTyped() && len(arg.Lengths) == 0 && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && (len(arg.DeclarationSpecifiers) == 0 || (len(arg.DeclarationSpecifiers) == 1 && arg.DeclarationSpecifiers[0] == constants.DeclarationBasic))

	return arg.Type.Primitive() && !arg.SliceTyped() && len(arg.Lengths) == 0 && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && (len(arg.DeclarationSpecifiers) == 0 || (len(arg.DeclarationSpecifiers) == 1 && arg.DeclarationSpecifiers[0] == opcodetable.DeclarationBasic))
}

// declarationSpecifiersForm reports whether specs are base followed by one or
// more dims, or by none when dims is 0, and then by a pointer when pointer
// is set.
func declarationSpecifiersForm(specs []int, base int, dims int, pointer bool) bool {
	if pointer {
		if len(specs) == 0 || specs[len(specs)-1] != opcodetable.DeclarationPointer {
			return false
		}
		specs = specs[:len(specs)-1]
	}
	if len(specs) == 0 || specs[0] != base || (dims != 0) != (len(specs) > 1) {
		return false
	}
	for _, spec := range specs[1:] {
		if spec != dims {
			return false
		}
	}
	return true
}

func CXArgumentTypePointerAtomic(arg *CXArgument) bool {
	return arg.Type.Primitive() && arg.PointerTargetType == 0 && arg.StructType == nil && !arg.SliceTyped() && len(arg.Lengths) == 0 && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && (len(arg.DeclarationSpecifiers) == 2 && arg.DeclarationSpecifiers[0] == opcodetable.DeclarationBasic && arg.DeclarationSpecifiers[1] == opcodetable.DeclarationPointer)
}

func CXArgumentTypeArrayAtomic(arg *CXArgument) bool {
	declPointerFound := false
	for _, decl := range arg.DeclarationSpecifiers {
		if decl == opcodetable.DeclarationPointer {
			declPointerFound = true
		}
	}

	return !declPointerFound && arg.Type.Primitive() && !arg.SliceTyped() && len(arg.Lengths) > 0 && !slices.Contains(arg.Lengths, 0) && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && declarationSpecifiersForm(arg.DeclarationSpecifiers, opcodetable.DeclarationBasic, opcodetable.DeclarationArray, false)
}

func CXArgumentTypePointerArrayAtomic(arg *CXArgument) bool {
	declPointerFound := false
	for _, decl := range arg.DeclarationSpecifiers {
		if decl == opcodetable.DeclarationPointer {
			declPointerFound = true
		}
	}

	return declPointerFound && arg.Type.Primitive() && !arg.SliceTyped() && len(arg.Lengths) > 0 && !slices.Contains(arg.Lengths, 0) && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && declarationSpecifiersForm(arg.DeclarationSpecifiers, opcodetable.DeclarationBasic, opcodetable.DeclarationArray, true)
}

func CXArgumentTypeSliceAtomic(arg *CXArgument) bool {
	declPointerFound := false
	for _, decl := range arg.DeclarationSpecifiers {
		if decl == opcodetable.DeclarationPointer {
			declPointerFound = true
		}
	}

	return !declPointerFound && arg.Type.Primitive() && arg.SliceTyped() && len(arg.Lengths) > 0 && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && declarationSpecifiersForm(arg.DeclarationSpecifiers, opcodetable.DeclarationBasic, opcodetable.DeclarationSlice, false)
}

func CXArgumentTypePointerSliceAtomic(arg *CXArgument) bool {
	declPointerFound := false
	for _, decl := range arg.DeclarationSpecifiers {
		if decl == opcodetable.DeclarationPointer {
			declPointerFound = true
		}
	}

	return declPointerFound && arg.Type.Primitive() && arg.SliceTyped() && len(arg.Lengths) > 0 && len(arg.Fields) == 0 && len(arg.DereferenceOperations) == 0 && declarationSpecifiersForm(arg.DeclarationSpecifiers, opcodetable.DeclarationBasic, opcodetable.DeclarationSlice, true)
}

func CXArgumentTypeStruct(arg *CXArgument) bool {
	declPointerFound := false
	declSliceFound := false
	for _, decl := range arg.DeclarationSpecifiers {
		if decl == opcodetable.DeclarationPointer {
			declPointerFound = true
		}

		if decl == opcodetable.DeclarationSlice {
			declSliceFound = true
		}
	}

	return !declPointerFound && !declSliceFound && !arg.SliceTyped() && len(arg.Lengths) == 0 && arg.StructType != nil && len(arg.DeclarationSpecifiers) >= 1 && arg.DeclarationSpecifiers[0] == opcodetable.DeclarationStruct
}

func CXArgumentTypePointerStruct(arg *CXArgument) bool {
	declPointerFound := false
	declSliceFound := false
	for _, decl := range arg.DeclarationSpecifiers {
		if decl == opcodetable.DeclarationPointer {
			declPointerFound = true
		}

		if decl == opcodetable.DeclarationSlice {
			declSliceFound = true
		}
	}

	return declPointerFound && !declSliceFound && !arg.SliceTyped() && len(arg.Lengths) == 0 && arg.StructType != nil && declarationSpecifiersForm(arg.DeclarationSpecifiers, opcodetable.DeclarationStruct, 0, true)
}

// argumentDebugAlloc allocates the debug record of an argument in prgrm's
// CXArgumentDebugMng. A nil prgrm, for a native, gets CXArgumentDebugGidNil.
func argumentDebugAlloc(prgrm *CXProgram, fileName SourceFileGid, fileLine int) CXArgumentDebugGid {
	if prgrm == nil {
		return CXArgumentDebugGidNil
	}
	argumentDebugMng := &prgrm.CxIrState.CXArgumentDebugMng
	gid := argumentDebugMng.CXArgumentDebugAlloc()
	details := argumentDebugMng.CXArgumentDebugGet(gid)
	details.FileName = fileName
	details.FileLine = fileLine
	details.FileColumn = prgrm.CXProgramCompileErrorColumn(fileName, fileLine)
	return gid
}

// CXArgumentDebugFileLine returns the file name and line that gid names in
// prgrm, the name viewed from the string pool. A native parameter has none
// and gets an empty name and -1, as Param gave it.
func CXArgumentDebugFileLine(prgrm *CXProgram, gid CXArgumentDebugGid) ([]byte, int) {
	if gid == CXArgumentDebugGidNil {
		return nil, -1
	}
	details := prgrm.CxIrState.CXArgumentDebugMng.CXArgumentDebugGet(gid)
	return prgrm.CXProgramSourceFileBytes(details.FileName), details.FileLine
}

// cxArgumentPointerField reports whether the field access arg is a pointer,
// also a *i32, which keeps the type i32 and ends its specifiers with it.
func cxArgumentPointerField(arg *CXArgument) bool {
	if arg.PointerTyped() {
		return true
	}
	specs := arg.DeclarationSpecifiers
	return arg.StructType == nil && len(arg.DereferenceOperations) == 0 && len(arg.Indexes) == 0 && len(specs) > 0 && specs[len(specs)-1] == opcodetable.DeclarationPointer
}

// ArraySliceDimensions returns the number of array dimensions of an array
// of slices, 1 for [N][]T, and 0 for any other type.
func (arg CXArgument) ArraySliceDimensions() int {
	for i, length := range arg.Lengths {
		if length == 0 {
			return i
		}
	}
	return 0
}

// ArraySliceMixed reports whether arg has both array and slice dimensions,
// as [2][]i32 and [][2][]i32 have.
func (arg CXArgument) ArraySliceMixed() bool {
	return arg.ArraySliceDimensions() > 0 || arg.SliceArrayDimensions() > 0
}

// SliceArrayDimensions returns the number of slice dimensions of a slice of
// arrays, 1 for [][N]T, and 0 for any other type.
func (arg CXArgument) SliceArrayDimensions() int {
	for i, length := range arg.Lengths {
		if length != 0 {
			return i
		}
	}
	return 0
}

// ValueSizeOfLengths returns the size of a value of the array and slice
// lengths, outermost first, whose innermost elements take elem bytes.
func ValueSizeOfLengths(lengths []valuetypes.Pointer, elem valuetypes.Pointer) valuetypes.Pointer {
	size := elem
	for i := len(lengths) - 1; i >= 0; i-- {
		if lengths[i] == 0 {
			size = valuetypes.PointerSize
		} else {
			size *= lengths[i]
		}
	}
	return size
}

// ArraySize returns the size in bytes of the array arg, whose elements are
// slice headers when it is an array of slices.
func (arg CXArgument) ArraySize() valuetypes.Pointer {
	if direct, behind := arg.DirectDimensions(); behind {
		return valuetypes.PointerSize * valuetypes.TotalLength(arg.Lengths[:direct])
	}
	if dims := arg.ArraySliceDimensions(); dims > 0 {
		return valuetypes.PointerSize * valuetypes.TotalLength(arg.Lengths[:dims])
	}
	return arg.Size * valuetypes.TotalLength(arg.Lengths)
}

// DirectDimensions returns how many dimensions of Lengths, outermost first,
// arg holds itself, and true when the rest are behind a pointer, as the
// second of [2]*[3]T is. An index strips its dimension from the specifiers.
func (arg CXArgument) DirectDimensions() (int, bool) {
	// (*a[i])[j] on a [N]*[M]T dereferences the pointer between its indexes.
	indexes := 0
	for i, op := range arg.DereferenceOperations {
		if op == opcodetable.DerefArray || op == opcodetable.DerefSlice {
			indexes++
			continue
		}
		if op == opcodetable.DerefPointer && indexes > 0 && indexes < len(arg.Lengths) && slices.ContainsFunc(arg.DereferenceOperations[i+1:], func(next int) bool { return next == opcodetable.DerefArray || next == opcodetable.DerefSlice }) {
			return indexes, true
		}
		break
	}
	specs := arg.DeclarationSpecifiers
	dims := 0
	for _, spec := range specs {
		if spec == opcodetable.DeclarationArray || spec == opcodetable.DeclarationSlice {
			dims++
		}
	}
	stripped := 0
	if len(arg.Indexes) > 0 && dims+len(arg.Indexes) == len(arg.Lengths) {
		stripped = len(arg.Indexes)
	}
	outer := 0
	i := len(specs) - 1
	for ; i >= 0 && (specs[i] == opcodetable.DeclarationArray || specs[i] == opcodetable.DeclarationSlice); i-- {
		outer++
	}
	if stripped == 0 && outer == 0 || i < 0 || specs[i] != opcodetable.DeclarationPointer || stripped+outer >= len(arg.Lengths) {
		return len(arg.Lengths), false
	}
	return stripped + outer, true
}
