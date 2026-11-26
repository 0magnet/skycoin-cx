package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	core "github.com/skycoin/skycoin-cx/t5-cx-runtime-core"
)

// A core with no program appended, as on js/wasm where os.Executable fails,
// takes the program as cx-core --program FILE.
func TestCoreProgramLoadFile(t *testing.T) {
	directory := t.TempDir()
	built := filepath.Join(directory, "built")
	if err := os.WriteFile(built, core.TrailerAppend([]byte("core"), []byte("program")), 0o644); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(directory, "plain")
	if err := os.WriteFile(plain, []byte("no trailer"), 0o644); err != nil {
		t.Fatal(err)
	}

	program, args, problem := coreProgramLoad([]string{coreProgramFileFlag, built, "a", "b"})
	if problem != "" || string(program) != "program" || strings.Join(args, " ") != "a b" {
		t.Errorf("built: %q %q %q", program, args, problem)
	}
	if _, _, problem = coreProgramLoad([]string{coreProgramFileFlag, plain}); !strings.Contains(problem, "holds no CX program") {
		t.Errorf("plain file: problem %q", problem)
	}
	if _, _, problem = coreProgramLoad([]string{"a"}); !strings.Contains(problem, coreProgramFileFlag) {
		t.Errorf("no program: problem %q does not name %s", problem, coreProgramFileFlag)
	}
}
