package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const buildTestEnv = "CX_BUILD_TEST_ARGS"

// TestBuildNative builds the compare benchmarks with -target=native and
// checks each prints the hash of expected.txt and its pass line. It needs
// clang and skips without it.
func TestBuildNative(t *testing.T) {
	if args := os.Getenv(buildTestEnv); args != "" {
		Run(strings.Split(args, "\x1f"))
		return
	}
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("no clang")
	}
	dir := "../t0-benchmarks/compare"
	file, err := os.Open(filepath.Join(dir, "expected.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	want := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if f := strings.Fields(scanner.Text()); len(f) == 3 && f[1] == "full" {
			want[f[0]] = f[2]
		}
	}
	for _, bench := range []string{"count", "fib", "array", "struct", "memtest", "primes"} {
		output := filepath.Join(t.TempDir(), bench)
		build := exec.Command(os.Args[0], "-test.run=^TestBuildNative$")
		build.Env = append(os.Environ(), buildTestEnv+"="+strings.Join([]string{"--target=native", "-o", output, filepath.Join(dir, "cx", bench+".cx")}, "\x1f"))
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("%s: build: %v\n%s", bench, err, out)
		}
		out, err := exec.Command(output).Output()
		if err != nil {
			t.Fatalf("%s: run: %v", bench, err)
		}
		lines := strings.Split(string(out), "\n")
		if len(lines) < 2 || lines[0] != want[bench] || lines[1] != bench+" pass" {
			t.Errorf("%s: got %q, want hash %s and a pass line", bench, out, want[bench])
		}
	}
}
