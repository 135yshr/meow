package formatter

import (
	"strings"
	"testing"
)

func format(t *testing.T, input string) string {
	t.Helper()
	return FormatSource(input, "test.nyan")
}

func TestFormatBasicIndentation(t *testing.T) {
	input := `meow greet(name string) string {
bring "hello"
}
`
	want := `meow greet(name string) string {
  bring "hello"
}
`
	got := format(t, input)
	if got != want {
		t.Errorf("indentation mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatNestedIndentation(t *testing.T) {
	input := `meow f(n int) int {
sniff (n > 0) {
bring n
}
bring 0
}
`
	want := `meow f(n int) int {
  sniff (n > 0) {
    bring n
  }
  bring 0
}
`
	got := format(t, input)
	if got != want {
		t.Errorf("nested indentation mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatOperatorSpacing(t *testing.T) {
	input := `nyan x=1+2
`
	want := `nyan x = 1 + 2
`
	got := format(t, input)
	if got != want {
		t.Errorf("operator spacing mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatCommentPreservation(t *testing.T) {
	input := `# this is a comment
nyan x = 1
`
	want := `# this is a comment
nyan x = 1
`
	got := format(t, input)
	if got != want {
		t.Errorf("comment mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatBlockCommentPreservation(t *testing.T) {
	input := `-~ block comment ~-
nyan x = 1
`
	want := `-~ block comment ~-
nyan x = 1
`
	got := format(t, input)
	if got != want {
		t.Errorf("block comment mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatScratchSameLine(t *testing.T) {
	input := `sniff (x > 0) {
  nya(x)
}
scratch {
  nya(0)
}
`
	want := `sniff (x > 0) {
  nya(x)
} scratch {
  nya(0)
}
`
	got := format(t, input)
	if got != want {
		t.Errorf("scratch same line mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatBlankLineNormalization(t *testing.T) {
	input := `nyan x = 1



nyan y = 2
`
	want := `nyan x = 1

nyan y = 2
`
	got := format(t, input)
	if got != want {
		t.Errorf("blank line normalization mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatIdempotent(t *testing.T) {
	input := `meow fib(n int) int {
  sniff (n <= 1) {
    bring n
  }
  bring fib(n - 1) + fib(n - 2)
}

nyan i = 0
purr (i < 10) {
  nya(fib(i))
  i = i + 1
}
`
	first := format(t, input)
	second := format(t, first)
	if first != second {
		t.Errorf("not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestFormatCommaSpacing(t *testing.T) {
	input := `nyan xs = [1,2,3]
`
	want := `nyan xs = [1, 2, 3]
`
	got := format(t, input)
	if got != want {
		t.Errorf("comma spacing mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatFunctionCallNoSpaceAfterName(t *testing.T) {
	input := `nya(42)
`
	want := `nya(42)
`
	got := format(t, input)
	if got != want {
		t.Errorf("function call mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatEmptyInput(t *testing.T) {
	got := format(t, "")
	if got != "" {
		t.Errorf("empty input should produce empty output, got: %q", got)
	}
}

func TestFormatPipeOperator(t *testing.T) {
	input := `nyan result = xs|=|lick(paw(x int) { x * 2 })
`
	got := format(t, input)
	if got == "" {
		t.Fatal("unexpected empty output")
	}
	// Should contain spaces around |=|
	if !strings.Contains(got, "|=|") {
		t.Errorf("expected pipe operator in output, got: %s", got)
	}
}
