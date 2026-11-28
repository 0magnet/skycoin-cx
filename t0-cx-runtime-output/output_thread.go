//go:build !tinygo && !js && !wasip1

package output

import (
	"io"
	"sync/atomic"
)

// ThreadAvailable says that this build runs a second thread for ModeThread.
// TinyGo, js and wasip1 builds run one, and ModeThread writes as ModeInstant.
const ThreadAvailable = true

// ringSize is the size of the ring of ModeThread, a power of two.
const ringSize = 1 << 16

// ring is the shared buffer of ModeThread. The program thread copies into it
// and advances ringWritten with an atomic store, and the writing thread
// writes up to it and advances ringDrained. A push waits only when the ring
// is full.
type ring struct {
	ringBuffer  [ringSize]byte
	ringWritten atomic.Uint64
	ringDrained atomic.Uint64
	ringWake    chan struct{}
	ringSpace   chan struct{}
	ringOut     io.Writer
}

func ringStart(out io.Writer) *ring {
	r := &ring{
		ringWake:  make(chan struct{}, 1),
		ringSpace: make(chan struct{}, 1),
		ringOut:   out,
	}
	go r.ringRun()
	return r
}

// ringSignal wakes the thread that waits on channel, if any, and never
// blocks.
func ringSignal(channel chan struct{}) {
	select {
	case channel <- struct{}{}:
	default:
	}
}

func (r *ring) ringPush(bytes []byte) {
	written := r.ringWritten.Load()
	for len(bytes) > 0 {
		free := ringSize - int(written-r.ringDrained.Load())
		if free == 0 {
			ringSignal(r.ringWake)
			<-r.ringSpace
			continue
		}
		count := min(free, len(bytes))
		start := int(written % ringSize)
		first := copy(r.ringBuffer[start:], bytes[:count])
		copy(r.ringBuffer[:], bytes[first:count])
		written += uint64(count)
		r.ringWritten.Store(written)
		bytes = bytes[count:]
	}
	ringSignal(r.ringWake)
}

// ringFlush returns once every byte pushed so far is written.
func (r *ring) ringFlush() {
	target := r.ringWritten.Load()
	for r.ringDrained.Load() < target {
		ringSignal(r.ringWake)
		<-r.ringSpace
	}
}

// ringRun is the writing thread. It lives as long as the process, which
// writes every byte with a flush before it ends.
func (r *ring) ringRun() {
	for range r.ringWake {
		r.ringDrain()
	}
}

// ringDrain writes everything pushed so far, one contiguous part at a
// time.
func (r *ring) ringDrain() {
	for {
		drained := r.ringDrained.Load()
		written := r.ringWritten.Load()
		if drained == written {
			return
		}
		start := drained % ringSize
		count := min(written-drained, ringSize-start)
		outWrite(r.ringOut, r.ringBuffer[start:start+count])
		r.ringDrained.Store(drained + count)
		ringSignal(r.ringSpace)
	}
}
