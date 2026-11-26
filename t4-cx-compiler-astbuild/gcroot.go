package astbuild

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// gcRootsDeclare adds a new local that holds a pointer, slice or str, also
// an escaped one, to the GC roots of fn with its declared type.
func gcRootsDeclare(prgrm *ast.CXProgram, fn *ast.CXFunction, typeSig *ast.CXTypeSignature, sym *ast.CXArgument) {
	offset := typeSig.Offset
	specs, typ, strct, lengths := ast.SignatureGCDeclaration(prgrm, typeSig)
	if sym != nil {
		offset = sym.Offset
		specs, typ, strct, lengths = ast.CXArgumentGCDeclaration(sym)
	}
	if offset >= prgrm.Stack.Size && !ast.OffsetEscaped(offset) {
		return
	}
	if !gcValueReferences(prgrm, specs, typ, strct, 0) {
		return
	}
	gcRootAdd(fn, offset, specs, typ, strct, lengths)
}

// gcRootAdd adds the local at offset to fn.ListOfSlots.
// GCREACH: adds a GC root to ListOfSlots,
// see docs/GarbageCollector.md.
func gcRootAdd(fn *ast.CXFunction, offset valuetypes.Pointer, specs []int, typ valuetypes.Code, strct *ast.CXStruct, lengths []valuetypes.Pointer) {
	fn.ListOfSlots = append(fn.ListOfSlots, ast.CXFunctionSlot{
		CXFunctionSlotOffset:     offset,
		CXFunctionSlotLengths:    lengths,
		CXFunctionSlotType:       typ,
		CXFunctionSlotSpecifiers: specs,
		CXFunctionSlotStruct:     strct,
	})
}

// gcRootAddEscapeSlot adds the frame slot of an escaped local, which holds
// its heap object, to fn.ListOfSlots.
func gcRootAddEscapeSlot(fn *ast.CXFunction, offset valuetypes.Pointer) {
	gcRootAdd(fn, offset, []int{opcodetable.DeclarationBasic, opcodetable.DeclarationPointer}, valuetypes.UNDEFINED, nil, nil)
}

// gcPointerDepthLimit bounds the struct nesting gcValueReferences follows.
const gcPointerDepthLimit = 16

// gcValueReferences reports whether a value of declaration specifiers specs
// holds a pointer, slice or str.
func gcValueReferences(prgrm *ast.CXProgram, specs []int, typ valuetypes.Code, strct *ast.CXStruct, depth int) bool {
	if depth > gcPointerDepthLimit || len(specs) == 0 {
		return false
	}
	switch specs[len(specs)-1] {
	case opcodetable.DeclarationPointer, opcodetable.DeclarationSlice:
		return true
	case opcodetable.DeclarationArray:
		return gcValueReferences(prgrm, specs[:len(specs)-1], typ, strct, depth+1)
	}
	if strct == nil {
		return typ == valuetypes.STR
	}
	for _, fieldIdx := range strct.Fields {
		specs, typ, fieldStruct, _ := ast.SignatureGCDeclaration(prgrm, prgrm.CXTypeSignatureFromArrayGet(fieldIdx))
		if gcValueReferences(prgrm, specs, typ, fieldStruct, depth+1) {
			return true
		}
	}
	return false
}
