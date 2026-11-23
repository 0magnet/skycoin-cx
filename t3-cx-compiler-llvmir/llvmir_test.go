package llvmir

import (
	"strings"
	"testing"

	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
)

// TestOpcodeTable checks that the families of OpInfos end where the image
// opcodes end, so every scalar opcode has its type and operation.
func TestOpcodeTable(t *testing.T) {
	if opcodeTableEnd != ir.OpcodeCount {
		t.Fatalf("table ends at %d, want %d", opcodeTableEnd, ir.OpcodeCount)
	}
	if got := OpName(ir.OpcodeI32Add); got != "i32.add" {
		t.Fatalf("OpName(OpcodeI32Add) = %q", got)
	}
	if got := OpName(ir.OpcodeF64ToF32); got != "f64.f32" {
		t.Fatalf("OpName(OpcodeF64ToF32) = %q", got)
	}
}

// TestTemplates expands every template and lists the opcodes left to the
// runtime. Run with -v to print the opcode to template table.
func TestTemplates(t *testing.T) {
	direct, runtime := 0, 0
	for op := ir.OpcodeBoolNot; op < ir.OpcodeCount; op++ {
		tmpl, found := Template(op)
		if !found {
			runtime++
			t.Logf("%-12s runtime %s", OpName(op), FallbackName(op))
			continue
		}
		direct++
		e := &emitter{decls: map[string]string{}, strs: map[string]string{}, image: &ir.Image{ImageFuncs: []ir.Func{{}}}}
		e.expand(tmpl, OpInfos[op], "%a", "%b", "%r", 0)
		text := e.body.String()
		if strings.ContainsAny(text, "{}") {
			t.Fatalf("%s: placeholder left in %q", OpName(op), text)
		}
		t.Logf("%-12s direct  %s", OpName(op), strings.ReplaceAll(strings.TrimSpace(tmpl), "\n", " ; "))
	}
	t.Logf("scalar opcodes: %d direct, %d runtime", direct, runtime)
}
