package compiler_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/135yshr/meow/compiler"
	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/interpreter"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

// runCompiled builds source and runs it, giving what it printed and the
// failure it reported, if any, from the file name on.
func runCompiled(t *testing.T, source string) (output, failure string) {
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
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	_ = cmd.Run()
	failure = strings.TrimSpace(stderr.String())
	// The path is the temporary directory's, so only the file name and what
	// follows it are compared.
	if i := strings.LastIndex(failure, "prog.nyan:"); i >= 0 {
		failure = failure[i:]
	}
	return stdout.String(), failure
}

// runInterpreted runs source the way the playground does.
func runInterpreted(t *testing.T, source string) (output, failure string) {
	t.Helper()
	prog, errs := parser.New(lexer.New(source, "prog.nyan").Tokens()).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ti, checkErrs := checker.New().Check(prog)
	if len(checkErrs) > 0 {
		t.Fatalf("checker errors: %v", checkErrs)
	}
	var stdout bytes.Buffer
	in := interpreter.New(&stdout)
	in.SetTypeInfo(ti)
	if err := in.RunSafe(prog); err != nil {
		failure = err.Error()
	}
	return stdout.String(), failure
}

// A function may read a top-level binding written below it, but calling the
// function before the binding's line has run reaches a name with nothing in it
// yet. Both backends used to call that undefined, which sent the reader looking
// for a missing declaration rather than at the order of the lines, and the
// playground said "undefined function" for a call where the compiled program
// said "undefined variable" (#161).
func TestABindingUsedBeforeItIsBoundSaysSo(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		output  string
		failure string
	}{
		{
			"read from a typed function",
			"meow greet() string { bring label }\nnya(greet())\nnyan label = \"tama\"\n",
			"",
			"prog.nyan:1:23: Hiss! label is used before it is bound, nya~",
		},
		{
			"read from an untyped function",
			"meow show() { nya(label) }\nnya(\"start\")\nshow()\nnyan label = \"tama\"\n",
			"start\n",
			"prog.nyan:1:15: Hiss! label is used before it is bound, nya~",
		},
		{
			"called",
			"meow run() { nya(other(\"x\")) }\nrun()\nnyan other = paw(s) { bring s }\n",
			"",
			"prog.nyan:1:14: Hiss! other is used before it is bound, nya~",
		},
		{
			"called from a typed function",
			"meow run() string { bring other(\"x\") }\nnya(run())\nnyan other = paw(s) { bring s }\n",
			"",
			"prog.nyan:1:21: Hiss! other is used before it is bound, nya~",
		},
		{
			"piped into",
			"meow run() { nya(\"x\" |=| other) }\nrun()\nnyan other = paw(s) { bring s }\n",
			"",
			"prog.nyan:1:14: Hiss! other is used before it is bound, nya~",
		},
		{
			// Since #154 a top-level binding can take a builtin's name, so an
			// ordinary builtin can be the name that is not bound yet.
			"called over a builtin's name",
			"meow run() { nya(trim(\"  x  \")) }\nrun()\nnyan trim = paw(s) { bring s }\n",
			"",
			"prog.nyan:1:14: Hiss! trim is used before it is bound, nya~",
		},
		{
			// A top-level binding takes its name for the whole program, above
			// its own line too, so `upper` here is the binding and not the
			// builtin. The compiled program always said so; the playground
			// reached the builtin whenever nothing was bound yet.
			"named as a value over a builtin's name, above the binding",
			"nyan speak = upper\nnya(speak(\"a\"))\nnyan upper = paw(s) { bring s + \"!\" }\n",
			"",
			"prog.nyan:1:1: Hiss! upper is used before it is bound, nya~",
		},
		{
			// The checker lets the top level itself name a binding written
			// further down, so it can get here without a function too.
			"read at the top level, above the binding",
			"nya(\"start\")\nnya(label)\nnyan label = \"tama\"\n",
			"start\n",
			"prog.nyan:2:1: Hiss! label is used before it is bound, nya~",
		},
		{
			// It is a Furball like any other, so it can be caught.
			"caught",
			"meow show() { nya(label ~> \"fallback\") }\nshow()\nnyan label = \"tama\"\nshow()\n",
			"fallback\ntama\n",
			"",
		},
		{
			"caught from a call",
			"meow show() { nya(other(\"x\") ~> \"fallback\") }\nshow()\nnyan other = paw(s) { bring s + \"!\" }\nshow()\n",
			"fallback\nx!\n",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, backend := range []struct {
				name string
				run  func(*testing.T, string) (string, string)
			}{
				{"compiled", runCompiled},
				{"interpreted", runInterpreted},
			} {
				output, failure := backend.run(t, tt.source)
				if output != tt.output {
					t.Errorf("%s printed %q, want %q", backend.name, output, tt.output)
				}
				if failure != tt.failure {
					t.Errorf("%s failed with %q, want %q", backend.name, failure, tt.failure)
				}
			}
		})
	}
}
