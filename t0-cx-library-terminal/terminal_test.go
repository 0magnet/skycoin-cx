package terminal

import (
	"strings"
	"testing"
)

func TestArgRead(t *testing.T) {
	ArgsSet([]string{"++1000", "0"})
	defer ArgsSet(nil)
	if count := ArgCount(); count != 2 {
		t.Fatalf("ArgCount() = %d, want 2", count)
	}
	var buffer [3]byte
	for _, want := range []struct {
		index  int32
		copied int32
		text   string
	}{{0, 3, "100"}, {1, 1, "0"}, {2, -1, ""}, {-1, -1, ""}} {
		copied := ArgRead(want.index, buffer[:])
		if copied != want.copied || (copied > 0 && string(buffer[:copied]) != want.text) {
			t.Errorf("ArgRead(%d) = %d %q, want %d %q", want.index, copied, buffer[:max(copied, 0)], want.copied, want.text)
		}
	}
}

func TestInputLineRead(t *testing.T) {
	long := strings.Repeat("x", 5000)
	InputSet(strings.NewReader("one\r\n\n" + long + "\nlast"))
	defer InputSet(strings.NewReader(""))
	var buffer [8]byte
	for _, want := range []struct {
		copied int32
		text   string
	}{{3, "one"}, {0, ""}, {8, "xxxxxxxx"}, {4, "last"}, {-1, ""}} {
		copied := InputLineRead(buffer[:])
		if copied != want.copied || (copied > 0 && string(buffer[:copied]) != want.text) {
			t.Errorf("InputLineRead() = %d %q, want %d %q", copied, buffer[:max(copied, 0)], want.copied, want.text)
		}
	}
}

// InputLineRead makes no Go allocation once the reader exists, as the vm
// runs it.
func TestInputLineReadAllocs(t *testing.T) {
	InputSet(strings.NewReader(strings.Repeat("12345\n", 1000)))
	defer InputSet(strings.NewReader(""))
	var buffer [16]byte
	if allocs := testing.AllocsPerRun(100, func() { InputLineRead(buffer[:]) }); allocs != 0 {
		t.Errorf("InputLineRead allocates %v times a call, want 0", allocs)
	}
}
