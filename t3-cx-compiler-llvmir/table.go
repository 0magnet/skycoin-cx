package llvmir

import (
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// Type is a scalar type of the image, the operand type of a scalar opcode.
type Type uint8

const (
	TypeNone Type = iota
	TypeBool
	TypeI8
	TypeI16
	TypeI32
	TypeI64
	TypeUI8
	TypeUI16
	TypeUI32
	TypeUI64
	TypeF32
	TypeF64
	TypeCount
)

// TypeNames are the CX names, used in opcode and runtime function names.
var TypeNames = [TypeCount]string{"", "bool", "i8", "i16", "i32", "i64", "ui8", "ui16", "ui32", "ui64", "f32", "f64"}

var typeLlvm = [TypeCount]string{"", "i8", "i8", "i16", "i32", "i64", "i8", "i16", "i32", "i64", "float", "double"}
var typeSize = [TypeCount]uint32{0, 1, 1, 2, 4, 8, 1, 2, 4, 8, 4, 8}
var typeBits = [TypeCount]uint32{0, 8, 8, 16, 32, 64, 8, 16, 32, 64, 32, 64}

func typeSigned(t Type) bool   { return t >= TypeI8 && t <= TypeI64 }
func typeUnsigned(t Type) bool { return t >= TypeUI8 && t <= TypeUI64 }
func typeFloat(t Type) bool    { return t == TypeF32 || t == TypeF64 }

// Kind is the operation of a scalar opcode.
type Kind uint8

const (
	KindNone Kind = iota
	KindAdd
	KindSub
	KindMul
	KindDiv
	KindMod
	KindNeg
	KindAbs
	KindMax
	KindMin
	KindEq
	KindUneq
	KindLt
	KindLteq
	KindGt
	KindGteq
	KindBitAnd
	KindBitOr
	KindBitXor
	KindBitClear
	KindBitShl
	KindBitShr
	KindPrint
	KindNot
	KindAnd
	KindOr
	KindConv
	KindCount
)

// KindNames are the CX operator names.
var KindNames = [KindCount]string{"", "add", "sub", "mul", "div", "mod", "neg", "abs", "max", "min",
	"eq", "uneq", "lt", "lteq", "gt", "gteq", "bitand", "bitor", "bitxor", "bitclear", "shl", "shr",
	"print", "not", "and", "or", "to"}

func kindUnary(k Kind) bool {
	return k == KindNeg || k == KindAbs || k == KindNot || k == KindPrint || k == KindConv
}
func kindCompare(k Kind) bool {
	return k >= KindEq && k <= KindGteq
}

// OpInfo describes a scalar opcode: its operand type, its operation and, for
// a conversion, the result type.
type OpInfo struct {
	OpType Type
	OpKind Kind
	OpTo   Type
}

// OpInfos is filled from the family order of t2-cx-runtime-ir, checked by a
// test against OpcodeCount. Read only.
var OpInfos, opcodeTableEnd = opInfosBuild()

// OpName names a scalar opcode as the language does, i32.add or i32.f64.
func OpName(op ir.Opcode) string {
	info := OpInfos[op]
	if info.OpKind == KindNone {
		return opcodeControlNames[op]
	}
	if info.OpKind == KindConv {
		return TypeNames[info.OpType] + "." + TypeNames[info.OpTo]
	}
	return TypeNames[info.OpType] + "." + KindNames[info.OpKind]
}

var opcodeControlNames = map[ir.Opcode]string{
	ir.OpcodeReturn: "return", ir.OpcodeJump: "jump", ir.OpcodeJumpIf: "jumpif", ir.OpcodeZero: "zero",
	ir.OpcodeMove: "move", ir.OpcodeCallFrame: "callframe", ir.OpcodeCallArg: "callarg", ir.OpcodeCall: "call",
	ir.OpcodeCallResult: "callresult", ir.OpcodeTestEqual: "testequal",
	ir.OpcodeAddressIndexStatic: "addressindexstatic", ir.OpcodeAddressIndex: "addressindex",
	ir.OpcodeAddressDerefStatic: "addressderefstatic", ir.OpcodeAddressDeref: "addressderef",
	ir.OpcodeAddressLoad: "addressload", ir.OpcodeAddressStore: "addressstore", ir.OpcodeAddressOf: "addressof",
	ir.OpcodeAddressPointer: "addresspointer", ir.OpcodePrintText: "printtext", ir.OpcodePrintValue: "printvalue",
	ir.OpcodePrintStr: "printstr",
}

// opInfosBuild fills the table in family order. It also returns the opcode
// after the last one it filled, OpcodeCount when the families match.
func opInfosBuild() (infos [ir.OpcodeCount]OpInfo, op ir.Opcode) {
	op = ir.OpcodeBoolNot
	for _, k := range []Kind{KindNot, KindAnd, KindOr, KindEq, KindUneq, KindPrint} {
		infos[op] = OpInfo{OpType: TypeBool, OpKind: k}
		op++
	}
	signed := []Kind{KindAdd, KindSub, KindMul, KindDiv, KindMod, KindNeg, KindAbs, KindMax, KindMin,
		KindEq, KindUneq, KindLt, KindLteq, KindGt, KindGteq,
		KindBitAnd, KindBitOr, KindBitXor, KindBitClear, KindBitShl, KindBitShr, KindPrint}
	unsigned := []Kind{KindAdd, KindSub, KindMul, KindDiv, KindMod, KindMax, KindMin,
		KindEq, KindUneq, KindLt, KindLteq, KindGt, KindGteq,
		KindBitAnd, KindBitOr, KindBitXor, KindBitClear, KindBitShl, KindBitShr, KindPrint}
	float := []Kind{KindAdd, KindSub, KindMul, KindDiv, KindMod, KindNeg, KindAbs, KindMax, KindMin,
		KindEq, KindUneq, KindLt, KindLteq, KindGt, KindGteq, KindPrint}
	families := []struct {
		types []Type
		kinds []Kind
	}{
		{[]Type{TypeI8, TypeI16, TypeI32, TypeI64}, signed},
		{[]Type{TypeUI8, TypeUI16, TypeUI32, TypeUI64}, unsigned},
		{[]Type{TypeF32, TypeF64}, float},
	}
	for _, family := range families {
		for _, t := range family.types {
			for _, k := range family.kinds {
				infos[op] = OpInfo{OpType: t, OpKind: k}
				op++
			}
		}
	}
	for from := TypeI8; from <= TypeF64; from++ {
		for to := TypeI8; to <= TypeF64; to++ {
			if from == to {
				continue
			}
			infos[op] = OpInfo{OpType: from, OpKind: KindConv, OpTo: to}
			op++
		}
	}
	return infos, op
}

// resultType is the type an opcode writes to C.
func resultType(info OpInfo) Type {
	switch {
	case info.OpKind == KindConv:
		return info.OpTo
	case kindCompare(info.OpKind):
		return TypeBool
	case info.OpKind == KindPrint:
		return TypeNone
	}
	return info.OpType
}
