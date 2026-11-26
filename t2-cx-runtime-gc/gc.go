// Package gc is the heap collector of the old loop. It sees a root table and
// the heap, and nothing else of the program. See docs/GarbageCollector.md.
package gc

import (
	"math/bits"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
)

// Kind is what a Type value is.
type Kind uint8

const (
	// KindPointer is a pointer, followed to the value it points to.
	KindPointer Kind = iota + 1
	// KindSlice is a slice, followed to each element.
	KindSlice
	// KindStr is a str, whose object holds no references.
	KindStr
	// KindArray is an array, walked element by element in place.
	KindArray
	// KindStruct is a struct, walked field by field in place.
	KindStruct
)

// Type describes where a value holds pointers, slices and strs. A nil
// *Type holds none.
type Type struct {
	TypeKind Kind
	// TypeElement is the pointee, slice element or array element, nil when
	// it holds no references.
	TypeElement *Type
	// TypeElementSize is the size of a slice or array element.
	TypeElementSize types.Pointer
	// TypeLength is the length of an array.
	TypeLength types.Pointer
	// TypeFields are the fields of a struct that hold references.
	TypeFields []TypeField
	// TypeIndex is the position of a pointer or slice type in
	// StateTypes plus one, set by StateTypeAdd. 0 is not added.
	TypeIndex types.Pointer
}

// TypeField is a field of a struct that holds references.
type TypeField struct {
	TypeFieldOffset types.Pointer
	TypeFieldType   *Type
}

// Root is a value the collection starts from. A frame root has a frame
// offset, flagged with EscapeOffsetFlag when the local escaped, and a
// global root an address. RootInner is added after the frame address.
type Root struct {
	RootOffset types.Pointer
	RootInner  types.Pointer
	RootType   *Type
}

// RootFrame is a live frame, its frame pointer and the roots of its
// function.
type RootFrame struct {
	RootFramePointer types.Pointer
	RootFrameRoots   []Root
}

// RootTable is everything the collection reads besides the heap: the
// roots of the globals, the live frames and the stack size, below which a
// frame offset is relative to its frame.
type RootTable struct {
	RootTableGlobals   []Root
	RootTableFrames    []RootFrame
	RootTableStackSize types.Pointer
}

// Visit is a pointer the mark has yet to follow, and its Type.
type Visit struct {
	VisitAddress types.Pointer
	VisitType    *Type
}

// The mark byte of an object header. gcMarkPending is an object the mark
// queued when StateWork was full, found again by a scan of the heap.
const (
	gcMarkLive    = 1
	gcMarkPending = 2
)

// State is the scratch memory of the collection. StatePrepare
// allocates it once for the heap cap, so a collection allocates nothing.
type State struct {
	// StateStarts has a bit for each byte of the heap, set where an
	// object starts.
	StateStarts []uint64
	// StateSlots has a bit for each byte of memory, set where the mark
	// read a reference. Its set bits are the references to update, in
	// address order.
	StateSlots []uint64
	// StateSlotWords has a bit for each word of StateSlots that holds a
	// set bit, and StateSlotHigh is the highest such word.
	StateSlotWords []uint64
	StateSlotHigh  int
	// StateWork holds heap objects to walk with the type that the
	// forwarding address field of their header names while the mark runs.
	StateWork []types.Pointer
	// StateWorkOverflow is set when an object did not fit StateWork
	// and was left gcMarkPending.
	StateWorkOverflow bool
	// StateCollecting is set while a collection runs. A collection that did
	// not finish leaves the bitmaps dirty, so the scratch is not reused.
	StateCollecting bool
	// StateViews holds the pointers into an object, and those to an object
	// already queued with another type, that the mark has yet to follow.
	StateViews []Visit
	// StateTypes are the pointer and slice GCTypes, by TypeIndex-1.
	StateTypes []*Type
}

// StateWorkSize is the entries of StateWork for a heap cap. A full
// StateWork leaves objects for a scan of the heap.
func StateWorkSize(heapCap types.Pointer) int {
	return max(256, int(heapCap/1024))
}

