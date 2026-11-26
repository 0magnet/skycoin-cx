package gc

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// gcTestHeap builds a memory with a root slot at 0 and a heap at 64 that
// holds a garbage str "dead" and then a live str "live" the root points to.
func gcTestHeap() (memory []byte, heapStartsAt valuetypes.Pointer, heapPointer valuetypes.Pointer, live valuetypes.Pointer) {
	heapStartsAt = 64
	size := valuetypes.ObjectHeaderSize + 4
	memory = make([]byte, 256)
	garbage := heapStartsAt + opcodetable.NullHeapAddressOffset
	live = garbage + size
	valuetypes.ObjectPayloadWrite(memory, garbage, size, []byte("dead"))
	valuetypes.ObjectPayloadWrite(memory, live, size, []byte("live"))
	valuetypes.PtrWrite(memory, 0, live)
	return memory, heapStartsAt, live + size - heapStartsAt, live
}

// gcTestState returns a State prepared for memory and its heap.
func gcTestState(memory []byte, heapStartsAt valuetypes.Pointer) *State {
	var state State
	StatePrepare(&state, valuetypes.IntToPtrCast(len(memory))-heapStartsAt, valuetypes.IntToPtrCast(len(memory)))
	return &state
}

func TestGarbageCollectorCollectMovesLiveObjectDown(t *testing.T) {
	memory, heapStartsAt, heapPointer, _ := gcTestHeap()
	roots := &RootTable{RootTableGlobals: []Root{{RootOffset: 0, RootType: &Type{TypeKind: KindStr}}}, RootTableStackSize: 32}
	pointer := GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, gcTestState(memory, heapStartsAt))
	moved := heapStartsAt + opcodetable.NullHeapAddressOffset
	if want := opcodetable.NullHeapAddressOffset + valuetypes.ObjectHeaderSize + 4; pointer != want {
		t.Fatalf("heap pointer %d, want %d", pointer, want)
	}
	if got := valuetypes.PtrRead(memory, 0); got != moved {
		t.Fatalf("root points to %d, want %d", got, moved)
	}
	if got := string(memory[moved+valuetypes.ObjectHeaderSize : moved+valuetypes.ObjectHeaderSize+4]); got != "live" {
		t.Fatalf("moved object holds %q, want live", got)
	}
}

func TestGarbageCollectorCollectFreesUnreachable(t *testing.T) {
	memory, heapStartsAt, heapPointer, _ := gcTestHeap()
	if pointer := GarbageCollectorCollect(memory, heapStartsAt, heapPointer, &RootTable{RootTableStackSize: 32}, gcTestState(memory, heapStartsAt)); pointer != opcodetable.NullHeapAddressOffset {
		t.Fatalf("heap pointer %d, want %d", pointer, opcodetable.NullHeapAddressOffset)
	}
}

// A frame root at a frame offset is read relative to its frame pointer.
func TestGarbageCollectorCollectFrameRoot(t *testing.T) {
	memory, heapStartsAt, heapPointer, live := gcTestHeap()
	valuetypes.PtrWrite(memory, 0, 0)
	valuetypes.PtrWrite(memory, 16+8, live)
	roots := &RootTable{
		RootTableFrames:    []RootFrame{{RootFramePointer: 16, RootFrameRoots: []Root{{RootOffset: 8, RootType: &Type{TypeKind: KindStr}}}}},
		RootTableStackSize: 32,
	}
	GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, gcTestState(memory, heapStartsAt))
	if got, want := valuetypes.PtrRead(memory, 16+8), heapStartsAt+opcodetable.NullHeapAddressOffset; got != want {
		t.Fatalf("frame slot points to %d, want %d", got, want)
	}
}

// gcTestList builds a heap at 64 of count garbage strs and count list nodes,
// each holding a pointer to the next node and a str, with a root at 0 that
// points to the first node. It returns the root type.
func gcTestList(state *State, count int) (memory []byte, heapStartsAt valuetypes.Pointer, heapPointer valuetypes.Pointer, root *Type) {
	heapStartsAt = 64
	node := &Type{TypeKind: KindStruct}
	pointer := &Type{TypeKind: KindPointer, TypeElement: node}
	StateTypeAdd(state, pointer)
	node.TypeFields = []TypeField{{TypeFieldOffset: 0, TypeFieldType: pointer}, {TypeFieldOffset: valuetypes.PointerSize, TypeFieldType: &Type{TypeKind: KindStr}}}
	strSize := valuetypes.ObjectHeaderSize + 1
	nodeSize := valuetypes.ObjectHeaderSize + 2*valuetypes.PointerSize
	memory = make([]byte, heapStartsAt+opcodetable.NullHeapAddressOffset+valuetypes.IntToPtrCast(count)*(2*strSize+nodeSize))
	c := heapStartsAt + opcodetable.NullHeapAddressOffset
	previous := valuetypes.Pointer(0)
	for i := 0; i < count; i++ {
		valuetypes.ObjectPayloadWrite(memory, c, strSize, []byte{'g'})
		c += strSize
		str := c
		valuetypes.ObjectPayloadWrite(memory, str, strSize, []byte{byte('a' + i%26)})
		c += strSize
		valuetypes.ObjectSizeWrite(memory, c, nodeSize)
		valuetypes.PtrWrite(memory, c+valuetypes.ObjectHeaderSize+valuetypes.PointerSize, str)
		if previous == 0 {
			valuetypes.PtrWrite(memory, 0, c)
		} else {
			valuetypes.PtrWrite(memory, previous+valuetypes.ObjectHeaderSize, c)
		}
		previous = c
		c += nodeSize
	}
	return memory, heapStartsAt, c - heapStartsAt, pointer
}

