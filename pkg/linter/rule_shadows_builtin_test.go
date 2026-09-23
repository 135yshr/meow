package linter

import (
	"strings"
	"testing"
)

// only returns the diagnostics this rule raised, so that a program written to
// trip it does not have to avoid tripping every other rule as well.
func only(diags []Diagnostic, rule string) []Diagnostic {
	var kept []Diagnostic
	for _, d := range diags {
		if d.Rule == rule {
			kept = append(kept, d)
		}
	}
	return kept
}

// A binding takes a builtin's name for as long as it is in scope, wherever the
// name is written (#154). That is the ordinary rule, and it means an accident
// nothing else can report: the program is correct, it just no longer has the
// builtin. Said at the binding, the reader sees both names at once.
func TestABindingOverABuiltinIsReported(t *testing.T) {
	diags := only(lint(t, `nyan xs = [3, 1, 2]
nyan len = len(xs)
nya(len)`), "shadows-builtin")
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %d: %v", len(diags), diags)
	}
	want := `binding "len" shadows the builtin len; calls to len in this scope reach the binding`
	if got := diags[0].Message; got != want {
		t.Errorf("says %q, want %q", got, want)
	}
	if diags[0].Pos.Line != 2 {
		t.Errorf("reported on line %d, want line 2 — where the name is taken", diags[0].Pos.Line)
	}
	if diags[0].Severity != Warning {
		t.Errorf("severity is %v, want a warning: the program is correct", diags[0].Severity)
	}
}

// A parameter is a binding too, and the one most likely to be written without
// noticing: `round`, `len` and `sort` are ordinary names for an argument.
func TestAParameterOverABuiltinIsReported(t *testing.T) {
	diags := only(lint(t, `meow summary(round int, xs litter) string {
  bring to_string(round)
}
nya(summary(1, [1, 2]))`), "shadows-builtin")
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %d: %v", len(diags), diags)
	}
	if got := diags[0].Message; !strings.Contains(got, `parameter "round" shadows the builtin round`) {
		t.Errorf("says %q, want it to name the parameter", got)
	}
}

func TestALambdaParameterOverABuiltinIsReported(t *testing.T) {
	diags := only(lint(t, `nyan f = paw(sort) { sort }
nya(f(1))`), "shadows-builtin")
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %d: %v", len(diags), diags)
	}
	if got := diags[0].Message; !strings.Contains(got, `parameter "sort"`) {
		t.Errorf("says %q, want it to name the lambda's parameter", got)
	}
}

func TestALoopVariableOverABuiltinIsReported(t *testing.T) {
	diags := only(lint(t, `purr head (3) {
  nya(head)
}`), "shadows-builtin")
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %d: %v", len(diags), diags)
	}
	if got := diags[0].Message; !strings.Contains(got, `loop variable "head"`) {
		t.Errorf("says %q, want it to name the loop variable", got)
	}
}

// A kitty or collar constructor is taken over by the same rule, so it is worth
// the same warning — and the declaration may be written after the binding.
func TestABindingOverAConstructorIsReported(t *testing.T) {
	diags := only(lint(t, `nyan Point = paw(n) { n }
kitty Point {
  x: int
}
collar Age = int
nyan Age = paw(n) { n }
nya(Point(1))
nya(Age(1))`), "shadows-builtin")
	if len(diags) != 2 {
		t.Fatalf("expected two diagnostics, got %d: %v", len(diags), diags)
	}
	if got := diags[0].Message; !strings.Contains(got, `shadows the kitty Point`) {
		t.Errorf("says %q, want it to name the kitty", got)
	}
	if got := diags[1].Message; !strings.Contains(got, `shadows the collar Age`) {
		t.Errorf("says %q, want it to name the collar", got)
	}
}

// A name that is nobody else's is not reported, which is nearly every name.
func TestAnOrdinaryBindingIsNotReported(t *testing.T) {
	diags := only(lint(t, `nyan total = 1
nyan names = ["a"]
meow add(a int, b int) int { bring a + b }
nya(add(total, 1))
nya(names)`), "shadows-builtin")
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

// Reading a builtin as a value is not shadowing it — nothing is bound.
func TestNamingABuiltinAsAValueIsNotReported(t *testing.T) {
	diags := only(lint(t, `nyan f = upper
nya(f("hi"))
nya(lick(["a"], upper))`), "shadows-builtin")
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}

// A binding written inside a body is reported there, where it takes the name.
func TestABindingInsideABodyIsReportedThere(t *testing.T) {
	diags := only(lint(t, `meow scoped(s string) string {
  nyan trim = paw(x) { x }
  bring trim(s)
}
nya(scoped("a"))`), "shadows-builtin")
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %d: %v", len(diags), diags)
	}
	if diags[0].Pos.Line != 2 {
		t.Errorf("reported on line %d, want line 2", diags[0].Pos.Line)
	}
}

// A top-level meow named after a builtin takes the name as surely as a binding
// does, and used to be unreachable instead.
func TestAFunctionNamedAfterABuiltinIsReported(t *testing.T) {
	diags := only(lint(t, `meow sort(a int, b int) int { bring a + b }
nya(sort(1, 2))`), "shadows-builtin")
	if len(diags) != 1 {
		t.Fatalf("expected one diagnostic, got %d: %v", len(diags), diags)
	}
	if got := diags[0].Message; !strings.Contains(got, `function "sort" shadows the builtin sort`) {
		t.Errorf("says %q, want it to name the function", got)
	}
}
