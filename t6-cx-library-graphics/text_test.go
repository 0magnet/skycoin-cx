//go:build cxfx

package graphics

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// textTestProgram returns a program whose memory holds a pointer at offset 0
// to a str object with text, and the value that reads it.
func textTestProgram(text string) (*ast.CXProgram, ast.CXValue) {
	memory := make([]byte, 64+len(text))
	heapOffset := valuetypes.Pointer(32)
	valuetypes.ObjectPayloadWrite(memory, heapOffset, valuetypes.ObjectSizeCompute([]byte(text)), []byte(text))
	valuetypes.PtrWrite(memory, 0, heapOffset)
	return &ast.CXProgram{Memory: memory}, ast.CXValue{Offset: 0}
}

func TestCxfxTextNulTerminated(t *testing.T) {
	prgrm, value := textTestProgram("projection")
	name := cxfxTextNulTerminated(ast.ProgramStringBytesRead(prgrm, &value))
	if got := string(cxfxState.CxfxTextBuffer[:len("projection")+1]); got != "projection\x00" || name != &cxfxState.CxfxTextBuffer[0] {
		t.Fatalf("the C string is %q", got)
	}
	long := make([]byte, 3*TextBufferSizeInitial)
	cxfxTextNulTerminated(long)
	if len(cxfxState.CxfxTextBuffer) < len(long)+1 || cxfxState.CxfxTextBuffer[len(long)] != 0 {
		t.Fatal("a long name did not grow the scratch")
	}
}

func TestCxfxTextNulTerminatedAllocations(t *testing.T) {
	prgrm, value := textTestProgram("projection")
	cxfxTextNulTerminated(ast.ProgramStringBytesRead(prgrm, &value))
	allocations := testing.AllocsPerRun(100, func() {
		cxfxTextNulTerminated(ast.ProgramStringBytesRead(prgrm, &value))
	})
	if allocations != 0 {
		t.Fatalf("a GL name allocates %v times", allocations)
	}
}

func TestCxfxTextViewAllocations(t *testing.T) {
	prgrm, value := textTestProgram("Score: 1234")
	var sink int
	allocations := testing.AllocsPerRun(100, func() {
		sink += len(stringpool.StringView(ast.ProgramStringBytesRead(prgrm, &value)))
	})
	if allocations != 0 {
		t.Fatalf("reading text allocates %v times", allocations)
	}
	if view := stringpool.StringView(ast.ProgramStringBytesRead(prgrm, &value)); view != "Score: 1234" {
		t.Fatalf("text is %q", view)
	}
}

func TestCxfxHandleTables(t *testing.T) {
	saved := cxfxState
	defer func() { cxfxState = saved }()
	cxfxState = State{}

	first := cxfxTexturePixelsAdd(TexturePixels{TexturePixelsValues: []float32{1}})
	second := cxfxTexturePixelsAdd(TexturePixels{TexturePixelsValues: []float32{2}})
	if first != 0 || second != 1 || cxfxTexturePixelsGet(second).TexturePixelsValues[0] != 2 {
		t.Fatalf("handles %d and %d", first, second)
	}
	*cxfxTexturePixelsGet(first) = TexturePixels{}
	if again := cxfxTexturePixelsAdd(TexturePixels{TexturePixelsValues: []float32{3}}); again != first {
		t.Fatalf("a freed slot was not reused, got %d", again)
	}
	if handle := cxfxCallbackI32I32Add(7); handle != 0 || cxfxCallbackI32I32Get(0) != 7 {
		t.Fatalf("callback handle %d", handle)
	}
}

func TestCxfxCallbackInputsAllocations(t *testing.T) {
	allocations := testing.AllocsPerRun(100, func() {
		cxfxCallbackInputsI32(0, 3)
		cxfxCallbackInputsI64(1, -1)
		cxfxCallbackInputsF64(2, 0.5)
	})
	if allocations != 0 {
		t.Fatalf("writing callback inputs allocates %v times", allocations)
	}
	if valuetypes.I32Read(cxfxState.CxfxCallbackInputs[0], 0) != 3 || valuetypes.I64Read(cxfxState.CxfxCallbackInputs[1], 0) != -1 || valuetypes.F64Read(cxfxState.CxfxCallbackInputs[2], 0) != 0.5 {
		t.Fatal("callback inputs read back wrong")
	}
}

func BenchmarkCxfxTextUniformName(b *testing.B) {
	prgrm, value := textTestProgram("projection")
	b.ReportAllocs()
	for b.Loop() {
		cxfxTextNulTerminated(ast.ProgramStringBytesRead(prgrm, &value))
	}
}
