package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"testing"
	"unsafe"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t4-execute"
)

// runHeapProgram is a heap heavy test program and the heap sizes it runs
// with. initHeap 0 collects after nearly every step that allocates.
type runHeapProgram struct {
	path     string
	initHeap valuetypes.Pointer
	heapCap  valuetypes.Pointer
}

// runHeapPrograms allocate, collect and grow the heap: the GC tests with an
// initial heap of 0 bytes, test-gc.cx at the default heap, and a live heap
// grown near a heap cap of 512K that collects there.
func runHeapPrograms() []runHeapProgram {
	programs := []runHeapProgram{{"test-gc.cx", opcodetable.HeapSizeInit, opcodetable.MaxHeapSize}}
	for _, path := range append(gcBenchPrograms, "test-gc.cx", "test-gc-slice-resize.cx", "test-gc-array-of-slices.cx") {
		programs = append(programs, runHeapProgram{path, 0, opcodetable.MaxHeapSize})
	}
	return append(programs, runHeapProgram{"test-heap-cap-near.cx", opcodetable.HeapSizeInit, 512 << 10})
}

// runHeapCompile compiles the test program path with the heap sizes initHeap
// and heapCap into a fresh actions.AST, prepared to run.
func runHeapCompile(tb testing.TB, path string, initHeap valuetypes.Pointer, heapCap valuetypes.Pointer) {
	source, err := os.ReadFile(filepath.Join("..", "tests", path))
	if err != nil {
		tb.Fatal(err)
	}
	initSize, maxSize := opcodetable.HeapSizeInit, opcodetable.MaxHeapSize
	opcodetable.HeapSizeInit, opcodetable.MaxHeapSize = initHeap, heapCap
	defer func() { opcodetable.HeapSizeInit, opcodetable.MaxHeapSize = initSize, maxSize }()
	compileRunProgram(tb, runBenchProgram{name: path, source: string(source)})
	execute.RunPrepare(astbuild.AST)
}

// TestRunHeapAllocs fails when a heap heavy program makes a Go allocation or
// starts a Go collection while it runs (D42). The heap, the scratch of the
// collector and its root table are allocated by RunPrepare.
func TestRunHeapAllocs(t *testing.T) {
	initProgram()
	for _, p := range runHeapPrograms() {
		t.Run(fmt.Sprintf("%s/heap-initial-%d", p.path, p.initHeap), func(t *testing.T) {
			discardStdout(t)
			// The run shrinks the heap no lower than HeapSizeInit, as cx does.
			initSize := opcodetable.HeapSizeInit
			defer func() { opcodetable.HeapSizeInit = initSize }()
			const runs = 2
			var compiled []*ast.CXProgram
			for i := 0; i <= runs; i++ {
				runHeapCompile(t, p.path, p.initHeap, p.heapCap)
				compiled = append(compiled, astbuild.AST)
			}
			opcodetable.HeapSizeInit = p.initHeap
			next := 0
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			gcPercent := debug.SetGCPercent(-1)
			allocs := testing.AllocsPerRun(runs, func() {
				if err := execute.RunCompiled(compiled[next], 0, nil); err != nil {
					t.Fatal(err)
				}
				if compiled[next].RuntimeError.RuntimeErrorRecordRaised {
					t.Fatalf("runtime error %d", compiled[next].RuntimeError.RuntimeErrorRecordCode)
				}
				next++
			})
			debug.SetGCPercent(gcPercent)
			runtime.ReadMemStats(&after)
			if allocs != 0 {
				t.Errorf("%.0f allocations per run, want 0 (D42)", allocs)
			}
			if after.NumGC != before.NumGC {
				t.Errorf("%d Go GC cycles during the runs, want 0", after.NumGC-before.NumGC)
			}
		})
	}
}

