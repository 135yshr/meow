package checker_test

import (
	"strings"
	"testing"
)

// checkErrors runs the checker over source and returns its errors as one string.
func checkErrors(t *testing.T, src string) string {
	t.Helper()
	_, errs := check(t, src)
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

// A name the program never bound used to reach the Go compiler untouched, so a
// misspelled builtin surfaced as `undefined: keys` — generated Go leaking
// through, with no Meow source position and no mention of the Meow name.
func TestUndefinedNamesAreReported(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"unknown variable", `nya(nosuchvar)`, "undefined variable nosuchvar"},
		{"unknown function", `nya(keys({"a": 1}))`, "undefined variable keys"},
		{"a Go builtin is not a Meow one", `nya(min(1, 2))`, "undefined variable min"},
		{"misspelled builtin", `nya(lenn("ab"))`, "undefined variable lenn"},
		{"unknown name inside a function", `meow f(n int) int { bring n + missing }
nya(f(1))`, "undefined variable missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkErrors(t, tt.src)
			if !strings.Contains(got, tt.want) {
				t.Errorf("got %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

// The check must not fire on anything a program may legitimately name.
func TestKnownNamesAreAccepted(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"builtin", `nya(len(["a"]))`},
		{"fuzz seed builtin", `meow fuzz_add(a int) {
  seed(1)
  expect(a, a)
}`},
		{"top-level binding", `nyan limit = 5
nya(limit)`},
		{"top-level binding from a function", `nyan limit = 5
meow under(n int) bool { bring n < limit }
nya(under(1))`},
		{"user function", `meow f(n int) int { bring n }
nya(f(1))`},
		{"nested function", `meow outer(x int) int {
  meow inner(y int) int { bring x + y }
  bring inner(10)
}
nya(outer(5))`},
		{"recursive nested function", `meow outer(x int) int {
  meow down(y int) int {
    sniff (y <= 0) { bring 0 }
    bring down(y - 1)
  }
  bring down(x)
}
nya(outer(3))`},
		{"lambda parameter", `nyan double = paw(n int) { n * 2 }
nya(double(5))`},
		{"loop variable", `purr w (["a"]) { nya(w) }`},
		{"imported package", `nab "clock"
nya(clock.now() > 0)`},
		// A function may be written above a binding it reads and still be called
		// after that binding runs — the compiler hoists top-level bindings to
		// package scope, and the interpreter runs the declarations first.
		{"a global bound after the function that reads it", `meow f() int { bring to_int(x) }
nyan x = 1
nya(f())`},
		// Functions written side by side can call each other in either order.
		{"a nested function calling a later sibling", `meow outer(n int) int {
  meow first(y int) int { bring second(y) + 1 }
  meow second(y int) int { bring y * 2 }
  bring first(n)
}
nya(outer(3))`},
		{"aliased package", `nab "clock" tag t
nya(t.now() > 0)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkErrors(t, tt.src); strings.Contains(got, "undefined variable") {
				t.Errorf("unexpected undefined-name error: %s", got)
			}
		})
	}
}
