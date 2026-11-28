package main

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-lower"
)

// TestBuildCallGraphReport checks the warning on a call that closes a cycle
// and the depth and stack line of a program without one (decision D144).
func TestBuildCallGraphReport(t *testing.T) {
	for _, c := range []struct {
		name   string
		source string
		quiet  bool
		want   string
	}{
		{"recursive.cx", `package main

func down(n i32) (out i32) {
	if n > 0 {
		out = up(n - 1)
	}
}

func up(n i32) (out i32) {
	out = down(n)
}

func main() {
	i32.print(down(3))
}
`, false, "cx build: warning: recursive.cx:10: recursive call of down from up, the call depth of this program is not known at build time\n"},
		{"chain.cx", `package main

func leaf(n i32) (out i32) {
	out = n + 1
}

func middle(n i32) (out i32) {
	out = leaf(n) * 2
}

func main() {
	i32.print(middle(3))
	i32.print(leaf(1))
}
`, false, "cx build: call depth 2 (main -> middle -> leaf), stack 28 of 1048576 bytes\n"},
		{"chain.cx", `package main

func main() {
	i32.print(1)
}
`, true, ""},
	} {
		initProgram()
		compileRunSources(t, []string{c.name}, []string{c.source})
		image, failures := lower.Program(astbuild.AST)
		if len(failures) > 0 {
			t.Fatalf("%s does not lower: %+v", c.name, failures)
		}
		if got := buildCallGraphReport(&image, c.quiet); got != c.want {
			t.Errorf("%s reports %q, want %q", c.name, got, c.want)
		}
	}
}