// A slice past the heap cap stops the run with an out of memory error that
// names the request and the cap, and a serialized program keeps its cap.
func TestRunHeapCapExceeded(t *testing.T) {
	initProgram()
	discardStdout(t)
	const heapCap = 64 << 10
	runHeapCompile(t, "test-heap-cap-exceeded.cx", opcodetable.HeapSizeInit, heapCap)
	if err := execute.RunCompiled(astbuild.AST, 0, nil); err != nil {
		t.Fatal(err)
	}
	record := astbuild.AST.RuntimeError
	if !record.RuntimeErrorRecordRaised || record.RuntimeErrorRecordCode != valuetypes.CXRuntimeHeapExhaustedError {
		t.Fatalf("runtime error %v code %d, want CX_RUNTIME_HEAP_EXHAUSTED_ERROR", record.RuntimeErrorRecordRaised, record.RuntimeErrorRecordCode)
	}
	outOfMemory, ok := record.RuntimeErrorRecordPanicValue.(ast.RuntimeOutOfMemory)
	want := uint64(valuetypes.ObjectHeaderSize+opcodetable.SliceHeaderSize) + 20000*4
	if !ok || outOfMemory.RuntimeOutOfMemoryCap != heapCap || outOfMemory.RuntimeOutOfMemoryRequest != want {
		t.Fatalf("panic value %#v, want a request of %d bytes and a cap of %d", record.RuntimeErrorRecordPanicValue, want, heapCap)
	}

	image, err := ast.CXProgramSerialize(astbuild.AST, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := ast.Deserialize(image, nil)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Heap.Max != heapCap {
		t.Fatalf("deserialized heap cap %d, want %d", loaded.Heap.Max, heapCap)
	}
}

// runHeapZeroSource reads a slice resized past the initial heap, which grows
// the heap, and the bytes the heap grows into must be zero.
const runHeapZeroSource = `package main

func main() {
	var s []i32
	s = resize(s, 1000000)
	var sum i32
	for i := 0; i < 1000000; i++ {
		sum = sum + s[i]
	}
	printf("%d %d\n", len(s), sum)
}
`

// runHeapOutput compiles and runs the test program path, or "zero" for
// runHeapZeroSource, into a fresh actions.AST and returns what it printed.
// poison fills the memory that the previous program leaves to the reserve
// with 0xa5 first.
func runHeapOutput(t *testing.T, path string, poison bool) string {
	source := runHeapZeroSource
	if path != "zero" {
		text, err := os.ReadFile(filepath.Join("..", "tests", path))
		if err != nil {
			t.Fatal(err)
		}
		source = string(text)
	}
	var reserved *byte
	if poison {
		memory := astbuild.AST.Memory[:cap(astbuild.AST.Memory)]
		for i := range memory {
			memory[i] = 0xa5
		}
		reserved = unsafe.SliceData(memory)
	}
	ast.ProgramMemoryRelease(astbuild.AST)
	out, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	stdout := os.Stdout
	os.Stdout = out
	defer func() { os.Stdout = stdout }()
	compileRunProgram(t, runBenchProgram{name: path, source: source})
	execute.RunPrepare(astbuild.AST)
	if poison && unsafe.SliceData(astbuild.AST.Memory) != reserved {
		t.Fatalf("%s: RunPrepare did not take the released memory", path)
	}
	if err := execute.RunCompiled(astbuild.AST, 0, nil); err != nil {
		t.Fatal(err)
	}
	os.Stdout = stdout
	text, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

// TestProgramMemoryRelease fails when a program prepared with the memory a
// released program left prints other output than with fresh memory (D106).
// The released memory is not cleared, so the heap clears what it grows into.
func TestProgramMemoryRelease(t *testing.T) {
	initProgram()
	for _, path := range []string{"zero", "test-gc.cx", "test-slices.cx", "test-gc-nested.cx", "test-array.cx"} {
		t.Run(path, func(t *testing.T) {
			want := runHeapOutput(t, path, false)
			if path == "zero" && want != "1000000 0\n" {
				t.Fatalf("output %q, want %q", want, "1000000 0\n")
			}
			runHeapOutput(t, "test-gc.cx", false)
			if got := runHeapOutput(t, path, true); got != want {
				t.Errorf("output with released memory:\n%s\nwant:\n%s", got, want)
			}
			if got := runHeapOutput(t, path, false); got != want {
				t.Errorf("output after a released program:\n%s\nwant:\n%s", got, want)
			}
		})
	}
	ast.ProgramMemoryRelease(astbuild.AST)
	if astbuild.AST.Memory != nil {
		t.Error("a released program keeps its memory")
	}
}
