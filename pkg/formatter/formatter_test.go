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

purr i (10) {
  nya(fib(i))
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

// An index belongs to what it follows. Set apart by a space it read as a
// litter standing on its own next to the thing it was meant to look inside,
// and every formatted file that reached into a basket came back changed.
func TestFormatIndexBracketHugsItsSubject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"basket lookup", "nya(resp[\"body\"])\n", "nya(resp[\"body\"])\n"},
		{"litter index", "nya(xs[0])\n", "nya(xs[0])\n"},
		{"indexing a litter literal", "nya([1, 2][0])\n", "nya([1, 2][0])\n"},
		{"indexing twice", "nya(resp[\"body\"][0])\n", "nya(resp[\"body\"][0])\n"},
		{"indexing a call's result", "nya(f()[0])\n", "nya(f()[0])\n"},
		// The same bracket opening a litter is a value, and keeps the spacing
		// of whatever it follows.
		{"litter after assignment", "nyan xs = [1, 2]\n", "nyan xs = [1, 2]\n"},
		{"litter after a keyword", "bring [1, 2]\n", "bring [1, 2]\n"},
		{"litter as an argument", "nya([1, 2])\n", "nya([1, 2])\n"},
		{"nested litters", "nyan p = [[1, 2], [3, 4]]\n", "nyan p = [[1, 2], [3, 4]]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(t, tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatEmptyInput(t *testing.T) {
	got := format(t, "")
	if got != "" {
		t.Errorf("empty input should produce empty output, got: %q", got)
	}
}

func TestFormatUnaryMinus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "negative literal",
			input: "nyan x = -1\n",
			want:  "nyan x = -1\n",
		},
		{
			name:  "negative in expression",
			input: "nyan x = 1 + -2\n",
			want:  "nyan x = 1 + -2\n",
		},
		{
			name:  "binary minus preserved",
			input: "nyan x = 3 - 1\n",
			want:  "nyan x = 3 - 1\n",
		},
		{
			name:  "negative at line start",
			input: "-1\n",
			want:  "-1\n",
		},
		{
			name:  "negative after assign",
			input: "nyan x = -42\n",
			want:  "nyan x = -42\n",
		},
		{
			name:  "negative after paren",
			input: "nya(-1)\n",
			want:  "nya(-1)\n",
		},
		{
			name:  "binary minus after block comment",
			input: "nyan x = a -~ c ~- - 1\n",
			want:  "nyan x = a -~ c ~- - 1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format(t, tt.input)
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestFormatInlineLambdaSingleLine(t *testing.T) {
	input := `nyan result = lick(paw(x int) { x * 2 })
`
	want := `nyan result = lick(paw(x int) { x * 2 })
`
	got := format(t, input)
	if got != want {
		t.Errorf("inline lambda formatting mismatch\ngot:\n%s\nwant:\n%s", got, want)
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

// A purr's subject is part of the loop's opening, not a call. Judged by the
// token right before the paren — the loop variable — the space was dropped,
// and every `purr i (10)` in the repository came back looking like a call
// to `i`.
func TestFormatPurrKeepsTheSpaceBeforeItsSubject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"count form", "purr i (5) {\n  nya(i)\n}\n", "purr i (5) {\n  nya(i)\n}\n"},
		{"element form", "purr x (xs) {\n  nya(x)\n}\n", "purr x (xs) {\n  nya(x)\n}\n"},
		{"with an index", "purr i, x (xs) {\n  nya(x)\n}\n", "purr i, x (xs) {\n  nya(x)\n}\n"},
		{"over a basket", "purr k, v (m) {\n  nya(k)\n}\n", "purr k, v (m) {\n  nya(k)\n}\n"},
		{"conditional form", "purr (ready) {\n  bolt\n}\n", "purr (ready) {\n  bolt\n}\n"},
		// A call is still a call.
		{"a call keeps none", "nya(f(1))\n", "nya(f(1))\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(t, tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// A range is one thing. Spaced out as though `..` were an operator between two
// numbers, `1..5` came back as `1 .. 5`, a form in no document and no source
// file here.
func TestFormatRangeStaysTight(t *testing.T) {
	tests := []struct{ input, want string }{
		{"purr i (1..5) {\n  nya(i)\n}\n", "purr i (1..5) {\n  nya(i)\n}\n"},
		{"nyan r = 1..10\n", "nyan r = 1..10\n"},
		{"purr i (a..b) {\n  nya(i)\n}\n", "purr i (a..b) {\n  nya(i)\n}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := format(t, tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// A basket literal is a value, not a body. Broken open like a block, a one-line
// `{"body": "hi"}` came back across three lines and an empty `{}` across two.
func TestFormatBasketLiteralKeepsItsShape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"one line stays one line", "nyan m = {\"a\": 1, \"b\": 2}\n", "nyan m = {\"a\": 1, \"b\": 2}\n"},
		{"empty", "nyan m = {}\n", "nyan m = {}\n"},
		{"as an argument", "nya({\"a\": 1})\n", "nya({\"a\": 1})\n"},
		{"as a loop subject", "purr k ({\"a\": 1}) {\n  nya(k)\n}\n", "purr k ({\"a\": 1}) {\n  nya(k)\n}\n"},
		{"nested", "nyan m = {\"h\": {\"k\": \"v\"}}\n", "nyan m = {\"h\": {\"k\": \"v\"}}\n"},
		// Spread over lines in the source, it keeps that shape and is indented.
		{
			"several lines keep their indent",
			"nyan m = {\n  \"a\": 1,\n  \"b\": 2\n}\n",
			"nyan m = {\n  \"a\": 1,\n  \"b\": 2\n}\n",
		},
		// A block brace is still a block brace.
		{"a body still opens", "meow f() {\n  bring 1\n}\n", "meow f() {\n  bring 1\n}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(t, tt.input); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
