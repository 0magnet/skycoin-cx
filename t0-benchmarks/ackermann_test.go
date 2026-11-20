package benchmarks

import "testing"

func BenchmarkAckermann(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ackermann(uint(3), uint(1))
	}
}

// See #151.
func TestAckermann(t *testing.T) {
	for _, c := range []struct{ m, n, want uint }{
		{0, 0, 1}, {1, 2, 4}, {2, 3, 9}, {3, 1, 13}, {3, 3, 61},
	} {
		if got := ackermann(c.m, c.n); got != c.want {
			t.Errorf("ackermann(%d, %d) = %d, want %d", c.m, c.n, got, c.want)
		}
	}
}
