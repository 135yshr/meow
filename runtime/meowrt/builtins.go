package meowrt

import (
	"fmt"
	"strings"
)

// Nya prints a value (the Meow print function). If any argument is an
// unhandled Furball, Nya returns it without printing so the error
// propagates instead of being stringified and silently swallowed.
func Nya(args ...Value) Value {
	for _, a := range args {
		if f, ok := a.(*Furball); ok && !f.Handled {
			return f
		}
	}
	parts := make([]string, len(args))
	for i, v := range args {
		parts[i] = v.String()
	}
	for i, p := range parts {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(p)
	}
	fmt.Println()
	return NewNil()
}

// Hiss raises an error with the given message, returning a Furball value
// that propagates through the runtime as a value (no panic).
func Hiss(args ...Value) Value {
	parts := make([]string, len(args))
	for i, v := range args {
		parts[i] = v.String()
	}
	if len(parts) == 0 {
		return &Furball{Message: "Hiss!"}
	}
	return &Furball{Message: "Hiss! " + strings.Join(parts, " ")}
}

// Call invokes a function value with the given arguments.
// If the function has a fixed arity and fewer arguments are supplied,
// it returns a partially applied function. Errors propagate as Furball.
func Call(fn Value, args ...Value) Value {
	if f, ok := fn.(*Furball); ok {
		return f
	}
	for _, a := range args {
		if f, ok := a.(*Furball); ok {
			return f
		}
	}
	f, ok := fn.(*Func)
	if !ok {
		return &Furball{Message: fmt.Sprintf("Hiss! %s is not callable, nya~", fn.Type())}
	}
	if f.Arity > 0 {
		if len(args) < f.Arity {
			return PartialApply(f, args...)
		}
		if len(args) > f.Arity {
			return &Furball{Message: fmt.Sprintf("Hiss! %s expects %d arguments but got %d, nya~", f.Name, f.Arity, len(args))}
		}
	}
	return f.Call(args...)
}

// Len returns the length of a string or list.
func Len(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	switch v := v.(type) {
	case *String:
		return NewInt(int64(len(v.Val)))
	case *List:
		return NewInt(int64(v.Len()))
	default:
		return &Furball{Message: fmt.Sprintf("Hiss! Cannot get length of %s, nya~", v.Type())}
	}
}

// ToInt converts a value to an integer.
func ToInt(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	switch v := v.(type) {
	case *Int:
		return v
	case *Float:
		return NewInt(int64(v.Val))
	case *Bool:
		if v.Val {
			return NewInt(1)
		}
		return NewInt(0)
	default:
		return &Furball{Message: fmt.Sprintf("Hiss! Cannot convert %s to Int, nya~", v.Type())}
	}
}

// ToFloat converts a value to a float.
func ToFloat(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	switch v := v.(type) {
	case *Float:
		return v
	case *Int:
		return NewFloat(float64(v.Val))
	default:
		return &Furball{Message: fmt.Sprintf("Hiss! Cannot convert %s to Float, nya~", v.Type())}
	}
}

// ToString converts a value to a string.
func ToString(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	return NewString(v.String())
}

// ToBytes converts a string value to a list of Byte values.
func ToBytes(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	s, ok := v.(*String)
	if !ok {
		return &Furball{Message: fmt.Sprintf("Hiss! Cannot convert %s to bytes, nya~", v.Type())}
	}
	bytes := []byte(s.Val)
	elems := make([]Value, len(bytes))
	for i, b := range bytes {
		elems[i] = NewByte(b)
	}
	return NewList(elems...)
}

// ToRunes converts a string value to a list of single-character strings.
func ToRunes(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	s, ok := v.(*String)
	if !ok {
		return &Furball{Message: fmt.Sprintf("Hiss! Cannot convert %s to runes, nya~", v.Type())}
	}
	runes := []rune(s.Val)
	elems := make([]Value, len(runes))
	for i, r := range runes {
		elems[i] = NewString(string(r))
	}
	return NewList(elems...)
}

// Gag calls fn (a zero-argument Func) and returns its result.
// Errors are converted into a Handled Furball — handled marks the value as
// already surfaced, so generated code's automatic short-circuit does not
// re-propagate it. This lets user code branch on `is_furball(result)`.
//
// Two error sources are unified: untyped runtime helpers return raw Furball
// values (which Gag re-flags as handled), and typed function bodies may
// panic internally — a deferred recover catches those and wraps them.
func Gag(fn Value) (result Value) {
	if f, ok := fn.(*Furball); ok {
		return markHandled(f)
	}
	f, ok := fn.(*Func)
	if !ok {
		return &Furball{Message: fmt.Sprintf("Hiss! gag expects a function, got %s, nya~", fn.Type()), Handled: true}
	}
	defer func() {
		if r := recover(); r != nil {
			result = &Furball{Message: fmt.Sprintf("%v", r), Handled: true}
		}
		if fb, ok := result.(*Furball); ok {
			result = markHandled(fb)
		}
	}()
	return f.Call()
}

// markHandled returns a new Furball with the same message marked as handled,
// preserving the original for any code path that still references it.
func markHandled(f *Furball) *Furball {
	if f.Handled {
		return f
	}
	return &Furball{Message: f.Message, Handled: true}
}

// GagOr calls fn (a zero-argument Func) and returns its result.
// If the result is a Furball, the fallback is used: if fallback is a Func
// it is called with the Furball as argument; otherwise fallback is returned.
func GagOr(fn, fallback Value) Value {
	return Recover(Gag(fn), fallback)
}

// IsFurball returns NewBool(true) if v is a Furball, NewBool(false) otherwise.
// Note: this intentionally does NOT propagate the Furball — it is the only
// builtin that can inspect a Furball without being short-circuited.
func IsFurball(v Value) Value {
	_, ok := v.(*Furball)
	return NewBool(ok)
}
