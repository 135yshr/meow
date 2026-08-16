package meowrt

import "iter"

// RangeSeq yields the pairs a `purr` walks over a value whose kind is not known
// until the program runs.
//
// A litter is walked element by element; anything else is read as a count and
// walked from zero, which is what the two written forms of `purr` mean. The
// choice has to be made here rather than while generating code because an
// expression like a call's result or a map lookup carries no static type — and
// guessing one of the two forms turns the other into a wrong answer.
//
// It matches what the playground interpreter does with the same program, so the
// two backends agree.
func RangeSeq(v Value) iter.Seq2[int, Value] {
	return func(yield func(int, Value) bool) {
		if list, ok := v.(*List); ok {
			for i, item := range list.Items {
				if !yield(i, item) {
					return
				}
			}
			return
		}
		n := AsInt(v)
		for i := int64(0); i < n; i++ {
			if !yield(int(i), NewInt(i)) {
				return
			}
		}
	}
}
