package stringpool

import (
	"testing"
	"unsafe"
)

func TestProgramTextStringViewShares(t *testing.T) {
	text := []byte("uniform mat4 projection;")
	view := StringView(text)
	if view != "uniform mat4 projection;" {
		t.Fatalf("view is %q", view)
	}
	if unsafe.StringData(view) != &text[0] {
		t.Fatal("view copied the bytes")
	}
}

func TestProgramTextStringViewEmpty(t *testing.T) {
	if view := StringView(nil); view != "" {
		t.Fatalf("view of nil is %q", view)
	}
	if view := StringView([]byte{}); view != "" {
		t.Fatalf("view of empty is %q", view)
	}
}

func TestProgramTextStringViewAllocations(t *testing.T) {
	text := []byte("void main() { gl_FragColor = vec4(1.0); }")
	var sink int
	allocations := testing.AllocsPerRun(100, func() {
		sink += len(StringView(text))
	})
	if allocations != 0 {
		t.Fatalf("StringView allocates %v times", allocations)
	}
}
