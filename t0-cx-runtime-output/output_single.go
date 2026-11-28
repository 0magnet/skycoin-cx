//go:build tinygo || js || wasip1

package output

import "io"

// ThreadAvailable says that this build runs one thread, so ModeThread writes
// as ModeInstant.
const ThreadAvailable = false

type ring struct{}

func ringStart(out io.Writer) *ring { return nil }

func (r *ring) ringPush(bytes []byte) {}

func (r *ring) ringFlush() {}
