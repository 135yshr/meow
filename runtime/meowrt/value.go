package meowrt

import (
	"fmt"
	"strings"
)

// Value is the core interface for all Meow values.
type Value interface {
	Type() string
	String() string
	IsTruthy() bool
}

// Int represents an integer value.
type Int struct{ Val int64 }

func NewInt(v int64) *Int       { return &Int{Val: v} }
func (i *Int) Type() string     { return "Int" }
func (i *Int) String() string   { return fmt.Sprintf("%d", i.Val) }
func (i *Int) IsTruthy() bool   { return i.Val != 0 }

// Float represents a floating-point value.
type Float struct{ Val float64 }

func NewFloat(v float64) *Float   { return &Float{Val: v} }
func (f *Float) Type() string     { return "Float" }
func (f *Float) String() string   { return fmt.Sprintf("%g", f.Val) }
func (f *Float) IsTruthy() bool   { return f.Val != 0 }

// String represents a string value.
type String struct{ Val string }

func NewString(v string) *String  { return &String{Val: v} }
func (s *String) Type() string    { return "String" }
func (s *String) String() string  { return s.Val }
func (s *String) IsTruthy() bool  { return s.Val != "" }

// Bool represents a boolean value.
type Bool struct{ Val bool }

func NewBool(v bool) *Bool      { return &Bool{Val: v} }
func (b *Bool) Type() string    { return "Bool" }
func (b *Bool) String() string  { return fmt.Sprintf("%t", b.Val) }
func (b *Bool) IsTruthy() bool  { return b.Val }

// NilValue represents a nil/catnap value.
type NilValue struct{}

func NewNil() *NilValue           { return &NilValue{} }
func (n *NilValue) Type() string  { return "Nil" }
func (n *NilValue) String() string { return "catnap" }
func (n *NilValue) IsTruthy() bool { return false }

// Func represents a function value.
type Func struct {
	Name string
	Fn   func(args ...Value) Value
}

func NewFunc(name string, fn func(args ...Value) Value) *Func {
	return &Func{Name: name, Fn: fn}
}

func (f *Func) Type() string    { return "Func" }
func (f *Func) String() string  { return fmt.Sprintf("<meow %s>", f.Name) }
func (f *Func) IsTruthy() bool  { return true }

func (f *Func) Call(args ...Value) Value {
	return f.Fn(args...)
}

// List represents a list value.
type List struct {
	Items []Value
}

func NewList(items ...Value) *List {
	return &List{Items: items}
}

func (l *List) Type() string   { return "List" }
func (l *List) IsTruthy() bool { return len(l.Items) > 0 }
func (l *List) String() string {
	parts := make([]string, len(l.Items))
	for i, v := range l.Items {
		parts[i] = v.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (l *List) Len() int {
	return len(l.Items)
}

func (l *List) Get(index int) Value {
	if index < 0 || index >= len(l.Items) {
		panic(fmt.Sprintf("Hiss! Index %d out of range, nya~", index))
	}
	return l.Items[index]
}
