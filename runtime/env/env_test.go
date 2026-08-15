package env_test

import (
	"os"
	"testing"

	"github.com/135yshr/meow/runtime/env"
	"github.com/135yshr/meow/runtime/meowrt"
)

// unset guarantees the variable is absent, and restores whatever the runner had
// once the test finishes.
func unset(t *testing.T, name string) {
	t.Helper()
	t.Setenv(name, "")
	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}
}

func TestHunt(t *testing.T) {
	t.Setenv("MEOW_TOKEN", "s3cret")

	got := env.Hunt(meowrt.NewString("MEOW_TOKEN"))
	if got.String() != "s3cret" {
		t.Errorf("got %q, want %q", got.String(), "s3cret")
	}
}

// An unset variable reads as catnap, so a caller can tell it apart from one set
// to the empty string.
func TestHuntUnsetIsNil(t *testing.T) {
	unset(t, "MEOW_DEFINITELY_UNSET")

	got := env.Hunt(meowrt.NewString("MEOW_DEFINITELY_UNSET"))
	if _, ok := got.(*meowrt.Furball); ok {
		t.Fatalf("expected catnap, got a Furball: %s", got.String())
	}
	if got.String() != "catnap" {
		t.Errorf("got %q, want %q", got.String(), "catnap")
	}
}

func TestHuntEmptyIsNotUnset(t *testing.T) {
	t.Setenv("MEOW_EMPTY", "")

	if got := env.Hunt(meowrt.NewString("MEOW_EMPTY")); got.String() != "" {
		t.Errorf("got %q, want an empty string", got.String())
	}
	if got := env.Hunt(meowrt.NewString("MEOW_EMPTY"), meowrt.NewString("fallback")); got.String() != "" {
		t.Errorf("a set-but-empty variable must not use the fallback, got %q", got.String())
	}
}

func TestHuntFallback(t *testing.T) {
	unset(t, "MEOW_DEFINITELY_UNSET")

	got := env.Hunt(meowrt.NewString("MEOW_DEFINITELY_UNSET"), meowrt.NewString("fallback"))
	if got.String() != "fallback" {
		t.Errorf("got %q, want %q", got.String(), "fallback")
	}
}

func TestHuntRejectsBadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []meowrt.Value
	}{
		{"no arguments", nil},
		{"too many arguments", []meowrt.Value{
			meowrt.NewString("A"), meowrt.NewString("b"), meowrt.NewString("c"),
		}},
		{"non-string name", []meowrt.Value{meowrt.NewInt(1)}},
		{"empty name", []meowrt.Value{meowrt.NewString("")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := env.Hunt(tt.args...).(*meowrt.Furball); !ok {
				t.Error("expected a Furball")
			}
		})
	}
}

func TestSniffed(t *testing.T) {
	t.Setenv("MEOW_EMPTY", "")

	if got := env.Sniffed(meowrt.NewString("MEOW_EMPTY")); got.String() != "true" {
		t.Errorf("a set-but-empty variable is still set, got %q", got.String())
	}
	unset(t, "MEOW_DEFINITELY_UNSET")
	if got := env.Sniffed(meowrt.NewString("MEOW_DEFINITELY_UNSET")); got.String() != "false" {
		t.Errorf("got %q, want false", got.String())
	}
}

func TestProwlListsNamesOnly(t *testing.T) {
	t.Setenv("MEOW_TOKEN", "s3cret")

	got := env.Prowl()
	l, ok := got.(*meowrt.List)
	if !ok {
		t.Fatalf("expected a List, got %T", got)
	}
	var found bool
	for v := range l.Iter() {
		if v.String() == "MEOW_TOKEN" {
			found = true
		}
		// Listing values would make it far too easy to print a secret.
		if v.String() == "s3cret" {
			t.Fatal("prowl must not expose values")
		}
	}
	if !found {
		t.Error("expected MEOW_TOKEN to be listed")
	}
}

func TestProwlIsSorted(t *testing.T) {
	l, ok := env.Prowl().(*meowrt.List)
	if !ok {
		t.Fatal("expected a List")
	}
	prev := ""
	for v := range l.Iter() {
		if v.String() < prev {
			t.Fatalf("names are not sorted: %q came after %q", v.String(), prev)
		}
		prev = v.String()
	}
}

// A wrong argument count must surface as a Furball. These are called from
// generated Go, where a fixed arity would instead be a compile error.
func TestArityErrorsAreFurballs(t *testing.T) {
	tests := []struct {
		name string
		got  meowrt.Value
	}{
		{"sniffed with no arguments", env.Sniffed()},
		{"sniffed with two arguments", env.Sniffed(meowrt.NewString("A"), meowrt.NewString("B"))},
		{"prowl with an argument", env.Prowl(meowrt.NewString("A"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.got.(*meowrt.Furball); !ok {
				t.Errorf("expected a Furball, got %s", tt.got.String())
			}
		})
	}
}
