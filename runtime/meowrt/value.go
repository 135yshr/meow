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

// Byte represents a byte value.
type Byte struct{ Val byte }

// NewByte creates a new Byte value.
func NewByte(v byte) *Byte     { return &Byte{Val: v} }
func (b *Byte) Type() string   { return "Byte" }
func (b *Byte) String() string { return fmt.Sprintf("%d", b.Val) }
func (b *Byte) IsTruthy() bool { return b.Val != 0 }

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
	// Arity is the number of expected arguments. -1 means variadic (no currying).
	Arity int
}

// NewFunc creates a new Func value with the given name and implementation.
// The function is variadic (Arity -1) and will not be auto-curried.
// Passing a nil fn is a programmer error (not a runtime Meow error),
// so it still panics — this is internal API contract, not Meow-level semantics.
func NewFunc(name string, fn func(args ...Value) Value) *Func {
	if fn == nil {
		panic("meowrt: NewFunc called with nil fn (internal contract violation)")
	}
	return &Func{Name: name, Fn: fn, Arity: -1}
}

// NewFuncWithArity creates a new Func value with a fixed arity.
// When called with fewer arguments than arity, the function is automatically
// partially applied (curried).
// Passing a nil fn is a programmer error (internal contract violation).
func NewFuncWithArity(name string, arity int, fn func(args ...Value) Value) *Func {
	if fn == nil {
		panic("meowrt: NewFuncWithArity called with nil fn (internal contract violation)")
	}
	return &Func{Name: name, Fn: fn, Arity: arity}
}

// PartialApply creates a new function that captures the given arguments and
// waits for the remaining ones. The returned function's arity is reduced
// by the number of supplied arguments.
func PartialApply(fn *Func, args ...Value) *Func {
	remaining := fn.Arity - len(args)
	name := fmt.Sprintf("<partial %s %d/%d>", fn.Name, len(args), fn.Arity)
	return NewFuncWithArity(name, remaining, func(moreArgs ...Value) Value {
		allArgs := make([]Value, 0, len(args)+len(moreArgs))
		allArgs = append(allArgs, args...)
		allArgs = append(allArgs, moreArgs...)
		if fn.Arity > 0 && len(allArgs) < fn.Arity {
			return PartialApply(fn, allArgs...)
		}
		return fn.Fn(allArgs...)
	})
}

func (f *Func) Type() string   { return "Func" }
func (f *Func) String() string { return fmt.Sprintf("<meow %s>", f.Name) }
func (f *Func) IsTruthy() bool { return true }

// Call invokes the function with the given arguments.
func (f *Func) Call(args ...Value) Value {
	return f.Fn(args...)
}

// Furball represents an error value. Errors raised by hiss(), failed runtime
// helpers, and propagated values all use *Furball.
//
// Handled marks a Furball that has been intentionally surfaced to user code
// via gag/~> — such Furballs are still reported by is_furball but do NOT
// trigger codegen's automatic propagation short-circuit. This is the bridge
// that lets user code inspect a caught error as a plain value without the
// surrounding statement immediately re-propagating it.
type Furball struct {
	Message string
	Handled bool
}

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

// Get returns the item at the given index. Returns a Furball if the index
// is out of range, propagated as a value through the runtime.
func (l *List) Get(index int) Value {
	if index < 0 || index >= len(l.Items) {
		return &Furball{Message: fmt.Sprintf("Hiss! Index %d out of range, nya~", index)}
	}
	return l.Items[index]
}

// Kitty represents a user-defined struct value.
type Kitty struct {
	TypeName   string
	FieldNames []string
	Fields     map[string]Value
}

// NewKitty creates a new Kitty value, returning a Value that is either
// the constructed Kitty or a Furball if the argument count does not match
// the field count (or if any argument is itself a Furball).
func NewKitty(typeName string, fieldNames []string, args ...Value) Value {
	if f := propagate(args...); f != nil {
		return f
	}
	if len(args) != len(fieldNames) {
		return &Furball{Message: fmt.Sprintf("Hiss! %s expects %d fields but got %d, nya~", typeName, len(fieldNames), len(args))}
	}
	fields := make(map[string]Value, len(fieldNames))
	for i, name := range fieldNames {
		fields[name] = args[i]
	}
	return &Kitty{TypeName: typeName, FieldNames: fieldNames, Fields: fields}
}

func (k *Kitty) Type() string   { return k.TypeName }
func (k *Kitty) IsTruthy() bool { return true }
func (k *Kitty) String() string {
	parts := make([]string, len(k.FieldNames))
	for i, name := range k.FieldNames {
		parts[i] = name + ": " + k.Fields[name].String()
	}
	return k.TypeName + "{" + strings.Join(parts, ", ") + "}"
}

// GetField returns the value of a field by name. Returns a Furball if the
// field does not exist.
func (k *Kitty) GetField(name string) Value {
	v, ok := k.Fields[name]
	if !ok {
		return &Furball{Message: fmt.Sprintf("Hiss! %s has no field %s, nya~", k.TypeName, name)}
	}
	return v
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
// On failure (unsupported type or Furball input) it returns the JSON-encoded
// error message string, prefixed with the Hiss marker. Callers that need to
// distinguish errors from valid JSON should use ToJSONValue.
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
	case *Kitty:
		parts := make([]string, len(val.FieldNames))
		for i, name := range val.FieldNames {
			kb, _ := json.Marshal(name)
			parts[i] = string(kb) + ":" + ToJSON(val.Fields[name])
		}
		return "{" + strings.Join(parts, ",") + "}"
	case *Furball:
		b, _ := json.Marshal(val.Message)
		return string(b)
	default:
		b, _ := json.Marshal(fmt.Sprintf("Hiss! cannot serialize %s to JSON, nya~", v.Type()))
		return string(b)
	}
}

