package packagetree

import (
	"testing"

	sourceargs "github.com/skycoin/skycoin-cx/t0-cx-compiler-sourceargs"
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader"
)

func TestTree(t *testing.T) {
	const want = "main\n|--sys/os\n|--testimport2\n|  |--testimport1\n|  |  `--sys/gl\n|  `--testimport3\n|     `--testimport1\n|        `--sys/gl\n`--testimport1\n   `--sys/gl\n"
	_, sourceCodes, _, perr := sourceargs.Parse([]string{"test_tree"}, true)
	if perr != nil {
		t.Fatal(perr)
	}
	if err := packageloader.CXProgramLoad("TestTree", sourceCodes); err != nil {
		t.Fatal(err)
	}
	output, err := ImportTreeGet("TestTree")
	if err != nil {
		t.Fatal(err)
	}
	if output != want {
		t.Errorf("GetImportTree:\n got %q\nwant %q", output, want)
	}
}

func TestAddDependenciesCycle(t *testing.T) {
	importMap := map[string][]string{"a": {"b"}, "b": {"a"}}
	got := DependenciesRecursiveAdd(importMap, []string{"a"}, 0, map[int]bool{})
	want := "`--a\n   `--b\n      `--a (import cycle)\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
