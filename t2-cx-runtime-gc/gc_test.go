package gc

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
)

// gcTestHeap builds a memory with a root slot at 0 and a heap at 64 that
// holds a garbage str "dead" and then a live str "live" the root points to.
func gcTestHeap() (memory []byte, heapStartsAt types.Pointer, heapPointer types.Pointer, live types.Pointer) {
	heapStartsAt = 64
	size := types.ObjectHeaderSize + 4
	memory = make([]byte, 256)
	garbage := heapStartsAt + constants.NullHeapAddressOffset
	live = garbage + size
	types.ObjectPayloadWrite(memory, garbage, size, []byte("dead"))
	types.ObjectPayloadWrite(memory, live, size, []byte("live"))
	types.PtrWrite(memory, 0, live)
	return memory, heapStartsAt, live + size - heapStartsAt, live
}

// gcTestState returns a State prepared for memory and its heap.
func gcTestState(memory []byte, heapStartsAt types.Pointer) *State {
	var state State
	StatePrepare(&state, types.IntToPtrCast(len(memory))-heapStartsAt, types.IntToPtrCast(len(memory)))
	return &state
}

func TestGarbageCollectorCollectMovesLiveObjectDown(t *testing.T) {
	memory, heapStartsAt, heapPointer, _ := gcTestHeap()
	roots := &RootTable{RootTableGlobals: []Root{{RootOffset: 0, RootType: &Type{TypeKind: KindStr}}}, RootTableStackSize: 32}
	pointer := GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, gcTestState(memory, heapStartsAt))
	moved := heapStartsAt + constants.NullHeapAddressOffset
	if want := constants.NullHeapAddressOffset + types.ObjectHeaderSize + 4; pointer != want {
		t.Fatalf("heap pointer %d, want %d", pointer, want)
	}
	if got := types.PtrRead(memory, 0); got != moved {
		t.Fatalf("root points to %d, want %d", got, moved)
	}
	if got := string(memory[moved+types.ObjectHeaderSize : moved+types.ObjectHeaderSize+4]); got != "live" {
		t.Fatalf("moved object holds %q, want live", got)
	}
}

func TestGarbageCollectorCollectFreesUnreachable(t *testing.T) {
	memory, heapStartsAt, heapPointer, _ := gcTestHeap()
	if pointer := GarbageCollectorCollect(memory, heapStartsAt, heapPointer, &RootTable{RootTableStackSize: 32}, gcTestState(memory, heapStartsAt)); pointer != constants.NullHeapAddressOffset {
		t.Fatalf("heap pointer %d, want %d", pointer, constants.NullHeapAddressOffset)
	}
}

// A frame root at a frame offset is read relative to its frame pointer.
func TestGarbageCollectorCollectFrameRoot(t *testing.T) {
	memory, heapStartsAt, heapPointer, live := gcTestHeap()
	types.PtrWrite(memory, 0, 0)
	types.PtrWrite(memory, 16+8, live)
	roots := &RootTable{
		RootTableFrames:    []RootFrame{{RootFramePointer: 16, RootFrameRoots: []Root{{RootOffset: 8, RootType: &Type{TypeKind: KindStr}}}}},
		RootTableStackSize: 32,
	}
	GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, gcTestState(memory, heapStartsAt))
	if got, want := types.PtrRead(memory, 16+8), heapStartsAt+constants.NullHeapAddressOffset; got != want {
		t.Fatalf("frame slot points to %d, want %d", got, want)
	}
}

