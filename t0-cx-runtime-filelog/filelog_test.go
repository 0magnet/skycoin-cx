package filelog

import (
	"os"
	"testing"
)

func TestFilelogCreateOpenStat(t *testing.T) {
	t.Chdir(t.TempDir())
	file, err := Create("a.txt")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := file.WriteString("hi"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	file, err = Open("a.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if fileStat, err := Stat("a.txt"); err != nil || fileStat.Size() != 2 {
		t.Errorf("Stat: %v, %v", fileStat, err)
	}
	if _, err := os.Stat("missing.txt"); err == nil {
		t.Error("missing.txt exists")
	}
}