// gcTestListCheck walks the list from the root at 0 and fails unless it has
// count nodes whose strs read a, b, c and on.
func gcTestListCheck(t *testing.T, memory []byte, count int) {
	t.Helper()
	node := valuetypes.PtrRead(memory, 0)
	for i := 0; i < count; i++ {
		if node == 0 {
			t.Fatalf("list ends after %d nodes, want %d", i, count)
		}
		str := valuetypes.PtrRead(memory, node+valuetypes.ObjectHeaderSize+valuetypes.PointerSize)
		if got, want := memory[str+valuetypes.ObjectHeaderSize], byte('a'+i%26); got != want {
			t.Fatalf("node %d holds %q, want %q", i, got, want)
		}
		node = valuetypes.PtrRead(memory, node+valuetypes.ObjectHeaderSize)
	}
}

// A list collects with no allocation, and with a StateWork of one entry,
// which leaves nodes for the scan of the heap, it keeps the same objects.
func TestGarbageCollectorCollectList(t *testing.T) {
	const count = 300
	for _, work := range []int{0, 1} {
		var state State
		memory, heapStartsAt, heapPointer, root := gcTestList(&state, count)
		StatePrepare(&state, valuetypes.IntToPtrCast(len(memory))-heapStartsAt, valuetypes.IntToPtrCast(len(memory)))
		if work != 0 {
			state.StateWork = make([]valuetypes.Pointer, 0, work)
		}
		roots := &RootTable{RootTableGlobals: []Root{{RootOffset: 0, RootType: root}}, RootTableStackSize: 32}
		var pointer valuetypes.Pointer
		allocs := testing.AllocsPerRun(1, func() {
			pointer = GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, &state)
		})
		if allocs != 0 {
			t.Errorf("work %d: %.0f allocations per collection, want 0", work, allocs)
		}
		live := opcodetable.NullHeapAddressOffset + count*(valuetypes.ObjectHeaderSize+1+valuetypes.ObjectHeaderSize+2*valuetypes.PointerSize)
		if pointer != live {
			t.Errorf("work %d: heap pointer %d, want %d", work, pointer, live)
		}
		gcTestListCheck(t, memory, count)
		// A second collection moves nothing and keeps the list.
		if again := GarbageCollectorCollect(memory, heapStartsAt, pointer, roots, &state); again != live {
			t.Errorf("work %d: second collection heap pointer %d, want %d", work, again, live)
		}
		gcTestListCheck(t, memory, count)
	}
}

// Pointers into an object past the capacity of StateViews grow it, and
// the collection keeps the object and moves the pointers into it.
func TestGarbageCollectorCollectViewsGrow(t *testing.T) {
	memory, heapStartsAt, heapPointer, live := gcTestHeap()
	inner := &Type{TypeKind: KindPointer, TypeElement: &Type{TypeKind: KindStr}}
	roots := &RootTable{RootTableStackSize: 32}
	for i := valuetypes.Pointer(0); i < 3; i++ {
		valuetypes.PtrWrite(memory, i*8, live+1+i)
		roots.RootTableGlobals = append(roots.RootTableGlobals, Root{RootOffset: i * 8, RootType: inner})
	}
	state := gcTestState(memory, heapStartsAt)
	state.StateViews = make([]Visit, 0, 2)
	GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, state)
	moved := heapStartsAt + opcodetable.NullHeapAddressOffset
	for i := valuetypes.Pointer(0); i < 3; i++ {
		if got, want := valuetypes.PtrRead(memory, i*8), moved+1+i; got != want {
			t.Errorf("pointer %d is %d, want %d", i, got, want)
		}
	}
	if got := string(memory[moved+valuetypes.ObjectHeaderSize : moved+valuetypes.ObjectHeaderSize+4]); got != "live" {
		t.Fatalf("moved object holds %q, want live", got)
	}
}
