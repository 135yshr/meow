package json

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
)

func TestUnravelScalars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"string", `"cat"`, "cat"},
		{"true", "true", "true"},
		{"false", "false", "false"},
		{"null", "null", "catnap"},
		{"whole number", "42", "42"},
		{"negative whole number", "-7", "-7"},
		{"fraction", "2.5", "2.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unravel(meowrt.NewString(tt.in))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// JSON has one number type and Meow has two. A whole value has to come back as
// an Int, or an id read from an API arrives reading 42.0.
func TestWholeNumbersBecomeInts(t *testing.T) {
	got := Unravel(meowrt.NewString(`{"id": 42}`))
	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %T, want a Map", got)
	}
	if _, ok := m.Items["id"].(*meowrt.Int); !ok {
		t.Errorf("id is %T, want *meowrt.Int", m.Items["id"])
	}
}

func TestFractionsStayFloats(t *testing.T) {
	got := Unravel(meowrt.NewString(`{"ratio": 2.5}`))
	m := got.(*meowrt.Map)
	if _, ok := m.Items["ratio"].(*meowrt.Float); !ok {
		t.Errorf("ratio is %T, want *meowrt.Float", m.Items["ratio"])
	}
}

func TestUnravelNested(t *testing.T) {
	got := Unravel(meowrt.NewString(`{"hits": [{"marker": "m1"}], "count": 1}`))
	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %T, want a Map", got)
	}
	hits, ok := m.Items["hits"].(*meowrt.List)
	if !ok {
		t.Fatalf("hits is %T, want a List", m.Items["hits"])
	}
	if hits.Len() != 1 {
		t.Fatalf("got %d hits, want 1", hits.Len())
	}
	first := hits.Items[0].(*meowrt.Map)
	if first.Items["marker"].String() != "m1" {
		t.Errorf("marker = %q, want m1", first.Items["marker"].String())
	}
}

// Text that is not JSON is a Furball rather than a panic: a reply turns out to
// be an HTML error page from a proxy often enough that a program has to be able
// to recover from it.
func TestUnravelRejectsNonJSON(t *testing.T) {
	tests := []struct {
		name string
		in   meowrt.Value
	}{
		{"html", meowrt.NewString("<html>nope</html>")},
		{"empty", meowrt.NewString("")},
		{"truncated object", meowrt.NewString("{unclosed")},
		{"not a string", meowrt.NewInt(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := Unravel(tt.in).(*meowrt.Furball); !ok {
				t.Error("expected a Furball")
			}
		})
	}
}

func TestWind(t *testing.T) {
	tests := []struct {
		name string
		in   meowrt.Value
		want string
	}{
		{"string", meowrt.NewString("text"), `"text"`},
		{"int", meowrt.NewInt(1), "1"},
		{"bool", meowrt.NewBool(true), "true"},
		{"catnap", meowrt.NewNil(), "null"},
		{"litter", meowrt.NewList(meowrt.NewInt(1), meowrt.NewInt(2)), "[1,2]"},
		{"basket", meowrt.NewMap(map[string]meowrt.Value{"a": meowrt.NewInt(1)}), `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wind(tt.in)
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// Saying so beats writing something that reads back as a different value.
func TestWindRejectsShapelessValues(t *testing.T) {
	got := Wind(&meowrt.Furball{Message: "boom"})
	if _, ok := got.(*meowrt.Furball); !ok {
		t.Error("expected a Furball")
	}
}

// A round trip preserves every value, but not the order the keys were written
// in: a basket is a Go map, which has no order, so the keys come back sorted.
// A litter's order is preserved, being a sequence.
func TestRoundTrip(t *testing.T) {
	original := `{"n":1,"nested":{"s":"two"},"list":[1,2]}`
	want := `{"list":[1,2],"n":1,"nested":{"s":"two"}}`

	got := Wind(Unravel(meowrt.NewString(original)))
	if got.String() != want {
		t.Errorf("got %q, want %q", got.String(), want)
	}
}

// A document may be nested as deeply as its source asks, which would recurse
// until the stack ran out.
func TestUnravelBoundsNesting(t *testing.T) {
	deep := strings.Repeat("[", maxDepth+50) + strings.Repeat("]", maxDepth+50)

	if _, ok := Unravel(meowrt.NewString(deep)).(*meowrt.Furball); !ok {
		t.Error("expected a Furball for a document nested past the bound")
	}
}

func TestArityErrorsAreFurballs(t *testing.T) {
	tests := []struct {
		name string
		got  meowrt.Value
	}{
		{"unravel with no argument", Unravel()},
		{"unravel with two", Unravel(meowrt.NewString("1"), meowrt.NewString("2"))},
		{"wind with no argument", Wind()},
		{"wind with two", Wind(meowrt.NewInt(1), meowrt.NewInt(2))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.got.(*meowrt.Furball); !ok {
				t.Errorf("expected a Furball, got %s", tt.got.String())
			}
		})
	}
}
