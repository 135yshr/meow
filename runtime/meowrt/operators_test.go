package meowrt

import (
	"strings"
	"testing"
)

func expectPanic(t *testing.T, msg string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, but no panic occurred", msg)
		}
		s, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(s, msg) {
			t.Fatalf("panic %q does not contain %q", s, msg)
		}
	}()
	fn()
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

func TestAddIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot add", func() {
		Add(NewInt(1), NewFloat(2.0))
	})
}

func TestAddIntString_Panics(t *testing.T) {
	expectPanic(t, "Cannot add", func() {
		Add(NewInt(1), NewString("hello"))
	})
}

func TestAddStringInt_Panics(t *testing.T) {
	expectPanic(t, "Cannot add", func() {
		Add(NewString("hello"), NewInt(1))
	})
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

func TestSubIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot subtract", func() {
		Sub(NewInt(1), NewFloat(2.0))
	})
}

func TestSubStringString_Panics(t *testing.T) {
	expectPanic(t, "Cannot subtract", func() {
		Sub(NewString("a"), NewString("b"))
	})
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

func TestMulIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot multiply", func() {
		Mul(NewInt(1), NewFloat(2.0))
	})
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

func TestDivIntByZero_Panics(t *testing.T) {
	expectPanic(t, "Division by zero", func() {
		Div(NewInt(1), NewInt(0))
	})
}

func TestDivFloatByZero_Panics(t *testing.T) {
	expectPanic(t, "Division by zero", func() {
		Div(NewFloat(1.0), NewFloat(0.0))
	})
}

func TestDivIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot divide", func() {
		Div(NewInt(1), NewFloat(2.0))
	})
}

// --- Mod ---

func TestModIntInt(t *testing.T) {
	result := Mod(NewInt(10), NewInt(3))
	if result.(*Int).Val != 1 {
		t.Errorf("expected 1, got %s", result)
	}
}

func TestModByZero_Panics(t *testing.T) {
	expectPanic(t, "Division by zero", func() {
		Mod(NewInt(10), NewInt(0))
	})
}

func TestModFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot modulo", func() {
		Mod(NewFloat(10.0), NewFloat(3.0))
	})
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

func TestEqualIntString_Panics(t *testing.T) {
	expectPanic(t, "Cannot compare", func() {
		Equal(NewInt(1), NewString("1"))
	})
}

func TestEqualIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot compare", func() {
		Equal(NewInt(1), NewFloat(1.0))
	})
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

func TestLessThanIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot compare", func() {
		LessThan(NewInt(1), NewFloat(2.0))
	})
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

func TestGreaterThanIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot compare", func() {
		GreaterThan(NewInt(1), NewFloat(2.0))
	})
}

func TestLessEqualIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot compare", func() {
		LessEqual(NewInt(1), NewFloat(2.0))
	})
}

func TestGreaterEqualIntFloat_Panics(t *testing.T) {
	expectPanic(t, "Cannot compare", func() {
		GreaterEqual(NewInt(1), NewFloat(2.0))
	})
}
