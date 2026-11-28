package output

import (
	"bytes"
	"strconv"
	"sync"
	"testing"
)

// lockedBuffer is a bytes.Buffer the writing thread and the test share.
type lockedBuffer struct {
	bufferLock sync.Mutex
	buffer     bytes.Buffer
}

func (locked *lockedBuffer) Write(text []byte) (int, error) {
	locked.bufferLock.Lock()
	defer locked.bufferLock.Unlock()
	return locked.buffer.Write(text)
}

func (locked *lockedBuffer) String() string {
	locked.bufferLock.Lock()
	defer locked.bufferLock.Unlock()
	return locked.buffer.String()
}

// Every mode writes every byte in order, more than the ring holds included,
// and a flush returns only once the bytes are written.
func TestWriterModes(t *testing.T) {
	var want bytes.Buffer
	var lines [][]byte
	for index := 0; index < 20000; index++ {
		line := strconv.AppendInt(nil, int64(index), 10)
		line = append(line, '\n')
		lines = append(lines, line)
		want.Write(line)
	}
	for mode := ModeBatch; mode < ModeCount; mode++ {
		out := &lockedBuffer{}
		writer := WriterNew(out, mode)
		for index, line := range lines {
			writer.WriterHandOff(line)
			if index == 100 {
				writer.WriterFlush(nil)
				if got := out.String(); got != want.String()[:len(got)] || len(got) < 101*2 {
					t.Fatalf("mode %s: after a flush %d bytes are written", ModeNames[mode], len(got))
				}
			}
		}
		writer.WriterFlush([]byte("end\n"))
		if got := out.String(); got != want.String()+"end\n" {
			t.Errorf("mode %s: %d bytes written, want %d", ModeNames[mode], len(got), want.Len()+4)
		}
	}
}

func TestModeByName(t *testing.T) {
	for mode := ModeBatch; mode < ModeCount; mode++ {
		if found, ok := ModeByName(ModeNames[mode]); !ok || found != mode {
			t.Errorf("ModeByName(%q) = %d, %v", ModeNames[mode], found, ok)
		}
	}
	if _, ok := ModeByName("fast"); ok {
		t.Error("ModeByName(\"fast\") found a mode")
	}
}
