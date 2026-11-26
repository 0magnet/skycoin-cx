package ast

import (
	"sync"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t1-constants"
	gc "github.com/skycoin/skycoin-cx/t2-cx-runtime-gc"
)

// MinHeapSize is the initial heap of a program with the heap cap
// MaxHeapSize, see MinHeapSizeOf.
func MinHeapSize() types.Pointer {
	return MinHeapSizeOf(constants.MaxHeapSize)
}

// MinHeapSizeOf is the initial heap of a program with the heap cap heapCap:
// HeapSizeInit, at most heapCap and at least NullHeapAddressOffset,
// the bytes of nil.
func MinHeapSizeOf(heapCap types.Pointer) types.Pointer {
	return max(min(constants.HeapSizeInit, heapCap), constants.NullHeapAddressOffset)
}

// HeapCap returns the heap cap of prgrm, Heap.Max, or MaxHeapSize for a
// program that has none.
func HeapCap(prgrm *CXProgram) types.Pointer {
	if prgrm.Heap.Max == 0 {
		return constants.MaxHeapSize
	}
	return prgrm.Heap.Max
}

// EnsureMinimumHeapSize sizes the memory of `prgrm` before a run. It
// allocates the memory for the heap cap once, with the scratch and root
// table of the collector (D42), gives the heap `MinHeapSizeOf` bytes after
// the data segment, and starts the GC pressure from 0. It allocates nothing
// for a program it has sized.
// GCREACH: sets the GC pressure and its limit and prepares the collector,
// see docs/GarbageCollector.md.
func (cxprogram *CXProgram) EnsureMinimumHeapSize() {
	if cxprogram.Memory == nil {
		panic("ast: the memory of the program was released")
	}
	heap := &cxprogram.Heap
	heap.Max = min(HeapCap(cxprogram), constants.EscapeOffsetFlag/2)
	heapMemoryReserve(cxprogram)
	currHeapSize := types.IntToPtrCast(len(cxprogram.Memory)) - heap.StartsAt
	if minHeapSize := MinHeapSizeOf(heap.Max); currHeapSize < minHeapSize {
		length := len(cxprogram.Memory)
		cxprogram.Memory = cxprogram.Memory[:heap.StartsAt+minHeapSize]
		clear(cxprogram.Memory[length:])
		heap.Size = minHeapSize
	}
	GarbageCollectorPrepare(cxprogram)
	heap.Pressure = 0
	heap.PressureLimit = heapPressureLimit(cxprogram)
}

// heapMemoryReserve gives prgrm.Memory the capacity of the stack, the data
// and the heap cap, which the heap grows into without allocating. It takes
// the memory a released program left in the reserve when it is big enough.
// The heap clears what it grows into, so the reserved bytes are not cleared.
func heapMemoryReserve(prgrm *CXProgram) {
	size := max(prgrm.Heap.StartsAt+HeapCap(prgrm), types.IntToPtrCast(len(prgrm.Memory)))
	if types.IntToPtrCast(cap(prgrm.Memory)) < size {
		memory := memoryReserveTake(prgrm, size)
		if memory == nil {
			memory = make([]byte, 0, size)
		}
		prgrm.Memory = append(memory, prgrm.Memory...)
	}
}

// memoryReserve is the memory and the collector scratch of the last program
// that ProgramMemoryRelease released, kept for the next program to prepare
// (D106).
type memoryReserve struct {
	memoryReserveLock      sync.Mutex
	memoryReserveMemory    []byte
	memoryReserveCollector gc.State
}

var programMemoryReserve memoryReserve

// memoryReserveTake returns the reserved memory, emptied, when it holds size
// bytes, or nil. It gives prgrm the reserved collector scratch when prgrm has
// none.
// GCREACH: hands over the collector scratch, see docs/GarbageCollector.md.
func memoryReserveTake(prgrm *CXProgram, size types.Pointer) []byte {
	reserve := &programMemoryReserve
	reserve.memoryReserveLock.Lock()
	defer reserve.memoryReserveLock.Unlock()
	if prgrm.Heap.Collector.StateStarts == nil {
		prgrm.Heap.Collector = reserve.memoryReserveCollector
		reserve.memoryReserveCollector = gc.State{}
	}
	memory := reserve.memoryReserveMemory
	if types.IntToPtrCast(cap(memory)) < size {
		return nil
	}
	reserve.memoryReserveMemory = nil
	return memory[:0]
}

// ProgramMemoryRelease gives the memory and the collector scratch of prgrm
// to the next program that RunPrepare sizes, so programs loaded one after
// another in a process allocate the heap cap once (D106). prgrm must not run
// or be read again, as its memory is gone.
// GCREACH: takes the collector scratch, see docs/GarbageCollector.md.
func ProgramMemoryRelease(prgrm *CXProgram) {
	if prgrm == nil {
		return
	}
	reserve := &programMemoryReserve
	reserve.memoryReserveLock.Lock()
	defer reserve.memoryReserveLock.Unlock()
	if cap(prgrm.Memory) > cap(reserve.memoryReserveMemory) {
		reserve.memoryReserveMemory = prgrm.Memory[:0]
	}
	collector := prgrm.Heap.Collector
	if !collector.StateCollecting && len(collector.StateSlots) > len(reserve.memoryReserveCollector.StateSlots) {
		clear(collector.StateTypes)
		collector.StateTypes = collector.StateTypes[:0]
		reserve.memoryReserveCollector = collector
	}
	prgrm.Memory = nil
	prgrm.Heap.Collector = gc.State{}
}

