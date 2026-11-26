package main

import (
	"errors"
	"path/filepath"
	"testing"

	globals "github.com/skycoin/skycoin-cx/t2-globals"
)

// CXPathSet fell back to os/user only and panicked where user.Current
// fails, as on js/wasm.
func TestCXPathSetWithoutUser(t *testing.T) {
	saved, savedBin := cxPathHomeDirectory, globals.BINPATH
	defer func() { cxPathHomeDirectory, globals.BINPATH = saved, savedBin }()
	t.Setenv("CXPATH", "")

	cxPathHomeDirectory = func() (string, error) { return "/home/tester", nil }
	if err := CXPathSet(cxCmdFlags{}); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/home/tester", "cx", "bin"); globals.BINPATH != want {
		t.Errorf("BINPATH = %q, want %q", globals.BINPATH, want)
	}

	cxPathHomeDirectory = func() (string, error) { return "", errors.New("no home") }
	if err := CXPathSet(cxCmdFlags{}); err == nil {
		t.Error("CXPathSet without CXPATH and home returned no error")
	}

	t.Setenv("CXPATH", "/opt/cx")
	if err := CXPathSet(cxCmdFlags{}); err != nil || globals.BINPATH != filepath.Join("/opt/cx", "bin") {
		t.Errorf("with CXPATH: %v, BINPATH %q", err, globals.BINPATH)
	}
}
