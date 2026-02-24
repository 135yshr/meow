package types

// Type represents a Meow type.
type Type interface {
	String() string
	Equals(Type) bool
}

// IntType represents the int type.
type IntType struct{}

func (IntType) String() string    { return "int" }
func (IntType) Equals(t Type) bool { _, ok := t.(IntType); return ok }

// FloatType represents the float type.
type FloatType struct{}

func (FloatType) String() string    { return "float" }
func (FloatType) Equals(t Type) bool { _, ok := t.(FloatType); return ok }

// StringType represents the string type.
type StringType struct{}

func (StringType) String() string    { return "string" }
func (StringType) Equals(t Type) bool { _, ok := t.(StringType); return ok }

// BoolType represents the bool type.
type BoolType struct{}

func (BoolType) String() string    { return "bool" }
func (BoolType) Equals(t Type) bool { _, ok := t.(BoolType); return ok }

// NilType represents the nil type.
type NilType struct{}

func (NilType) String() string    { return "nil" }
func (NilType) Equals(t Type) bool { _, ok := t.(NilType); return ok }

// AnyType represents an untyped value (backward compatibility).
// Operations involving AnyType skip static type checks.
type AnyType struct{}

func (AnyType) String() string    { return "any" }
func (AnyType) Equals(t Type) bool { _, ok := t.(AnyType); return ok }

// ListType represents a list type with element type.
type ListType struct{ Elem Type }

func (l ListType) String() string { return "list[" + l.Elem.String() + "]" }
func (l ListType) Equals(t Type) bool {
	o, ok := t.(ListType)
	return ok && l.Elem.Equals(o.Elem)
}

// FuncType represents a function type.
type FuncType struct {
	Params []Type
	Return Type
}

func (f FuncType) String() string {
	s := "("
	for i, p := range f.Params {
		if i > 0 {
			s += ", "
		}
		s += p.String()
	}
	s += ") " + f.Return.String()
	return s
}

func (f FuncType) Equals(t Type) bool {
	o, ok := t.(FuncType)
	if !ok || len(f.Params) != len(o.Params) {
		return false
	}
	for i := range f.Params {
		if !f.Params[i].Equals(o.Params[i]) {
			return false
		}
	}
	return f.Return.Equals(o.Return)
}

// IsAny reports whether t is AnyType.
func IsAny(t Type) bool {
	_, ok := t.(AnyType)
	return ok
}

// IsNumeric reports whether t is IntType or FloatType.
func IsNumeric(t Type) bool {
	switch t.(type) {
	case IntType, FloatType:
		return true
	}
	return false
}
