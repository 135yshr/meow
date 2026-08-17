package meowrt

import "testing"

// solo drains RangeSolo into the strings a one-variable `purr` would bind.
func solo(v Value) (bound []string) {
	for item := range RangeSolo(v) {
		bound = append(bound, item.String())
	}
	return bound
}

// pair drains RangePair into the two strings a two-variable `purr` would bind.
func pair(v Value) (first, second []string) {
	for a, b := range RangePair(v) {
		first = append(first, a.String())
		second = append(second, b.String())
	}
	return first, second
}

func TestRangeSoloWalksALitterElementwise(t *testing.T) {
	got := solo(NewList(NewString("a"), NewString("b")))

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

// A basket yields keys rather than values, because the value is one lookup
// away and the key is not recoverable from the value.
func TestRangeSoloWalksABasketByKey(t *testing.T) {
	got := solo(NewMap(map[string]Value{"b": NewInt(2), "a": NewInt(1)}))

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

// Anything that is not walkable is read as a count, which is the other written
// form of `purr`. Both have to be served because the form cannot be told apart
// when the subject's type is unknown until the program runs.
func TestRangeSoloReadsANumberAsACount(t *testing.T) {
	got := solo(NewInt(3))

	if len(got) != 3 || got[0] != "0" || got[2] != "2" {
		t.Errorf("got %v, want [0 1 2]", got)
	}
}

func TestRangePairBindsIndexAndElement(t *testing.T) {
	idx, elems := pair(NewList(NewString("a"), NewString("b")))

	if len(idx) != 2 || idx[0] != "0" || idx[1] != "1" {
		t.Errorf("indices = %v, want [0 1]", idx)
	}
	if len(elems) != 2 || elems[0] != "a" || elems[1] != "b" {
		t.Errorf("elements = %v, want [a b]", elems)
	}
}

func TestRangePairBindsKeyAndValue(t *testing.T) {
	keys, vals := pair(NewMap(map[string]Value{"b": NewInt(2), "a": NewInt(1)}))

	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Errorf("keys = %v, want [a b]", keys)
	}
	if len(vals) != 2 || vals[0] != "1" || vals[1] != "2" {
		t.Errorf("values = %v, want [1 2]", vals)
	}
}

// Go randomizes map iteration, so a basket walked in its own order would give
// a program output that differs run to run. Sorting buys a testable program.
func TestBasketWalkIsSortedByKey(t *testing.T) {
	m := NewMap(map[string]Value{
		"delta": NewInt(4), "alpha": NewInt(1), "charlie": NewInt(3), "bravo": NewInt(2),
	})
	want := []string{"alpha", "bravo", "charlie", "delta"}

	// Repeated so a chance agreement with Go's randomized order is not mistaken
	// for the sort working.
	for range 20 {
		got := solo(m)
		for i, k := range want {
			if got[i] != k {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	}
}

func TestRangeEmptyCases(t *testing.T) {
	tests := []struct {
		name string
		v    Value
	}{
		{"empty litter", NewList()},
		{"empty basket", NewMap(map[string]Value{})},
		{"zero count", NewInt(0)},
		{"negative count", NewInt(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := solo(tt.v); len(got) != 0 {
				t.Errorf("RangeSolo gave %v, want no turns", got)
			}
			if first, _ := pair(tt.v); len(first) != 0 {
				t.Errorf("RangePair gave %v, want no turns", first)
			}
		})
	}
}

// A `bring` or an early exit stops the Go loop, which must stop the sequence
// rather than run it to the end.
func TestRangeStopsWhenTheLoopBreaks(t *testing.T) {
	list := NewList(NewString("a"), NewString("b"), NewString("c"))

	seen := 0
	for range RangeSolo(list) {
		seen++
		break
	}
	if seen != 1 {
		t.Errorf("RangeSolo ran %d turns after a break, want 1", seen)
	}

	seen = 0
	for range RangePair(list) {
		seen++
		break
	}
	if seen != 1 {
		t.Errorf("RangePair ran %d turns after a break, want 1", seen)
	}
}
