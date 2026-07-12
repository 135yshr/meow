package meowrt_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
)

func TestToJSONString(t *testing.T) {
	got := meowrt.ToJSON(meowrt.NewString("hello"))
	if got != `"hello"` {
		t.Errorf("expected %q, got %q", `"hello"`, got)
	}
}

func TestToJSONStringEscape(t *testing.T) {
	got := meowrt.ToJSON(meowrt.NewString(`say "hi"`))
	if got != `"say \"hi\""` {
		t.Errorf("expected %q, got %q", `"say \"hi\""`, got)
	}
}

func TestToJSONInt(t *testing.T) {
	got := meowrt.ToJSON(meowrt.NewInt(42))
	if got != "42" {
		t.Errorf("expected %q, got %q", "42", got)
	}
}

func TestToJSONFloat(t *testing.T) {
	got := meowrt.ToJSON(meowrt.NewFloat(3.14))
	if got != "3.14" {
		t.Errorf("expected %q, got %q", "3.14", got)
	}
}

func TestToJSONBool(t *testing.T) {
	if got := meowrt.ToJSON(meowrt.NewBool(true)); got != "true" {
		t.Errorf("expected %q, got %q", "true", got)
	}
	if got := meowrt.ToJSON(meowrt.NewBool(false)); got != "false" {
		t.Errorf("expected %q, got %q", "false", got)
	}
}

func TestToJSONNil(t *testing.T) {
	got := meowrt.ToJSON(meowrt.NewNil())
	if got != "null" {
		t.Errorf("expected %q, got %q", "null", got)
	}
}

func TestToJSONList(t *testing.T) {
	list := meowrt.NewList(meowrt.NewInt(1), meowrt.NewString("two"), meowrt.NewBool(true))
	got := meowrt.ToJSON(list)
	expected := `[1,"two",true]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestToJSONMap(t *testing.T) {
	m := meowrt.NewMap(map[string]meowrt.Value{
		"name": meowrt.NewString("Nyantyu"),
		"age":  meowrt.NewInt(3),
	})
	got := meowrt.ToJSON(m)
	// Keys are sorted alphabetically
	expected := `{"age":3,"name":"Nyantyu"}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestToJSONNestedMapList(t *testing.T) {
	m := meowrt.NewMap(map[string]meowrt.Value{
		"cats": meowrt.NewList(
			meowrt.NewString("Nyantyu"),
			meowrt.NewString("Tyako"),
		),
		"count": meowrt.NewInt(2),
	})
	got := meowrt.ToJSON(m)
	expected := `{"cats":["Nyantyu","Tyako"],"count":2}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestAsIntSuccess(t *testing.T) {
	got := meowrt.AsInt(meowrt.NewInt(42))
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestTryAsIntFurball(t *testing.T) {
	_, f := meowrt.TryAsInt(meowrt.NewString("hello"))
	if f == nil {
		t.Fatal("expected Furball, got nil")
	}
	if !strings.Contains(f.Message, "Hiss!") || !strings.Contains(f.Message, "expected int") {
		t.Errorf("unexpected message: %q", f.Message)
	}
}

func TestAsIntMismatchPanics(t *testing.T) {
	// AsInt panics on type mismatch (including Furball input) so failures
	// don't silently turn into zero values — Gag's recover converts the
	// panic back into a Furball at the typed-path boundary.
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on type mismatch")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "expected int") {
			t.Errorf("unexpected panic value: %v", r)
		}
	}()
	meowrt.AsInt(meowrt.NewString("hello"))
}

func TestAsIntFurballPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on Furball input")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "boom") {
			t.Errorf("expected Furball message in panic, got %v", r)
		}
	}()
	meowrt.AsInt(&meowrt.Furball{Message: "Hiss! boom, nya~"})
}

func TestAsFloatSuccess(t *testing.T) {
	got := meowrt.AsFloat(meowrt.NewFloat(3.14))
	if got != 3.14 {
		t.Errorf("expected 3.14, got %g", got)
	}
}

func TestAsStringSuccess(t *testing.T) {
	got := meowrt.AsString(meowrt.NewString("hello"))
	if got != "hello" {
		t.Errorf("expected hello, got %s", got)
	}
}

func TestAsBoolSuccess(t *testing.T) {
	got := meowrt.AsBool(meowrt.NewBool(true))
	if got != true {
		t.Errorf("expected true, got %t", got)
	}
}

func TestToJSONUnsupportedType(t *testing.T) {
	// Unsupported types now serialize to a JSON-encoded Hiss error message
	// instead of panicking. This keeps ToJSON total over Value.
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value { return meowrt.NewNil() })
	got := meowrt.ToJSON(fn)
	if !strings.Contains(got, "Hiss!") || !strings.Contains(got, "Func") {
		t.Errorf("expected JSON-quoted Hiss error about Func, got %q", got)
	}
}
