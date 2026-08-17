package meowrt

import (
	"iter"
	"sort"
)

// sortedKeys returns a basket's keys in order.
//
// Go randomizes map iteration, so walking a basket in its own order would make
// a program's output differ run to run. Sorting costs a little and buys a
// program that can be tested.
func sortedKeys(m *Map) []string {
	keys := make([]string, 0, len(m.Items))
	for k := range m.Items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// RangeSolo yields what the one-variable `purr x (v)` binds on each turn:
// a litter's elements, a basket's keys, or the numbers counted up to.
//
// A basket yields keys rather than values because that is what a program has
// to have — the value is one lookup away, and the key is not recoverable from
// the value.
func RangeSolo(v Value) iter.Seq[Value] {
	return func(yield func(Value) bool) {
		switch v := v.(type) {
		case *List:
			for _, item := range v.Items {
				if !yield(item) {
					return
				}
			}
		case *Map:
			for _, k := range sortedKeys(v) {
				if !yield(NewString(k)) {
					return
				}
			}
		default:
			n := AsInt(v)
			for i := int64(0); i < n; i++ {
				if !yield(NewInt(i)) {
					return
				}
			}
		}
	}
}

// RangePair yields what the two-variable `purr a, b (v)` binds: a litter's
// index and element, or a basket's key and value.
//
// Counting is offered too, so that a subject whose kind is only known at run
// time has an answer in either form; both variables take the counter, as there
// is nothing else to give them.
func RangePair(v Value) iter.Seq2[Value, Value] {
	return func(yield func(Value, Value) bool) {
		switch v := v.(type) {
		case *List:
			for i, item := range v.Items {
				if !yield(NewInt(int64(i)), item) {
					return
				}
			}
		case *Map:
			for _, k := range sortedKeys(v) {
				if !yield(NewString(k), v.Items[k]) {
					return
				}
			}
		default:
			n := AsInt(v)
			for i := int64(0); i < n; i++ {
				if !yield(NewInt(i), NewInt(i)) {
					return
				}
			}
		}
	}
}