// TryAsInt extracts an int64 from a Value, returning a Furball on type mismatch.
func TryAsInt(v Value) (int64, *Furball) {
	if f, ok := v.(*Furball); ok {
		return 0, f
	}
	if i, ok := v.(*Int); ok {
		return i.Val, nil
	}
	if v == nil {
		return 0, &Furball{Message: "Hiss! expected int but got nil, nya~"}
	}
	return 0, &Furball{Message: fmt.Sprintf("Hiss! expected int but got %s, nya~", v.Type())}
}

// TryAsByte extracts a byte from a Value, returning a Furball on type mismatch.
func TryAsByte(v Value) (byte, *Furball) {
	if f, ok := v.(*Furball); ok {
		return 0, f
	}
	if b, ok := v.(*Byte); ok {
		return b.Val, nil
	}
	if v == nil {
		return 0, &Furball{Message: "Hiss! expected byte but got nil, nya~"}
	}
	return 0, &Furball{Message: fmt.Sprintf("Hiss! expected byte but got %s, nya~", v.Type())}
}

// TryAsFloat extracts a float64 from a Value, returning a Furball on type mismatch.
func TryAsFloat(v Value) (float64, *Furball) {
	if fb, ok := v.(*Furball); ok {
		return 0, fb
	}
	if f, ok := v.(*Float); ok {
		return f.Val, nil
	}
	if v == nil {
		return 0, &Furball{Message: "Hiss! expected float but got nil, nya~"}
	}
	return 0, &Furball{Message: fmt.Sprintf("Hiss! expected float but got %s, nya~", v.Type())}
}

// TryAsString extracts a string from a Value, returning a Furball on type mismatch.
func TryAsString(v Value) (string, *Furball) {
	if f, ok := v.(*Furball); ok {
		return "", f
	}
	if s, ok := v.(*String); ok {
		return s.Val, nil
	}
	if v == nil {
		return "", &Furball{Message: "Hiss! expected string but got nil, nya~"}
	}
	return "", &Furball{Message: fmt.Sprintf("Hiss! expected string but got %s, nya~", v.Type())}
}

// TryAsList extracts a *List from a Value, returning a Furball on type mismatch.
func TryAsList(v Value) (*List, *Furball) {
	if f, ok := v.(*Furball); ok {
		return nil, f
	}
	if l, ok := v.(*List); ok {
		return l, nil
	}
	if v == nil {
		return nil, &Furball{Message: "Hiss! expected list but got nil, nya~"}
	}
	return nil, &Furball{Message: fmt.Sprintf("Hiss! expected list but got %s, nya~", v.Type())}
}

// TryAsBool extracts a bool from a Value, returning a Furball on type mismatch.
func TryAsBool(v Value) (bool, *Furball) {
	if f, ok := v.(*Furball); ok {
		return false, f
	}
	if b, ok := v.(*Bool); ok {
		return b.Val, nil
	}
	if v == nil {
		return false, &Furball{Message: "Hiss! expected bool but got nil, nya~"}
	}
	return false, &Furball{Message: fmt.Sprintf("Hiss! expected bool but got %s, nya~", v.Type())}
}

// AsInt extracts an int64 from a Value. A Furball input panics with the
// Hiss message so callers that haven't short-circuited can't silently
// receive a zero value — Gag's deferred recover converts the panic back
// into a Furball at the typed-path boundary. Non-Furball type mismatch
// also panics for the same loud-failure reason.
func AsInt(v Value) int64 {
	n, f := TryAsInt(v)
	if f != nil {
		panic(f.Message)
	}
	return n
}

// AsByte extracts a byte from a Value. Panics on Furball input or type
// mismatch (see AsInt for the rationale).
func AsByte(v Value) byte {
	n, f := TryAsByte(v)
	if f != nil {
		panic(f.Message)
	}
	return n
}

// AsFloat extracts a float64 from a Value. Panics on Furball input or
// type mismatch (see AsInt for the rationale).
func AsFloat(v Value) float64 {
	n, f := TryAsFloat(v)
	if f != nil {
		panic(f.Message)
	}
	return n
}

// AsString extracts a string from a Value. Panics on Furball input or
// type mismatch (see AsInt for the rationale).
func AsString(v Value) string {
	s, f := TryAsString(v)
	if f != nil {
		panic(f.Message)
	}
	return s
}

// AsList extracts a *List from a Value. Panics on Furball input or type
// mismatch (see AsInt for the rationale).
func AsList(v Value) *List {
	l, f := TryAsList(v)
	if f != nil {
		panic(f.Message)
	}
	return l
}

// AsBool extracts a bool from a Value. Panics on Furball input or type
// mismatch (see AsInt for the rationale).
func AsBool(v Value) bool {
	b, f := TryAsBool(v)
	if f != nil {
		panic(f.Message)
	}
	return b
}
