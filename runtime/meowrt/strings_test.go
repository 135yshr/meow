package meowrt_test

import (
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
)

func TestWhiff(t *testing.T) {
	tests := []struct {
		name string
		s    string
		sub  string
		want string
	}{
		{"present", "hello,world", "world", "true"},
		{"absent", "hello,world", "dog", "false"},
		{"empty needle", "hello", "", "true"},
		{"empty haystack", "", "x", "false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Whiff(meowrt.NewString(tt.s), meowrt.NewString(tt.sub))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestTrack(t *testing.T) {
	tests := []struct {
		name string
		s    string
		sub  string
		want string
	}{
		{"found", "hello,world", "world", "6"},
		{"at start", "hello", "he", "0"},
		{"absent reports -1", "hello", "dog", "-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Track(meowrt.NewString(tt.s), meowrt.NewString(tt.sub))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestShred(t *testing.T) {
	tests := []struct {
		name string
		s    string
		sep  string
		want string
	}{
		{"comma separated", "a,b,c", ",", "[a, b, c]"},
		{"separator absent", "abc", ",", "[abc]"},
		{"empty separator splits characters", "abc", "", "[a, b, c]"},
		{"empty input", "", ",", "[]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Shred(meowrt.NewString(tt.s), meowrt.NewString(tt.sep))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// Track and Nibble must agree on their unit, so that the offset Track reports
// can be handed straight to Nibble. Byte offsets would break that on
// multi-byte text.
func TestTrackAgreesWithNibble(t *testing.T) {
	s := meowrt.NewString("にゃんこ meow")
	sub := meowrt.NewString("meow")

	at := meowrt.Track(s, sub)
	if at.String() != "5" {
		t.Fatalf("expected a character offset of 5, got %s", at.String())
	}
	got := meowrt.Nibble(s, at, meowrt.NewInt(9))
	if got.String() != "meow" {
		t.Errorf("got %q, want %q", got.String(), "meow")
	}
}

func TestTangleInvertsShred(t *testing.T) {
	original := "a,b,c"
	parts := meowrt.Shred(meowrt.NewString(original), meowrt.NewString(","))
	got := meowrt.Tangle(parts, meowrt.NewString(","))
	if got.String() != original {
		t.Errorf("got %q, want %q", got.String(), original)
	}
}

func TestTangleRejectsNonStringElements(t *testing.T) {
	list := meowrt.NewList(meowrt.NewString("a"), meowrt.NewInt(1))
	if _, ok := meowrt.Tangle(list, meowrt.NewString(",")).(*meowrt.Furball); !ok {
		t.Error("expected a Furball for a non-String element")
	}
}

func TestNibble(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		start, end int64
		want       string
	}{
		{"prefix", "hello,world", 0, 5, "hello"},
		{"middle", "hello,world", 6, 11, "world"},
		{"negative start counts from the end", "hello,world", -5, 11, "world"},
		{"end clamped past the length", "hello", 0, 99, "hello"},
		{"start clamped before zero", "hello", -99, 2, "he"},
		{"inverted range is empty", "hello", 4, 2, ""},
		{"equal bounds are empty", "hello", 2, 2, ""},
		// Counted in characters, so multi-byte text behaves the way it reads.
		{"multibyte", "にゃんこ", 1, 3, "ゃん"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meowrt.Nibble(meowrt.NewString(tt.s), meowrt.NewInt(tt.start), meowrt.NewInt(tt.end))
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestStringBuiltinsRejectNonStrings(t *testing.T) {
	num := meowrt.NewInt(1)
	str := meowrt.NewString("a")

	tests := []struct {
		name string
		got  meowrt.Value
	}{
		{"whiff haystack", meowrt.Whiff(num, str)},
		{"whiff needle", meowrt.Whiff(str, num)},
		{"track haystack", meowrt.Track(num, str)},
		{"shred value", meowrt.Shred(num, str)},
		{"tangle list", meowrt.Tangle(num, str)},
		{"nibble value", meowrt.Nibble(num, meowrt.NewInt(0), meowrt.NewInt(1))},
		{"nibble bounds", meowrt.Nibble(str, str, meowrt.NewInt(1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.got.(*meowrt.Furball); !ok {
				t.Errorf("expected a Furball, got %s", tt.got.String())
			}
		})
	}
}

func TestStringBuiltinsPropagateFurball(t *testing.T) {
	boom := meowrt.NewFurball("Hiss! boom, nya~")
	str := meowrt.NewString("a")

	for name, got := range map[string]meowrt.Value{
		"whiff":  meowrt.Whiff(boom, str),
		"track":  meowrt.Track(str, boom),
		"shred":  meowrt.Shred(boom, str),
		"tangle": meowrt.Tangle(boom, str),
		"nibble": meowrt.Nibble(boom, meowrt.NewInt(0), meowrt.NewInt(1)),
	} {
		if got != meowrt.Value(boom) {
			t.Errorf("%s: expected the Furball to propagate, got %s", name, got.String())
		}
	}
}

// to_string reassembles a Byte list, making it the inverse of to_bytes.
func TestToStringInvertsToBytes(t *testing.T) {
	original := "hello"
	bytes := meowrt.ToBytes(meowrt.NewString(original))
	if got := meowrt.ToString(bytes); got.String() != original {
		t.Errorf("got %q, want %q", got.String(), original)
	}
}

// Only Byte lists are reassembled; every other list keeps its display form.
func TestToStringLeavesOtherListsAlone(t *testing.T) {
	list := meowrt.NewList(meowrt.NewInt(1), meowrt.NewInt(2))
	if got := meowrt.ToString(list); got.String() != "[1, 2]" {
		t.Errorf("got %q, want %q", got.String(), "[1, 2]")
	}
	if got := meowrt.ToString(meowrt.NewList()); got.String() != "[]" {
		t.Errorf("got %q, want %q", got.String(), "[]")
	}
}
