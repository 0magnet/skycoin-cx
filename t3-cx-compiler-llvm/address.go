package llvm

import (
	"fmt"

	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// An array whose address never leaves its function is an alloca, and each
// global is its own LLVM global, so neither shares cx_mem with the rest of
// the program (optimizations/0005). An address slot that points into such
// a region holds the offset into the region instead of the vm address.

// addressRegion is a frame region of the current function, or a global
// region when addressRegionGlobal is set.
type addressRegion struct {
	addressRegionGlobal bool
	addressRegionIndex  int
}

// addressProducer is an OpcodeAddressIndexStatic of a region, or an
// OpcodeAddressIndex of a slot that one wrote, with the largest offset into
// the region its slot can hold.
type addressProducer struct {
	addressProducerRegion    addressRegion
	addressProducerOffsetMax uint64
}

// funcLeaders returns the pcs that start a basic block of fn.
func funcLeaders(fn *ir.Func) map[uint32]bool {
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
	return leaders
}

// accessWritten reports whether in writes its operand which. Every other
// operand of accesses is read.
func accessWritten(in *ir.Instr, which int) bool {
	switch in.InstrOpcode {
	case ir.OpcodeZero:
		return which == 0
	case ir.OpcodeAddressStore, ir.OpcodeTestEqual:
		return false
	}
	return which == 2
}

// globalsPartition joins globalRanges into the global regions. A region
// whose address is taken, or that holds a literal, stays in cx_mem, and
// all stay when an address falls outside every region.
func (e *emitter) globalsPartition() {
	e.globals = partition(e.globalRanges)
	for _, offset := range e.globalAddressed {
		i := find(e.globals, offset)
		if i < 0 {
			for j := range e.globals {
				e.globals[j].regionShadow = true
			}
			break
		}
		e.globals[i].regionShadow = true
	}
	for _, span := range e.image.ImageReadOnly {
		for i := range e.globals {
			if span.SpanStart < e.globals[i].regionEnd && e.globals[i].regionStart < span.SpanEnd {
				e.globals[i].regionShadow = true
			}
		}
	}
	for i := range e.globals {
		e.globals[i].regionName = fmt.Sprintf("@cx_global.%d", e.globals[i].regionStart)
	}
}

// addressRegionOf returns the region of r in layout.
func (e *emitter) addressRegionOf(layout *funcLayout, r addressRegion) *region {
	if r.addressRegionGlobal {
		return &e.globals[r.addressRegionIndex]
	}
	return &layout.funcLayoutFrame[r.addressRegionIndex]
}

// addressAnalyze finds the address slots of fn that hold an offset into a
// region. The slot read by an address opcode gets its value from the last
// write before it in the same basic block. A region stays in cx_mem when
// such a slot is read any other way, when the read cannot be proven inside
// the region, or when the slot may come from another block.
func (e *emitter) addressAnalyze(fn *ir.Func, layout *funcLayout) {
	leaders := funcLeaders(fn)
	layout.funcLayoutProducers = map[uint32]addressProducer{}
	layout.funcLayoutUses = map[uint32]addressProducer{}
	escape := func(r addressRegion) { e.addressRegionOf(layout, r).regionShadow = true }
	var unknown []region
	for pc := range fn.FuncCode {
		in := &fn.FuncCode[pc]
		var base *addressProducer
		for _, ac := range accesses(in) {
			offset, inFrame := operand(in, ac.accessWhich)
			if inFrame || offset < e.dataEnd || ac.accessAddress || accessWritten(in, ac.accessWhich) {
				continue
			}
			read := region{regionStart: offset, regionEnd: offset + ac.accessSize}
			writer, exact, covers := addressWriter(fn, leaders, uint32(pc), read)
			if writer < 0 {
				unknown = append(unknown, read)
				continue
			}
			producer, found := layout.funcLayoutProducers[uint32(writer)]
			if !found {
				if !covers {
					unknown = append(unknown, read)
				}
				continue
			}
			size, allowed := addressUseSize(in, ac.accessWhich)
			if !exact || !allowed {
				escape(producer.addressProducerRegion)
				continue
			}
			if in.InstrOpcode == ir.OpcodeAddressIndex {
				base = &producer
				continue
			}
			r := e.addressRegionOf(layout, producer.addressProducerRegion)
			if producer.addressProducerOffsetMax+uint64(in.InstrM)+uint64(size) > uint64(r.regionEnd-r.regionStart) {
				escape(producer.addressProducerRegion)
				continue
			}
			layout.funcLayoutUses[uint32(pc)] = producer
		}
		switch in.InstrOpcode {
		case ir.OpcodeAddressIndexStatic:
			offset, inFrame := operand(in, 0)
			r := addressRegion{addressRegionGlobal: !inFrame, addressRegionIndex: -1}
			if inFrame {
				r.addressRegionIndex = find(layout.funcLayoutFrame, offset)
			} else if offset >= e.image.ImageDataStart {
				r.addressRegionIndex = find(e.globals, offset)
			}
			if r.addressRegionIndex < 0 {
				continue
			}
			if in.InstrAux == 0 {
				escape(r)
				continue
			}
			start := e.addressRegionOf(layout, r).regionStart
			layout.funcLayoutProducers[uint32(pc)] = addressProducer{
				addressProducerRegion:    r,
				addressProducerOffsetMax: uint64(offset-start) + uint64(in.InstrAux-1)*uint64(in.InstrN),
			}
		case ir.OpcodeAddressIndex:
			if base == nil {
				continue
			}
			if in.InstrAux == 0 {
				escape(base.addressProducerRegion)
				continue
			}
			layout.funcLayoutUses[uint32(pc)] = *base
			layout.funcLayoutProducers[uint32(pc)] = addressProducer{
				addressProducerRegion:    base.addressProducerRegion,
				addressProducerOffsetMax: base.addressProducerOffsetMax + uint64(in.InstrM) + uint64(in.InstrAux-1)*uint64(in.InstrN),
			}
		}
	}
	for _, read := range unknown {
		for pc, producer := range layout.funcLayoutProducers {
			slot, _ := operand(&fn.FuncCode[pc], 2)
			if slot < read.regionEnd && read.regionStart < slot+4 {
				escape(producer.addressProducerRegion)
			}
		}
	}
}

// addressUseSize returns the bytes an address opcode reads or writes at the
// slot its operand which holds, false when the operand is not that slot.
func addressUseSize(in *ir.Instr, which int) (uint32, bool) {
	switch {
	case in.InstrOpcode == ir.OpcodeAddressLoad && which == 0:
		return in.InstrN, true
	case in.InstrOpcode == ir.OpcodeAddressStore && which == 2:
		return in.InstrN, true
	case in.InstrOpcode == ir.OpcodeAddressDeref && which == 0:
		return pointerSize, true
	case in.InstrOpcode == ir.OpcodeAddressIndex && which == 0:
		return 0, true
	}
	return 0, false
}

// addressWriter returns the last pc before pc in its basic block that
// writes scratch overlapping read, or -1, and whether it writes exactly
// read and whether it covers read.
func addressWriter(fn *ir.Func, leaders map[uint32]bool, pc uint32, read region) (int, bool, bool) {
	if leaders[pc] {
		return -1, false, false
	}
	for q := int(pc) - 1; q >= 0; q-- {
		in := &fn.FuncCode[q]
		for _, ac := range accesses(in) {
			offset, inFrame := operand(in, ac.accessWhich)
			if inFrame || !accessWritten(in, ac.accessWhich) {
				continue
			}
			if offset < read.regionEnd && read.regionStart < offset+ac.accessSize {
				end := offset + ac.accessSize
				return q, offset == read.regionStart && end == read.regionEnd, offset <= read.regionStart && end >= read.regionEnd
			}
		}
		if leaders[uint32(q)] {
			break
		}
	}
	return -1, false, false
}

// addressProducerLocal returns the producer that pc is, when its region is
// kept out of cx_mem.
func (e *emitter) addressProducerLocal(pc uint32) (addressProducer, bool) {
	layout := &e.layouts[e.function]
	producer, found := layout.funcLayoutProducers[pc]
	if !found || e.addressRegionOf(layout, producer.addressProducerRegion).regionShadow {
		return addressProducer{}, false
	}
	return producer, true
}

// addressAt returns a pointer to size bytes at the address slot value addr
// plus offset, read by the instruction at pc. A slot that holds an offset
// into a region kept out of cx_mem points into the region, checked
// against its end as memAt checks a vm address. addressAnalyze proved the
// check never fails, but without it LLVM moved the constant stores of
// structserialize out of its loop and could not forward them, see
// optimizations/0005.
func (e *emitter) addressAt(addr string, offset uint32, size uint32, pc uint32) string {
	layout := &e.layouts[e.function]
	producer, found := layout.funcLayoutUses[pc]
	if !found {
		return e.memAt(addr, offset, size, pc)
	}
	r := e.addressRegionOf(layout, producer.addressProducerRegion)
	if r.regionShadow {
		return e.memAt(addr, offset, size, pc)
	}
	a64 := e.localNew()
	e.line("%s = zext i32 %s to i64", a64, addr)
	sum := e.localNew()
	e.line("%s = add i64 %s, %d", sum, a64, offset)
	bad := e.localNew()
	e.line("%s = icmp ugt i64 %s, %d", bad, sum, int64(r.regionEnd-r.regionStart)-int64(size))
	e.check(bad, StopAddressInvalid, pc)
	p := e.localNew()
	e.line("%s = getelementptr inbounds i8, ptr %s, i64 %s", p, r.regionName, sum)
	return p
}

// globalsDefine writes an LLVM global for each global region kept out of
// cx_mem, with its bytes of ImageData.
func (e *emitter) globalsDefine() {
	for _, r := range e.globals {
		if r.regionShadow {
			continue
		}
		initial := e.image.ImageData[r.regionStart-e.image.ImageDataStart : r.regionEnd-e.image.ImageDataStart]
		value := "zeroinitializer"
		for _, b := range initial {
			if b != 0 {
				value = llvmBytes(initial)
				break
			}
		}
		fmt.Fprintf(&e.out, "%s = internal global [%d x i8] %s, align 16\n", r.regionName, len(initial), value)
	}
}
