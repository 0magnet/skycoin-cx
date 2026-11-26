package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/skycoin/skycoin-cx/t2-cx-core-environment"
)

// CXPathSet fell back to os/user only and panicked where user.Current
// fails, as on js/wasm.
func TestCXPathSetWithoutUser(t *testing.T) {
	saved, savedBin := cxPathHomeDirectory, environment.BINPATH
	defer func() { cxPathHomeDirectory, environment.BINPATH = saved, savedBin }()
	t.Setenv("CXPATH", "")

	cxPathHomeDirectory = func() (string, error) { return "/home/tester", nil }
	if err := CXPathSet(cxCmdFlags{}); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("/home/tester", "cx", "bin"); environment.BINPATH != want {
		t.Errorf("BINPATH = %q, want %q", environment.BINPATH, want)
	}

	cxPathHomeDirectory = func() (string, error) { return "", errors.New("no home") }
	if err := CXPathSet(cxCmdFlags{}); err == nil {
		t.Error("CXPathSet without CXPATH and home returned no error")
	}

	t.Setenv("CXPATH", "/opt/cx")
	if err := CXPathSet(cxCmdFlags{}); err != nil || environment.BINPATH != filepath.Join("/opt/cx", "bin") {
		t.Errorf("with CXPATH: %v, BINPATH %q", err, environment.BINPATH)
	}
}
