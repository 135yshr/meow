package meowrt

import (
	"encoding/json"
	"fmt"
	"sort"
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

// NewInt creates a new Int value.
func NewInt(v int64) *Int     { return &Int{Val: v} }
func (i *Int) Type() string   { return "Int" }
func (i *Int) String() string { return fmt.Sprintf("%d", i.Val) }
func (i *Int) IsTruthy() bool { return i.Val != 0 }

// Float represents a floating-point value.
type Float struct{ Val float64 }

// NewFloat creates a new Float value.
func NewFloat(v float64) *Float { return &Float{Val: v} }
func (f *Float) Type() string   { return "Float" }
func (f *Float) String() string { return fmt.Sprintf("%g", f.Val) }
func (f *Float) IsTruthy() bool { return f.Val != 0 }

// String represents a string value.
type String struct{ Val string }

// NewString creates a new String value.
func NewString(v string) *String { return &String{Val: v} }
func (s *String) Type() string   { return "String" }
func (s *String) String() string { return s.Val }
func (s *String) IsTruthy() bool { return s.Val != "" }

// Bool represents a boolean value.
type Bool struct{ Val bool }

// NewBool creates a new Bool value.
func NewBool(v bool) *Bool     { return &Bool{Val: v} }
func (b *Bool) Type() string   { return "Bool" }
func (b *Bool) String() string { return fmt.Sprintf("%t", b.Val) }
func (b *Bool) IsTruthy() bool { return b.Val }

// NilValue represents a nil/catnap value.
type NilValue struct{}

// NewNil creates a new NilValue.
func NewNil() *NilValue            { return &NilValue{} }
func (n *NilValue) Type() string   { return "Nil" }
func (n *NilValue) String() string { return "catnap" }
func (n *NilValue) IsTruthy() bool { return false }

// Func represents a function value.
type Func struct {
	// Name is the function name for display purposes.
	Name string
	// Fn is the underlying Go function implementation.
	Fn func(args ...Value) Value
}

// NewFunc creates a new Func value with the given name and implementation.
// fn must not be nil; passing nil will panic.
func NewFunc(name string, fn func(args ...Value) Value) *Func {
	if fn == nil {
		panic("Hiss! fn must not be nil, nya~")
	}
	return &Func{Name: name, Fn: fn}
}

func (f *Func) Type() string   { return "Func" }
func (f *Func) String() string { return fmt.Sprintf("<meow %s>", f.Name) }
func (f *Func) IsTruthy() bool { return true }

// Call invokes the function with the given arguments.
func (f *Func) Call(args ...Value) Value {
	return f.Fn(args...)
}

// Furball represents an error value caught by Gag.
type Furball struct{ Message string }

func (e *Furball) Type() string   { return "Furball" }
func (e *Furball) String() string { return e.Message }
func (e *Furball) IsTruthy() bool { return false }

// List represents a list value.
type List struct {
	// Items is the slice of values in the list.
	Items []Value
}

// NewList creates a new List value from the given items.
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

// Len returns the number of items in the list.
func (l *List) Len() int {
	return len(l.Items)
}

// Get returns the item at the given index.
// It panics if index is out of range.
func (l *List) Get(index int) Value {
	if index < 0 || index >= len(l.Items) {
		panic(fmt.Sprintf("Hiss! Index %d out of range, nya~", index))
	}
	return l.Items[index]
}

// Map represents a map value with string keys.
type Map struct {
	Items map[string]Value
}

// NewMap creates a new Map value from the given items.
func NewMap(items map[string]Value) *Map {
	return &Map{Items: items}
}

func (m *Map) Type() string   { return "Map" }
func (m *Map) IsTruthy() bool { return len(m.Items) > 0 }
func (m *Map) String() string {
	parts := make([]string, 0, len(m.Items))
	for k, v := range m.Items {
		parts = append(parts, k+": "+v.String())
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// Get returns the value for the given key and whether it was found.
func (m *Map) Get(key string) (Value, bool) {
	v, ok := m.Items[key]
	return v, ok
}

// ToJSON serializes a Value to its JSON string representation.
func ToJSON(v Value) string {
	switch val := v.(type) {
	case *String:
		b, _ := json.Marshal(val.Val)
		return string(b)
	case *Int:
		return fmt.Sprintf("%d", val.Val)
	case *Float:
		return fmt.Sprintf("%g", val.Val)
	case *Bool:
		if val.Val {
			return "true"
		}
		return "false"
	case *NilValue:
		return "null"
	case *List:
		parts := make([]string, len(val.Items))
		for i, item := range val.Items {
			parts[i] = ToJSON(item)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case *Map:
		keys := make([]string, 0, len(val.Items))
		for k := range val.Items {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(val.Items))
		for _, k := range keys {
			kb, _ := json.Marshal(k)
			parts = append(parts, string(kb)+":"+ToJSON(val.Items[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"
	default:
		panic(fmt.Sprintf("Hiss! cannot serialize %s to JSON, nya~", v.Type()))
	}
}

// AsInt extracts an int64 from a Value, panicking with a descriptive message on type mismatch.
func AsInt(v Value) int64 {
	if i, ok := v.(*Int); ok {
		return i.Val
	}
	panic(fmt.Sprintf("Hiss! expected int but got %s, nya~", v.Type()))
}

// AsFloat extracts a float64 from a Value, panicking with a descriptive message on type mismatch.
func AsFloat(v Value) float64 {
	if f, ok := v.(*Float); ok {
		return f.Val
	}
	panic(fmt.Sprintf("Hiss! expected float but got %s, nya~", v.Type()))
}

// AsString extracts a string from a Value, panicking with a descriptive message on type mismatch.
func AsString(v Value) string {
	if s, ok := v.(*String); ok {
		return s.Val
	}
	panic(fmt.Sprintf("Hiss! expected string but got %s, nya~", v.Type()))
}

// AsBool extracts a bool from a Value, panicking with a descriptive message on type mismatch.
func AsBool(v Value) bool {
	if b, ok := v.(*Bool); ok {
		return b.Val
	}
	panic(fmt.Sprintf("Hiss! expected bool but got %s, nya~", v.Type()))
}