// StateViewsSize is the entries of StateViews for a memory size. More
// pointers into objects pending at once grow it, the one Go allocation a
// collection can make (D97).
func StateViewsSize(memorySize types.Pointer) int {
	return max(256, int(memorySize/512))
}

// StatePrepare allocates the scratch of state for a heap of at most
// heapCap bytes in a memory of at most memorySize bytes. It allocates
// nothing when state already fits them.
func StatePrepare(state *State, heapCap types.Pointer, memorySize types.Pointer) {
	if words := int(heapCap/64) + 1; len(state.StateStarts) < words {
		state.StateStarts = make([]uint64, words)
	}
	if words := int(memorySize/64) + 1; len(state.StateSlots) < words {
		state.StateSlots = make([]uint64, words)
		state.StateSlotWords = make([]uint64, words/64+1)
	}
	if size := StateWorkSize(heapCap); cap(state.StateWork) < size {
		state.StateWork = make([]types.Pointer, 0, size)
	}
	if size := StateViewsSize(memorySize); cap(state.StateViews) < size {
		state.StateViews = make([]Visit, 0, size)
	}
}

// StateTypeAdd gives the pointer or slice type t its TypeIndex.
func StateTypeAdd(state *State, t *Type) {
	state.StateTypes = append(state.StateTypes, t)
	t.TypeIndex = types.IntToPtrCast(len(state.StateTypes))
}

// gcCollection is one collection of the heap from heapStartsAt to
// heapStartsAt+heapPointer in memory.
type gcCollection struct {
	memory       []byte
	heapStartsAt types.Pointer
	heapPointer  types.Pointer
	state        *State
}

// GarbageCollectorCollect collects the heap of memory that starts at
// heapStartsAt and is used up to heapStartsAt+heapPointer. It marks every
// object reachable from roots, walking each root by its Type and
// recording each reference it reads, then gives each live object its new
// address, moves the recorded references to the new addresses and slides
// the objects down. It returns the new heap pointer. state is prepared by
// StatePrepare for the heap and memory.
func GarbageCollectorCollect(memory []byte, heapStartsAt types.Pointer, heapPointer types.Pointer, roots *RootTable, state *State) types.Pointer {
	collection := gcCollection{memory: memory, heapStartsAt: heapStartsAt, heapPointer: heapPointer, state: state}
	state.StateCollecting = true
	collection.gcObjectStartsBuild()
	state.StateWork = state.StateWork[:0]
	state.StateViews = state.StateViews[:0]
	state.StateWorkOverflow = false
	state.StateSlotHigh = 0

	for _, root := range roots.RootTableGlobals {
		collection.gcMarkValue(root.RootOffset, root.RootType)
	}
	for _, frame := range roots.RootTableFrames {
		for _, root := range frame.RootFrameRoots {
			if address, found := collection.gcFrameAddress(roots, frame.RootFramePointer, root.RootOffset); found {
				collection.gcMarkValue(address+root.RootInner, root.RootType)
			}
		}
	}
	collection.gcMarkDrain()

	forward := constants.NullHeapAddressOffset
	end := heapStartsAt + heapPointer
	for object := heapStartsAt + constants.NullHeapAddressOffset; object < end; {
		size := types.ObjectSizeRead(memory, object)
		if size == 0 {
			break
		}
		if memory[object] == gcMarkLive {
			types.ObjectForwardingAddressWrite(memory, object, heapStartsAt+forward)
			forward += size
		}
		object += size
	}

	collection.gcSlotsUpdate()

	for object := heapStartsAt + constants.NullHeapAddressOffset; object < end; {
		size := types.ObjectSizeRead(memory, object)
		if size == 0 {
			break
		}
		if memory[object] == gcMarkLive {
			memory[object] = 0
			to := types.ObjectForwardingAddressRead(memory, object)
			copy(memory[to:to+size], memory[object:object+size])
		}
		object += size
	}
	clear(state.StateStarts[:heapPointer/64+1])
	state.StateCollecting = false
	return forward
}

