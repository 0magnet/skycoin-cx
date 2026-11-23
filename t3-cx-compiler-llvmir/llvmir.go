// Package llvmir is a prototype that turns an ir.Image into textual LLVM IR,
// the llvmir target of skycoin/skycoin-cx. It is pure Go and links no LLVM.
// Scalar opcodes with a template become direct IR, the rest call the C
// runtime in runtime.c. Frame variables whose address is never taken become
// allocas that LLVM keeps in registers. Variables whose address is taken stay
// in a shadow stack inside cx_mem at the same offsets as on the vm, so every
// pointer keeps its vm value.
package llvmir

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// RuntimeC is the C runtime every emitted module links with.
//
//go:embed runtime.c.txt
var RuntimeC []byte

const pointerSize = uint32(types.POINTER_SIZE)

// Stats counts the static instructions of a module by how they were emitted.
type Stats struct {
	StatsDirect   map[ir.Opcode]int
	StatsFallback map[ir.Opcode]int
	StatsShadow   int
	StatsAlloca   int
}

// region is a byte range of a frame or of the scratch that becomes one
// alloca, or stays in cx_mem when shadow is set.
type region struct {
	start, end uint32
	declared   bool
	shadow     bool
	name       string
}

type access struct {
	which    int
	size     uint32
	address  bool
	declared bool
}

type pendingCall struct {
	active bool
	frame  uint32
	args   map[uint32]string
	sizes  map[uint32]uint32
}

type emitter struct {
	image    *ir.Image
	out      bytes.Buffer
	body     bytes.Buffer
	cold     bytes.Buffer
	decls    map[string]string
	strs     map[string]string
	stats    Stats
	dataEnd  uint32
	fid      uint32
	fn       *ir.Func
	tmp      int
	frame    []region
	scratch  []region
	call     pendingCall
	result   string
	resultOf uint32
	readOnly []bool
}

// Emit returns the LLVM IR module of image, or an error for an instruction
// the prototype does not translate.
func Emit(image *ir.Image, options Options) ([]byte, Stats, error) {
	e := &emitter{
		image:   image,
		decls:   map[string]string{},
		strs:    map[string]string{},
		dataEnd: ir.ImageDataEnd(image),
		stats:   Stats{StatsDirect: map[ir.Opcode]int{}, StatsFallback: map[ir.Opcode]int{}},
	}
	if opcodeTableEnd != ir.OpcodeCount {
		return nil, e.stats, fmt.Errorf("opcode table ends at %d, image has %d opcodes", opcodeTableEnd, ir.OpcodeCount)
	}
	e.markReadOnly(options)
	for id := range image.ImageFuncs {
		fn := &image.ImageFuncs[id]
		if len(fn.FuncCode) == 0 {
			continue
		}
		if err := e.function(uint32(id), fn); err != nil {
			return nil, e.stats, fmt.Errorf("%s: %w", image.ImageFuncNames[id], err)
		}
	}
	e.header()
	e.out.Write(e.body.Bytes())
	e.entry()
	names := make([]string, 0, len(e.decls))
	for name := range e.decls {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		e.out.WriteString(e.decls[name] + "\n")
	}
	e.out.WriteString("\n!0 = !{!\"branch_weights\", i32 1, i32 1048576}\n")
	return e.out.Bytes(), e.stats, nil
}

func (e *emitter) header() {
	img := e.image
	fmt.Fprintf(&e.out, "; CX image as LLVM IR, prototype llvmir target\n")
	size := e.dataEnd + img.ImageScratchSize
	fmt.Fprintf(&e.out, "@cx_mem = internal global [%d x i8] zeroinitializer, align 16\n", size)
	for i, run := range e.dataRuns() {
		b := img.ImageData[run[0]:run[1]]
		fmt.Fprintf(&e.out, "@cx_init.%d = private unnamed_addr constant [%d x i8] %s\n", i, len(b), llvmBytes(b))
	}
	fmt.Fprintf(&e.out, "@cx_text = internal constant [%d x i8] %s\n", len(img.ImageText), llvmBytes(img.ImageText))
	keys := make([]string, 0, len(e.strs))
	for text := range e.strs {
		keys = append(keys, text)
	}
	sort.Strings(keys)
	for _, text := range keys {
		b := append([]byte(text), 0)
		fmt.Fprintf(&e.out, "%s = private unnamed_addr constant [%d x i8] %s\n", e.strs[text], len(b), llvmBytes(b))
	}
	e.out.WriteString("\n")
}

func (e *emitter) entry() {
	img := e.image
	e.out.WriteString("define i32 @main() {\nentry:\n")
	for i, run := range e.dataRuns() {
		fmt.Fprintf(&e.out, "  call void @llvm.memcpy.p0.p0.i64(ptr getelementptr (i8, ptr @cx_mem, i64 %d), ptr @cx_init.%d, i64 %d, i1 false)\n", img.ImageDataStart+uint32(run[0]), i, run[1]-run[0])
	}
	fp := uint32(0)
	if img.ImageEntryInit != ir.ImageFuncNone {
		fmt.Fprintf(&e.out, "  call %s %s(i32 0, i32 0)\n", e.retType(&img.ImageFuncs[img.ImageEntryInit]), e.funcName(img.ImageEntryInit))
		fp = img.ImageFuncs[img.ImageEntryInit].FuncFrameSize
	}
	if img.ImageEntryMain != ir.ImageFuncNone {
		fmt.Fprintf(&e.out, "  call %s %s(i32 %d, i32 0)\n", e.retType(&img.ImageFuncs[img.ImageEntryMain]), e.funcName(img.ImageEntryMain), fp)
	}
	e.declare("cx_rt_finish", "declare i32 @cx_rt_finish()")
	e.out.WriteString("  %code = call i32 @cx_rt_finish()\n  ret i32 %code\n}\n\n")
}

