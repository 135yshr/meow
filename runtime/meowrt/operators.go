package meowrt

import "fmt"

// propagate returns the first Furball found among the arguments, or nil if
// none is present. Operators use this to short-circuit error propagation
// without panic.
func propagate(args ...Value) *Furball {
	for _, a := range args {
		if f, ok := a.(*Furball); ok {
			return f
		}
	}
	return nil
}

// Add performs addition on same-type operands only.
// int+int→int, float+float→float, string+string→string.
func Add(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewInt(a.Val + b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewByte(a.Val + b.Val)
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
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot add %s and %s, nya~", a.Type(), b.Type())}
}

// Sub performs subtraction on same-type operands only.
// int-int→int, float-float→float.
func Sub(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewInt(a.Val - b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewByte(a.Val - b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewFloat(a.Val - b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot subtract %s and %s, nya~", a.Type(), b.Type())}
}

// Mul performs multiplication on same-type operands only.
// int*int→int, float*float→float.
func Mul(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewInt(a.Val * b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewByte(a.Val * b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewFloat(a.Val * b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot multiply %s and %s, nya~", a.Type(), b.Type())}
}

// Div performs division on same-type operands only.
// int/int→int, float/float→float.
func Div(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			if b.Val == 0 {
				return &Furball{Message: "Hiss! Division by zero, nya~"}
			}
			return NewInt(a.Val / b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			if b.Val == 0 {
				return &Furball{Message: "Hiss! Division by zero, nya~"}
			}
			return NewByte(a.Val / b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			if b.Val == 0 {
				return &Furball{Message: "Hiss! Division by zero, nya~"}
			}
			return NewFloat(a.Val / b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot divide %s and %s, nya~", a.Type(), b.Type())}
}

// Mod performs modulo on integers only.
func Mod(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	if ai, aok := a.(*Int); aok {
		if bi, bok := b.(*Int); bok {
			if bi.Val == 0 {
				return &Furball{Message: "Hiss! Division by zero, nya~"}
			}
			return NewInt(ai.Val % bi.Val)
		}
	}
	if ab, aok := a.(*Byte); aok {
		if bb, bok := b.(*Byte); bok {
			if bb.Val == 0 {
				return &Furball{Message: "Hiss! Division by zero, nya~"}
			}
			return NewByte(ab.Val % bb.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot modulo %s and %s, nya~", a.Type(), b.Type())}
}

// Negate negates a value.
func Negate(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	switch v := v.(type) {
	case *Int:
		return NewInt(-v.Val)
	case *Float:
		return NewFloat(-v.Val)
	default:
		return &Furball{Message: fmt.Sprintf("Hiss! Cannot negate %s, nya~", v.Type())}
	}
}

// Not performs logical NOT.
func Not(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	return NewBool(!v.IsTruthy())
}

// Equal checks equality between same-type operands only.
func Equal(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val == b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
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
			eq, err := listEqual(a, b)
			if err != nil {
				return err
			}
			return NewBool(eq)
		}
	case *Kitty:
		if b, ok := b.(*Kitty); ok {
			eq, err := kittyEqual(a, b)
			if err != nil {
				return err
			}
			return NewBool(eq)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type())}
}

func listEqual(a, b *List) (bool, *Furball) {
	if len(a.Items) != len(b.Items) {
		return false, nil
	}
	for i := range a.Items {
		r := Equal(a.Items[i], b.Items[i])
		if f, ok := r.(*Furball); ok {
			return false, f
		}
		if !r.(*Bool).Val {
			return false, nil
		}
	}
	return true, nil
}

func kittyEqual(a, b *Kitty) (bool, *Furball) {
	if a.TypeName != b.TypeName {
		return false, nil
	}
	if len(a.FieldNames) != len(b.FieldNames) {
		return false, nil
	}
	for _, name := range a.FieldNames {
		r := Equal(a.Fields[name], b.Fields[name])
		if f, ok := r.(*Furball); ok {
			return false, f
		}
		if !r.(*Bool).Val {
			return false, nil
		}
	}
	return true, nil
}

// NotEqual checks inequality between same-type operands only.
func NotEqual(a, b Value) Value {
	r := Equal(a, b)
	if f, ok := r.(*Furball); ok {
		return f
	}
	return NewBool(!r.(*Bool).Val)
}

// LessThan performs less-than comparison on same-type operands only.
func LessThan(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val < b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewBool(a.Val < b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val < b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type())}
}

// GreaterThan performs greater-than comparison on same-type operands only.
func GreaterThan(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val > b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewBool(a.Val > b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val > b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type())}
}

// LessEqual performs less-than-or-equal comparison on same-type operands only.
func LessEqual(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val <= b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewBool(a.Val <= b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val <= b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type())}
}

// GreaterEqual performs greater-than-or-equal comparison on same-type operands only.
func GreaterEqual(a, b Value) Value {
	if f := propagate(a, b); f != nil {
		return f
	}
	switch a := a.(type) {
	case *Int:
		if b, ok := b.(*Int); ok {
			return NewBool(a.Val >= b.Val)
		}
	case *Byte:
		if b, ok := b.(*Byte); ok {
			return NewBool(a.Val >= b.Val)
		}
	case *Float:
		if b, ok := b.(*Float); ok {
			return NewBool(a.Val >= b.Val)
		}
	}
	return &Furball{Message: fmt.Sprintf("Hiss! Cannot compare %s and %s, nya~", a.Type(), b.Type())}
}

// And performs logical AND (short-circuit).
func And(a, b Value) Value {
	if f, ok := a.(*Furball); ok {
		return f
	}
	if !a.IsTruthy() {
		return a
	}
	if f, ok := b.(*Furball); ok {
		return f
	}
	return b
}

// Or performs logical OR (short-circuit).
func Or(a, b Value) Value {
	if f, ok := a.(*Furball); ok {
		return f
	}
	if a.IsTruthy() {
		return a
	}
	if f, ok := b.(*Furball); ok {
		return f
	}
	return b
}
