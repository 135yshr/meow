package compiler_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/135yshr/meow/compiler"
)

// captureStdio runs fn with os.Stdout and os.Stderr replaced by a pipe, and
// answers with everything it wrote. RunFuzz reports through the process's own
// streams — it hands them to the `go test` it spawns — so this is the only way
// to read what a fuzz run said, and it keeps the noise out of `go test`'s own
// output.
func captureStdio(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
		w.Close()
		r.Close()
	}()
	fn()
	os.Stdout, os.Stderr = oldStdout, oldStderr
	w.Close()
	out := <-done
	r.Close()
	return out
}

// fuzzTime keeps the fuzzer to a handful of executions: enough to run every
// seed and a few mutations, not enough to be worth waiting for. The count form
// rather than a duration, so a slow machine does not run longer.
const fuzzTime = "10x"

// Nothing else in the Go test suite builds a fuzz target, which is how
// `meow test -fuzz` came to ship generating Go that did not compile (#143).
// This drives the whole path: generate, build, and run the targets in
// examples/fuzz_arithmetic.nyan.
func TestFuzzTargetsBuildAndRun(t *testing.T) {
	c := compiler.New(nil)

	var err error
	out := captureStdio(t, func() {
		err = c.RunFuzz(filepath.Join("..", "examples", "fuzz_arithmetic.nyan"), fuzzTime)
	})

	if err != nil {
		t.Fatalf("fuzz run failed: %v\n%s", err, out)
	}
	for _, target := range []string{"FuzzAdd_commutative", "FuzzAdd_identity", "FuzzAbs_non_negative"} {
		if !strings.Contains(out, target) {
			t.Errorf("expected %s to have been run, got:\n%s", target, out)
		}
	}
}

// A failing assertion inside a fuzz target fails the fuzz case, and says which
// target, which input and which assertion.
func TestAFuzzFailureIsReported(t *testing.T) {
	dir := t.TempDir()
	src := `meow half(x int) int {
  bring x / 2
}

meow fuzz_half_round_trip(x int) {
  seed(7)

  expect(half(x) * 2, x)
}
`
	path := filepath.Join(dir, "fuzz_half.nyan")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	c := compiler.New(nil)

	var err error
	out := captureStdio(t, func() {
		err = c.RunFuzz(path, fuzzTime)
	})

	if err == nil {
		t.Fatalf("expected the fuzz run to fail, got:\n%s", out)
	}
	for _, want := range []string{
		"FAIL: FuzzHalf_round_trip",
		"Hiss! fuzz_half_round_trip(x=7) failed, nya~: expected 7, got 6",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected the report to contain %q, got:\n%s", want, out)
		}
	}
}
