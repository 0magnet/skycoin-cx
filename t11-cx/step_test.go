package main

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t4-actions"

	"github.com/skycoin/skycoin-cx/t4-execute"
	cxparsing "github.com/skycoin/skycoin-cx/t9-cxparsing"
)

// The search for the next line used an unsigned counter with c >= 0, so it
// ran past the bottom of the call stack. See #81.
func TestRunCompiledStepsOutOfNestedCall(t *testing.T) {
	initProgram()
	src := "package main\nfunc f() {\n\tstr.print(\"in f\")\n}\nfunc main() {\n\tf()\n}\n"
	if err := cxparsing.ParseSourceStrings([]string{src}, []string{"step.cx"}); err != nil {
		t.Fatal(err)
	}
	if err := cxparsing.AddInitFunction(actions.AST); err != nil {
		t.Fatal(err)
	}
	if err := execute.RunCompiled(actions.AST, 1000, nil); err != nil {
		t.Fatal(err)
	}
}
