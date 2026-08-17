package compiler_test

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/135yshr/meow/compiler"
)

var update = flag.Bool("update", false, "update golden files")

func TestGoldenFiles(t *testing.T) {
	entries, err := filepath.Glob(filepath.Join("..", "testdata", "*.nyan"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no test data files found")
	}

	c := compiler.New(nil)

	for _, nyanFile := range entries {
		base := filepath.Base(nyanFile)
		if strings.HasSuffix(base, "_test.nyan") {
			continue
		}
		name := strings.TrimSuffix(base, ".nyan")
		t.Run(name, func(t *testing.T) {
			goldenFile := strings.TrimSuffix(nyanFile, ".nyan") + ".golden"

			// Compile to binary
			tmpBin, err := os.CreateTemp("", "meow-test-*")
			if err != nil {
				t.Fatal(err)
			}
			tmpBin.Close()
			defer os.Remove(tmpBin.Name())

			if err := c.Build(nyanFile, tmpBin.Name()); err != nil {
				t.Fatalf("build failed: %v", err)
			}

			// Run binary and capture output
			cmd := exec.Command(tmpBin.Name())
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Run(); err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got := stdout.String()

			if *update {
				if err := os.WriteFile(goldenFile, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("cannot read golden file: %v", err)
			}

			if got != string(want) {
				t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, string(want))
			}
		})
	}
}

func TestGoldenTestFiles(t *testing.T) {
	entries, err := filepath.Glob(filepath.Join("..", "testdata", "*_test.nyan"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Skip("no test data files found")
	}

	c := compiler.New(nil)

	for _, nyanFile := range entries {
		name := strings.TrimSuffix(filepath.Base(nyanFile), ".nyan")
		t.Run(name, func(t *testing.T) {
			goldenFile := strings.TrimSuffix(nyanFile, ".nyan") + ".golden"

			tmpBin, err := os.CreateTemp("", "meow-test-golden-*")
			if err != nil {
				t.Fatal(err)
			}
			tmpBin.Close()
			defer os.Remove(tmpBin.Name())

			if err := c.BuildTest(nyanFile, tmpBin.Name()); err != nil {
				t.Fatalf("build failed: %v", err)
			}

			cmd := exec.Command(tmpBin.Name())
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Run(); err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got := stdout.String()

			if *update {
				if err := os.WriteFile(goldenFile, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("cannot read golden file: %v", err)
			}

			if got != string(want) {
				t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, string(want))
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	c := compiler.New(nil)
	_, err := c.CompileToGo(`nyan = 42`, "bad.nyan")
	if err == nil {
		t.Fatal("expected error for bad syntax")
	}
	if !strings.Contains(err.Error(), "Hiss!") {
		t.Errorf("expected cat-themed error, got: %s", err)
	}
}

// A program's exit status is how it tells a shell, cron or a CI step what it
// found, so it has to survive being compiled and run. The golden files cannot
// cover this: their harness treats a non-zero status as a failed run.
func TestScramSetsTheExitStatus(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int
	}{
		{"a status", "nya(\"checked\")\nscram(3)\n", 3},
		{"success", "nya(\"checked\")\nscram(0)\n", 0},
		{"no argument means success", "scram()\n", 0},
		// Reaching the end is success, the same as a process that ran out of
		// statements.
		{"never scrammed", "nya(\"checked\")\n", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runSource(t, tt.source); got != tt.want {
				t.Errorf("exited with %d, want %d", got, tt.want)
			}
		})
	}
}

// What follows scram must not run: a check that decided it is done has decided.
func TestScramStopsTheProgram(t *testing.T) {
	dir := t.TempDir()
	nyanPath := filepath.Join(dir, "prog.nyan")
	source := "nya(\"before\")\nscram(0)\nnya(\"after\")\n"
	if err := os.WriteFile(nyanPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "prog")
	if err := compiler.New(nil).Build(nyanPath, binPath); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	out, err := exec.Command(binPath).Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if string(out) != "before\n" {
		t.Errorf("got %q, want %q", string(out), "before\n")
	}
}

// env.haul reads what the program was started with, so a program can be told
// what to work on rather than having it written into its source.
func TestHaulReadsTheCommandLine(t *testing.T) {
	dir := t.TempDir()
	nyanPath := filepath.Join(dir, "prog.nyan")
	source := "nab \"env\"\npurr a (env.haul()) { nya(a) }\n"
	if err := os.WriteFile(nyanPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "prog")
	if err := compiler.New(nil).Build(nyanPath, binPath); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// -v among them because a program is entitled to its own flags.
	out, err := exec.Command(binPath, "--target", "https://example.test", "-v").Output()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	want := "--target\nhttps://example.test\n-v\n"
	if string(out) != want {
		t.Errorf("got %q, want %q", string(out), want)
	}
}

// runSource builds source and reports the status the program exited with.
func runSource(t *testing.T, source string) int {
	t.Helper()
	dir := t.TempDir()
	nyanPath := filepath.Join(dir, "prog.nyan")
	if err := os.WriteFile(nyanPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "prog")
	if err := compiler.New(nil).Build(nyanPath, binPath); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	cmd := exec.Command(binPath)
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	return 0
}

// Run is what `meow run` calls, and it is the piece that hands the program its
// arguments and reports back the status it ended on. Running the binary
// directly, as the tests above do, would leave both untested.
func TestRunForwardsArgumentsAndReportsTheStatus(t *testing.T) {
	dir := t.TempDir()
	nyanPath := filepath.Join(dir, "prog.nyan")
	// Ends on the number of arguments it was handed, so one run proves both
	// that they arrived and that the status came back.
	source := "nab \"env\"\nscram(len(env.haul()))\n"
	if err := os.WriteFile(nyanPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	err := compiler.New(nil).Run(nyanPath, "--target", "https://example.test", "-v")

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("got %v, want an exit status", err)
	}
	if exitErr.ExitCode() != 3 {
		t.Errorf("exited with %d, want 3", exitErr.ExitCode())
	}
}

// A program that ends well is not an error, so nothing is reported.
func TestRunReportsNoErrorWhenTheProgramSucceeds(t *testing.T) {
	dir := t.TempDir()
	nyanPath := filepath.Join(dir, "prog.nyan")
	if err := os.WriteFile(nyanPath, []byte("scram(0)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := compiler.New(nil).Run(nyanPath); err != nil {
		t.Errorf("got %v, want no error", err)
	}
}

// A fully typed function returns a native Go type and cannot pass a Furball
// back, so a refused status is raised there rather than dropped. Emitting a
// bare call left the program running as if nothing had been asked: this printed
// 7 and succeeded.
func TestScramRefusedInsideATypedFunction(t *testing.T) {
	dir := t.TempDir()
	nyanPath := filepath.Join(dir, "prog.nyan")
	source := "meow f() int {\n  scram(300)\n  bring 7\n}\nnya(to_string(f()))\n"
	if err := os.WriteFile(nyanPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "prog")
	if err := compiler.New(nil).Build(nyanPath, binPath); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	out, err := exec.Command(binPath).CombinedOutput()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("got %v and output %q, want a failure", err, string(out))
	}
	if !strings.Contains(string(out), "0 to 255") {
		t.Errorf("output %q, want the reason the status was refused", string(out))
	}
}

// A failure says what went wrong; without a position it does not say where, and
// finding which of two hundred lines asked for that number is the reader's
// problem. The golden harness cannot cover this: it captures stdout, and a
// failure is reported on stderr.
func TestAFailureSaysWhereItHappened(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			"a runtime failure at the top level",
			"nya(\"starting\")\nnyan raw = \"not a number\"\nnya(to_string(to_float(raw)))\n",
			"prog.nyan:3:1: Hiss! Cannot read \"not a number\" as a Float, nya~",
		},
		{
			// Inside a fully typed function, where failure travels as a panic
			// rather than as a value.
			"a failure inside a typed function",
			"meow ratio(a int, b int) float {\n  sniff (b == 0) {\n    hiss(\"cannot divide by zero\")\n  }\n  bring to_float(a) / to_float(b)\n}\nnya(to_string(ratio(1, 0)))\n",
			"prog.nyan:3:5: Hiss! cannot divide by zero",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			nyanPath := filepath.Join(dir, "prog.nyan")
			if err := os.WriteFile(nyanPath, []byte(tt.source), 0o644); err != nil {
				t.Fatal(err)
			}
			binPath := filepath.Join(dir, "prog")
			if err := compiler.New(nil).Build(nyanPath, binPath); err != nil {
				t.Fatalf("build failed: %v", err)
			}

			cmd := exec.Command(binPath)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				t.Fatal("expected the program to fail")
			}

			// The path is the temporary directory's, so only the file name and
			// position are compared.
			got := strings.TrimSpace(stderr.String())
			if !strings.HasSuffix(got, tt.want) {
				t.Errorf("got %q, want it to end with %q", got, tt.want)
			}
		})
	}
}

