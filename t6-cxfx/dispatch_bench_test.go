//go:build cxfx

package cxfx

import (
	"runtime"
	"testing"

	"github.com/go-gl/gl/v3.2-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// benchWindow is the hidden window whose GL context the benchmarks use.
var benchWindow *glfw.Window

// benchContext makes the context of a hidden window current on the locked
// thread of b, creating the window once, and skips b without a display.
func benchContext(b *testing.B) {
	runtime.LockOSThread()
	b.Cleanup(func() {
		glfw.DetachCurrentContext()
		runtime.UnlockOSThread()
	})
	if benchWindow == nil {
		if err := glfw.Init(); err != nil {
			b.Skip("no display:", err)
		}
		glfw.WindowHint(glfw.Visible, glfw.False)
		window, err := glfw.CreateWindow(64, 64, "bench", nil, nil)
		if err != nil {
			b.Skip("no GL window:", err)
		}
		benchWindow = window
		benchWindow.MakeContextCurrent()
		if err := gl.Init(); err != nil {
			b.Skip("no GL:", err)
		}
	}
	benchWindow.MakeContextCurrent()
}

// benchCode returns the opcode named name.
func benchCode(b *testing.B, name string) int {
	for _, entry := range opcodetable.NativeNameTable {
		if entry.NativeNameText == name {
			return entry.NativeNameCode
		}
	}
	b.Fatalf("no opcode %q", name)
	return 0
}

// BenchmarkCxfxOpcodeDispatch times CxfxOpcodeDispatch alone on operands
// already in program memory, for simple GL and GLFW calls (#432).
func BenchmarkCxfxOpcodeDispatch(b *testing.B) {
	benches := []struct {
		name   string
		inputs []valuetypes.Pointer
		output bool
	}{
		{"gl.Color3f", []valuetypes.Pointer{0, 4, 8}, false},
		{"gl.Normal3f", []valuetypes.Pointer{0, 4, 8}, false},
		{"gl.TexCoord2f", []valuetypes.Pointer{0, 4}, false},
		{"gl.GetError", nil, true},
		{"glfw.GetTime", nil, true},
		{"glfw.ShouldClose", []valuetypes.Pointer{16}, true},
	}
	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			benchContext(b)
			handle := cxfxWindowAdd(benchWindow)
			defer func() { cxfxState.CxfxWindowTable[handle] = nil }()
			prgrm := &ast.CXProgram{Memory: make([]byte, 64)}
			valuetypes.F32Write(prgrm.Memory, 0, 0.25)
			valuetypes.F32Write(prgrm.Memory, 4, 0.5)
			valuetypes.F32Write(prgrm.Memory, 8, 0.75)
			valuetypes.I32Write(prgrm.Memory, 16, handle)
			inputs := make([]ast.CXValue, len(bench.inputs))
			for i, offset := range bench.inputs {
				inputs[i].Offset = offset
			}
			outputs := []ast.CXValue{{Offset: 32}}
			if !bench.output {
				outputs = nil
			}
			code := benchCode(b, bench.name)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				CxfxOpcodeDispatch(prgrm, code, inputs, outputs)
			}
		})
	}
}
