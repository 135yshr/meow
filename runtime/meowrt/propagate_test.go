package meowrt

import (
	"strings"
	"testing"
)

// A call written as a statement inside a typed function has its value
// discarded. Without raising it, a Furball would leave no trace at all and the
// function would report success.
func TestPropagateRaisesAFurball(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "boom") {
			t.Errorf("panicked with %v, want the Furball's message", r)
		}
	}()

	Propagate(&Furball{Message: "Hiss! boom, nya~"})
}

func TestPropagatePassesEverythingElseThrough(t *testing.T) {
	tests := []struct {
		name string
		in   Value
	}{
		{"an int", NewInt(3)},
		{"a litter", NewList(NewInt(1))},
		{"catnap", NewNil()},
		// gag has already answered for a handled Furball, and a program may
		// still be holding it as an ordinary value.
		{"a handled furball", &Furball{Message: "caught", Handled: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Propagate(tt.in); got != tt.in {
				t.Errorf("got %v, want the value back unchanged", got)
			}
		})
	}
}
