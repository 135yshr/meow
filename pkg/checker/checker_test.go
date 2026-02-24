package checker_test

import (
	"testing"

	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
	"github.com/135yshr/meow/pkg/types"
)

func check(t *testing.T, input string) (*checker.TypeInfo, []*checker.TypeError) {
	t.Helper()
	l := lexer.New(input, "test.nyan")
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	c := checker.New()
	return c.Check(prog)
}

func TestInferIntLiteral(t *testing.T) {
	info, errs := check(t, `nyan x = 42`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.IntType); !ok {
		t.Errorf("expected int, got %v", info.VarTypes["x"])
	}
}

func TestInferFloatLiteral(t *testing.T) {
	info, errs := check(t, `nyan x = 3.14`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.FloatType); !ok {
		t.Errorf("expected float, got %v", info.VarTypes["x"])
	}
}

func TestInferStringLiteral(t *testing.T) {
	info, errs := check(t, `nyan x = "hello"`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.StringType); !ok {
		t.Errorf("expected string, got %v", info.VarTypes["x"])
	}
}

func TestInferBoolLiteral(t *testing.T) {
	info, errs := check(t, `nyan x = yarn`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.BoolType); !ok {
		t.Errorf("expected bool, got %v", info.VarTypes["x"])
	}
}

func TestTypedVarMatchesLiteral(t *testing.T) {
	_, errs := check(t, `nyan x int = 42`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestTypedVarMismatch(t *testing.T) {
	_, errs := check(t, `nyan x int = "hello"`)
	if len(errs) == 0 {
		t.Fatal("expected type error, got none")
	}
}

func TestInferAddIntInt(t *testing.T) {
	info, errs := check(t, `nyan x = 1 + 2`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.IntType); !ok {
		t.Errorf("expected int, got %v", info.VarTypes["x"])
	}
}

func TestInferAddStringString(t *testing.T) {
	info, errs := check(t, `nyan x = "a" + "b"`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.StringType); !ok {
		t.Errorf("expected string, got %v", info.VarTypes["x"])
	}
}

func TestErrorAddIntString(t *testing.T) {
	_, errs := check(t, `nyan x = 1 + "hello"`)
	if len(errs) == 0 {
		t.Fatal("expected type error, got none")
	}
}

func TestInferComparison(t *testing.T) {
	info, errs := check(t, `nyan x = 1 < 2`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.BoolType); !ok {
		t.Errorf("expected bool, got %v", info.VarTypes["x"])
	}
}

func TestUntypedCodePassesCheck(t *testing.T) {
	// Existing untyped code should pass without errors
	_, errs := check(t, `
meow greet(name) {
  bring "Hello, " + name
}
nyan msg = greet("World")
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestTypedFuncCallTypeCheck(t *testing.T) {
	_, errs := check(t, `
meow add(a int, b int) int {
  bring a + b
}
nyan result = add(1, 2)
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestTypedFuncCallTypeMismatch(t *testing.T) {
	_, errs := check(t, `
meow add(a int, b int) int {
  bring a + b
}
nyan result = add(1, "two")
`)
	if len(errs) == 0 {
		t.Fatal("expected type error for argument mismatch")
	}
}

func TestInferFuncReturnType(t *testing.T) {
	info, errs := check(t, `
meow double(x int) int {
  bring x + x
}
nyan result = double(5)
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["result"].(types.IntType); !ok {
		t.Errorf("expected int for result, got %v", info.VarTypes["result"])
	}
}

func TestToIntReturnsInt(t *testing.T) {
	info, errs := check(t, `nyan x = to_int(3.14)`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.IntType); !ok {
		t.Errorf("expected int, got %v", info.VarTypes["x"])
	}
}

func TestToStringReturnsString(t *testing.T) {
	info, errs := check(t, `nyan x = to_string(42)`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.StringType); !ok {
		t.Errorf("expected string, got %v", info.VarTypes["x"])
	}
}
