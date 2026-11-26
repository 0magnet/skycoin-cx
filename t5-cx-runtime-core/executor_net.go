//go:build !cxcore || cxcorelibrarynet

package core

import (
	"github.com/skycoin/skycoin-cx/t1-cx-library-networkbatch"
)

// executorNetRunStart closes what an earlier run left open and starts the
// net library in netbatch.NetRunMode.
func executorNetRunStart() error {
	return networkbatch.NetRunStart(networkbatch.NetRunMode, networkbatch.NetRunRecordPath)
}
