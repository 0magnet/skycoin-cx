package bolt

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
