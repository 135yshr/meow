package meowrt

import (
	"strings"
	"testing"
)

// expectFurball asserts that fn returns a Furball whose Message contains msg.
// Replaces the previous expectPanic helper after the panic-to-Furball
// migration: operators now return *Furball values instead of panicking.
func expectFurball(t *testing.T, msg string, fn func() Value) {
	t.Helper()
	v := fn()
	f, ok := v.(*Furball)
	if !ok {
		t.Fatalf("expected *Furball containing %q, got %T (%v)", msg, v, v)
	}
	if !strings.Contains(f.Message, msg) {
		t.Fatalf("Furball %q does not contain %q", f.Message, msg)
	}
}

// --- Add ---

func TestAddIntInt(t *testing.T) {
	result := Add(NewInt(1), NewInt(2))
	if result.(*Int).Val != 3 {
		t.Errorf("expected 3, got %s", result)
	}
}

func TestAddFloatFloat(t *testing.T) {
	result := Add(NewFloat(1.5), NewFloat(2.5))
	if result.(*Float).Val != 4.0 {
		t.Errorf("expected 4.0, got %s", result)
	}
}

func TestAddStringString(t *testing.T) {
	result := Add(NewString("hello"), NewString(" world"))
	if result.(*String).Val != "hello world" {
		t.Errorf("expected 'hello world', got %s", result)
	}
}

func TestAddIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot add", func() Value { return Add(NewInt(1), NewFloat(2.0)) })
}

func TestAddIntString_Furball(t *testing.T) {
	expectFurball(t, "Cannot add", func() Value { return Add(NewInt(1), NewString("hello")) })
}

func TestAddStringInt_Furball(t *testing.T) {
	expectFurball(t, "Cannot add", func() Value { return Add(NewString("hello"), NewInt(1)) })
}

// --- Sub ---

func TestSubIntInt(t *testing.T) {
	result := Sub(NewInt(5), NewInt(3))
	if result.(*Int).Val != 2 {
		t.Errorf("expected 2, got %s", result)
	}
}

func TestSubFloatFloat(t *testing.T) {
	result := Sub(NewFloat(5.5), NewFloat(2.5))
	if result.(*Float).Val != 3.0 {
		t.Errorf("expected 3.0, got %s", result)
	}
}

func TestSubIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot subtract", func() Value { return Sub(NewInt(1), NewFloat(2.0)) })
}

func TestSubStringString_Furball(t *testing.T) {
	expectFurball(t, "Cannot subtract", func() Value { return Sub(NewString("a"), NewString("b")) })
}

// --- Mul ---

func TestMulIntInt(t *testing.T) {
	result := Mul(NewInt(3), NewInt(4))
	if result.(*Int).Val != 12 {
		t.Errorf("expected 12, got %s", result)
	}
}

func TestMulFloatFloat(t *testing.T) {
	result := Mul(NewFloat(2.0), NewFloat(3.0))
	if result.(*Float).Val != 6.0 {
		t.Errorf("expected 6.0, got %s", result)
	}
}

func TestMulIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot multiply", func() Value { return Mul(NewInt(1), NewFloat(2.0)) })
}

// --- Div ---

func TestDivIntInt(t *testing.T) {
	result := Div(NewInt(10), NewInt(2))
	if result.(*Int).Val != 5 {
		t.Errorf("expected 5, got %s", result)
	}
}

func TestDivFloatFloat(t *testing.T) {
	result := Div(NewFloat(7.5), NewFloat(2.5))
	if result.(*Float).Val != 3.0 {
		t.Errorf("expected 3.0, got %s", result)
	}
}

func TestDivIntByZero_Furball(t *testing.T) {
	expectFurball(t, "Division by zero", func() Value { return Div(NewInt(1), NewInt(0)) })
}

func TestDivFloatByZero_Furball(t *testing.T) {
	expectFurball(t, "Division by zero", func() Value { return Div(NewFloat(1.0), NewFloat(0.0)) })
}

func TestDivIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot divide", func() Value { return Div(NewInt(1), NewFloat(2.0)) })
}

// --- Mod ---

func TestModIntInt(t *testing.T) {
	result := Mod(NewInt(10), NewInt(3))
	if result.(*Int).Val != 1 {
		t.Errorf("expected 1, got %s", result)
	}
}

func TestModByZero_Furball(t *testing.T) {
	expectFurball(t, "Division by zero", func() Value { return Mod(NewInt(10), NewInt(0)) })
}

func TestModFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot modulo", func() Value { return Mod(NewFloat(10.0), NewFloat(3.0)) })
}

// --- Equal ---

func TestEqualIntInt(t *testing.T) {
	result := Equal(NewInt(1), NewInt(1))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
	result = Equal(NewInt(1), NewInt(2))
	if result.(*Bool).Val {
		t.Error("expected false")
	}
}

func TestEqualStringString(t *testing.T) {
	result := Equal(NewString("cat"), NewString("cat"))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestEqualBoolBool(t *testing.T) {
	result := Equal(NewBool(true), NewBool(true))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestEqualNilNil(t *testing.T) {
	result := Equal(NewNil(), NewNil())
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestEqualIntString_Furball(t *testing.T) {
	expectFurball(t, "Cannot compare", func() Value { return Equal(NewInt(1), NewString("1")) })
}

func TestEqualIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot compare", func() Value { return Equal(NewInt(1), NewFloat(1.0)) })
}

// --- NotEqual ---

func TestNotEqualIntInt(t *testing.T) {
	result := NotEqual(NewInt(1), NewInt(2))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

// --- Comparison ---

func TestLessThanIntInt(t *testing.T) {
	result := LessThan(NewInt(1), NewInt(2))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestLessThanFloatFloat(t *testing.T) {
	result := LessThan(NewFloat(1.0), NewFloat(2.0))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestLessThanIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot compare", func() Value { return LessThan(NewInt(1), NewFloat(2.0)) })
}

func TestGreaterThanIntInt(t *testing.T) {
	result := GreaterThan(NewInt(2), NewInt(1))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestLessEqualIntInt(t *testing.T) {
	result := LessEqual(NewInt(1), NewInt(1))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestGreaterEqualIntInt(t *testing.T) {
	result := GreaterEqual(NewInt(2), NewInt(1))
	if !result.(*Bool).Val {
		t.Error("expected true")
	}
}

func TestGreaterThanIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot compare", func() Value { return GreaterThan(NewInt(1), NewFloat(2.0)) })
}

func TestLessEqualIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot compare", func() Value { return LessEqual(NewInt(1), NewFloat(2.0)) })
}

func TestGreaterEqualIntFloat_Furball(t *testing.T) {
	expectFurball(t, "Cannot compare", func() Value { return GreaterEqual(NewInt(1), NewFloat(2.0)) })
}
