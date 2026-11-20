package graph

import (
	"fmt"
	"os"
	"testing"

	"github.com/skycoin/skycoin-cx/t0-plbolt"
)

func TestMain(m *testing.M) {
	if err := bolt.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
