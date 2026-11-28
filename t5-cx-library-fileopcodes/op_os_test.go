//go:build !nocxos

package fileopcodes

import (
	"reflect"
	"testing"

	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

// os.Run split the command on single spaces. See #118.
func TestOsCommandSplit(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		ok   bool
	}{
		{"", nil, true},
		{"   ", nil, true},
		{"echo a", []string{"echo", "a"}, true},
		{"echo  a\tb", []string{"echo", "a", "b"}, true},
		{"sh -c 'echo hi'", []string{"sh", "-c", "echo hi"}, true},
		{`printf "%s|" "a b" c`, []string{"printf", "%s|", "a b", "c"}, true},
		{`echo ""`, []string{"echo", ""}, true},
		{`echo a'b c'd`, []string{"echo", "ab cd"}, true},
		{`C:\bin\tool.exe x`, []string{`C:\bin\tool.exe`, "x"}, true},
		{"echo 'a", nil, false},
	}
	for _, c := range cases {
		var pool stringpool.StringPool
		words, closed := osCommandSplit(&pool, []byte(c.in), nil)
		if closed != c.ok {
			t.Errorf("osCommandSplit(%q) closed %v, want %v", c.in, closed, c.ok)
			continue
		}
		var got []string
		for _, word := range words {
			got = append(got, string(pool.StringPoolBytes(word)))
		}
		if c.ok && !reflect.DeepEqual(got, c.want) {
			t.Errorf("osCommandSplit(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