// MemoryResize sets the heap of prgrm to newMemSize bytes, at most its cap.
// The heap grows into the capacity of prgrm.Memory, which zeroes the new
// bytes, and shrinks by reslicing.
func MemoryResize(prgrm *CXProgram, newMemSize types.Pointer, expand bool) {
	heapCap := HeapCap(prgrm)
	if newMemSize > heapCap {
		newMemSize = heapCap
	}

	if newMemSize == prgrm.Heap.Size {
		// Then we're at the limit; we can't expand anymore.
		// We can only hope that the free memory is enough for the CX program to continue running.
		return
	}

	end := prgrm.Heap.StartsAt + newMemSize
	if expand {
		if types.IntToPtrCast(cap(prgrm.Memory)) < end {
			// A program that EnsureMinimumHeapSize did not size.
			heapMemoryReserve(prgrm)
		}
		length := len(prgrm.Memory)
		prgrm.Memory = prgrm.Memory[:end]
		clear(prgrm.Memory[length:])
	} else {
		prgrm.Memory = prgrm.Memory[:end]
	}
	prgrm.Heap.Size = newMemSize
}

// SequenceAllocate allocates size bytes on the heap and returns their address.
// It never collects. When the heap is full it grows it, up to the heap
// cap, and past that panics with RuntimeOutOfMemory. The bytes count toward
// Heap.Pressure, which the next GarbageCollectorCheck reads.
// GCREACH: grows the heap and adds to the GC pressure,
// see docs/GarbageCollector.md.
func SequenceAllocate(prgrm *CXProgram, size types.Pointer) types.Pointer {
	address := prgrm.Heap.Pointer
	free := address + size
	if free > prgrm.Heap.Size {
		if free > HeapCap(prgrm) || free < address {
			HeapOutOfMemory(prgrm, uint64(size))
		}
		// Growing to leave MinHeapFreeRatio free keeps the next
		// allocations from growing it again before the next collection.
		grown := max(free, types.F32ToPtrCast(float32(free)/(1.0-constants.MinHeapFreeRatio)))
		MemoryResize(prgrm, grown, true)
	}
	prgrm.Heap.Pointer = free
	prgrm.Heap.Pressure += size
	return address + prgrm.Heap.StartsAt
}

// HeapOutOfMemory stops the run with an out of memory error for an
// allocation of request bytes that does not fit the heap cap.
func HeapOutOfMemory(prgrm *CXProgram, request uint64) {
	panic(RuntimeOutOfMemory{RuntimeOutOfMemoryRequest: request, RuntimeOutOfMemoryTaken: prgrm.Heap.Pointer, RuntimeOutOfMemoryCap: HeapCap(prgrm)})
}

// StrAllocate returns a new str object holding data. It writes the object
// after the allocation, which may give prgrm a new Memory.
// GCREACH: allocates through SequenceAllocate, which adds to the GC pressure,
// see docs/GarbageCollector.md.
// VIOLATION STRHOT-281: B, only the boundary opcodes use Go strings, #412.
func StrAllocate(prgrm *CXProgram, text string) types.Pointer {
	size := types.ObjectHeaderSize + types.IntToPtrCast(len(text))
	object := SequenceAllocate(prgrm, size)
	types.ObjectSizeWrite(prgrm.Memory, object, size)
	copy(prgrm.Memory[object+types.ObjectHeaderSize:object+size], text)
	return object
}

// StrAllocateAt writes a new str object holding data and stores its address
// at address. Nothing collects inside it, so address stays valid.
// GCREACH: allocates through SequenceAllocate, which adds to the GC pressure,
// see docs/GarbageCollector.md.
// VIOLATION STRHOT-282: B, only the boundary opcodes use Go strings, #412.
func StrAllocateAt(prgrm *CXProgram, address types.Pointer, text string) {
	types.PtrWrite(prgrm.Memory, address, StrAllocate(prgrm, text))
}

// StrBytesAllocateAt writes a new str object holding first followed by second
// and stores its address at address. Both may be views into prgrm.Memory,
// which keeps its array while the heap grows.
// GCREACH: allocates through SequenceAllocate, which adds to the GC pressure,
// see docs/GarbageCollector.md.
func StrBytesAllocateAt(prgrm *CXProgram, address types.Pointer, first []byte, second []byte) {
	size := types.ObjectHeaderSize + types.IntToPtrCast(len(first)+len(second))
	object := SequenceAllocate(prgrm, size)
	types.ObjectSizeWrite(prgrm.Memory, object, size)
	payload := prgrm.Memory[object+types.ObjectHeaderSize : object+size]
	copy(payload[copy(payload, first):], second)
	types.PtrWrite(prgrm.Memory, address, object)
}

// heapPressureLimit is the bytes the program may allocate before the next
// collection: the free heap, but no more than half of what is left below the
// heap cap, so that a heap near its cap collects before the heap is full
// rather than stopping with an out of memory error.
// GCREACH: sets the GC pressure limit, see docs/GarbageCollector.md.
func heapPressureLimit(prgrm *CXProgram) types.Pointer {
	heap := &prgrm.Heap
	room := HeapCap(prgrm) - min(heap.Pointer, HeapCap(prgrm))
	return min(heap.Size-heap.Pointer, room/2)
}
