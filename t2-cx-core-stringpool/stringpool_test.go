package stringpool

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	assert "github.com/skycoin/skycoin-cx/t0-cx-core-assert"
	objmodel "github.com/skycoin/skycoin-cx/t1-cx-core-objmodel"
)

func stringPoolTestNew() *StringPool {
	pool := &StringPool{}
	pool.StringPoolInit(4, 16)
	return pool
}

func TestStringPoolIntern(t *testing.T) {
	pool := stringPoolTestNew()
	main := pool.StringPoolIntern([]byte("main"))
	other := pool.StringPoolIntern([]byte("other"))
	if main == StringGidNil || other == StringGidNil || main == other {
		t.Fatalf("Gids %#x and %#x, want two distinct non-nil Gids", main, other)
	}
	if objmodel.GidTypeTag(int64(main)) != StringGidTypeTag || objmodel.GidGeneration(int64(main)) != objmodel.GidGenerationFirst {
		t.Errorf("Gid %#x lacks the String tag or the first generation", main)
	}
	if again := pool.StringPoolIntern([]byte("main")); !StringPoolEqual(again, main) {
		t.Errorf("second intern of main = %#x, want %#x", again, main)
	}
	if got := pool.StringPoolIntern(nil); got != StringGidNil {
		t.Errorf("intern of the empty string = %#x, want StringGidNil", got)
	}
	if pool.StringPoolLen() != 2 {
		t.Errorf("Len = %d, want 2", pool.StringPoolLen())
	}
	// Grow the arena and the slots past their Init capacity.
	gids := make([]StringGid, 100)
	for i := range gids {
		gids[i] = pool.StringPoolIntern([]byte(fmt.Sprintf("name%d", i)))
	}
	pool.StringPoolFreeze()
	if pool.StringPoolIndex != nil {
		t.Error("the index survives StringPoolFreeze")
	}
	for i, gid := range gids {
		if got := pool.StringPoolBytes(gid); string(got) != fmt.Sprintf("name%d", i) {
			t.Errorf("Bytes(%#x) = %q, want name%d", gid, got, i)
		}
	}
	if got := pool.StringPoolBytes(main); !bytes.Equal(got, []byte("main")) || cap(got) != len(got) {
		t.Errorf("Bytes(main) = %q with cap %d", got, cap(got))
	}
	if got := pool.StringPoolBytes(StringGidNil); len(got) != 0 {
		t.Errorf("Bytes(StringGidNil) = %q, want empty", got)
	}
}

func TestStringPoolFind(t *testing.T) {
	pool := stringPoolTestNew()
	main := pool.StringPoolIntern([]byte("main"))
	other := pool.StringPoolIntern([]byte("other"))
	for _, frozen := range []bool{false, true} {
		if frozen {
			pool.StringPoolFreeze()
		}
		if got, found := pool.StringPoolFind([]byte("other")); !found || got != other {
			t.Errorf("frozen %v: Find(other) = %#x, %v, want %#x, true", frozen, got, found, other)
		}
		if got, found := pool.StringPoolFind([]byte("main")); !found || got != main {
			t.Errorf("frozen %v: Find(main) = %#x, %v, want %#x, true", frozen, got, found, main)
		}
		if _, found := pool.StringPoolFind([]byte("absent")); found {
			t.Errorf("frozen %v: Find(absent) found a name", frozen)
		}
	}
	if pool.StringPoolLen() != 2 {
		t.Errorf("Find interned a name, Len = %d, want 2", pool.StringPoolLen())
	}
}

func TestStringPoolAllocs(t *testing.T) {
	pool := stringPoolTestNew()
	name := []byte("a.longer.package.name")
	gid := pool.StringPoolIntern(name)
	if allocs := testing.AllocsPerRun(100, func() { pool.StringPoolIntern(name) }); allocs != 0 {
		t.Errorf("intern of an existing string: %v allocations, want 0", allocs)
	}
	pool.StringPoolFreeze()
	if allocs := testing.AllocsPerRun(100, func() { pool.StringPoolBytes(gid) }); allocs != 0 {
		t.Errorf("Bytes: %v allocations, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = pool.StringPoolText(gid) }); allocs != 0 {
		t.Errorf("Text: %v allocations, want 0", allocs)
	}
	if text := pool.StringPoolText(gid); text != string(name) {
		t.Errorf("Text = %q, want %q", text, name)
	}
}

const stringPoolTestCaseEnv = "CX_STRINGPOOL_TEST_CASE"

var stringPoolTestCases = map[string]func(){
	"intern frozen": func() {
		pool := stringPoolTestNew()
		pool.StringPoolFreeze()
		pool.StringPoolIntern([]byte("x"))
	},
	"bytes tag": func() {
		pool := stringPoolTestNew()
		pool.StringPoolIntern([]byte("x"))
		pool.StringPoolBytes(StringGid(objmodel.GidCompose(0x02, 1, 0)))
	},
	"bytes index": func() {
		pool := stringPoolTestNew()
		pool.StringPoolIntern([]byte("x"))
		pool.StringPoolBytes(StringGid(objmodel.GidCompose(StringGidTypeTag, 1, 1)))
	},
	"bytes generation": func() {
		pool := stringPoolTestNew()
		pool.StringPoolIntern([]byte("x"))
		pool.StringPoolBytes(StringGid(objmodel.GidCompose(StringGidTypeTag, 2, 0)))
	},
	"bytes negative": func() {
		pool := stringPoolTestNew()
		pool.StringPoolBytes(-1)
	},
}

func TestStringPoolAssertChild(t *testing.T) {
	name := os.Getenv(stringPoolTestCaseEnv)
	if name == "" {
		t.Skip("subprocess only")
	}
	stringPoolTestCases[name]()
	t.Fatal("returned after a failed assert")
}

func TestStringPoolAssert(t *testing.T) {
	for name := range stringPoolTestCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(os.Args[0], "-test.run=^TestStringPoolAssertChild$")
			cmd.Env = append(os.Environ(), stringPoolTestCaseEnv+"="+name)
			var stderr strings.Builder
			cmd.Stderr = &stderr
			err := cmd.Run()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != assert.ExitCode {
				t.Fatalf("exit = %v, want status %d\nstderr:\n%s", err, assert.ExitCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), "assert failed: precondition") {
				t.Errorf("stderr lacks the precondition assert:\n%s", stderr.String())
			}
		})
	}
}
