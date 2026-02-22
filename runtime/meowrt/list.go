package meowrt

import "iter"

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

// Lick maps a function over a list (like map).
func Lick(lst Value, fn Value) Value {
	l, ok := lst.(*List)
	if !ok {
		panic("Hiss! lick requires a List, nya~")
	}
	f := fn.(*Func)
	result := make([]Value, 0, l.Len())
	for v := range l.Iter() {
		result = append(result, f.Call(v))
	}
	return NewList(result...)
}

// Picky filters a list (like filter).
func Picky(lst Value, fn Value) Value {
	l, ok := lst.(*List)
	if !ok {
		panic("Hiss! picky requires a List, nya~")
	}
	f := fn.(*Func)
	result := make([]Value, 0)
	for v := range l.Iter() {
		if f.Call(v).IsTruthy() {
			result = append(result, v)
		}
	}
	return NewList(result...)
}

// Curl reduces a list (like fold/reduce).
func Curl(lst Value, init Value, fn Value) Value {
	l, ok := lst.(*List)
	if !ok {
		panic("Hiss! curl requires a List, nya~")
	}
	f := fn.(*Func)
	acc := init
	for v := range l.Iter() {
		acc = f.Call(acc, v)
	}
	return acc
}

// Append appends a value to a list, returning a new list.
func Append(lst Value, v Value) Value {
	l, ok := lst.(*List)
	if !ok {
		panic("Hiss! append requires a List, nya~")
	}
	items := make([]Value, len(l.Items)+1)
	copy(items, l.Items)
	items[len(l.Items)] = v
	return NewList(items...)
}

// Head returns the first element of a list.
func Head(lst Value) Value {
	l, ok := lst.(*List)
	if !ok {
		panic("Hiss! head requires a List, nya~")
	}
	if l.Len() == 0 {
		return NewNil()
	}
	return l.Items[0]
}

// Tail returns all elements except the first.
func Tail(lst Value) Value {
	l, ok := lst.(*List)
	if !ok {
		panic("Hiss! tail requires a List, nya~")
	}
	if l.Len() <= 1 {
		return NewList()
	}
	return NewList(l.Items[1:]...)
}