// gcTestList builds a heap at 64 of count garbage strs and count list nodes,
// each holding a pointer to the next node and a str, with a root at 0 that
// points to the first node. It returns the root type.
func gcTestList(state *State, count int) (memory []byte, heapStartsAt types.Pointer, heapPointer types.Pointer, root *Type) {
	heapStartsAt = 64
	node := &Type{TypeKind: KindStruct}
	pointer := &Type{TypeKind: KindPointer, TypeElement: node}
	StateTypeAdd(state, pointer)
	node.TypeFields = []TypeField{{TypeFieldOffset: 0, TypeFieldType: pointer}, {TypeFieldOffset: types.PointerSize, TypeFieldType: &Type{TypeKind: KindStr}}}
	strSize := types.ObjectHeaderSize + 1
	nodeSize := types.ObjectHeaderSize + 2*types.PointerSize
	memory = make([]byte, heapStartsAt+constants.NullHeapAddressOffset+types.IntToPtrCast(count)*(2*strSize+nodeSize))
	c := heapStartsAt + constants.NullHeapAddressOffset
	previous := types.Pointer(0)
	for i := 0; i < count; i++ {
		types.ObjectPayloadWrite(memory, c, strSize, []byte{'g'})
		c += strSize
		str := c
		types.ObjectPayloadWrite(memory, str, strSize, []byte{byte('a' + i%26)})
		c += strSize
		types.ObjectSizeWrite(memory, c, nodeSize)
		types.PtrWrite(memory, c+types.ObjectHeaderSize+types.PointerSize, str)
		if previous == 0 {
			types.PtrWrite(memory, 0, c)
		} else {
			types.PtrWrite(memory, previous+types.ObjectHeaderSize, c)
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
	node := types.PtrRead(memory, 0)
	for i := 0; i < count; i++ {
		if node == 0 {
			t.Fatalf("list ends after %d nodes, want %d", i, count)
		}
		str := types.PtrRead(memory, node+types.ObjectHeaderSize+types.PointerSize)
		if got, want := memory[str+types.ObjectHeaderSize], byte('a'+i%26); got != want {
			t.Fatalf("node %d holds %q, want %q", i, got, want)
		}
		node = types.PtrRead(memory, node+types.ObjectHeaderSize)
	}
}

// A list collects with no allocation, and with a StateWork of one entry,
// which leaves nodes for the scan of the heap, it keeps the same objects.
func TestGarbageCollectorCollectList(t *testing.T) {
	const count = 300
	for _, work := range []int{0, 1} {
		var state State
		memory, heapStartsAt, heapPointer, root := gcTestList(&state, count)
		StatePrepare(&state, types.IntToPtrCast(len(memory))-heapStartsAt, types.IntToPtrCast(len(memory)))
		if work != 0 {
			state.StateWork = make([]types.Pointer, 0, work)
		}
		roots := &RootTable{RootTableGlobals: []Root{{RootOffset: 0, RootType: root}}, RootTableStackSize: 32}
		var pointer types.Pointer
		allocs := testing.AllocsPerRun(1, func() {
			pointer = GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, &state)
		})
		if allocs != 0 {
			t.Errorf("work %d: %.0f allocations per collection, want 0", work, allocs)
		}
		live := constants.NullHeapAddressOffset + count*(types.ObjectHeaderSize+1+types.ObjectHeaderSize+2*types.PointerSize)
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
	for i := types.Pointer(0); i < 3; i++ {
		types.PtrWrite(memory, i*8, live+1+i)
		roots.RootTableGlobals = append(roots.RootTableGlobals, Root{RootOffset: i * 8, RootType: inner})
	}
	state := gcTestState(memory, heapStartsAt)
	state.StateViews = make([]Visit, 0, 2)
	GarbageCollectorCollect(memory, heapStartsAt, heapPointer, roots, state)
	moved := heapStartsAt + constants.NullHeapAddressOffset
	for i := types.Pointer(0); i < 3; i++ {
		if got, want := types.PtrRead(memory, i*8), moved+1+i; got != want {
			t.Errorf("pointer %d is %d, want %d", i, got, want)
		}
	}
	if got := string(memory[moved+types.ObjectHeaderSize : moved+types.ObjectHeaderSize+4]); got != "live" {
		t.Fatalf("moved object holds %q, want live", got)
	}
}
