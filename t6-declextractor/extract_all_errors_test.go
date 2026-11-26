package declaration_extractor_test

import (
	"testing"
	"time"

	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader" // thin: only File; no depth gain
	"github.com/skycoin/skycoin-cx/t6-declextractor"
)

// A bad package line makes every extractor of the file fail. See #123.
func TestDeclarationExtractor_ExtractAllDeclarationsManyErrors(t *testing.T) {
	files := []*packageloader.File{
		{FileName: "a.cx", Content: []byte("package\n")},
		{FileName: "b.cx", Content: []byte("package\n")},
	}

	done := make(chan error, 1)
	go func() {
		_, _, _, _, _, _, err := declaration_extractor.AllDeclarationsExtract(files)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("want an error for a bad package line, got nil")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("ExtractAllDeclarations did not return")
	}
}