func llvmBytes(b []byte) string {
	var sb strings.Builder
	sb.WriteString("c\"")
	for _, c := range b {
		if c >= 0x20 && c < 0x7f && c != '"' && c != '\\' {
			sb.WriteByte(c)
		} else {
			fmt.Fprintf(&sb, "\\%02X", c)
		}
	}
	sb.WriteString("\"")
	return sb.String()
}

func (e *emitter) declare(name, text string) { e.decls[name] = text }

func (e *emitter) str(text string) string {
	if name, found := e.strs[text]; found {
		return name
	}
	name := "@cx_str." + strconv.Itoa(len(e.strs))
	e.strs[text] = name
	return name
}

func (e *emitter) funcName(id uint32) string {
	return fmt.Sprintf("@\"cx.%d.%s\"", id, e.image.ImageFuncNames[id])
}

// valueType is the LLVM type that carries a slot of size bytes.
func valueType(size uint32) string {
	switch size {
	case 1, 2, 4, 8, 16:
		return "i" + strconv.Itoa(int(size*8))
	}
	return fmt.Sprintf("[%d x i8]", size)
}

func zeroValue(size uint32) string {
	switch size {
	case 1, 2, 4, 8, 16:
		return "0"
	}
	return "zeroinitializer"
}

func (e *emitter) retType(fn *ir.Func) string {
	if len(fn.FuncOutputs) == 0 {
		return "void"
	}
	parts := make([]string, len(fn.FuncOutputs))
	for i, slot := range fn.FuncOutputs {
		parts[i] = valueType(slot.SlotSize)
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func (e *emitter) newTmp() string {
	e.tmp++
	return "%t" + strconv.Itoa(e.tmp)
}

func (e *emitter) newLabel(prefix string) string {
	e.tmp++
	return prefix + strconv.Itoa(e.tmp)
}

func (e *emitter) line(format string, args ...any) {
	e.body.WriteString("  ")
	fmt.Fprintf(&e.body, format, args...)
	e.body.WriteByte('\n')
}

// accesses lists the operands of in with their sizes. An address access is
// an operand whose address is taken, the start of a variable that must stay
// in memory.
func accesses(in *ir.Instr) []access {
	n, m := in.InstrN, in.InstrM
	switch in.InstrOpcode {
	case ir.OpcodeJumpIf:
		return []access{{which: 0, size: 1}}
	case ir.OpcodeZero:
		return []access{{which: 0, size: n, declared: true}}
	case ir.OpcodeMove:
		return []access{{which: 0, size: n}, {which: 2, size: n}}
	case ir.OpcodeCallArg:
		return []access{{which: 0, size: n}}
	case ir.OpcodeCallResult:
		return []access{{which: 2, size: n}}
	case ir.OpcodeTestEqual:
		return []access{{which: 0, size: n}, {which: 1, size: m}, {which: 2, size: pointerSize}}
	case ir.OpcodeAddressIndexStatic:
		return []access{{which: 0, address: true}, {which: 1, size: 4}, {which: 2, size: 4}}
	case ir.OpcodeAddressIndex:
		return []access{{which: 0, size: 4}, {which: 1, size: 4}, {which: 2, size: 4}}
	case ir.OpcodeAddressDerefStatic:
		return []access{{which: 0, size: pointerSize}, {which: 2, size: 4}}
	case ir.OpcodeAddressDeref:
		return []access{{which: 0, size: 4}, {which: 2, size: 4}}
	case ir.OpcodeAddressLoad:
		return []access{{which: 0, size: 4}, {which: 2, size: n}}
	case ir.OpcodeAddressStore:
		return []access{{which: 0, size: n}, {which: 2, size: 4}}
	case ir.OpcodeAddressOf:
		return []access{{which: 0, address: true}, {which: 2, size: pointerSize}}
	case ir.OpcodeAddressPointer:
		return []access{{which: 0, size: 4}, {which: 2, size: pointerSize}}
	case ir.OpcodePrintValue:
		return []access{{which: 0, size: printValueSize(types.Code(m))}}
	case ir.OpcodePrintStr:
		return []access{{which: 0, size: pointerSize}}
	}
	info := OpInfos[in.InstrOpcode]
	if info.OpKind == KindNone {
		return nil
	}
	list := []access{{which: 0, size: typeSize[info.OpType]}}
	if !kindUnary(info.OpKind) {
		list = append(list, access{which: 1, size: typeSize[info.OpType]})
	}
	if r := resultType(info); r != TypeNone {
		list = append(list, access{which: 2, size: typeSize[r]})
	}
	return list
}

func printValueSize(code types.Code) uint32 {
	switch code {
	case types.BOOL, types.I8, types.UI8:
		return 1
	case types.I16, types.UI16:
		return 2
	case types.I32, types.UI32, types.F32:
		return 4
	case types.I64, types.UI64, types.F64:
		return 8
	}
	return 0
}

func operand(in *ir.Instr, which int) (uint32, bool) {
	switch which {
	case 0:
		return in.InstrA, in.InstrMode&ir.InstrModeFrameA != 0
	case 1:
		return in.InstrB, in.InstrMode&ir.InstrModeFrameB != 0
	case 2:
		return in.InstrC, in.InstrMode&ir.InstrModeFrameC != 0
	}
	return in.InstrD, in.InstrMode&ir.InstrModeFrameD != 0
}

type interval struct {
	start, end uint32
	declared   bool
}

// partition merges overlapping intervals into regions.
func partition(list []interval) []region {
	sort.Slice(list, func(i, j int) bool { return list[i].start < list[j].start })
	var out []region
	for _, iv := range list {
		if iv.end <= iv.start {
			continue
		}
		if n := len(out); n > 0 && iv.start < out[n-1].end {
			out[n-1].end = max(out[n-1].end, iv.end)
			out[n-1].declared = out[n-1].declared || iv.declared
			continue
		}
		out = append(out, region{start: iv.start, end: iv.end, declared: iv.declared})
	}
	return out
}

func find(regions []region, offset uint32) int {
	i := sort.Search(len(regions), func(i int) bool { return regions[i].end > offset })
	if i < len(regions) && regions[i].start <= offset {
		return i
	}
	return -1
}

// analyze splits the frame and the scratch of fn into regions.
func (e *emitter) analyze(fn *ir.Func) error {
	var frame, scratch []interval
	var addressed []uint32
	for _, slot := range fn.FuncInputs {
		frame = append(frame, interval{slot.SlotOffset, slot.SlotOffset + slot.SlotSize, true})
	}
	for _, slot := range fn.FuncOutputs {
		frame = append(frame, interval{slot.SlotOffset, slot.SlotOffset + slot.SlotSize, true})
	}
	for pc := range fn.FuncCode {
		in := &fn.FuncCode[pc]
		for _, ac := range accesses(in) {
			offset, inFrame := operand(in, ac.which)
			switch {
			case inFrame && ac.address:
				addressed = append(addressed, offset)
			case inFrame:
				frame = append(frame, interval{offset, offset + ac.size, ac.declared})
			case ac.address:
				if offset >= e.dataEnd {
					return fmt.Errorf("pc %d takes the address of scratch", pc)
				}
			case offset >= e.dataEnd:
				scratch = append(scratch, interval{offset, offset + ac.size, false})
			}
		}
	}
	e.frame = partition(frame)
	e.scratch = partition(scratch)
	all := false
	for _, offset := range addressed {
		i := find(e.frame, offset)
		if i < 0 || !e.frame[i].declared || e.frame[i].start != offset {
			all = true
			break
		}
		e.frame[i].shadow = true
	}
	if all {
		e.frame = []region{{start: 0, end: fn.FuncFrameSize, shadow: true}}
	}
	for i := range e.frame {
		if e.frame[i].shadow {
			e.stats.StatsShadow++
		} else {
			e.stats.StatsAlloca++
			e.frame[i].name = fmt.Sprintf("%%f%d", e.frame[i].start)
		}
	}
	for i := range e.scratch {
		e.scratch[i].name = fmt.Sprintf("%%s%d", e.scratch[i].start-e.dataEnd)
	}
	return nil
}

// ptr returns a pointer to operand which of in, size bytes.
func (e *emitter) ptr(in *ir.Instr, which int, size uint32) (string, error) {
	offset, inFrame := operand(in, which)
	return e.ptrAt(offset, inFrame, size)
}

func (e *emitter) ptrAt(offset uint32, inFrame bool, size uint32) (string, error) {
	regions := e.scratch
	if inFrame {
		regions = e.frame
	} else if offset < e.dataEnd {
		t := e.newTmp()
		e.line("%s = getelementptr inbounds i8, ptr @cx_mem, i64 %d", t, offset)
		return t, nil
	}
	i := find(regions, offset)
	if i < 0 || offset+size > regions[i].end {
		return "", fmt.Errorf("operand %d size %d outside its region", offset, size)
	}
	r := &regions[i]
	if r.shadow {
		t := e.newTmp()
		e.line("%s = getelementptr inbounds i8, ptr %%fpp, i64 %d", t, offset)
		return t, nil
	}
	if offset == r.start {
		return r.name, nil
	}
	t := e.newTmp()
	e.line("%s = getelementptr inbounds i8, ptr %s, i64 %d", t, r.name, offset-r.start)
	return t, nil
}

func (e *emitter) load(in *ir.Instr, which int, typ string, size uint32) (string, error) {
	offset, inFrame := operand(in, which)
	if v, found := e.constant(offset, inFrame, typ, size); found {
		return v, nil
	}
	p, err := e.ptr(in, which, size)
	if err != nil {
		return "", err
	}
	t := e.newTmp()
	e.line("%s = load %s, ptr %s, align 1", t, typ, p)
	return t, nil
}

func (e *emitter) store(in *ir.Instr, which int, typ string, size uint32, value string) error {
	p, err := e.ptr(in, which, size)
	if err != nil {
		return err
	}
	e.line("store %s %s, ptr %s, align 1", typ, value, p)
	return nil
}

// vmAddress returns the vm address of operand which, the frame pointer added
// for a frame operand.
func (e *emitter) vmAddress(in *ir.Instr, which int) string {
	offset, inFrame := operand(in, which)
	if !inFrame {
		return strconv.FormatUint(uint64(offset), 10)
	}
	t := e.newTmp()
	e.line("%s = add i32 %%fp, %d", t, offset)
	return t
}

// stopText is what the vm prints for a stop at pc, the first line of
// PrinterVmErrorPrint.
func (e *emitter) stopText(pc uint32, code int, message string) string {
	file, number := "", uint32(0)
	if line, found := ir.LineAt(e.image, e.fid, pc); found {
		file, number = e.image.ImageFileNames[line.LineFile], line.LineNumber
	}
	return fmt.Sprintf("error: %s:%d, %s, %s", file, number, constants.ErrorStrings[code], message)
}

var stopCodes = map[string]struct {
	code    int
	message string
}{
	"DivideByZero":   {types.CX_RUNTIME_ERROR, "runtime error: integer divide by zero"},
	"StackOverflow":  {types.CX_RUNTIME_STACK_OVERFLOW_ERROR, constants.STACK_OVERFLOW_ERROR},
	"IndexNegative":  {types.CX_RUNTIME_INVALID_CAST, strconv.Itoa(types.CX_RUNTIME_INVALID_CAST)},
	"AddressInvalid": {types.CX_RUNTIME_ERROR, "address outside the stack and data segment"},
}

// check branches to a cold stop block when cond is true.
func (e *emitter) check(cond string, stop string, pc uint32) {
	sc := stopCodes[stop]
	bad, ok := e.newLabel("stop"), e.newLabel("ok")
	e.line("br i1 %s, label %%%s, label %%%s, !prof !0", cond, bad, ok)
	e.body.WriteString(ok + ":\n")
	text := e.str(e.stopText(pc, sc.code, sc.message))
	e.declare("cx_rt_stop", "declare void @cx_rt_stop(i32, ptr) noreturn cold nounwind")
	fmt.Fprintf(&e.cold, "%s:\n  call void @cx_rt_stop(i32 %d, ptr %s)\n  unreachable\n", bad, sc.code, text)
}

func (e *emitter) function(id uint32, fn *ir.Func) error {
	e.fid, e.fn, e.tmp = id, fn, 0
	e.cold.Reset()
	e.call = pendingCall{}
	if err := e.analyze(fn); err != nil {
		return err
	}
	params := []string{"i32 %fp", "i32 %depth"}
	for i, slot := range fn.FuncInputs {
		params = append(params, fmt.Sprintf("%s %%in%d", valueType(slot.SlotSize), i))
	}
	fmt.Fprintf(&e.body, "define internal %s %s(%s) nounwind {\nentry:\n", e.retType(fn), e.funcName(id), strings.Join(params, ", "))
	e.line("%%fp64 = zext i32 %%fp to i64")
	e.line("%%fpp = getelementptr inbounds i8, ptr @cx_mem, i64 %%fp64")
	e.line("%%sp = add i32 %%fp, %d", fn.FuncFrameSize)
	e.declare("llvm.memset", "declare void @llvm.memset.p0.i64(ptr, i8, i64, i1)")
	e.declare("llvm.memcpy", "declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)")
	for _, r := range e.frame {
		if r.shadow {
			t := e.newTmp()
			e.line("%s = getelementptr inbounds i8, ptr %%fpp, i64 %d", t, r.start)
			e.line("call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 %d, i1 false)", t, r.end-r.start)
			continue
		}
		e.line("%s = alloca [%d x i8], align 8", r.name, r.end-r.start)
		e.line("call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 %d, i1 false)", r.name, r.end-r.start)
	}
	for _, r := range e.scratch {
		e.line("%s = alloca [%d x i8], align 8", r.name, r.end-r.start)
		e.line("call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 %d, i1 false)", r.name, r.end-r.start)
	}
	for i, slot := range fn.FuncInputs {
		p, err := e.ptrAt(slot.SlotOffset, true, slot.SlotSize)
		if err != nil {
			return err
		}
		e.line("store %s %%in%d, ptr %s, align 1", valueType(slot.SlotSize), i, p)
	}
	e.line("br label %%pc0")

	leaders := map[uint32]bool{0: true}
	for pc := range fn.FuncCode {
		in := &fn.FuncCode[pc]
		switch in.InstrOpcode {
		case ir.OpcodeJump:
			leaders[in.InstrM] = true
			leaders[uint32(pc)+1] = true
		case ir.OpcodeJumpIf:
			leaders[in.InstrM] = true
			leaders[in.InstrN] = true
			leaders[uint32(pc)+1] = true
		case ir.OpcodeReturn:
			leaders[uint32(pc)+1] = true
		}
	}
	terminated := true
	for pc := range fn.FuncCode {
		upc := uint32(pc)
		if leaders[upc] {
			if !terminated {
				e.line("br label %%pc%d", pc)
			}
			fmt.Fprintf(&e.body, "pc%d:\n", pc)
		}
		in := &fn.FuncCode[pc]
		terminated = in.InstrOpcode == ir.OpcodeReturn || in.InstrOpcode == ir.OpcodeJump || in.InstrOpcode == ir.OpcodeJumpIf
		if err := e.instr(upc, in); err != nil {
			return fmt.Errorf("pc %d %s: %w", pc, OpName(in.InstrOpcode), err)
		}
	}
	e.body.Write(e.cold.Bytes())
	e.body.WriteString("}\n\n")
	return nil
}

func (e *emitter) copyBytes(dst, src string, size uint32) {
	switch size {
	case 1, 2, 4, 8:
		t := e.newTmp()
		typ := valueType(size)
		e.line("%s = load %s, ptr %s, align 1", t, typ, src)
		e.line("store %s %s, ptr %s, align 1", typ, t, dst)
	default:
		e.line("call void @llvm.memcpy.p0.p0.i64(ptr %s, ptr %s, i64 %d, i1 false)", dst, src, size)
	}
}

// memAt returns a pointer into cx_mem at the vm address addr (an i32 value)
// plus offset, after checking that size bytes fit below the data end.
func (e *emitter) memAt(addr string, offset uint32, size uint32, pc uint32) string {
	a64 := e.newTmp()
	e.line("%s = zext i32 %s to i64", a64, addr)
	sum := e.newTmp()
	e.line("%s = add i64 %s, %d", sum, a64, offset)
	bad := e.newTmp()
	e.line("%s = icmp ugt i64 %s, %d", bad, sum, int64(e.dataEnd)-int64(size))
	e.check(bad, "AddressInvalid", pc)
	p := e.newTmp()
	e.line("%s = getelementptr inbounds i8, ptr @cx_mem, i64 %s", p, sum)
	return p
}

func (e *emitter) instr(pc uint32, in *ir.Instr) error {
	op := in.InstrOpcode
	direct := true
	defer func() {
		if direct {
			e.stats.StatsDirect[op]++
		} else {
			e.stats.StatsFallback[op]++
		}
	}()
	switch op {
	case ir.OpcodeReturn:
		fn := e.fn
		if len(fn.FuncOutputs) == 0 {
			e.line("ret void")
			return nil
		}
		agg := "undef"
		rt := e.retType(fn)
		for i, slot := range fn.FuncOutputs {
			p, err := e.ptrAt(slot.SlotOffset, true, slot.SlotSize)
			if err != nil {
				return err
			}
			v := e.newTmp()
			e.line("%s = load %s, ptr %s, align 1", v, valueType(slot.SlotSize), p)
			next := e.newTmp()
			e.line("%s = insertvalue %s %s, %s %s, %d", next, rt, agg, valueType(slot.SlotSize), v, i)
			agg = next
		}
		e.line("ret %s %s", rt, agg)
		return nil
	case ir.OpcodeJump:
		e.line("br label %%pc%d", in.InstrM)
		return nil
	case ir.OpcodeJumpIf:
		v, err := e.load(in, 0, "i8", 1)
		if err != nil {
			return err
		}
		c := e.newTmp()
		e.line("%s = icmp ne i8 %s, 0", c, v)
		e.line("br i1 %s, label %%pc%d, label %%pc%d", c, in.InstrM, in.InstrN)
		return nil
	case ir.OpcodeZero:
		p, err := e.ptr(in, 0, in.InstrN)
		if err != nil {
			return err
		}
		e.line("call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 %d, i1 false)", p, in.InstrN)
		return nil
	case ir.OpcodeMove:
		if offset, inFrame := operand(in, 0); in.InstrN <= 8 {
			if v, found := e.constant(offset, inFrame, valueType(in.InstrN), in.InstrN); found {
				return e.store(in, 2, valueType(in.InstrN), in.InstrN, v)
			}
		}
		src, err := e.ptr(in, 0, in.InstrN)
		if err != nil {
			return err
		}
		dst, err := e.ptr(in, 2, in.InstrN)
		if err != nil {
			return err
		}
		e.copyBytes(dst, src, in.InstrN)
		return nil
	case ir.OpcodeCallFrame:
		if e.call.active {
			return fmt.Errorf("nested call")
		}
		e.call = pendingCall{active: true, frame: in.InstrN, args: map[uint32]string{}, sizes: map[uint32]uint32{}}
		d := e.newTmp()
		e.line("%s = icmp uge i32 %%depth, %d", d, ir.ImageCallDepthMax-1)
		e.check(d, "StackOverflow", pc)
		end := e.newTmp()
		e.line("%s = add i64 %d, %d", end, e.fn.FuncFrameSize, in.InstrN)
		end2 := e.newTmp()
		e.line("%s = add i64 %%fp64, %s", end2, end)
		s := e.newTmp()
		e.line("%s = icmp ugt i64 %s, %d", s, end2, e.image.ImageStackSize)
		e.check(s, "StackOverflow", pc)
		return nil
	case ir.OpcodeCallArg:
		if !e.call.active {
			return fmt.Errorf("argument outside a call")
		}
		v, err := e.load(in, 0, valueType(in.InstrN), in.InstrN)
		if err != nil {
			return err
		}
		e.call.args[in.InstrC] = v
		e.call.sizes[in.InstrC] = in.InstrN
		return nil
	case ir.OpcodeCall:
		callee := &e.image.ImageFuncs[in.InstrM]
		if len(callee.FuncCode) == 0 {
			return fmt.Errorf("call of %s, which has no code", e.image.ImageFuncNames[in.InstrM])
		}
		if callee.FuncFrameSize != e.call.frame {
			return fmt.Errorf("call frame %d, callee frame %d", e.call.frame, callee.FuncFrameSize)
		}
		args := []string{"i32 %sp"}
		d := e.newTmp()
		e.line("%s = add i32 %%depth, 1", d)
		args = append(args, "i32 "+d)
		used := 0
		for _, slot := range callee.FuncInputs {
			typ := valueType(slot.SlotSize)
			if v, found := e.call.args[slot.SlotOffset]; found {
				if e.call.sizes[slot.SlotOffset] != slot.SlotSize {
					return fmt.Errorf("argument size %d for a slot of %d", e.call.sizes[slot.SlotOffset], slot.SlotSize)
				}
				args = append(args, typ+" "+v)
				used++
				continue
			}
			args = append(args, typ+" "+zeroValue(slot.SlotSize))
		}
		if used != len(e.call.args) {
			return fmt.Errorf("argument outside the callee inputs")
		}
		rt := e.retType(callee)
		if rt == "void" {
			e.line("call void %s(%s)", e.funcName(in.InstrM), strings.Join(args, ", "))
			e.result = ""
		} else {
			e.result = e.newTmp()
			e.line("%s = call %s %s(%s)", e.result, rt, e.funcName(in.InstrM), strings.Join(args, ", "))
		}
		e.resultOf = in.InstrM
		e.call = pendingCall{}
		return nil
	case ir.OpcodeCallResult:
		callee := &e.image.ImageFuncs[e.resultOf]
		for i, slot := range callee.FuncOutputs {
			if slot.SlotOffset != in.InstrA {
				continue
			}
			if slot.SlotSize != in.InstrN {
				return fmt.Errorf("result size %d of a slot of %d", in.InstrN, slot.SlotSize)
			}
			v := e.newTmp()
			e.line("%s = extractvalue %s %s, %d", v, e.retType(callee), e.result, i)
			return e.store(in, 2, valueType(slot.SlotSize), slot.SlotSize, v)
		}
		return fmt.Errorf("result %d is not a callee output", in.InstrA)
	case ir.OpcodeAddressIndexStatic, ir.OpcodeAddressIndex:
		var base string
		if op == ir.OpcodeAddressIndexStatic {
			base = e.vmAddress(in, 0)
		} else {
			slot, err := e.load(in, 0, "i32", 4)
			if err != nil {
				return err
			}
			base = e.newTmp()
			e.line("%s = add i32 %s, %d", base, slot, in.InstrM)
		}
		idx, err := e.load(in, 1, "i32", 4)
		if err != nil {
			return err
		}
		neg := e.newTmp()
		e.line("%s = icmp slt i32 %s, 0", neg, idx)
		e.check(neg, "IndexNegative", pc)
		mul := e.newTmp()
		e.line("%s = mul i32 %s, %d", mul, idx, in.InstrN)
		sum := e.newTmp()
		e.line("%s = add i32 %s, %s", sum, base, mul)
		return e.store(in, 2, "i32", 4, sum)
	case ir.OpcodeAddressDerefStatic, ir.OpcodeAddressDeref:
		var p string
		var err error
		if op == ir.OpcodeAddressDerefStatic {
			if p, err = e.ptr(in, 0, pointerSize); err != nil {
				return err
			}
		} else {
			slot, err := e.load(in, 0, "i32", 4)
			if err != nil {
				return err
			}
			p = e.memAt(slot, in.InstrM, pointerSize, pc)
		}
		v := e.newTmp()
		e.line("%s = load %s, ptr %s, align 1", v, valueType(pointerSize), p)
		bad := e.newTmp()
		e.line("%s = icmp uge %s %s, %d", bad, valueType(pointerSize), v, e.dataEnd)
		e.check(bad, "AddressInvalid", pc)
		if pointerSize != 4 {
			t := e.newTmp()
			e.line("%s = trunc %s %s to i32", t, valueType(pointerSize), v)
			v = t
		}
		return e.store(in, 2, "i32", 4, v)
	case ir.OpcodeAddressLoad:
		slot, err := e.load(in, 0, "i32", 4)
		if err != nil {
			return err
		}
		src := e.memAt(slot, in.InstrM, in.InstrN, pc)
		dst, err := e.ptr(in, 2, in.InstrN)
		if err != nil {
			return err
		}
		e.copyBytes(dst, src, in.InstrN)
		return nil
	case ir.OpcodeAddressStore:
		slot, err := e.load(in, 2, "i32", 4)
		if err != nil {
			return err
		}
		dst := e.memAt(slot, in.InstrM, in.InstrN, pc)
		src, err := e.ptr(in, 0, in.InstrN)
		if err != nil {
			return err
		}
		e.copyBytes(dst, src, in.InstrN)
		return nil
	case ir.OpcodeAddressOf, ir.OpcodeAddressPointer:
		var v string
		if op == ir.OpcodeAddressOf {
			v = e.vmAddress(in, 0)
		} else {
			slot, err := e.load(in, 0, "i32", 4)
			if err != nil {
				return err
			}
			v = e.newTmp()
			e.line("%s = add i32 %s, %d", v, slot, in.InstrM)
		}
		if pointerSize != 4 {
			t := e.newTmp()
			e.line("%s = zext i32 %s to %s", t, v, valueType(pointerSize))
			v = t
		}
		return e.store(in, 2, valueType(pointerSize), pointerSize, v)
	case ir.OpcodePrintText:
		direct = false
		e.declare("cx_rt_print_text", "declare void @cx_rt_print_text(ptr, i32) nounwind memory(argmem: read, inaccessiblemem: readwrite)")
		p := e.newTmp()
		e.line("%s = getelementptr inbounds i8, ptr @cx_text, i64 %d", p, in.InstrN)
		e.line("call void @cx_rt_print_text(ptr %s, i32 %d)", p, in.InstrM)
		return nil
	case ir.OpcodePrintStr:
		direct = false
		v, err := e.load(in, 0, valueType(pointerSize), pointerSize)
		if err != nil {
			return err
		}
		if pointerSize != 4 {
			t := e.newTmp()
			e.line("%s = trunc i64 %s to i32", t, v)
			v = t
		}
		e.declare("cx_rt_print_str", "declare i32 @cx_rt_print_str(ptr, i32, i32, i32, i32) nounwind memory(argmem: read, inaccessiblemem: readwrite)")
		ok := e.newTmp()
		e.line("%s = call i32 @cx_rt_print_str(ptr @cx_mem, i32 %d, i32 %d, i32 %s, i32 %d)", ok, e.dataEnd, pointerSize, v, in.InstrN)
		bad := e.newTmp()
		e.line("%s = icmp eq i32 %s, 0", bad, ok)
		e.check(bad, "AddressInvalid", pc)
		return nil
	case ir.OpcodePrintValue:
		direct = false
		code := types.Code(in.InstrM)
		size := printValueSize(code)
		if size == 0 {
			return fmt.Errorf("printvalue of type code %d", code)
		}
		typ := map[types.Code]string{types.F32: "float", types.F64: "double"}[code]
		if typ == "" {
			typ = valueType(size)
		}
		v, err := e.load(in, 0, typ, size)
		if err != nil {
			return err
		}
		name := "cx_rt_print_value_" + map[types.Code]string{types.BOOL: "bool", types.I8: "i8", types.I16: "i16", types.I32: "i32", types.I64: "i64",
			types.UI8: "ui8", types.UI16: "ui16", types.UI32: "ui32", types.UI64: "ui64", types.F32: "f32", types.F64: "f64"}[code]
		e.declare(name, fmt.Sprintf("declare void @%s(%s, i32) nounwind memory(inaccessiblemem: readwrite)", name, typ))
		e.line("call void @%s(%s %s, i32 %d)", name, typ, v, in.InstrN)
		return nil
	case ir.OpcodeTestEqual:
		direct = false
		a, err := e.ptr(in, 0, in.InstrN)
		if err != nil {
			return err
		}
		b, err := e.ptr(in, 1, in.InstrM)
		if err != nil {
			return err
		}
		msg, err := e.load(in, 2, valueType(pointerSize), pointerSize)
		if err != nil {
			return err
		}
		if pointerSize != 4 {
			t := e.newTmp()
			e.line("%s = trunc i64 %s to i32", t, msg)
			msg = t
		}
		file, number := "", uint32(0)
		if line, found := ir.LineAt(e.image, e.fid, pc); found {
			file, number = e.image.ImageFileNames[line.LineFile], line.LineNumber
		}
		where := e.str(fmt.Sprintf("%s: %d", file, number))
		e.declare("cx_rt_test_equal", "declare void @cx_rt_test_equal(ptr, i32, ptr, i32, ptr, i32, i32, i32, ptr) nounwind")
		e.line("call void @cx_rt_test_equal(ptr %s, i32 %d, ptr %s, i32 %d, ptr @cx_mem, i32 %d, i32 %d, i32 %s, ptr %s)", a, in.InstrN, b, in.InstrM, e.dataEnd, pointerSize, msg, where)
		return nil
	}

	info := OpInfos[op]
	if info.OpKind == KindNone {
		return fmt.Errorf("opcode %d not translated", op)
	}
	t := info.OpType
	T := typeLlvm[t]
	a, err := e.load(in, 0, T, typeSize[t])
	if err != nil {
		return err
	}
	b := ""
	if !kindUnary(info.OpKind) {
		if b, err = e.load(in, 1, T, typeSize[t]); err != nil {
			return err
		}
	}
	rt := resultType(info)
	R := typeLlvm[rt]
	r := e.newTmp()
	tmpl, found := Template(op)
	if !found {
		direct = false
		name := FallbackName(op)
		args := T + " " + a
		if b != "" {
			args += ", " + T + " " + b
		}
		params := T
		if b != "" {
			params += ", " + T
		}
		if rt == TypeNone {
			e.declare(name, fmt.Sprintf("declare void @%s(%s) nounwind memory(inaccessiblemem: readwrite)", name, params))
			e.line("call void @%s(%s)", name, args)
			return nil
		}
		e.declare(name, fmt.Sprintf("declare %s @%s(%s) nounwind willreturn memory(none)", R, name, params))
		e.line("%s = call %s @%s(%s)", r, R, name, args)
		return e.store(in, 2, R, typeSize[rt], r)
	}
	e.expand(tmpl, info, a, b, r, pc)
	return e.store(in, 2, R, typeSize[rt], r)
}

// expand writes a template with its placeholders filled.
func (e *emitter) expand(tmpl string, info OpInfo, a, b, r string, pc uint32) {
	t := info.OpType
	rt := resultType(info)
	temps := map[string]string{}
	for i := 1; i <= 9; i++ {
		key := "{t" + strconv.Itoa(i) + "}"
		if strings.Contains(tmpl, key) {
			temps[key] = e.newTmp()
		}
	}
	fsuffix := map[Type]string{TypeF32: "f32", TypeF64: "f64"}[t]
	pairs := []string{"{a}", a, "{b}", b, "{r}", r, "{T}", typeLlvm[t], "{R}", typeLlvm[rt], "{F}", fsuffix,
		"{W1}", strconv.Itoa(int(typeBits[t]) - 1), "{W}", strconv.Itoa(int(typeBits[t]))}
	for k, v := range temps {
		pairs = append(pairs, k, v)
	}
	text := strings.NewReplacer(pairs...).Replace(tmpl)
	for _, l := range strings.Split(text, "\n") {
		if strings.HasPrefix(l, "check ") {
			f := strings.Fields(l)
			e.check(f[1], f[2], pc)
			continue
		}
		e.line("%s", l)
	}
	for _, intrinsic := range []string{"llvm.abs", "llvm.smax", "llvm.smin", "llvm.umax", "llvm.umin", "llvm.fabs"} {
		if !strings.Contains(text, "@"+intrinsic+".") {
			continue
		}
		T := typeLlvm[t]
		suffix := T
		if intrinsic == "llvm.fabs" {
			suffix = fsuffix
			e.declare(intrinsic+"."+suffix, fmt.Sprintf("declare %s @%s.%s(%s)", T, intrinsic, suffix, T))
		} else if intrinsic == "llvm.abs" {
			e.declare(intrinsic+"."+suffix, fmt.Sprintf("declare %s @%s.%s(%s, i1)", T, intrinsic, suffix, T))
		} else {
			e.declare(intrinsic+"."+suffix, fmt.Sprintf("declare %s @%s.%s(%s, %s)", T, intrinsic, suffix, T, T))
		}
	}
}

// dataRuns returns the runs of the data segment that are not zero, as
// [start, end) offsets into ImageData, joining runs less than 64 bytes apart.
func (e *emitter) dataRuns() [][2]int {
	data := e.image.ImageData
	var runs [][2]int
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			continue
		}
		if n := len(runs); n > 0 && i-runs[n-1][1] < 64 {
			runs[n-1][1] = i + 1
			continue
		}
		runs = append(runs, [2]int{i, i + 1})
	}
	return runs
}

