package checker_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/types"
)

// The spec has always said that a binding takes the name over "whatever it
// holds and whether the name is read or called". Both backends answered a call
// from their own tables first, so the binding won the read and lost the call —
// the same name meaning two things in one scope (#154). Which declaration a
// call reaches is settled here, and recorded so that neither backend has to
// work it out again.

// A call of a name a binding has taken is typed by what the binding holds, not
// by the builtin it is named after.
func TestACallOfAShadowedBuiltinIsTypedByTheBinding(t *testing.T) {
	// The lambda's body is a trailing expression rather than a `bring`, so its
	// return type is concrete: a block-bodied lambda that returns by `bring`
	// declares nothing, and any would prove nothing here either way.
	info, errs := check(t, `nyan len = paw(s) { "counted" }
nyan answer = len("x")`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["answer"].(types.StringType); !ok {
		// len's own return type is int, and reading it off the name is what
		// used to happen.
		t.Errorf("answer is %v, want string — the type of what the binding holds", info.VarTypes["answer"])
	}
}

// And a builtin nothing has bound over is still typed as the builtin.
func TestACallOfAnUnshadowedBuiltinIsTypedByTheBuiltin(t *testing.T) {
	info, errs := check(t, `nyan answer = len([1, 2])`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["answer"].(types.IntType); !ok {
		t.Errorf("answer is %v, want int", info.VarTypes["answer"])
	}
}

// The builtin's arity is not imposed on a call that reaches a binding: the
// count belongs to whatever the call actually reaches.
func TestAShadowedBuiltinIsNotHeldToItsOwnArity(t *testing.T) {
	src := `nyan replace = paw(a, b) { bring a + b }
nya(replace("x", "y"))`
	if _, errs := check(t, src); len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

// A constructor is taken over by the same rule as a builtin.
func TestACallOfAShadowedConstructorIsTypedByTheBinding(t *testing.T) {
	info, errs := check(t, `kitty Point {
  x: int
}
nyan Point = paw(n) { "not the constructor" }
nyan p = Point(2)`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["p"].(types.StringType); !ok {
		t.Errorf("p is %v, want string — the type of what the binding holds", info.VarTypes["p"])
	}
}

// A binding made inside a body takes the name only for that body.
func TestABindingTakesTheNameOnlyWhereItIsInScope(t *testing.T) {
	src := `meow scoped(s string) string {
  nyan trim = paw(x) { "mine" }
  bring trim(s)
}
nyan outside = trim("  a  ")`
	info, errs := check(t, src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["outside"].(types.StringType); !ok {
		t.Errorf("outside is %v, want string", info.VarTypes["outside"])
	}
}

// What a call reaches is recorded for the backends, because the name alone
// cannot tell them: `upper` is the builtin's name whether or not a lambda is
// bound under it.
func TestTheResolutionOfACallIsRecordedForTheBackends(t *testing.T) {
	info, errs := check(t, `nyan upper = paw(s) { bring s }
nya(upper("hi"))
nya(lower("HI"))`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	taken := make(map[string]bool)
	for ident := range info.TakenByBinding {
		taken[ident.Name] = true
	}
	if !taken["upper"] {
		t.Error("upper is bound over, want the call recorded as taken by the binding")
	}
	if taken["lower"] {
		t.Error("nothing is bound over lower, want its call left to the builtin")
	}
}

// A pure function may call a binding that has taken an impure builtin's name:
// the builtin is not what the call reaches, and what the binding holds is
// checked where it is written. Judging by the name rather than by the
// declaration it reaches is what #137 and #138 fixed for functions.
func TestAPureFunctionMayCallABindingThatTookAnImpureBuiltinsName(t *testing.T) {
	src := `trill meow clean(s string) string {
  nyan gag = paw(x) { bring "safe:" + x }
  bring gag(s)
}`
	if _, errs := check(t, src); len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

// The impure builtin itself is still refused.
func TestAPureFunctionStillMayNotCallAnImpureBuiltin(t *testing.T) {
	_, errs := check(t, `trill meow noisy(s string) string {
  bring gag(s)
}`)
	if len(errs) != 1 {
		t.Fatalf("expected one error, got %d: %v", len(errs), errs)
	}
	if want := "must not call impure builtin gag"; !strings.Contains(errs[0].Message, want) {
		t.Errorf("says %q, want %q", errs[0].Message, want)
	}
}

// And a binding holding something impure is refused where the impurity is
// written, so taking a pure builtin's name is not a way round the check.
func TestABindingHoldingSomethingImpureIsStillRefused(t *testing.T) {
	_, errs := check(t, `trill meow sneaky(s string) string {
  nyan len = paw(x) { nya("side effect") bring x }
  bring len(s)
}`)
	if len(errs) == 0 {
		t.Fatal("expected an error about the side effect, got none")
	}
	if want := "must not call impure builtin nya"; !strings.Contains(errs[0].Message, want) {
		t.Errorf("says %q, want %q", errs[0].Message, want)
	}
}

// A name nothing has bound is not recorded as taken, so a program with no
// shadowing gives the backends nothing extra to consult.
func TestAnUnshadowedProgramRecordsNothing(t *testing.T) {
	info, errs := check(t, `kitty Point {
  x: int
}
nyan p = Point(1)
nya(len([1, 2]))
nya(p.x)`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(info.TakenByBinding) != 0 {
		var names []string
		for ident := range info.TakenByBinding {
			names = append(names, ident.Name)
		}
		t.Errorf("recorded %v, want nothing taken", names)
	}
}
