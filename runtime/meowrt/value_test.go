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
		"name": meowrt.NewString("Tama"),
		"age":  meowrt.NewInt(3),
	})
	got := meowrt.ToJSON(m)
	// Keys are sorted alphabetically
	expected := `{"age":3,"name":"Tama"}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestToJSONNestedMapList(t *testing.T) {
	m := meowrt.NewMap(map[string]meowrt.Value{
		"cats": meowrt.NewList(
			meowrt.NewString("Tama"),
			meowrt.NewString("Mochi"),
		),
		"count": meowrt.NewInt(2),
	})
	got := meowrt.ToJSON(m)
	expected := `{"cats":["Tama","Mochi"],"count":2}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestToJSONUnsupportedType(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, "Hiss!") || !strings.Contains(msg, "Func") {
			t.Errorf("expected Hiss error about Func, got %q", msg)
		}
	}()
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value { return meowrt.NewNil() })
	meowrt.ToJSON(fn)
}
