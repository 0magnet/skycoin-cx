// Package output writes the output of a vm run to the terminal in one of
// the output modes of decision D142. The machine appends every print to its
// own buffer, and a Writer takes that buffer at each hand off and flush. It
// uses only the standard library.
package output

import "io"

// Mode is how the prints of a run reach the terminal, chosen once before
// the run.
type Mode uint8

const (
	// ModeBatch writes the output when the buffer fills, at an explicit
	// flush and at the end of the run. The program decides when else.
	ModeBatch Mode = iota
	// ModeInstant writes every print as it happens.
	ModeInstant
	// ModeThread queues every print in a ring that a second thread writes,
	// so a print does not wait for the terminal.
	ModeThread
	// ModeCount is one past the last mode.
	ModeCount
)

// ModeNames are the names --output-mode takes, by Mode.
var ModeNames = [ModeCount]string{"batch", "instant", "thread"}

// ModeByName returns the mode named name.
func ModeByName(name string) (Mode, bool) {
	for mode, modeName := range ModeNames {
		if modeName == name {
			return Mode(mode), true
		}
	}
	return ModeBatch, false
}

// Writer writes the output of one run. Its hand off and flush are bound to
// the mode when it is made, so no call tests the mode.
type Writer struct {
	writerOut     io.Writer
	writerRing    *ring
	writerHandOff func(writer *Writer, bytes []byte)
	writerFlush   func(writer *Writer, bytes []byte)
}

// WriterNew returns a writer of out in mode. ModeThread starts the writing
// thread, and a build with one thread writes as ModeInstant instead, see
// ThreadAvailable.
func WriterNew(out io.Writer, mode Mode) *Writer {
	writer := &Writer{writerOut: out, writerHandOff: writerWrite, writerFlush: writerWrite}
	if mode == ModeThread && ThreadAvailable {
		writer.writerRing = ringStart(out)
		writer.writerHandOff = writerQueue
		writer.writerFlush = writerQueueFlush
	}
	return writer
}

// WriterHandOff takes the bytes of the prints since the last hand off.
func (writer *Writer) WriterHandOff(bytes []byte) {
	writer.writerHandOff(writer, bytes)
}

// WriterFlush takes the remaining bytes and returns once every byte taken
// so far is written.
func (writer *Writer) WriterFlush(bytes []byte) {
	writer.writerFlush(writer, bytes)
}

// writerWrite writes bytes now.
func writerWrite(writer *Writer, bytes []byte) {
	outWrite(writer.writerOut, bytes)
}

// outWrite writes bytes to out. A failed write is dropped, as cx drops a
// failed write to stdout.
func outWrite(out io.Writer, bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	if _, err := out.Write(bytes); err != nil {
		return
	}
}

func writerQueue(writer *Writer, bytes []byte) {
	writer.writerRing.ringPush(bytes)
}

func writerQueueFlush(writer *Writer, bytes []byte) {
	writer.writerRing.ringPush(bytes)
	writer.writerRing.ringFlush()
}