// gcSlotsUpdate moves each reference the mark recorded, in address order,
// to the new address of the object it points into, and clears the record.
func (collection *gcCollection) gcSlotsUpdate() {
	state := collection.state
	memory := collection.memory
	for summary := 0; summary <= state.StateSlotHigh/64; summary++ {
		words := state.StateSlotWords[summary]
		state.StateSlotWords[summary] = 0
		for words != 0 {
			word := summary*64 + bits.TrailingZeros64(words)
			words &= words - 1
			slotBits := state.StateSlots[word]
			state.StateSlots[word] = 0
			for slotBits != 0 {
				slot := types.IntToPtrCast(word*64 + bits.TrailingZeros64(slotBits))
				slotBits &= slotBits - 1
				pointer := types.PtrRead(memory, slot)
				if object, found := collection.gcObjectContaining(pointer); found {
					types.PtrWrite(memory, slot, pointer-object+types.ObjectForwardingAddressRead(memory, object))
				}
			}
		}
	}
}

// gcMarkValue marks what the value of Type t at address refers to.
func (collection *gcCollection) gcMarkValue(address types.Pointer, t *Type) {
	switch t.TypeKind {
	case KindPointer, KindSlice, KindStr:
		collection.gcMarkReference(address, t)
	case KindArray:
		for i := types.Pointer(0); i < t.TypeLength; i++ {
			collection.gcMarkValue(address+i*t.TypeElementSize, t.TypeElement)
		}
	case KindStruct:
		for _, field := range t.TypeFields {
			collection.gcMarkValue(address+field.TypeFieldOffset, field.TypeFieldType)
		}
	}
}

// gcMarkReference marks the object the pointer, slice or str at slot points
// into, records slot for the update, and queues what the object holds. A
// slot read before is skipped: CX gives a memory location one type, so the
// first read queued all it refers to.
func (collection *gcCollection) gcMarkReference(slot types.Pointer, t *Type) {
	if uint64(slot)+uint64(types.PointerSize) > uint64(len(collection.memory)) {
		return
	}
	pointer := types.PtrRead(collection.memory, slot)
	object, found := collection.gcObjectContaining(pointer)
	if !found {
		return
	}
	state := collection.state
	word, bit := int(slot/64), slot%64
	if state.StateSlots[word]&(1<<bit) != 0 {
		return
	}
	state.StateSlots[word] |= 1 << bit
	state.StateSlotWords[word/64] |= 1 << (word % 64)
	state.StateSlotHigh = max(state.StateSlotHigh, word)
	if collection.memory[object] == 0 {
		collection.memory[object] = gcMarkLive
	}
	if t.TypeElement == nil {
		return
	}
	if pointer == object && t.TypeIndex != 0 {
		// The forwarding address field holds the type the object is queued
		// with until the mark ends.
		switch types.ObjectForwardingAddressRead(collection.memory, object) {
		case 0:
			types.ObjectForwardingAddressWrite(collection.memory, object, t.TypeIndex)
			if len(state.StateWork) == cap(state.StateWork) {
				collection.memory[object] = gcMarkPending
				state.StateWorkOverflow = true
				return
			}
			state.StateWork = append(state.StateWork, object)
			return
		case t.TypeIndex:
			return
		}
	}
	// VIOLATION VMALLOC-011: grows StateViews past StateViewsSize, D97, #407.
	state.StateViews = append(state.StateViews, Visit{VisitAddress: pointer, VisitType: t})
}

// gcMarkDrain walks the queued objects and pointers until none is left,
// then scans the heap for objects left gcMarkPending, until none is.
func (collection *gcCollection) gcMarkDrain() {
	state := collection.state
	memory := collection.memory
	for {
		collection.gcMarkQueuesDrain()
		if !state.StateWorkOverflow {
			return
		}
		state.StateWorkOverflow = false
		end := collection.heapStartsAt + collection.heapPointer
		for object := collection.heapStartsAt + constants.NullHeapAddressOffset; object < end; {
			size := types.ObjectSizeRead(memory, object)
			if size == 0 {
				break
			}
			if memory[object] == gcMarkPending {
				memory[object] = gcMarkLive
				collection.gcVisit(object, state.StateTypes[types.ObjectForwardingAddressRead(memory, object)-1])
				collection.gcMarkQueuesDrain()
			}
			object += size
		}
	}
}