// A call that comes back must leave the program where the call was made, and a
// call that fails must not. Both are checked here because the position is a
// single note the runtime keeps, and a call is what can leave it stale.
func TestAFailureIsBlamedOnTheRightLine(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			// The call succeeds; the failure is in the next argument.
			"after a call that succeeded",
			"meow ok(n int) int {\n  nyan doubled = n * 2\n  bring doubled\n}\nnya(\"starting\")\nnya(to_string(ok(1)), to_string(to_float(\"bad\")))\n",
			"prog.nyan:6:1: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			// The failure is inside the call, so that is the line to report.
			"inside a call",
			"meow bad(s string) float {\n  nya(\"converting\")\n  bring to_float(s)\n}\nnya(to_string(bad(\"nope\")))\n",
			"prog.nyan:3:3: Hiss! Cannot read \"nope\" as a Float, nya~",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			nyanPath := filepath.Join(dir, "prog.nyan")
			if err := os.WriteFile(nyanPath, []byte(tt.source), 0o644); err != nil {
				t.Fatal(err)
			}
			binPath := filepath.Join(dir, "prog")
			if err := compiler.New(nil).Build(nyanPath, binPath); err != nil {
				t.Fatalf("build failed: %v", err)
			}

			cmd := exec.Command(binPath)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				t.Fatal("expected the program to fail")
			}

			if got := strings.TrimSpace(stderr.String()); !strings.HasSuffix(got, tt.want) {
				t.Errorf("got %q, want it to end with %q", got, tt.want)
			}
		})
	}
}
