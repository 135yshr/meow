package meowrt

import (
	"fmt"
	"strings"
)

// Nya prints a value (the Meow print function).
func Nya(args ...Value) Value {
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

// Hiss raises an error with the given message.
func Hiss(args ...Value) Value {
	parts := make([]string, len(args))
	for i, v := range args {
		parts[i] = v.String()
	}
	if len(parts) == 0 {
		panic("Hiss!")
	}
	panic(fmt.Sprintf("Hiss! %s", strings.Join(parts, " ")))
}

// Call invokes a function value with the given arguments.
// If the function has a fixed arity and fewer arguments are supplied,
// it returns a partially applied function.
func Call(fn Value, args ...Value) Value {
	f, ok := fn.(*Func)
	if !ok {
		panic(fmt.Sprintf("Hiss! %s is not callable, nya~", fn.Type()))
	}
	if f.Arity > 0 && len(args) < f.Arity {
		return PartialApply(f, args...)
	}
	return f.Call(args...)
}

// Len returns the length of a string or list.
func Len(v Value) Value {
	switch v := v.(type) {
	case *String:
		return NewInt(int64(len(v.Val)))
	case *List:
		return NewInt(int64(v.Len()))
	default:
		panic(fmt.Sprintf("Hiss! Cannot get length of %s, nya~", v.Type()))
	}
}

// ToInt converts a value to an integer.
func ToInt(v Value) Value {
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
		panic(fmt.Sprintf("Hiss! Cannot convert %s to Int, nya~", v.Type()))
	}
}

// ToFloat converts a value to a float.
func ToFloat(v Value) Value {
	switch v := v.(type) {
	case *Float:
		return v
	case *Int:
		return NewFloat(float64(v.Val))
	default:
		panic(fmt.Sprintf("Hiss! Cannot convert %s to Float, nya~", v.Type()))
	}
}

// ToString converts a value to a string.
func ToString(v Value) Value {
	return NewString(v.String())
}

// Gag calls fn (a zero-argument Func) and recovers from any panic,
// returning the panic message wrapped in a Furball.
// If fn succeeds, its return value is returned as-is.
func Gag(fn Value) Value {
	f, ok := fn.(*Func)
	if !ok {
		panic(fmt.Sprintf("Hiss! gag expects a function, got %s, nya~", fn.Type()))
	}
	var result Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = &Furball{Message: fmt.Sprintf("%v", r)}
			}
		}()
		result = f.Call()
	}()
	return result
}

// GagOr calls fn (a zero-argument Func) and recovers from any panic.
// If the call succeeds, its return value is returned as-is.
// If it panics, the fallback is used: if fallback is a Func it is called
// with the Furball as argument; otherwise fallback is returned directly.
func GagOr(fn Value, fallback Value) Value {
	result := Gag(fn)
	if _, ok := result.(*Furball); ok {
		if f, fok := fallback.(*Func); fok {
			return f.Call(result)
		}
		return fallback
	}
	return result
}

// IsFurball returns NewBool(true) if v is a Furball, NewBool(false) otherwise.
func IsFurball(v Value) Value {
	_, ok := v.(*Furball)
	return NewBool(ok)
}
