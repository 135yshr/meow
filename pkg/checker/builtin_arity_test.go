package checker_test

import (
	"strings"
	"testing"
)

// Calling a builtin with the wrong number of arguments used to be nobody's
// error in Meow: the CLI spliced the arguments into a Go call and let the Go
// compiler refuse it — naming `meow.Lower` at a line of generated code — while
// the playground's interpreter counted them itself and answered with a
// Furball. The same program was a build failure on one backend and a
// recoverable value on the other (#155). The count is the checker's to report,
// so that both backends are handed a program whose builtin calls already add
// up.
func TestCallingABuiltinWithTooFewArgumentsIsReportedHere(t *testing.T) {
	_, errs := check(t, `nya(lower())`)
	if len(errs) != 1 {
		t.Fatalf("expected one error, got %d: %v", len(errs), errs)
	}
	want := "lower requires 1 argument(s), got 0"
	if !strings.Contains(errs[0].Message, want) {
		t.Errorf("expected %q, got %q", want, errs[0].Message)
	}
}

// The wording is the one the runtime already uses for a builtin held as a
// value, so a program is told the same thing wherever the count is noticed.
func TestTheArityErrorIsWordedAsTheRuntimeWordsIt(t *testing.T) {
	_, errs := check(t, `nya(replace("a", "b"))`)
	if len(errs) != 1 {
		t.Fatalf("expected one error, got %d: %v", len(errs), errs)
	}
	want := "replace requires 3 argument(s), got 2"
	if !strings.Contains(errs[0].Message, want) {
		t.Errorf("expected %q, got %q", want, errs[0].Message)
	}
}

func TestCallingABuiltinWithTooManyArgumentsIsReportedHere(t *testing.T) {
	_, errs := check(t, `nya(len([1, 2], 3))`)
	if len(errs) != 1 {
		t.Fatalf("expected one error, got %d: %v", len(errs), errs)
	}
	want := "len requires 1 argument(s), got 2"
	if !strings.Contains(errs[0].Message, want) {
		t.Errorf("expected %q, got %q", want, errs[0].Message)
	}
}

// The error carries the position of the call that is wrong, which is the whole
// point of reporting it here rather than letting `go build` name a line of
// generated Go.
func TestTheArityErrorCarriesThePositionOfTheCall(t *testing.T) {
	_, errs := check(t, "nyan xs = [1, 2]\nnya(len())")
	if len(errs) != 1 {
		t.Fatalf("expected one error, got %d: %v", len(errs), errs)
	}
	if errs[0].Pos.Line != 2 {
		t.Errorf("expected the error on line 2, got line %d (%s)", errs[0].Pos.Line, errs[0])
	}
}

// A builtin that takes what it is given has no count to be wrong about.
func TestAVariadicBuiltinTakesAnyNumberOfArguments(t *testing.T) {
	for _, src := range []string{
		`nya()`,
		`nya(1)`,
		`nya(1, 2, 3)`,
		`hiss("boom")`,
	} {
		if _, errs := check(t, src); len(errs) > 0 {
			t.Errorf("%s: unexpected errors: %v", src, errs)
		}
	}
}

// A pipe hands the builtin on its right the value on its left, so that value
// counts towards the arity. Counting only the written arguments would refuse
// the form the tutorial teaches.
func TestAPipeSuppliesAnArgumentToTheCallOnItsRight(t *testing.T) {
	for _, src := range []string{
		`nya([1, 2, 3] |=| lick(paw(x) { bring x * 2 }))`,
		`nya([3, 1, 2] |=| sort)`,
		`nya("  a  " |=| trim |=| upper)`,
		`nya([1, 2, 3] |=| nya)`,
	} {
		if _, errs := check(t, src); len(errs) > 0 {
			t.Errorf("%s: unexpected errors: %v", src, errs)
		}
	}
}

// And a pipe into a builtin that wants more than the one value it is handed is
// still the wrong count. This is the form that used to reach `go build` as
// "not enough arguments in call to meow.Round".
func TestAPipeIntoABuiltinWantingMoreIsStillWrong(t *testing.T) {
	_, errs := check(t, `nya(3.14159 |=| round)`)
	if len(errs) != 1 {
		t.Fatalf("expected one error, got %d: %v", len(errs), errs)
	}
	want := "round requires 2 argument(s), got 1"
	if !strings.Contains(errs[0].Message, want) {
		t.Errorf("expected %q, got %q", want, errs[0].Message)
	}
}

// A binding of a builtin's name is called by the binding's own shape, not the
// builtin's, wherever the call reaches the binding rather than the builtin.
// Today a call always reaches the builtin (#154), so a binding is not a way
// out of the count — but a *read* of the name is the binding, and calling
// through that read is the binding's arity. The builtin's must not be imposed
// on it.
func TestABuiltinsArityIsNotImposedOnAValueHoldingSomethingElse(t *testing.T) {
	src := `nyan f = paw(a, b) { bring a + b }
nya(f(1, 2))`
	if _, errs := check(t, src); len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

// The count is reported once for each call that is wrong, wherever the call is
// written. Some expressions are walked more than the once that is obvious — a
// peek arm's pattern is inferred as a reference as well as matched — and a
// count reported per walk would say the same thing twice about one call.
func TestAWrongCountIsReportedOncePerCall(t *testing.T) {
	src := `meow inAFunction() string { bring lower() }
trill meow inAPureFunction() int { bring len() }
nyan inALambda = paw() { bring upper() }
meow inALoop() int {
  purr i (3) {
    nya(trim())
  }
  bring 1
}
kitty Cat { name: string }
groom Cat {
  meow inAMethod() litter { bring reverse() }
}
nyan inAnArm = peek(1) {
  1 => sort(),
  _ => []
}`
	_, errs := check(t, src)
	if len(errs) != 6 {
		t.Fatalf("expected one error for each of the six calls, got %d: %v", len(errs), errs)
	}
	seen := make(map[string]int)
	for _, e := range errs {
		seen[e.Message]++
	}
	for _, name := range []string{"lower", "len", "upper", "trim", "reverse", "sort"} {
		want := name + " requires 1 argument(s), got 0"
		if seen[want] != 1 {
			t.Errorf("%q reported %d times, want once (all: %v)", want, seen[want], errs)
		}
	}
}
