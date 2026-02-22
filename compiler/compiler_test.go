package compiler_test

import (
	"bytes"
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
		name := strings.TrimSuffix(filepath.Base(nyanFile), ".nyan")
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
