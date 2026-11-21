//go:build cxfx && !mobile

package cxfx

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t0-util"
	"github.com/skycoin/skycoin-cx/t3-ast"
)

// textTestProgram returns a program whose memory holds a pointer at offset 0
// to a str object with text, and the value that reads it.
func textTestProgram(text string) (*ast.CXProgram, ast.CXValue) {
	memory := make([]byte, 64+len(text))
	heapOffset := types.Pointer(32)
	types.Write_obj_data(memory, heapOffset, types.Compute_obj_size([]byte(text)), []byte(text))
	types.Write_ptr(memory, 0, heapOffset)
	return &ast.CXProgram{Memory: memory}, ast.CXValue{Offset: 0}
}

func TestCxglCStringFindKeepsCache(t *testing.T) {
	prgrm, value := textTestProgram("projection")
	defer freeCString("projection")
	defer freeCString("csources")

	first := cxglCStringFind(ast.ProgramStringBytesRead(prgrm, &value))
	if cSources["projection"] != first {
		t.Fatal("a miss is not cached under the text")
	}
	if again := cxglCStringFind(ast.ProgramStringBytesRead(prgrm, &value)); again != first {
		t.Fatal("a hit returned a different C string")
	}

	stored := getCString("csources", "void main() {}")
	if found := cxglCStringFind([]byte("csources")); found != stored {
		t.Fatal("a key stored by gl.Strs was not found")
	}
}

func TestCxglCStringFindHitAllocations(t *testing.T) {
	prgrm, value := textTestProgram("projection")
	defer freeCString("projection")
	cxglCStringFind(ast.ProgramStringBytesRead(prgrm, &value))

	allocations := testing.AllocsPerRun(100, func() {
		cxglCStringFind(ast.ProgramStringBytesRead(prgrm, &value))
	})
	if allocations != 0 {
		t.Fatalf("a cached lookup allocates %v times", allocations)
	}
}

func TestCxfxTextViewAllocations(t *testing.T) {
	prgrm, value := textTestProgram("Score: 1234")
	var sink int
	allocations := testing.AllocsPerRun(100, func() {
		sink += len(util.ProgramTextStringView(ast.ProgramStringBytesRead(prgrm, &value)))
	})
	if allocations != 0 {
		t.Fatalf("reading text allocates %v times", allocations)
	}
	if view := util.ProgramTextStringView(ast.ProgramStringBytesRead(prgrm, &value)); view != "Score: 1234" {
		t.Fatalf("text is %q", view)
	}
}

func BenchmarkCxfxTextUniformName(b *testing.B) {
	prgrm, value := textTestProgram("projection")
	defer freeCString("projection")
	b.Run("Get_str", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			name := value.Get_str(prgrm)
			getCString(name, name)
		}
	})
	b.Run("View", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cxglCStringFind(ast.ProgramStringBytesRead(prgrm, &value))
		}
	})
}

func BenchmarkCxfxTextPrint(b *testing.B) {
	prgrm, value := textTestProgram("Score: 1234567890, lives: 3, level: 12")
	var sink int
	b.Run("Get_str", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sink += len(value.Get_str(prgrm))
		}
	})
	b.Run("View", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sink += len(util.ProgramTextStringView(ast.ProgramStringBytesRead(prgrm, &value)))
		}
	})
}
