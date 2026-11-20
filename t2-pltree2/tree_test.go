package tree

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t1-plloader2"
)

func TestTree(t *testing.T) {
	const want = "main\n|--os\n|--testimport2\n|  |--testimport1\n|  |  `--gl\n|  `--testimport3\n|     `--testimport1\n|        `--gl\n`--testimport1\n   `--gl\n"
	_, sourceCodes, _, perr := loader.ParseArgsForCX([]string{"test_tree"}, true)
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
