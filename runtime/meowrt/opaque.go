package meowrt

import "fmt"

// Opaque carries a Go value that Meow has no shape for.
//
// A Go library is rarely a set of functions over plain data. It hands back a
// client, a connection, a handle — something to call the next function on. With
// nowhere to put such a thing, every library had to be wrapped by hand, in Go,
// before Meow could reach it. Held opaquely it can be passed straight back into
// the next call, and the wrapping stops being necessary.
//
// Meow can do nothing with what is inside but carry it. That is the point: it
// does not have to understand a *sts.Client to hold one.
type Opaque struct {
	// What names the thing for a reader, since its own type name is Go's and
	// often says little.
	What string
	V    any
}

// NewOpaque wraps a Go value so Meow can hold it.
func NewOpaque(what string, v any) *Opaque { return &Opaque{What: what, V: v} }

func (o *Opaque) Type() string { return "Opaque" }

func (o *Opaque) String() string {
	if o.What != "" {
		return fmt.Sprintf("<%s>", o.What)
	}
	return fmt.Sprintf("<%T>", o.V)
}

// IsTruthy reports whether anything is being held. A handle that failed to be
// made is nothing, and reads as false.
func (o *Opaque) IsTruthy() bool { return o.V != nil }

// AsOpaque reads v as a held Go value.
func AsOpaque(v Value) (*Opaque, bool) {
	o, ok := v.(*Opaque)
	return o, ok
}
