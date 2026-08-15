package random

import (
	"errors"
	"strings"
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
)

// stubInt pins the integer source so the assertions are about the conversion
// and the bounds, not about which value happened to come out.
func stubInt(t *testing.T, value int64) *int64 {
	t.Helper()
	original := randomInt
	t.Cleanup(func() { randomInt = original })
	var bound int64
	randomInt = func(n int64) int64 {
		bound = n
		return value
	}
	return &bound
}

func TestRoll(t *testing.T) {
	bound := stubInt(t, 7)

	if got := Roll(meowrt.NewInt(10)); got.String() != "7" {
		t.Errorf("got %q, want %q", got.String(), "7")
	}
	if *bound != 10 {
		t.Errorf("bound: got %d, want 10", *bound)
	}
}

// The bound is exclusive and must be positive: roll(0) has no valid answer.
func TestRollRejectsNonPositiveBound(t *testing.T) {
	for _, n := range []int64{0, -1} {
		if _, ok := Roll(meowrt.NewInt(n)).(*meowrt.Furball); !ok {
			t.Errorf("bound %d: expected a Furball", n)
		}
	}
}

func TestDrift(t *testing.T) {
	original := randomFloat
	t.Cleanup(func() { randomFloat = original })
	randomFloat = func() float64 { return 0.25 }

	if got := Drift(); got.String() != "0.25" {
		t.Errorf("got %q, want %q", got.String(), "0.25")
	}
}

func TestPick(t *testing.T) {
	bound := stubInt(t, 1)

	list := meowrt.NewList(meowrt.NewString("a"), meowrt.NewString("b"), meowrt.NewString("c"))
	if got := Pick(list); got.String() != "b" {
		t.Errorf("got %q, want %q", got.String(), "b")
	}
	if *bound != 3 {
		t.Errorf("bound: got %d, want 3", *bound)
	}
}

func TestPickRejectsEmptyList(t *testing.T) {
	if _, ok := Pick(meowrt.NewList()).(*meowrt.Furball); !ok {
		t.Error("expected a Furball for an empty litter")
	}
}

func TestTuft(t *testing.T) {
	original := randomBytes
	t.Cleanup(func() { randomBytes = original })
	randomBytes = func(b []byte) (int, error) {
		for i := range b {
			b[i] = 0xAB
		}
		return len(b), nil
	}

	got := Tuft(meowrt.NewInt(4))
	if got.String() != "abababab" {
		t.Errorf("got %q, want %q", got.String(), "abababab")
	}
}

// Hex doubles the length, which callers sizing an identifier need to know.
func TestTuftLengthIsTwiceTheBytes(t *testing.T) {
	got := Tuft(meowrt.NewInt(8))
	if len(got.String()) != 16 {
		t.Errorf("got %d characters, want 16", len(got.String()))
	}
}

func TestTuftBounds(t *testing.T) {
	for _, n := range []int64{0, -1, maxTuftBytes + 1} {
		if _, ok := Tuft(meowrt.NewInt(n)).(*meowrt.Furball); !ok {
			t.Errorf("length %d: expected a Furball", n)
		}
	}
}

func TestTuftReportsSourceFailure(t *testing.T) {
	original := randomBytes
	t.Cleanup(func() { randomBytes = original })
	randomBytes = func([]byte) (int, error) { return 0, errors.New("entropy exhausted") }

	got := Tuft(meowrt.NewInt(4))
	f, ok := got.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected a Furball, got %T", got)
	}
	if !strings.Contains(f.Message, "entropy exhausted") {
		t.Errorf("expected the cause in %q", f.Message)
	}
}

func TestArityAndTypeErrorsAreFurballs(t *testing.T) {
	tests := []struct {
		name string
		got  meowrt.Value
	}{
		{"roll with no argument", Roll()},
		{"roll with a non-int", Roll(meowrt.NewString("ten"))},
		{"drift with an argument", Drift(meowrt.NewInt(1))},
		{"pick with no argument", Pick()},
		{"pick with a non-list", Pick(meowrt.NewString("abc"))},
		{"tuft with no argument", Tuft()},
		{"tuft with a non-int", Tuft(meowrt.NewString("eight"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.got.(*meowrt.Furball); !ok {
				t.Errorf("expected a Furball, got %s", tt.got.String())
			}
		})
	}
}
