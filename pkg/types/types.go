package types

// Type represents a Meow type.
type Type interface {
	String() string
	Equals(Type) bool
}

// IntType represents the int type.
type IntType struct{}

func (IntType) String() string    { return "int" }
func (IntType) Equals(t Type) bool { _, ok := Unwrap(t).(IntType); return ok }

// FloatType represents the float type.
type FloatType struct{}

func (FloatType) String() string    { return "float" }
func (FloatType) Equals(t Type) bool { _, ok := Unwrap(t).(FloatType); return ok }

// StringType represents the string type.
type StringType struct{}

func (StringType) String() string    { return "string" }
func (StringType) Equals(t Type) bool { _, ok := Unwrap(t).(StringType); return ok }

// BoolType represents the bool type.
type BoolType struct{}

func (BoolType) String() string    { return "bool" }
func (BoolType) Equals(t Type) bool { _, ok := Unwrap(t).(BoolType); return ok }

// NilType represents the nil type.
type NilType struct{}

func (NilType) String() string    { return "nil" }
func (NilType) Equals(t Type) bool { _, ok := Unwrap(t).(NilType); return ok }

// FurballType represents an error (Furball) type.
type FurballType struct{}

func (FurballType) String() string    { return "furball" }
func (FurballType) Equals(t Type) bool { _, ok := Unwrap(t).(FurballType); return ok }

// AnyType is an internal fallback type for built-in operations whose types
// cannot be statically determined (e.g. head, tail, gag). User code must
// provide explicit type annotations; AnyType is not part of the user-facing
// type system.
type AnyType struct{}

func (AnyType) String() string    { return "any" }
func (AnyType) Equals(t Type) bool { _, ok := t.(AnyType); return ok }

// ListType represents a list type with element type.
type ListType struct{ Elem Type }

func (l ListType) String() string { return "list[" + l.Elem.String() + "]" }
func (l ListType) Equals(t Type) bool {
	o, ok := t.(ListType)
	if !ok {
		return false
	}
	// list[any] matches any list type (covariant)
	if IsAny(l.Elem) || IsAny(o.Elem) {
		return true
	}
	return l.Elem.Equals(o.Elem)
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

// KittyFieldType represents a field in a kitty type.
type KittyFieldType struct {
	Name string
	Type Type
}

// KittyType represents a user-defined struct type.
type KittyType struct {
	Name   string
	Fields []KittyFieldType
}

func (k KittyType) String() string { return k.Name }
func (k KittyType) Equals(t Type) bool {
	o, ok := t.(KittyType)
	return ok && k.Name == o.Name
}

// Unwrap resolves AliasType wrappers recursively, returning the underlying type.
// Non-alias types are returned unchanged.
func Unwrap(t Type) Type {
	for {
		a, ok := t.(AliasType)
		if !ok {
			return t
		}
		t = a.Underlying
	}
}

// AliasType represents a type alias (breed). It is transparent: an AliasType
// equals its underlying type.
type AliasType struct {
	Name       string
	Underlying Type
}

func (a AliasType) String() string { return a.Name }
func (a AliasType) Equals(t Type) bool {
	return Unwrap(a).Equals(Unwrap(t))
}

// CollarType represents a newtype (collar). It is nominal: two CollarTypes
// are equal only if they share the same name.
type CollarType struct {
	Name       string
	Underlying Type
}

func (c CollarType) String() string { return c.Name }
func (c CollarType) Equals(t Type) bool {
	o, ok := t.(CollarType)
	return ok && c.Name == o.Name
}

// TrickMethodSig represents a method signature in a trick.
type TrickMethodSig struct {
	Name       string
	ParamTypes []Type
	ReturnType Type
}

// TrickType represents a trick (interface) definition.
type TrickType struct {
	Name    string
	Methods []TrickMethodSig
}

func (t TrickType) String() string { return t.Name }
func (t TrickType) Equals(other Type) bool {
	o, ok := other.(TrickType)
	return ok && t.Name == o.Name
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
