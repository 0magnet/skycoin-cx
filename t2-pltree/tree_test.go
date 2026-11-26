package tree

import (
	"testing"

	sourceargs "github.com/skycoin/skycoin-cx/t0-cx-compiler-sourceargs"
	loader "github.com/skycoin/skycoin-cx/t1-plloader"
)

func TestTree(t *testing.T) {
	const want = "main\n|--sys/os\n|--testimport2\n|  |--testimport1\n|  |  `--sys/gl\n|  `--testimport3\n|     `--testimport1\n|        `--sys/gl\n`--testimport1\n   `--sys/gl\n"
	_, sourceCodes, _, perr := sourceargs.Parse([]string{"test_tree"}, true)
	if perr != nil {
		t.Fatal(perr)
	}
	if err := loader.LoadCXProgram("TestTree", sourceCodes); err != nil {
		t.Fatal(err)
	}
	output, err := GetImportTree("TestTree")
	if err != nil {
		t.Fatal(err)
	}
	if output != want {
		t.Errorf("GetImportTree:\n got %q\nwant %q", output, want)
	}
}

func TestAddDependenciesCycle(t *testing.T) {
	importMap := map[string][]string{"a": {"b"}, "b": {"a"}}
	got := AddDependenciesRecur(importMap, []string{"a"}, 0, map[int]bool{})
	want := "`--a\n   `--b\n      `--a (import cycle)\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