// gcMarkQueuesDrain walks what StateViews and StateWork hold.
func (collection *gcCollection) gcMarkQueuesDrain() {
	state := collection.state
	for {
		if n := len(state.StateViews); n > 0 {
			visit := state.StateViews[n-1]
			state.StateViews = state.StateViews[:n-1]
			collection.gcVisit(visit.VisitAddress, visit.VisitType)
			continue
		}
		n := len(state.StateWork)
		if n == 0 {
			return
		}
		object := state.StateWork[n-1]
		state.StateWork = state.StateWork[:n-1]
		collection.gcVisit(object, state.StateTypes[types.ObjectForwardingAddressRead(collection.memory, object)-1])
	}
}

// gcVisit marks what the pointer or slice of Type t refers to. A pointer
// to the heap is the address of the value less the object header.
func (collection *gcCollection) gcVisit(pointer types.Pointer, t *Type) {
	if t.TypeKind == KindPointer {
		collection.gcMarkValue(pointer+types.ObjectHeaderSize, t.TypeElement)
		return
	}
	object, _ := collection.gcObjectContaining(pointer)
	elements := pointer + types.ObjectHeaderSize + constants.SliceHeaderSize
	end := object + types.ObjectSizeRead(collection.memory, object)
	if elements > end || t.TypeElementSize == 0 {
		return
	}
	length := types.PtrRead(collection.memory, pointer+types.ObjectHeaderSize+types.PointerSize)
	length = min(length, (end-elements)/t.TypeElementSize)
	for i := types.Pointer(0); i < length; i++ {
		collection.gcMarkValue(elements+i*t.TypeElementSize, t.TypeElement)
	}
}

// gcFrameAddress returns the address of the frame variable at offset of the
// frame at framePointer. It returns false for an escaped local whose
// declaration has not run yet. An offset outside the stack is an address.
func (collection *gcCollection) gcFrameAddress(roots *RootTable, framePointer types.Pointer, offset types.Pointer) (types.Pointer, bool) {
	if offset&constants.EscapeOffsetFlag == 0 {
		if offset >= roots.RootTableStackSize {
			return offset, true
		}
		return offset + framePointer, true
	}
	object := types.PtrRead(collection.memory, framePointer+offset&^constants.EscapeOffsetFlag)
	if object < collection.heapStartsAt+constants.NullHeapAddressOffset || object >= collection.heapStartsAt+collection.heapPointer {
		return 0, false
	}
	return object + types.ObjectHeaderSize, true
}

// gcObjectStartsBuild sets the bit of each heap object in StateStarts and
// clears its mark and forwarding address, which a new object may not.
func (collection *gcCollection) gcObjectStartsBuild() {
	starts := collection.state.StateStarts
	end := collection.heapStartsAt + collection.heapPointer
	for c := collection.heapStartsAt + constants.NullHeapAddressOffset; c < end; {
		offset := c - collection.heapStartsAt
		starts[offset/64] |= 1 << (offset % 64)
		collection.memory[c] = 0
		objSize := types.ObjectSizeRead(collection.memory, c)
		if objSize == 0 {
			break
		}
		types.ObjectForwardingAddressWrite(collection.memory, c, 0)
		c += objSize
	}
}

// gcObjectContaining returns the address of the heap object that pointer
// points into. A pointer inside an object, as &s[1] is, keeps it alive.
func (collection *gcCollection) gcObjectContaining(pointer types.Pointer) (types.Pointer, bool) {
	if pointer < collection.heapStartsAt+constants.NullHeapAddressOffset || pointer >= collection.heapStartsAt+collection.heapPointer {
		return 0, false
	}
	starts := collection.state.StateStarts
	offset := pointer - collection.heapStartsAt
	word := int(offset / 64)
	// The starts at or below offset in its word.
	startBits := starts[word] & (2<<(offset%64) - 1)
	for startBits == 0 {
		if word == 0 {
			return 0, false
		}
		word--
		startBits = starts[word]
	}
	start := collection.heapStartsAt + types.IntToPtrCast(word*64+63-bits.LeadingZeros64(startBits))
	if pointer >= start+types.ObjectSizeRead(collection.memory, start) {
		return 0, false
	}
	return start, true
}
