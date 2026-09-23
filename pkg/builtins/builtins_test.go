package builtins_test

import (
	"testing"

	"github.com/135yshr/meow/pkg/builtins"
)

func TestKnownAnswersForABuiltinAndNothingElse(t *testing.T) {
	for _, name := range []string{"nya", "len", "replace", "seed"} {
		if !builtins.Known(name) {
			t.Errorf("%s is a builtin, want Known to say so", name)
		}
	}
	for _, name := range []string{"", "keys", "Len", "file.snoop"} {
		if builtins.Known(name) {
			t.Errorf("%s is not a builtin, want Known to say so", name)
		}
	}
}

func TestArityAnswersTheCountAndWhetherThereIsOne(t *testing.T) {
	for name, want := range map[string]int{
		"len":     1,
		"pad":     2,
		"replace": 3,
		"nya":     builtins.Variadic,
	} {
		got, ok := builtins.Arity(name)
		if !ok {
			t.Errorf("%s is a builtin, want an arity for it", name)
			continue
		}
		if got != want {
			t.Errorf("%s takes %d, want %d", name, got, want)
		}
	}
	if _, ok := builtins.Arity("keys"); ok {
		t.Error("keys is not a builtin, want no arity for it")
	}
}

func TestWrongRefusesOnlyACountItCanJudge(t *testing.T) {
	cases := []struct {
		name  string
		got   int
		want  int
		wrong bool
	}{
		{"len", 1, 1, false},
		{"len", 0, 1, true},
		{"len", 2, 1, true},
		{"replace", 3, 3, false},
		// A builtin that takes what it is given has no count to be wrong about.
		{"nya", 0, builtins.Variadic, false},
		{"nya", 7, builtins.Variadic, false},
		// Neither has a name that is no builtin: refusing it is not this
		// table's to do, and something else has already said it is undefined.
		{"keys", 2, 0, false},
	}
	for _, c := range cases {
		want, wrong := builtins.Wrong(c.name, c.got)
		if wrong != c.wrong {
			t.Errorf("Wrong(%q, %d) says wrong=%v, want %v", c.name, c.got, wrong, c.wrong)
		}
		if wrong && want != c.want {
			t.Errorf("Wrong(%q, %d) wants %d, want %d", c.name, c.got, want, c.want)
		}
	}
}

// Every builtin takes a number of arguments that a program could actually
// write. A negative count other than Variadic would have genBuiltinValue
// building a Go call with a negative number of arguments.
func TestEveryArityIsAWritableCount(t *testing.T) {
	for _, name := range builtins.Names() {
		n, _ := builtins.Arity(name)
		if n < 0 && n != builtins.Variadic {
			t.Errorf("%s takes %d arguments, which is no count at all", name, n)
		}
	}
}

func TestNamesListsEveryBuiltinInOrder(t *testing.T) {
	names := builtins.Names()
	if len(names) == 0 {
		t.Fatal("no builtins at all")
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Fatalf("names are not in order: %q before %q", names[i-1], names[i])
		}
	}
	for _, name := range names {
		if !builtins.Known(name) {
			t.Errorf("Names lists %q but Known does not answer for it", name)
		}
	}
}
