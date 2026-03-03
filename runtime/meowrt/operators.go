package meowrt

import "fmt"

// Add performs addition on same-type operands only.
// int+int→int, float+float→float, string+string→string.
func Add(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewInt(a.Val + b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewFloat(a.Val + b.Val)
		}
	case *String:
		if b, ok := b.(*String); ok {
			return NewString(a.Val + b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot add %s and %s, nya~", a.Type(), b.Type()))
}

// Sub performs subtraction on same-type operands only.
// int-int→int, float-float→float.
func Sub(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewInt(a.Val - b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewFloat(a.Val - b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot subtract %s and %s, nya~", a.Type(), b.Type()))
}

// Mul performs multiplication on same-type operands only.
// int*int→int, float*float→float.
func Mul(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewInt(a.Val * b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewFloat(a.Val * b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot multiply %s and %s, nya~", a.Type(), b.Type()))
}

// Div performs division on same-type operands only.
// int/int→int, float/float→float.
func Div(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			if b.Val == 0 {
				panic("Hiss! Division by zero, nya~")
			}
			return NewInt(a.Val / b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			if b.Val == 0 {
				panic("Hiss! Division by zero, nya~")
			}
			return NewFloat(a.Val / b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot divide %s and %s, nya~", a.Type(), b.Type()))
}

// Mod performs modulo on integers only.
func Mod(a, b Value) Value {
	ai, aok := a.(*Int)
	bi, bok := b.(*Int)
	if aok && bok {
		if bi.Val == 0 {
			panic("Hiss! Division by zero, nya~")
		}
		return NewInt(ai.Val % bi.Val)
	}
	panic(fmt.Sprintf("Hiss! Cannot modulo %s and %s, nya~", a.Type(), b.Type()))
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

// Equal checks equality between same-type operands only.
func Equal(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val == b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val == b.Val)
		}
	case *String:
		if b, ok := b.(*String); ok {
			return NewBool(a.Val == b.Val)
		}
	case *Bool:
		if b, ok := b.(*Bool); ok {
			return NewBool(a.Val == b.Val)
		}
	case *NilValue:
		if _, ok := b.(*NilValue); ok {
			return NewBool(true)
		}
	case *List:
		if b, ok := b.(*List); ok {
			return NewBool(listEqual(a, b))
		}
	case *Kitty:
		if b, ok := b.(*Kitty); ok {
			return NewBool(kittyEqual(a, b))
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type()))
}

func listEqual(a, b *List) bool {
	if len(a.Items) != len(b.Items) {
		return false
	}
	for i := range a.Items {
		eq := Equal(a.Items[i], b.Items[i]).(*Bool)
		if !eq.Val {
			return false
		}
	}
	return true
}

func kittyEqual(a, b *Kitty) bool {
	if a.TypeName != b.TypeName {
		return false
	}
	if len(a.FieldNames) != len(b.FieldNames) {
		return false
	}
	for _, name := range a.FieldNames {
		eq := Equal(a.Fields[name], b.Fields[name]).(*Bool)
		if !eq.Val {
			return false
		}
	}
	return true
}

// NotEqual checks inequality between same-type operands only.
func NotEqual(a, b Value) Value {
	eq := Equal(a, b).(*Bool)
	return NewBool(!eq.Val)
}

// LessThan performs less-than comparison on same-type operands only.
func LessThan(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val < b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val < b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type()))
}

// GreaterThan performs greater-than comparison on same-type operands only.
func GreaterThan(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val > b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val > b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type()))
}

// LessEqual performs less-than-or-equal comparison on same-type operands only.
func LessEqual(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val <= b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val <= b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type()))
}

// GreaterEqual performs greater-than-or-equal comparison on same-type operands only.
func GreaterEqual(a, b Value) Value {
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val >= b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val >= b.Val)
		}
	}
	panic(fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type()))
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
