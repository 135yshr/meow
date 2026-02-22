package meowrt

import "fmt"

func toFloat(v Value) float64 {
	switch v := v.(type) {
	case *Int:
		return float64(v.Val)
	case *Float:
		return v.Val
	default:
		panic(fmt.Sprintf("Hiss! Cannot convert %s to number, nya~", v.Type()))
	}
}

func isNumeric(v Value) bool {
	switch v.(type) {
	case *Int, *Float:
		return true
	default:
		return false
	}
}

func bothInt(a, b Value) (*Int, *Int, bool) {
	ai, aok := a.(*Int)
	bi, bok := b.(*Int)
	return ai, bi, aok && bok
}

// Add performs addition or string concatenation.
func Add(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewInt(ai.Val + bi.Val)
	}
	if isNumeric(a) && isNumeric(b) {
		return NewFloat(toFloat(a) + toFloat(b))
	}
	// String concatenation
	return NewString(a.String() + b.String())
}

// Sub performs subtraction.
func Sub(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewInt(ai.Val - bi.Val)
	}
	return NewFloat(toFloat(a) - toFloat(b))
}

// Mul performs multiplication.
func Mul(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewInt(ai.Val * bi.Val)
	}
	return NewFloat(toFloat(a) * toFloat(b))
}

// Div performs division.
func Div(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		if bi.Val == 0 {
			panic("Hiss! Division by zero, nya~")
		}
		return NewInt(ai.Val / bi.Val)
	}
	fb := toFloat(b)
	if fb == 0 {
		panic("Hiss! Division by zero, nya~")
	}
	return NewFloat(toFloat(a) / fb)
}

// Mod performs modulo.
func Mod(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		if bi.Val == 0 {
			panic("Hiss! Division by zero, nya~")
		}
		return NewInt(ai.Val % bi.Val)
	}
	panic("Hiss! Modulo requires integers, nya~")
}

// Negate negates a value.
func Negate(v Value) Value {
	switch v := v.(type) {
	case *Int:
		return NewInt(-v.Val)
	case *Float:
		return NewFloat(-v.Val)
	default:
		panic(fmt.Sprintf("Hiss! Cannot negate %s, nya~", v.Type()))
	}
}

// Not performs logical NOT.
func Not(v Value) Value {
	return NewBool(!v.IsTruthy())
}

// Equal checks equality.
func Equal(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewBool(ai.Val == bi.Val)
	}
	if isNumeric(a) && isNumeric(b) {
		return NewBool(toFloat(a) == toFloat(b))
	}
	return NewBool(a.String() == b.String())
}

// NotEqual checks inequality.
func NotEqual(a, b Value) Value {
	eq := Equal(a, b).(*Bool)
	return NewBool(!eq.Val)
}

// LessThan performs less-than comparison.
func LessThan(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewBool(ai.Val < bi.Val)
	}
	return NewBool(toFloat(a) < toFloat(b))
}

// GreaterThan performs greater-than comparison.
func GreaterThan(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewBool(ai.Val > bi.Val)
	}
	return NewBool(toFloat(a) > toFloat(b))
}

// LessEqual performs less-than-or-equal comparison.
func LessEqual(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewBool(ai.Val <= bi.Val)
	}
	return NewBool(toFloat(a) <= toFloat(b))
}

// GreaterEqual performs greater-than-or-equal comparison.
func GreaterEqual(a, b Value) Value {
	if ai, bi, ok := bothInt(a, b); ok {
		return NewBool(ai.Val >= bi.Val)
	}
	return NewBool(toFloat(a) >= toFloat(b))
}

// And performs logical AND (short-circuit).
func And(a, b Value) Value {
	if !a.IsTruthy() {
		return a
	}
	return b
}

// Or performs logical OR (short-circuit).
func Or(a, b Value) Value {
	if a.IsTruthy() {
		return a
	}
	return b
}
