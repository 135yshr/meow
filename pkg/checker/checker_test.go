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

func TestUntypedParamRequiresAnnotation(t *testing.T) {
	_, errs := check(t, `
meow greet(name) {
  bring "Hello, " + name
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for missing type annotation, got none")
	}
}

func TestMissingReturnTypeWithBring(t *testing.T) {
	_, errs := check(t, `
meow greet(name string) {
  bring "Hello, " + name
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for missing return type, got none")
	}
}

func TestVoidFunctionNoReturnType(t *testing.T) {
	// Void functions (no bring) don't need return type
	_, errs := check(t, `
meow noop() {
}
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

func TestReturnTypeMismatch(t *testing.T) {
	_, errs := check(t, `
meow greet(name string) int {
  bring "hello"
}
`)
	if len(errs) == 0 {
		t.Fatal("expected return type mismatch error, got none")
	}
}

func TestReturnTypeMatch(t *testing.T) {
	_, errs := check(t, `
meow add(a int, b int) int {
  bring a + b
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestUntypedFunctionRejected(t *testing.T) {
	_, errs := check(t, `
meow identity(x) {
  bring x
}
`)
	if len(errs) == 0 {
		t.Fatal("expected errors for untyped function, got none")
	}
}

func TestTypedIdentityPasses(t *testing.T) {
	_, errs := check(t, `
meow identity(x int) int {
  bring x
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestBringOutsideFunction(t *testing.T) {
	_, errs := check(t, `bring 42`)
	if len(errs) == 0 {
		t.Fatal("expected error for bring outside function, got none")
	}
}

func TestBareBringWithReturnType(t *testing.T) {
	_, errs := check(t, `
meow f(x int) int {
  bring
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for bare bring with return type, got none")
	}
}

func TestIfNonBoolCondition(t *testing.T) {
	_, errs := check(t, `sniff (42) {}`)
	if len(errs) == 0 {
		t.Fatal("expected error for non-bool if condition, got none")
	}
}

func TestRangeLoopCountForm(t *testing.T) {
	_, errs := check(t, `purr i (10) {
  nya(i)
}`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestRangeLoopRangeForm(t *testing.T) {
	_, errs := check(t, `purr i (1..20) {
  nya(i)
}`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestRangeLoopNonIntBound(t *testing.T) {
	_, errs := check(t, `purr i ("hello") {
  nya(i)
}`)
	if len(errs) == 0 {
		t.Fatal("expected error for non-int range bound, got none")
	}
}

func TestRangeLoopNonIntStart(t *testing.T) {
	_, errs := check(t, `purr i (1.5..10) {
  nya(i)
}`)
	if len(errs) == 0 {
		t.Fatal("expected error for non-int range start, got none")
	}
}

func TestSameScopeRedeclaration(t *testing.T) {
	_, errs := check(t, `
nyan x = 1
nyan x = 2
`)
	if len(errs) == 0 {
		t.Fatal("expected error for same-scope redeclaration, got none")
	}
}

func TestCrossScopeShadowingAllowed(t *testing.T) {
	_, errs := check(t, `
nyan x = 1
sniff (yarn) {
  nyan x = 2
  nya(x)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for cross-scope shadowing: %v", errs)
	}
}

func TestMatchArmTypeMismatch(t *testing.T) {
	_, errs := check(t, `
nyan x = 1
nyan y = peek(x) {
  1 => 42,
  2 => "hello",
  _ => 0
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for match arm type mismatch, got none")
	}
}

func TestAndNonBoolOperands(t *testing.T) {
	_, errs := check(t, `nyan x = 1 && 2`)
	if len(errs) == 0 {
		t.Fatal("expected error for non-bool AND operands, got none")
	}
}

func TestFuncArityMismatch(t *testing.T) {
	_, errs := check(t, `
meow add(a int, b int) int {
  bring a + b
}
nyan x = add(1)
`)
	if len(errs) == 0 {
		t.Fatal("expected error for arity mismatch, got none")
	}
}

func TestLambdaUntypedParam(t *testing.T) {
	_, errs := check(t, `nyan f = paw(x) { x }`)
	if len(errs) == 0 {
		t.Fatal("expected error for untyped lambda parameter, got none")
	}
}

func TestFuncNotAllPathsReturn(t *testing.T) {
	_, errs := check(t, `
meow abs(x int) int {
  sniff (x < 0) {
    bring -x
  }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for not returning on all paths, got none")
	}
}

func TestFuncAllPathsReturn(t *testing.T) {
	_, errs := check(t, `
meow abs(x int) int {
  sniff (x < 0) {
    bring -x
  } scratch {
    bring x
  }
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestListMixedTypes(t *testing.T) {
	_, errs := check(t, `nyan xs = [1, "hello"]`)
	if len(errs) == 0 {
		t.Fatal("expected error for mixed-type list, got none")
	}
}

func TestNotOnTruthyValue(t *testing.T) {
	// NOT operates on truthiness, so it accepts any type and returns bool
	info, errs := check(t, `nyan x = !123`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["x"].(types.BoolType); !ok {
		t.Errorf("expected bool, got %v", info.VarTypes["x"])
	}
}

func TestNotBoolOperand(t *testing.T) {
	_, errs := check(t, `nyan x = !yarn`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestBreedForwardReference(t *testing.T) {
	_, errs := check(t, `
breed Score = Points
breed Points = int
nyan s Score = 42
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for breed forward reference: %v", errs)
	}
}

func TestUnknownNamedType(t *testing.T) {
	_, errs := check(t, `nyan x Nonexistent = 42`)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown named type, got none")
	}
}

func TestBreedAliasInCondition(t *testing.T) {
	_, errs := check(t, `
breed Flag = bool
nyan f Flag = yarn
sniff (f) {
  nya(f)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for breed alias in condition: %v", errs)
	}
}

func TestBreedAliasInRange(t *testing.T) {
	_, errs := check(t, `
breed Count = int
nyan n Count = 5
purr i (n) {
  nya(i)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for breed alias in range: %v", errs)
	}
}

func TestBreedAliasUnaryMinus(t *testing.T) {
	_, errs := check(t, `
breed Num = int
nyan x Num = 42
nyan y = -x
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for breed alias unary minus: %v", errs)
	}
}

func TestCollarForwardReferenceToAlias(t *testing.T) {
	_, errs := check(t, `
collar Wrapper = Points
breed Points = int
nyan w = Wrapper(42)
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for collar forward reference to alias: %v", errs)
	}
}

func TestBreedAliasToCollarMemberAccess(t *testing.T) {
	_, errs := check(t, `
collar UserId = int
breed MyId = UserId
nyan id MyId = UserId(42)
nyan v = id.value
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for breed alias to collar member access: %v", errs)
	}
}

func TestBreedAliasToKittyMemberAccess(t *testing.T) {
	info, errs := check(t, `
kitty Cat {
  name: string,
  age: int
}
breed Pet = Cat
nyan p Pet = Cat("Nyantyu", 3)
nyan n = p.name
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for breed alias to kitty member access: %v", errs)
	}
	if _, ok := info.VarTypes["n"].(types.StringType); !ok {
		t.Errorf("expected string for n, got %v", info.VarTypes["n"])
	}
}

func TestAliasToCollarForwardChain(t *testing.T) {
	// breed -> collar -> breed chain with forward references
	_, errs := check(t, `
breed Wrapped = MyCollar
collar MyCollar = Points
breed Points = int
nyan w Wrapped = MyCollar(42)
nyan v = w.value
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for alias->collar->alias forward chain: %v", errs)
	}
}

func TestLearnUnknownType(t *testing.T) {
	_, errs := check(t, `
learn Unknown {
    meow show() string {
        bring "hello"
    }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for learn on unknown type, got none")
	}
}

func TestLearnDuplicateMethod(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
learn Cat {
    meow show() string {
        bring self.name
    }
    meow show() string {
        bring self.name
    }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate method, got none")
	}
}

func TestTrickSatisfaction(t *testing.T) {
	// Cat has show() string, so it structurally satisfies Showable
	info, errs := check(t, `
trick Showable {
    meow show() string
}
kitty Cat {
    name: string
}
learn Cat {
    meow show() string {
        bring self.name
    }
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	// Verify trick type was registered
	if _, ok := info.TrickTypes["Showable"]; !ok {
		t.Error("expected Showable trick to be registered")
	}
	// Verify learn method was registered
	if methods, ok := info.LearnImpls["Cat"]; !ok {
		t.Error("expected Cat learn impls to be registered")
	} else if _, ok := methods["show"]; !ok {
		t.Error("expected show method in Cat learn impls")
	}
}

func TestLearnMemberExprType(t *testing.T) {
	info, errs := check(t, `
kitty Cat {
    name: string
}
learn Cat {
    meow show() string {
        bring self.name
    }
}
nyan c = Cat("Nyantyu")
nyan s = c.show()
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := info.VarTypes["s"].(types.StringType); !ok {
		t.Errorf("expected string for s, got %v", info.VarTypes["s"])
	}
}

func TestCollarToCollarForwardRef(t *testing.T) {
	// collar whose underlying is another collar (resolved later)
	// Outer wraps Inner, so Outer(Inner(42)) is valid but Outer(42) is not
	_, errs := check(t, `
collar Outer = Inner
collar Inner = int
nyan o = Outer(Inner(42))
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for collar->collar forward ref: %v", errs)
	}
}
