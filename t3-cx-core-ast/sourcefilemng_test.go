package ast

import (
	"testing"

	objmodel "github.com/skycoin/skycoin-cx/t1-cx-core-objmodel"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

func TestSourceFileMng(t *testing.T) {
	var state CxIrState
	state.CxIrStateInit()
	pool := &state.StringPool
	manager := &state.SourceFileMng
	mainPath := pool.StringPoolIntern([]byte("main.cx"))
	otherPath := pool.StringPoolIntern([]byte("lib/other.cx"))
	mainFile := manager.SourceFileIntern(mainPath)
	otherFile := manager.SourceFileIntern(otherPath)
	if mainFile == SourceFileGidNil || otherFile == SourceFileGidNil || mainFile == otherFile {
		t.Fatalf("Gids %#x and %#x, want two distinct non-nil Gids", mainFile, otherFile)
	}
	if objmodel.GidTypeTag(int64(mainFile)) != SourceFileGidTypeTag || objmodel.GidGeneration(int64(mainFile)) != objmodel.GidGenerationFirst {
		t.Errorf("Gid %#x lacks the SourceFile tag or the first generation", mainFile)
	}
	if again := manager.SourceFileIntern(pool.StringPoolIntern([]byte("main.cx"))); again != mainFile {
		t.Errorf("second intern of main.cx = %#x, want %#x", again, mainFile)
	}
	if manager.SourceFileLen() != 2 {
		t.Errorf("Len = %d, want 2", manager.SourceFileLen())
	}
	if manager.SourceFilePath(otherFile) != otherPath {
		t.Errorf("path of %#x = %#x, want %#x", otherFile, manager.SourceFilePath(otherFile), otherPath)
	}
	if manager.SourceFileIntern(stringpool.StringGidNil) != SourceFileGidNil || manager.SourceFilePath(SourceFileGidNil) != stringpool.StringGidNil {
		t.Error("the empty path and SourceFileGidNil must map to each other")
	}
	if objmodel.GidTypeTagNameTable[SourceFileGidTypeTag] != "SourceFile" {
		t.Errorf("tag %#x is named %q, want SourceFile", SourceFileGidTypeTag, objmodel.GidTypeTagNameTable[SourceFileGidTypeTag])
	}
}

func TestSourceFileMngAssertChild(t *testing.T) {
	if objmodelTestCaseName() != "path of an unknown file" {
		t.Skip("subprocess only")
	}
	var state CxIrState
	state.CxIrStateInit()
	state.SourceFileMng.SourceFilePath(SourceFileGid(objmodel.GidCompose(SourceFileGidTypeTag, objmodel.GidGenerationFirst, 3)))
	t.Fatal("returned after a failed assert")
}

func TestSourceFileMngAssert(t *testing.T) {
	objmodelTestAssertSubprocess(t, "TestSourceFileMngAssertChild", "path of an unknown file", "SourceFile Gid")
}
