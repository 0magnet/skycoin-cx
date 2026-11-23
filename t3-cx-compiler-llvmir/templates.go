package llvmir

import (
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// A template is the direct LLVM IR of one scalar opcode. {a} and {b} are the
// loaded operands, {r} the result, {T} the operand type, {R} the result type,
// {W} and {W1} the bit width and one less, {t1} to {t9} fresh temporaries.
// A line "check {tN} Stop" branches to a run time stop when {tN} is true.
// An opcode without a template calls the runtime function cx_rt_<type>_<op>
// with the operands as values, see FallbackName.

// Template returns the direct IR of op, false when op has none.
func Template(op ir.Opcode) (string, bool) {
	info := OpInfos[op]
	t := info.OpType
	switch info.OpKind {
	case KindNone, KindPrint:
		return "", false
	case KindConv:
		return convTemplate(info.OpType, info.OpTo)
	}
	if t == TypeBool {
		switch info.OpKind {
		case KindNot:
			return "{t1} = icmp eq i8 {a}, 0\n{r} = zext i1 {t1} to i8", true
		case KindAnd, KindOr, KindEq, KindUneq:
			ops := map[Kind]string{KindAnd: "and i1", KindOr: "or i1", KindEq: "icmp eq i1", KindUneq: "icmp ne i1"}
			return "{t1} = icmp ne i8 {a}, 0\n{t2} = icmp ne i8 {b}, 0\n{t3} = " + ops[info.OpKind] + " {t1}, {t2}\n{r} = zext i1 {t3} to i8", true
		}
		return "", false
	}
	if typeFloat(t) {
		switch info.OpKind {
		case KindAdd:
			return "{r} = fadd {T} {a}, {b}", true
		case KindSub:
			return "{r} = fsub {T} {a}, {b}", true
		case KindMul:
			return "{r} = fmul {T} {a}, {b}", true
		case KindDiv:
			return "{r} = fdiv {T} {a}, {b}", true
		case KindNeg:
			return "{r} = fneg {T} {a}", true
		case KindAbs:
			return "{r} = call {T} @llvm.fabs.{F}({T} {a})", true
		case KindEq, KindUneq, KindLt, KindLteq, KindGt, KindGteq:
			preds := map[Kind]string{KindEq: "oeq", KindUneq: "une", KindLt: "olt", KindLteq: "ole", KindGt: "ogt", KindGteq: "oge"}
			return "{t1} = fcmp " + preds[info.OpKind] + " {T} {a}, {b}\n{r} = zext i1 {t1} to i8", true
		}
		// Mod, Max and Min follow math.Mod, math.Max and math.Min, whose NaN,
		// infinity and signed zero cases differ from frem and llvm.maximum.
		return "", false
	}
	signed := typeSigned(t)
	switch info.OpKind {
	case KindAdd:
		return "{r} = add {T} {a}, {b}", true
	case KindSub:
		return "{r} = sub {T} {a}, {b}", true
	case KindMul:
		return "{r} = mul {T} {a}, {b}", true
	case KindDiv:
		if signed {
			// Go defines MinInt / -1 as MinInt, LLVM sdiv makes it undefined.
			return "{t1} = icmp eq {T} {b}, 0\ncheck {t1} DivideByZero\n{t2} = icmp eq {T} {b}, -1\n{t3} = select i1 {t2}, {T} 1, {T} {b}\n{t4} = sdiv {T} {a}, {t3}\n{t5} = sub {T} 0, {a}\n{r} = select i1 {t2}, {T} {t5}, {T} {t4}", true
		}
		return "{t1} = icmp eq {T} {b}, 0\ncheck {t1} DivideByZero\n{r} = udiv {T} {a}, {b}", true
	case KindMod:
		if signed {
			return "{t1} = icmp eq {T} {b}, 0\ncheck {t1} DivideByZero\n{t2} = icmp eq {T} {b}, -1\n{t3} = select i1 {t2}, {T} 1, {T} {b}\n{r} = srem {T} {a}, {t3}", true
		}
		return "{t1} = icmp eq {T} {b}, 0\ncheck {t1} DivideByZero\n{r} = urem {T} {a}, {b}", true
	case KindNeg:
		return "{r} = sub {T} 0, {a}", true
	case KindAbs:
		return "{r} = call {T} @llvm.abs.{T}({T} {a}, i1 false)", true
	case KindMax:
		if signed {
			return "{r} = call {T} @llvm.smax.{T}({T} {a}, {T} {b})", true
		}
		return "{r} = call {T} @llvm.umax.{T}({T} {a}, {T} {b})", true
	case KindMin:
		if signed {
			return "{r} = call {T} @llvm.smin.{T}({T} {a}, {T} {b})", true
		}
		return "{r} = call {T} @llvm.umin.{T}({T} {a}, {T} {b})", true
	case KindEq, KindUneq, KindLt, KindLteq, KindGt, KindGteq:
		preds := map[Kind]string{KindEq: "eq", KindUneq: "ne", KindLt: "ult", KindLteq: "ule", KindGt: "ugt", KindGteq: "uge"}
		if signed {
			preds = map[Kind]string{KindEq: "eq", KindUneq: "ne", KindLt: "slt", KindLteq: "sle", KindGt: "sgt", KindGteq: "sge"}
		}
		return "{t1} = icmp " + preds[info.OpKind] + " {T} {a}, {b}\n{r} = zext i1 {t1} to i8", true
	case KindBitAnd:
		return "{r} = and {T} {a}, {b}", true
	case KindBitOr:
		return "{r} = or {T} {a}, {b}", true
	case KindBitXor:
		return "{r} = xor {T} {a}, {b}", true
	case KindBitClear:
		return "{t1} = xor {T} {b}, -1\n{r} = and {T} {a}, {t1}", true
	case KindBitShl:
		// Go shifts by the count as unsigned, a count of the width or more
		// gives 0. LLVM makes such a shift poison, so the select picks 0.
		return "{t1} = icmp uge {T} {b}, {W}\n{t2} = shl {T} {a}, {b}\n{r} = select i1 {t1}, {T} 0, {T} {t2}", true
	case KindBitShr:
		if signed {
			return "{t1} = icmp uge {T} {b}, {W}\n{t2} = select i1 {t1}, {T} {W1}, {T} {b}\n{r} = ashr {T} {a}, {t2}", true
		}
		return "{t1} = icmp uge {T} {b}, {W}\n{t2} = lshr {T} {a}, {b}\n{r} = select i1 {t1}, {T} 0, {T} {t2}", true
	}
	return "", false
}

func convTemplate(from, to Type) (string, bool) {
	switch {
	case typeFloat(from) && typeFloat(to):
		if from == TypeF32 {
			return "{r} = fpext float {a} to double", true
		}
		return "{r} = fptrunc double {a} to float", true
	case typeFloat(from):
		// Go leaves an out of range conversion implementation defined. freeze
		// turns LLVM poison into some value, which is no stronger.
		if typeSigned(to) {
			return "{t1} = fptosi {T} {a} to {R}\n{r} = freeze {R} {t1}", true
		}
		return "{t1} = fptoui {T} {a} to {R}\n{r} = freeze {R} {t1}", true
	case typeFloat(to):
		if typeSigned(from) {
			return "{r} = sitofp {T} {a} to {R}", true
		}
		return "{r} = uitofp {T} {a} to {R}", true
	}
	switch fb, tb := typeBits[from], typeBits[to]; {
	case fb == tb:
		return "{r} = add {T} {a}, 0", true
	case fb > tb:
		return "{r} = trunc {T} {a} to {R}", true
	case typeSigned(from):
		return "{r} = sext {T} {a} to {R}", true
	default:
		return "{r} = zext {T} {a} to {R}", true
	}
}

// FallbackName is the runtime function an opcode without a template calls.
func FallbackName(op ir.Opcode) string {
	info := OpInfos[op]
	if info.OpKind == KindConv {
		return "cx_rt_" + TypeNames[info.OpType] + "_to_" + TypeNames[info.OpTo]
	}
	return "cx_rt_" + TypeNames[info.OpType] + "_" + KindNames[info.OpKind]
}
