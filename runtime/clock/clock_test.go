package clock

import (
	"math"
	"testing"
	"time"

	"github.com/135yshr/meow/runtime/meowrt"
)

// freeze pins the clock so the assertions below are about the conversion, not
// about whatever the wall clock happened to read.
func freeze(t *testing.T, at time.Time) {
	t.Helper()
	original := now
	t.Cleanup(func() { now = original })
	now = func() time.Time { return at }
}

// captureSleep replaces the sleep so tests need not actually wait.
func captureSleep(t *testing.T) *time.Duration {
	t.Helper()
	original := sleep
	t.Cleanup(func() { sleep = original })
	var slept time.Duration
	sleep = func(d time.Duration) { slept = d }
	return &slept
}

func TestNow(t *testing.T) {
	freeze(t, time.Unix(1755266400, 500_000_000).UTC())

	if got := Now(); got.String() != "1755266400" {
		t.Errorf("got %q, want %q", got.String(), "1755266400")
	}
}

func TestNanos(t *testing.T) {
	freeze(t, time.Unix(1755266400, 500_000_000).UTC())

	if got := Nanos(); got.String() != "1755266400500000000" {
		t.Errorf("got %q, want %q", got.String(), "1755266400500000000")
	}
}

// The stamp is always UTC, so two machines in different zones agree.
func TestStampIsUTC(t *testing.T) {
	tokyo := time.FixedZone("JST", 9*60*60)
	freeze(t, time.Unix(1755266400, 0).In(tokyo))

	if got := Stamp(); got.String() != "2025-08-15T14:00:00Z" {
		t.Errorf("got %q, want %q", got.String(), "2025-08-15T14:00:00Z")
	}
}

func TestNap(t *testing.T) {
	slept := captureSleep(t)

	if got := Nap(meowrt.NewInt(250)); got.String() != "catnap" {
		t.Errorf("got %q, want %q", got.String(), "catnap")
	}
	if *slept != 250*time.Millisecond {
		t.Errorf("slept %v, want 250ms", *slept)
	}
}

func TestNapZeroIsAllowed(t *testing.T) {
	slept := captureSleep(t)

	if _, ok := Nap(meowrt.NewInt(0)).(*meowrt.Furball); ok {
		t.Error("expected zero to be allowed")
	}
	if *slept != 0 {
		t.Errorf("slept %v, want 0", *slept)
	}
}

// A negative delay almost always means the caller computed it wrongly, so it is
// reported rather than silently treated as zero.
func TestNapRejectsNegative(t *testing.T) {
	captureSleep(t)

	if _, ok := Nap(meowrt.NewInt(-1)).(*meowrt.Furball); !ok {
		t.Error("expected a Furball for a negative duration")
	}
}

func TestArityAndTypeErrorsAreFurballs(t *testing.T) {
	captureSleep(t)

	tests := []struct {
		name string
		got  meowrt.Value
	}{
		{"now with an argument", Now(meowrt.NewInt(1))},
		{"nanos with an argument", Nanos(meowrt.NewInt(1))},
		{"stamp with an argument", Stamp(meowrt.NewInt(1))},
		{"nap with no argument", Nap()},
		{"nap with two arguments", Nap(meowrt.NewInt(1), meowrt.NewInt(2))},
		{"nap with a non-int", Nap(meowrt.NewString("soon"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.got.(*meowrt.Furball); !ok {
				t.Errorf("expected a Furball, got %s", tt.got.String())
			}
		})
	}
}

// Milliseconds are multiplied into a time.Duration, which counts nanoseconds.
// Past the ceiling the product wraps negative and the sleep returns at once —
// the opposite of what was asked for — so the bound is checked beforehand.
func TestNapBoundary(t *testing.T) {
	slept := captureSleep(t)
	// Computed here rather than read from the implementation, so the test
	// states the contract independently of how it is enforced.
	ceiling := int64(math.MaxInt64) / int64(time.Millisecond)

	t.Run("largest accepted value", func(t *testing.T) {
		if _, ok := Nap(meowrt.NewInt(ceiling)).(*meowrt.Furball); ok {
			t.Fatal("expected the ceiling itself to be accepted")
		}
		if *slept <= 0 {
			t.Errorf("slept %v, want a positive duration", *slept)
		}
	})

	t.Run("first rejected value", func(t *testing.T) {
		if _, ok := Nap(meowrt.NewInt(ceiling + 1)).(*meowrt.Furball); !ok {
			t.Error("expected one past the ceiling to be rejected")
		}
	})

	t.Run("maximum int", func(t *testing.T) {
		if _, ok := Nap(meowrt.NewInt(math.MaxInt64)).(*meowrt.Furball); !ok {
			t.Error("expected MaxInt64 milliseconds to be rejected")
		}
	})
}
