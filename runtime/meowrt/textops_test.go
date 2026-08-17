package meowrt

import (
	"math"
	"strings"
	"testing"
)

func TestUpperLowerTrim(t *testing.T) {
	tests := []struct {
		name string
		got  Value
		want string
	}{
		{"upper", Upper(NewString("cat")), "CAT"},
		{"upper of empty", Upper(NewString("")), ""},
		{"lower", Lower(NewString("CAT")), "cat"},
		{"trim spaces", Trim(NewString("  cat  ")), "cat"},
		{"trim tabs and newlines", Trim(NewString("\t cat \n")), "cat"},
		{"trim of blank", Trim(NewString("   ")), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.String() != tt.want {
				t.Errorf("got %q, want %q", tt.got.String(), tt.want)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	// Every occurrence, not just the first.
	got := Replace(NewString("a-b-c"), NewString("-"), NewString("+"))
	if got.String() != "a+b+c" {
		t.Errorf("got %q, want a+b+c", got.String())
	}
}

// An empty search string would have ReplaceAll insert the replacement between
// every character, which is never what was meant.
func TestReplaceRejectsAnEmptySearch(t *testing.T) {
	if _, ok := Replace(NewString("cat"), NewString(""), NewString("x")).(*Furball); !ok {
		t.Error("expected a Furball")
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		width int64
		want  string
	}{
		{"right", "ab", 5, "ab   "},
		{"left", "ab", -5, "   ab"},
		{"exactly wide enough", "abc", 3, "abc"},
		{"already wider is kept whole", "abcdef", 3, "abcdef"},
		{"zero width", "ab", 0, "ab"},
		// Counted in characters, as nibble and track are, so a column lines up
		// for text that is not all ASCII.
		{"counts characters not bytes", "あい", 4, "あい  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Pad(NewString(tt.s), NewInt(tt.width))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestPadRejectsAnAbsurdWidth(t *testing.T) {
	if _, ok := Pad(NewString("a"), NewInt(math.MaxInt64)).(*Furball); !ok {
		t.Error("expected a Furball")
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		name string
		in   *List
		want string
	}{
		{"ints", NewList(NewInt(3), NewInt(1), NewInt(2)), "[1, 2, 3]"},
		{"strings", NewList(NewString("pear"), NewString("apple")), "[apple, pear]"},
		{"floats", NewList(NewFloat(2.5), NewFloat(1.5)), "[1.5, 2.5]"},
		{"ints and floats together", NewList(NewFloat(1.5), NewInt(1)), "[1, 1.5]"},
		{"empty", NewList(), "[]"},
		{"one element", NewList(NewString("only")), "[only]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sort(tt.in); got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// Sorting must not disturb the litter it was given, which a program may still
// be holding.
func TestSortLeavesTheOriginalAlone(t *testing.T) {
	original := NewList(NewInt(3), NewInt(1), NewInt(2))

	Sort(original)

	if original.String() != "[3, 1, 2]" {
		t.Errorf("original became %q", original.String())
	}
}

// There is no answer to "is 1 before \"a\"", and inventing one would put a
// program's output at the mercy of which happened to come first.
func TestSortRefusesWhatItCannotOrder(t *testing.T) {
	tests := []struct {
		name string
		in   Value
	}{
		{"numbers and strings", NewList(NewInt(1), NewString("a"))},
		{"baskets", NewList(NewMap(map[string]Value{}), NewMap(map[string]Value{}))},
		{"not a litter", NewString("cat")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := Sort(tt.in).(*Furball); !ok {
				t.Error("expected a Furball")
			}
		})
	}
}

func TestReverse(t *testing.T) {
	if got := Reverse(NewList(NewInt(1), NewInt(2), NewInt(3))); got.String() != "[3, 2, 1]" {
		t.Errorf("got %q, want [3, 2, 1]", got.String())
	}
	if got := Reverse(NewList()); got.String() != "[]" {
		t.Errorf("got %q, want []", got.String())
	}
	if _, ok := Reverse(NewString("cat")).(*Furball); !ok {
		t.Error("expected a Furball for a non-litter")
	}
}

func TestReverseLeavesTheOriginalAlone(t *testing.T) {
	original := NewList(NewInt(1), NewInt(2))

	Reverse(original)

	if original.String() != "[1, 2]" {
		t.Errorf("original became %q", original.String())
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		name   string
		x      Value
		places int64
		want   string
	}{
		{"two places", NewFloat(3.14159), 2, "3.14"},
		{"no places", NewFloat(3.14159), 0, "3"},
		// Half away from zero, the arithmetic convention, rather than Go's
		// half-to-even: a reader expects 2.5 to print as 3.
		{"half up", NewFloat(2.5), 0, "3"},
		{"half up from an odd number", NewFloat(3.5), 0, "4"},
		{"half away from zero when negative", NewFloat(-2.5), 0, "-3"},
		// A whole number is already rounded, and stays an Int rather than
		// printing 42.0.
		{"an int is unchanged", NewInt(42), 2, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Round(tt.x, NewInt(tt.places))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestRoundRejectsPlacesOutOfRange(t *testing.T) {
	for _, places := range []int64{-1, maxRoundPlaces + 1} {
		if _, ok := Round(NewFloat(1.5), NewInt(places)).(*Furball); !ok {
			t.Errorf("expected a Furball for %d places", places)
		}
	}
}

func TestRoundRejectsANonNumber(t *testing.T) {
	if _, ok := Round(NewString("cat"), NewInt(2)).(*Furball); !ok {
		t.Error("expected a Furball")
	}
}

// A Furball handed in comes back out, so a failure earlier in a chain is not
// swallowed by the helper it flows into.
func TestHelpersPropagateFurballs(t *testing.T) {
	boom := &Furball{Message: "boom"}
	tests := []struct {
		name string
		got  Value
	}{
		{"upper", Upper(boom)},
		{"lower", Lower(boom)},
		{"trim", Trim(boom)},
		{"replace", Replace(boom, NewString("a"), NewString("b"))},
		{"pad", Pad(boom, NewInt(2))},
		{"sort", Sort(boom)},
		{"reverse", Reverse(boom)},
		{"round", Round(boom, NewInt(2))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ok := tt.got.(*Furball)
			if !ok {
				t.Fatalf("got %T, want a Furball", tt.got)
			}
			if !strings.Contains(f.Message, "boom") {
				t.Errorf("got %q, want the original message", f.Message)
			}
		})
	}
}
