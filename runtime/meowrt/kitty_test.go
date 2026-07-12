package meowrt

import (
	"strings"
	"testing"
)

func mustKitty(t *testing.T, v Value) *Kitty {
	t.Helper()
	k, ok := v.(*Kitty)
	if !ok {
		t.Fatalf("expected *Kitty, got %T (%v)", v, v)
	}
	return k
}

func TestNewKitty(t *testing.T) {
	k := mustKitty(t, NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"), NewInt(3)))
	if k.TypeName != "Cat" {
		t.Errorf("TypeName = %q, want %q", k.TypeName, "Cat")
	}
	if k.Fields["name"].String() != "Nyantyu" {
		t.Errorf("name = %q, want %q", k.Fields["name"].String(), "Nyantyu")
	}
	if k.Fields["age"].(*Int).Val != 3 {
		t.Errorf("age = %d, want %d", k.Fields["age"].(*Int).Val, 3)
	}
}

func TestNewKittyArgCountMismatch(t *testing.T) {
	v := NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"))
	f, ok := v.(*Furball)
	if !ok {
		t.Fatalf("expected *Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "expects 2 fields but got 1") {
		t.Errorf("unexpected message: %q", f.Message)
	}
}

func TestKittyType(t *testing.T) {
	k := mustKitty(t, NewKitty("Dog", []string{"breed"}, NewString("Shiba")))
	if k.Type() != "Dog" {
		t.Errorf("Type() = %q, want %q", k.Type(), "Dog")
	}
}

func TestKittyString(t *testing.T) {
	k := mustKitty(t, NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"), NewInt(3)))
	want := "Cat{name: Nyantyu, age: 3}"
	if k.String() != want {
		t.Errorf("String() = %q, want %q", k.String(), want)
	}
}

func TestKittyIsTruthy(t *testing.T) {
	k := mustKitty(t, NewKitty("Cat", []string{"name"}, NewString("Nyantyu")))
	if !k.IsTruthy() {
		t.Error("IsTruthy() = false, want true")
	}
}

func TestKittyGetField(t *testing.T) {
	k := mustKitty(t, NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"), NewInt(3)))
	if v := k.GetField("name"); v.String() != "Nyantyu" {
		t.Errorf("GetField(name) = %q, want %q", v.String(), "Nyantyu")
	}
	if v := k.GetField("age").(*Int); v.Val != 3 {
		t.Errorf("GetField(age) = %d, want %d", v.Val, 3)
	}
}

func TestKittyGetFieldMissingReturnsFurball(t *testing.T) {
	k := mustKitty(t, NewKitty("Cat", []string{"name"}, NewString("Nyantyu")))
	v := k.GetField("unknown")
	f, ok := v.(*Furball)
	if !ok {
		t.Fatalf("expected *Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "has no field unknown") {
		t.Errorf("unexpected message: %q", f.Message)
	}
}

func TestKittyEqual(t *testing.T) {
	a := NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"), NewInt(3))
	b := NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"), NewInt(3))
	c := NewKitty("Cat", []string{"name", "age"}, NewString("Tyako"), NewInt(5))

	if eq := Equal(a, b).(*Bool); !eq.Val {
		t.Error("Equal(a, b) = false, want true")
	}
	if eq := Equal(a, c).(*Bool); eq.Val {
		t.Error("Equal(a, c) = true, want false")
	}
}

func TestKittyToJSON(t *testing.T) {
	k := NewKitty("Cat", []string{"name", "age"}, NewString("Nyantyu"), NewInt(3))
	want := `{"name":"Nyantyu","age":3}`
	if got := ToJSON(k); got != want {
		t.Errorf("ToJSON() = %q, want %q", got, want)
	}
}
