package main

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

func TestInlineCase(t *testing.T) {
	for _, test := range []struct {
		row  opcodetable.OpcodeDefinition
		call string
		want string
	}{
		{opcodetable.OpcodeDefinition{Name: "gl.Clear", Inputs: "i32"}, "gl.Clear(uint32(%0))",
			"gl.Clear(uint32(valuetypes.I32Read(prgrm.Memory, inputs[0].Offset)))"},
		{opcodetable.OpcodeDefinition{Name: "gl.End"}, "gl.End()", "gl.End()"},
		{opcodetable.OpcodeDefinition{Name: "glfw.ShouldClose", Inputs: "i32", Outputs: "bool"}, "cxfxWindowGet(%0).ShouldClose()",
			"valuetypes.BoolWrite(prgrm.Memory, outputs[0].Offset, cxfxWindowGet(valuetypes.I32Read(prgrm.Memory, inputs[0].Offset)).ShouldClose())"},
		{opcodetable.OpcodeDefinition{Name: "gl.Uniform1f", Inputs: "i32 f32"}, "gl.Uniform1f(%0, %1)",
			"gl.Uniform1f(valuetypes.I32Read(prgrm.Memory, inputs[0].Offset), valuetypes.F32Read(prgrm.Memory, inputs[1].Offset))"},
	} {
		got, err := inlineCase(test.row, test.call)
		if err != nil || got != test.want {
			t.Errorf("%s: got %q, %v", test.row.Name, got, err)
		}
	}
	for _, test := range []struct {
		row  opcodetable.OpcodeDefinition
		call string
	}{
		{opcodetable.OpcodeDefinition{Name: "past the inputs", Inputs: "i32"}, "f(%1)"},
		{opcodetable.OpcodeDefinition{Name: "a str input", Inputs: "str"}, "f(%0)"},
		{opcodetable.OpcodeDefinition{Name: "two outputs", Outputs: "i32 i32"}, "f()"},
		{opcodetable.OpcodeDefinition{Name: "a slice output", Outputs: "[]i32"}, "f()"},
	} {
		if _, err := inlineCase(test.row, test.call); err == nil {
			t.Errorf("%s: no error", test.row.Name)
		}
	}
	used := map[string]bool{}
	inlinePackages("int32(cxfxWindowGet(%0).GetKey(glfw.Key(%1)))", used)
	if !used["glfw"] || used["gl"] || len(used) != 1 {
		t.Errorf("packages %v", used)
	}
}
