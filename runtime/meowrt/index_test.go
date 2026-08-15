package meowrt_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
)

func TestIndexList(t *testing.T) {
	lst := meowrt.NewList(meowrt.NewInt(10), meowrt.NewInt(20), meowrt.NewInt(30))

	tests := []struct {
		name  string
		index int64
		want  string
	}{
		{"first", 0, "10"},
		{"middle", 1, "20"},
		{"last", 2, "30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Index(lst, meowrt.NewInt(tt.index))
			if got.String() != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got.String())
			}
		})
	}
}

func TestIndexListOutOfRange(t *testing.T) {
	lst := meowrt.NewList(meowrt.NewInt(10))

	for _, index := range []int64{-1, 1, 99} {
		got := meowrt.Index(lst, meowrt.NewInt(index))
		f, ok := got.(*meowrt.Furball)
		if !ok {
			t.Fatalf("index %d: expected a Furball, got %T", index, got)
		}
		if !strings.Contains(f.Message, "out of range") {
			t.Errorf("index %d: expected an out-of-range message, got %q", index, f.Message)
		}
	}
}

func TestIndexMap(t *testing.T) {
	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Nyantyu"),
		"age":  meowrt.NewInt(3),
	})

	if got := meowrt.Index(m, meowrt.NewString("name")); got.String() != "Nyantyu" {
		t.Errorf("expected %q, got %q", "Nyantyu", got.String())
	}
	if got := meowrt.Index(m, meowrt.NewString("age")); got.String() != "3" {
		t.Errorf("expected %q, got %q", "3", got.String())
	}
}

// A missing key is a lookup miss, not an error: it reads as catnap so callers
// can test for it without needing to recover from a Furball.
func TestIndexMapMissingKeyIsNil(t *testing.T) {
	m := meowrt.NewMap(map[string]meowrt.Value{"name": meowrt.NewString("Nyantyu")})

	got := meowrt.Index(m, meowrt.NewString("collar"))
	if _, ok := got.(*meowrt.Furball); ok {
		t.Fatalf("expected a Nil value, got a Furball: %s", got.String())
	}
	if got.String() != "catnap" {
		t.Errorf("expected %q, got %q", "catnap", got.String())
	}
}

func TestIndexWrongKeyType(t *testing.T) {
	tests := []struct {
		name      string
		container meowrt.Value
		key       meowrt.Value
	}{
		{"list with string key", meowrt.NewList(meowrt.NewInt(1)), meowrt.NewString("a")},
		{"map with int key", meowrt.NewMap(map[string]meowrt.Value{"a": meowrt.NewInt(1)}), meowrt.NewInt(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Index(tt.container, tt.key)
			if _, ok := got.(*meowrt.Furball); !ok {
				t.Errorf("expected a Furball, got %T (%s)", got, got.String())
			}
		})
	}
}

// Indexing something that is not a container reports a Furball rather than
// panicking, so `~>` can recover from it.
func TestIndexUnindexable(t *testing.T) {
	tests := []struct {
		name      string
		container meowrt.Value
	}{
		{"int", meowrt.NewInt(1)},
		{"string", meowrt.NewString("hi")},
		{"nil value", meowrt.NewNil()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Index(tt.container, meowrt.NewInt(0))
			f, ok := got.(*meowrt.Furball)
			if !ok {
				t.Fatalf("expected a Furball, got %T", got)
			}
			if !strings.Contains(f.Message, "cannot index") {
				t.Errorf("expected a 'cannot index' message, got %q", f.Message)
			}
		})
	}
}

// A Furball reaching an index expression propagates instead of being reported
// as a bogus type error against the Furball itself.
func TestIndexPropagatesFurball(t *testing.T) {
	boom := meowrt.NewFurball("Hiss! boom, nya~")

	if got := meowrt.Index(boom, meowrt.NewInt(0)); got != meowrt.Value(boom) {
		t.Errorf("expected the container Furball to propagate, got %s", got.String())
	}
	lst := meowrt.NewList(meowrt.NewInt(1))
	if got := meowrt.Index(lst, boom); got != meowrt.Value(boom) {
		t.Errorf("expected the index Furball to propagate, got %s", got.String())
	}
}
