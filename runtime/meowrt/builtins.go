package meowrt

import "fmt"

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

// Call invokes a function value with the given arguments.
func Call(fn Value, args ...Value) Value {
	f, ok := fn.(*Func)
	if !ok {
		panic(fmt.Sprintf("Hiss! %s is not callable, nya~", fn.Type()))
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
