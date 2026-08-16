package meowrt

import "testing"

// collect drains a RangeSeq into parallel slices, so the assertions below can
// speak about what the loop would have bound on each turn.
func collect(v Value) (indices []int, items []string) {
	for i, item := range RangeSeq(v) {
		indices = append(indices, i)
		items = append(items, item.String())
	}
	return indices, items
}

func TestRangeSeqWalksAListElementwise(t *testing.T) {
	idx, items := collect(NewList(NewString("a"), NewString("b")))

	if len(items) != 2 || items[0] != "a" || items[1] != "b" {
		t.Errorf("items = %v, want [a b]", items)
	}
	if len(idx) != 2 || idx[0] != 0 || idx[1] != 1 {
		t.Errorf("indices = %v, want [0 1]", idx)
	}
}

// Anything that is not a litter is read as a count, which is the other written
// form of `purr`. Both have to be served because the form cannot be told apart
// when the subject's type is unknown until the program runs.
func TestRangeSeqWalksANumberAsACount(t *testing.T) {
	idx, items := collect(NewInt(3))

	if len(items) != 3 || items[0] != "0" || items[2] != "2" {
		t.Errorf("items = %v, want [0 1 2]", items)
	}
	if len(idx) != 3 || idx[2] != 2 {
		t.Errorf("indices = %v, want [0 1 2]", idx)
	}
}

func TestRangeSeqEmptyCases(t *testing.T) {
	tests := []struct {
		name string
		v    Value
	}{
		{"empty litter", NewList()},
		{"zero count", NewInt(0)},
		{"negative count", NewInt(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, items := collect(tt.v); len(items) != 0 {
				t.Errorf("got %v, want no turns", items)
			}
		})
	}
}

// A `bring` or an early exit stops the Go loop, which must stop the sequence
// rather than run it to the end.
func TestRangeSeqStopsWhenTheLoopBreaks(t *testing.T) {
	seen := 0
	for range RangeSeq(NewList(NewString("a"), NewString("b"), NewString("c"))) {
		seen++
		break
	}
	if seen != 1 {
		t.Errorf("ran %d turns after a break, want 1", seen)
	}
}
