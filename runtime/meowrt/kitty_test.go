package meowrt

import (
	"testing"
)

func TestNewKitty(t *testing.T) {
	k := NewKitty("Cat", []string{"name", "age"}, NewString("Tama"), NewInt(3))
	if k.TypeName != "Cat" {
		t.Errorf("TypeName = %q, want %q", k.TypeName, "Cat")
	}
	if k.Fields["name"].String() != "Tama" {
		t.Errorf("name = %q, want %q", k.Fields["name"].String(), "Tama")
	}
	if k.Fields["age"].(*Int).Val != 3 {
		t.Errorf("age = %d, want %d", k.Fields["age"].(*Int).Val, 3)
	}
}

func TestNewKittyArgCountMismatch(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg, ok := r.(string)
		if !ok || msg == "" {
			t.Fatalf("expected string panic, got %v", r)
		}
	}()
	NewKitty("Cat", []string{"name", "age"}, NewString("Tama"))
}

func TestKittyType(t *testing.T) {
	k := NewKitty("Dog", []string{"breed"}, NewString("Shiba"))
	if k.Type() != "Dog" {
		t.Errorf("Type() = %q, want %q", k.Type(), "Dog")
	}
}

func TestKittyString(t *testing.T) {
	k := NewKitty("Cat", []string{"name", "age"}, NewString("Tama"), NewInt(3))
	want := "Cat{name: Tama, age: 3}"
	if k.String() != want {
		t.Errorf("String() = %q, want %q", k.String(), want)
	}
}

func TestKittyIsTruthy(t *testing.T) {
	k := NewKitty("Cat", []string{"name"}, NewString("Tama"))
	if !k.IsTruthy() {
		t.Error("IsTruthy() = false, want true")
	}
}

func TestKittyGetField(t *testing.T) {
	k := NewKitty("Cat", []string{"name", "age"}, NewString("Tama"), NewInt(3))
	if v := k.GetField("name"); v.String() != "Tama" {
		t.Errorf("GetField(name) = %q, want %q", v.String(), "Tama")
	}
	if v := k.GetField("age").(*Int); v.Val != 3 {
		t.Errorf("GetField(age) = %d, want %d", v.Val, 3)
	}
}

func TestKittyGetFieldPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	k := NewKitty("Cat", []string{"name"}, NewString("Tama"))
	k.GetField("unknown")
}

func TestKittyEqual(t *testing.T) {
	a := NewKitty("Cat", []string{"name", "age"}, NewString("Tama"), NewInt(3))
	b := NewKitty("Cat", []string{"name", "age"}, NewString("Tama"), NewInt(3))
	c := NewKitty("Cat", []string{"name", "age"}, NewString("Mike"), NewInt(5))

	if eq := Equal(a, b).(*Bool); !eq.Val {
		t.Error("Equal(a, b) = false, want true")
	}
	if eq := Equal(a, c).(*Bool); eq.Val {
		t.Error("Equal(a, c) = true, want false")
	}
}

func TestKittyToJSON(t *testing.T) {
	k := NewKitty("Cat", []string{"name", "age"}, NewString("Tama"), NewInt(3))
	want := `{"name":"Tama","age":3}`
	if got := ToJSON(k); got != want {
		t.Errorf("ToJSON() = %q, want %q", got, want)
	}
}
