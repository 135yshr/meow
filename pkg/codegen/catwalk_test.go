package codegen_test

import (
	"testing"

	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
)

func TestExtractCatwalkSingleLine(t *testing.T) {
	src := `meow catwalk_hello() {
  nya("Hello, Tama!")
  # Output:
  # Hello, Tama!
}`
	l := lexer.New(src, "test.nyan")
	co := codegen.ExtractCatwalkOutputs(l.Tokens())

	want := "Hello, Tama!\n"
	if got, ok := co["catwalk_hello"]; !ok {
		t.Fatal("catwalk_hello not found")
	} else if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractCatwalkMultiLine(t *testing.T) {
	src := `meow catwalk_multi() {
  nya("line1")
  nya("line2")
  # Output:
  # line1
  # line2
}`
	l := lexer.New(src, "test.nyan")
	co := codegen.ExtractCatwalkOutputs(l.Tokens())

	want := "line1\nline2\n"
	if got, ok := co["catwalk_multi"]; !ok {
		t.Fatal("catwalk_multi not found")
	} else if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractCatwalkMultipleFunctions(t *testing.T) {
	src := `meow catwalk_a() {
  nya("aaa")
  # Output:
  # aaa
}

meow catwalk_b() {
  nya("bbb")
  # Output:
  # bbb
}`
	l := lexer.New(src, "test.nyan")
	co := codegen.ExtractCatwalkOutputs(l.Tokens())

	if len(co) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(co))
	}
	if co["catwalk_a"] != "aaa\n" {
		t.Errorf("catwalk_a: got %q, want %q", co["catwalk_a"], "aaa\n")
	}
	if co["catwalk_b"] != "bbb\n" {
		t.Errorf("catwalk_b: got %q, want %q", co["catwalk_b"], "bbb\n")
	}
}

func TestExtractCatwalkNoOutput(t *testing.T) {
	src := `meow catwalk_no_output() {
  nya("hello")
}`
	l := lexer.New(src, "test.nyan")
	co := codegen.ExtractCatwalkOutputs(l.Tokens())

	if len(co) != 0 {
		t.Errorf("expected 0 entries, got %d: %v", len(co), co)
	}
}

func TestExtractCatwalkMixedWithTest(t *testing.T) {
	src := `meow test_foo() {
  judge(yarn)
}

meow catwalk_bar() {
  nya("bar")
  # Output:
  # bar
}`
	l := lexer.New(src, "test.nyan")
	co := codegen.ExtractCatwalkOutputs(l.Tokens())

	if len(co) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(co))
	}
	if co["catwalk_bar"] != "bar\n" {
		t.Errorf("catwalk_bar: got %q, want %q", co["catwalk_bar"], "bar\n")
	}
}
