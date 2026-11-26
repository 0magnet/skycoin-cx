package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// inlineImports are the packages an inline call may name, by the name it
// uses for them.
var inlineImports = map[string]string{
	"gl":   "github.com/go-gl/gl/v3.2-compatibility/gl",
	"glfw": "github.com/go-gl/glfw/v3.3/glfw",
}

// inlineReadTypes are the operand types an inline call reads, with the type
// part of the name of their types read and write functions, I32Read.
var inlineReadTypes = map[string]string{"bool": "Bool", "i32": "I32", "ui32": "UI32", "f32": "F32", "f64": "F64"}

// goFunctionInline reports whether the function part of a GoFunction is a call the
// switch makes inline rather than the name of a handler.
func goFunctionInline(function string) bool {
	return strings.Contains(function, "(")
}

// inlineCase returns the body of the case of row, whose GoFunction is an inline
// call: each %N becomes the read of input N at its offset, and the value of
// the call is written to the one output when the row has one.
func inlineCase(row opcodetable.OpcodeDefinition, call string) (string, error) {
	inputs := strings.Fields(row.Inputs)
	outputs := strings.Fields(row.Outputs)
	var b strings.Builder
	for i := 0; i < len(call); i++ {
		if call[i] != '%' {
			b.WriteByte(call[i])
			continue
		}
		j := i + 1
		for j < len(call) && call[j] >= '0' && call[j] <= '9' {
			j++
		}
		var index int
		if _, err := fmt.Sscan(call[i+1:j], &index); err != nil || index >= len(inputs) || inlineReadTypes[inputs[index]] == "" {
			return "", fmt.Errorf("definition %q: inline operand %q is not an atomic input", row.Name, call[i:j])
		}
		fmt.Fprintf(&b, "valuetypes.%sRead(prgrm.Memory, inputs[%d].Offset)", inlineReadTypes[inputs[index]], index)
		i = j - 1
	}
	switch {
	case len(outputs) == 0:
		return b.String(), nil
	case len(outputs) == 1 && inlineReadTypes[outputs[0]] != "":
		return fmt.Sprintf("valuetypes.%sWrite(prgrm.Memory, outputs[0].Offset, %s)", inlineReadTypes[outputs[0]], b.String()), nil
	}
	return "", fmt.Errorf("definition %q: an inline call writes at most one atomic output", row.Name)
}

// inlinePackage matches a package an inline call names.
var inlinePackage = regexp.MustCompile(`\b([a-z]+)\.`)

// inlinePackages adds to used the import names of inlineImports that call
// names.
func inlinePackages(call string, used map[string]bool) {
	for _, match := range inlinePackage.FindAllStringSubmatch(call, -1) {
		if _, known := inlineImports[match[1]]; known {
			used[match[1]] = true
		}
	}
}
