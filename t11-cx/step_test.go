package main

import (
	"io"
	"os"
	"testing"

	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"

	"github.com/skycoin/skycoin-cx/t4-cx-runtime-interpreter"
	"github.com/skycoin/skycoin-cx/t9-cx-compiler-parsing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		b, _ := io.ReadAll(r)
		done <- b
	}()
	defer func() { os.Stdout = stdout }()
	fn()
	w.Close()
	return string(<-done)
}

// RunCompiled with an op limit steps out of nested calls (#81), stops at
// the limit, and neither traces to stdout nor grows the program (#110).
func TestRunCompiledMaxOps(t *testing.T) {
	initProgram()
	src := "package main\nimport \"sys/string\"\nfunc f() {\n\tstring.PrintLine(\"in f\")\n}\nfunc main() {\n\tf()\n\tf()\n\tfor true {\n\t}\n}\n"
	if err := parsing.SourceStringsParse([]string{src}, []string{"step.cx"}); err != nil {
		t.Fatal(err)
	}
	if err := parsing.SysInitFunctionAdd(astbuild.AST); err != nil {
		t.Fatal(err)
	}
	astbuild.AST.CxIrState.StringPool.StringPoolFreeze()
	before := len(astbuild.AST.CXAtomicOps)
	out := captureStdout(t, func() {
		if err := interpreter.RunCompiled(astbuild.AST, 1000, nil); err != nil {
			t.Fatal(err)
		}
	})
	if out != "in f\nin f\n" {
		t.Errorf("stdout %q, want %q", out, "in f\nin f\n")
	}
	if after := len(astbuild.AST.CXAtomicOps); after != before {
		t.Errorf("CXAtomicOps grew from %d to %d", before, after)
	}
}