// Span is a byte range of the data segment, [SpanStart, SpanEnd) in vm
// addresses.
type Span struct {
	SpanStart, SpanEnd uint32
}

// Options changes how Emit translates an image.
type Options struct {
	// ReadOnly lists the data the program never writes, its literals. The
	// image does not say which data is a literal, so the caller finds it in
	// the program. A load from it becomes a constant.
	ReadOnly []Span
}

// constant returns the value of size bytes at the absolute address offset as
// an LLVM constant of typ, false when the bytes may change at run time.
func (e *emitter) constant(offset uint32, inFrame bool, typ string, size uint32) (string, bool) {
	if inFrame || size > 8 || offset < e.image.ImageDataStart || offset+size > e.dataEnd {
		return "", false
	}
	for a := offset; a < offset+size; a++ {
		if !e.readOnly[a-e.image.ImageDataStart] {
			return "", false
		}
	}
	var v uint64
	for i := uint32(0); i < size; i++ {
		v |= uint64(e.image.ImageData[offset-e.image.ImageDataStart+i]) << (8 * i)
	}
	switch typ {
	case "float":
		t := e.newTmp()
		e.line("%s = bitcast i32 %d to float", t, int32(uint32(v)))
		return t, true
	case "double":
		t := e.newTmp()
		e.line("%s = bitcast i64 %d to double", t, int64(v))
		return t, true
	}
	switch size {
	case 1:
		return strconv.Itoa(int(int8(v))), true
	case 2:
		return strconv.Itoa(int(int16(v))), true
	case 4:
		return strconv.Itoa(int(int32(v))), true
	case 8:
		return strconv.FormatInt(int64(v), 10), true
	}
	return "", false
}

// markReadOnly fills e.readOnly from options, leaving out every byte an
// instruction writes by its absolute address.
func (e *emitter) markReadOnly(options Options) {
	start := e.image.ImageDataStart
	e.readOnly = make([]bool, len(e.image.ImageData))
	for _, span := range options.ReadOnly {
		for a := max(span.SpanStart, start); a < span.SpanEnd && a < e.dataEnd; a++ {
			e.readOnly[a-start] = true
		}
	}
	for id := range e.image.ImageFuncs {
		for pc := range e.image.ImageFuncs[id].FuncCode {
			in := &e.image.ImageFuncs[id].FuncCode[pc]
			for _, ac := range accesses(in) {
				offset, inFrame := operand(in, ac.which)
				written := ac.which == 2 || in.InstrOpcode == ir.OpcodeZero
				if inFrame || !written || offset < start || offset >= e.dataEnd {
					continue
				}
				end := offset + ac.size
				for a := offset; a < end && a < e.dataEnd; a++ {
					e.readOnly[a-start] = false
				}
			}
		}
	}
}
