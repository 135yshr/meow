package meowrt

import (
	"fmt"
	"iter"
)

// Iter returns an iterator over the list items.
func (l *List) Iter() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		for _, item := range l.Items {
			if !yield(item) {
				return
			}
		}
	}
}

// requireList returns the value as *List, or a Furball if it's not a list.
// If the value is itself a Furball, that Furball is returned for propagation.
func requireList(name string, v Value) (*List, *Furball) {
	if f, ok := v.(*Furball); ok {
		return nil, f
	}
	l, ok := v.(*List)
	if !ok {
		return nil, &Furball{Message: fmt.Sprintf("Hiss! %s requires a List, got %s, nya~", name, v.Type())}
	}
	return l, nil
}

// requireFunc returns the value as *Func, or a Furball if it's not callable.
func requireFunc(name string, v Value) (*Func, *Furball) {
	if f, ok := v.(*Furball); ok {
		return nil, f
	}
	fn, ok := v.(*Func)
	if !ok {
		return nil, &Furball{Message: fmt.Sprintf("Hiss! %s requires a Func, got %s, nya~", name, v.Type())}
	}
	return fn, nil
}

// Lick maps a function over a list (like map).
func Lick(lst Value, fn Value) Value {
	l, fb := requireList("lick", lst)
	if fb != nil {
		return fb
	}
	f, fb := requireFunc("lick", fn)
	if fb != nil {
		return fb
	}
	result := make([]Value, 0, l.Len())
	for v := range l.Iter() {
		r := f.Call(v)
		if rf, ok := r.(*Furball); ok {
			return rf
		}
		result = append(result, r)
	}
	return NewList(result...)
}

// Picky filters a list (like filter).
func Picky(lst Value, fn Value) Value {
	l, fb := requireList("picky", lst)
	if fb != nil {
		return fb
	}
	f, fb := requireFunc("picky", fn)
	if fb != nil {
		return fb
	}
	result := make([]Value, 0)
	for v := range l.Iter() {
		r := f.Call(v)
		if rf, ok := r.(*Furball); ok {
			return rf
		}
		if r.IsTruthy() {
			result = append(result, v)
		}
	}
	return NewList(result...)
}

// Curl reduces a list (like fold/reduce).
func Curl(lst Value, init Value, fn Value) Value {
	l, fb := requireList("curl", lst)
	if fb != nil {
		return fb
	}
	if f, ok := init.(*Furball); ok {
		return f
	}
	f, fb := requireFunc("curl", fn)
	if fb != nil {
		return fb
	}
	acc := init
	for v := range l.Iter() {
		acc = f.Call(acc, v)
		if af, ok := acc.(*Furball); ok {
			return af
		}
	}
	return acc
}

// Append appends a value to a list, returning a new list.
func Append(lst Value, v Value) Value {
	l, fb := requireList("append", lst)
	if fb != nil {
		return fb
	}
	if f, ok := v.(*Furball); ok {
		return f
	}
	items := make([]Value, len(l.Items)+1)
	copy(items, l.Items)
	items[len(l.Items)] = v
	return NewList(items...)
}

// Head returns the first element of a list.
func Head(lst Value) Value {
	l, fb := requireList("head", lst)
	if fb != nil {
		return fb
	}
	if l.Len() == 0 {
		return NewNil()
	}
	return l.Items[0]
}

// Tail returns all elements except the first.
func Tail(lst Value) Value {
	l, fb := requireList("tail", lst)
	if fb != nil {
		return fb
	}
	if l.Len() <= 1 {
		return NewList()
	}
	return NewList(l.Items[1:]...)
}
