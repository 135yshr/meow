package checker_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/ast"
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
	_, errs := check(t, `purr i (yarn) {
  nya(i)
}`)
	if len(errs) == 0 {
		t.Fatal("expected error for non-int range bound, got none")
	}
}

func TestRangeLoopStringFormDirect(t *testing.T) {
	_, errs := check(t, `purr ch ("hello") {
  nya(ch)
}`)
	if len(errs) == 0 {
		t.Fatal("expected error for direct string iteration, got none")
	}
}

func TestRangeLoopWithToRunes(t *testing.T) {
	_, errs := check(t, `purr ch (to_runes("hello")) {
  nya(ch)
}`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestRangeLoopWithToRunesAndIndex(t *testing.T) {
	_, errs := check(t, `purr i, ch (to_runes("hello")) {
  nya(i)
  nya(ch)
}`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestRangeLoopTwoVarNonList(t *testing.T) {
	_, errs := check(t, `purr i, ch (10) {
  nya(i)
}`)
	if len(errs) == 0 {
		t.Fatal("expected error for two-variable form with non-list, got none")
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

func TestFuncArityTooMany(t *testing.T) {
	_, errs := check(t, `
meow add(a int, b int) int {
  bring a + b
}
nyan x = add(1, 2, 3)
`)
	if len(errs) == 0 {
		t.Fatal("expected error for too many arguments, got none")
	}
}

func TestFuncPartialApplication(t *testing.T) {
	info, errs := check(t, `
meow add(a int, b int) int {
  bring a + b
}
nyan x = add(1)
nyan y = x(2)
`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors for partial application: %v", errs)
	}
	// x should be a FuncType (int) -> int
	xType := info.VarTypes["x"]
	if xType == nil {
		t.Fatal("x should have a type")
	}
	ft, ok := xType.(types.FuncType)
	if !ok {
		t.Fatalf("x should be FuncType, got %T", xType)
	}
	if len(ft.Params) != 1 {
		t.Fatalf("partial add(1) should have 1 remaining param, got %d", len(ft.Params))
	}
	// y should be IntType (result of calling x(2))
	if _, ok := info.VarTypes["y"].(types.IntType); !ok {
		t.Fatalf("y should be int, got %T", info.VarTypes["y"])
	}
}

func TestLambdaUntypedParam(t *testing.T) {
	_, errs := check(t, `nyan f = paw(x) { x }`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors for untyped lambda parameter: %v", errs)
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

// A litter of mixed things is a litter of anything. `[Cat(...), Dog(...)]` was
// always read that way; primitives alone were refused on top of it, so the two
// halves of the same idea disagreed — and the playground, which does not run
// the checker, ran both.
func TestListOfMixedThingsIsAListOfAnything(t *testing.T) {
	inputs := []string{
		`nyan xs = [1, "hello"]`,
		`nyan xs = [1, yarn, "a", 2.5]`,
		`nyan xs = [1, catnap]`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			info, errs := check(t, input)
			if len(errs) > 0 {
				t.Fatalf("got %v, want no errors", errs)
			}
			list, ok := info.VarTypes["xs"].(types.ListType)
			if !ok {
				t.Fatalf("got %v, want a litter", info.VarTypes["xs"])
			}
			if !types.IsAny(list.Elem) {
				t.Errorf("elements are %v, want anything", list.Elem)
			}
		})
	}
}

// One kind all the way through still says which kind it is.
func TestListOfOneKindKeepsIt(t *testing.T) {
	info, errs := check(t, `nyan xs = [1, 2, 3]`)
	if len(errs) > 0 {
		t.Fatalf("got %v, want no errors", errs)
	}
	list, ok := info.VarTypes["xs"].(types.ListType)
	if !ok {
		t.Fatalf("got %v, want a litter", info.VarTypes["xs"])
	}
	if _, ok := list.Elem.(types.IntType); !ok {
		t.Errorf("elements are %v, want int", list.Elem)
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
groom Unknown {
    meow show() string {
        bring "hello"
    }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for groom on unknown type, got none")
	}
}

func TestLearnDuplicateMethod(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
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
pose Showable {
    meow show() string
}
kitty Cat {
    name: string
}
groom Cat {
    meow show() string {
        bring self.name
    }
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	// Verify pose type was registered
	if _, ok := info.TrickTypes["Showable"]; !ok {
		t.Error("expected Showable pose to be registered")
	}
	// Verify groom method was registered
	if methods, ok := info.LearnImpls["Cat"]; !ok {
		t.Error("expected Cat groom impls to be registered")
	} else if _, ok := methods["show"]; !ok {
		t.Error("expected show method in Cat groom impls")
	}
}

func TestLearnMemberExprType(t *testing.T) {
	info, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
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

func TestSelfOutsideLearn(t *testing.T) {
	_, errs := check(t, `
nyan x = self.name
`)
	if len(errs) == 0 {
		t.Fatal("expected error for self outside groom, got none")
	}
	found := false
	for _, e := range errs {
		if contains(e.Message, "self can only be used inside groom methods") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'self can only be used inside groom methods' error, got: %v", errs)
	}
}

// A method read rather than called is that method, bound to what it was
// reached through, and so has the method's own type. Writing one without
// calling it used to be an error, which made a method the one thing in the
// language that could not be piped into or handed to lick.
func TestAMethodReadRatherThanCalledHasTheMethodsType(t *testing.T) {
	info, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
    meow older(by int) string {
        bring self.name
    }
}
nyan c = Cat("Nyantyu")
nyan f = c.older
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	ft, ok := info.VarTypes["f"].(types.FuncType)
	if !ok {
		t.Fatalf("expected a function for f, got %v", info.VarTypes["f"])
	}
	if len(ft.Params) != 1 {
		t.Errorf("takes %d arguments, want the one the method takes", len(ft.Params))
	}
	if _, isString := ft.Return.(types.StringType); !isString {
		t.Errorf("gives back %v, want what the method gives back", ft.Return)
	}
}

// A member that is neither a field nor a method is still an error, and still
// says so where it is written.
func TestAMemberThatIsNeitherFieldNorMethod(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
    meow show() string {
        bring self.name
    }
}
nyan c = Cat("Nyantyu")
nyan f = c.purr
`)
	if len(errs) == 0 {
		t.Fatal("expected an error for a member that is not there, got none")
	}
	found := false
	for _, e := range errs {
		if contains(e.Message, "has no field or method purr") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected it to name the member, got: %v", errs)
	}
}

func TestLearnMethodMissingParamType(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
    meow greet(msg) string {
        bring msg
    }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for missing param type in groom method, got none")
	}
}

func TestLearnMethodMissingReturnType(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
    meow show() {
        bring self.name
    }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for missing return type in groom method with bring, got none")
	}
}

func TestLearnMethodNotAllPathsReturn(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string,
    age: int
}
groom Cat {
    meow describe() string {
        sniff (self.age < 1) {
            bring "kitten"
        }
    }
}
`)
	if len(errs) == 0 {
		t.Fatal("expected error for not returning on all paths in groom method, got none")
	}
}

func TestLearnMethodCallArityMismatch(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
groom Cat {
    meow greet(prefix string) string {
        bring prefix + self.name
    }
}
nyan c = Cat("Nyantyu")
nyan s = c.greet("Hi", "extra")
`)
	if len(errs) == 0 {
		t.Fatal("expected error for method arity mismatch, got none")
	}
}

func TestLearnMethodCallUnknownMethod(t *testing.T) {
	_, errs := check(t, `
kitty Cat {
    name: string
}
nyan c = Cat("Nyantyu")
nyan s = c.nonexistent()
`)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown method call, got none")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
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

func TestImportVarCollision(t *testing.T) {
	_, errs := check(t, `
nab "file"
nyan file = 1
`)
	if len(errs) == 0 {
		t.Fatal("expected error for import-variable name collision, got none")
	}
	found := false
	for _, e := range errs {
		if contains(e.Message, "shadows imported package") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'shadows imported package' error, got: %v", errs)
	}
}

func TestImportAliasAvoidCollision(t *testing.T) {
	_, errs := check(t, `
nab "file" tag f
nyan file = 1
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: alias should avoid collision: %v", errs)
	}
}

func TestImportAliasCollision(t *testing.T) {
	_, errs := check(t, `
nab "file" tag f
nyan f = 1
`)
	if len(errs) == 0 {
		t.Fatal("expected error for alias-variable collision, got none")
	}
	found := false
	for _, e := range errs {
		if contains(e.Message, "shadows imported package") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'shadows imported package' error, got: %v", errs)
	}
}

func TestImportDuplicateAlias(t *testing.T) {
	_, errs := check(t, `
nab "file" tag f
nab "http" tag f
`)
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate import alias, got none")
	}
	found := false
	for _, e := range errs {
		if contains(e.Message, "import name 'f' already used") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'import name already used' error, got: %v", errs)
	}
}

func TestImportBlockScopeNoCollision(t *testing.T) {
	// Block-level variable should not collide with import
	_, errs := check(t, `
nab "file"
meow test_fn() {
  nyan file = 1
  nya(file)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: block-level var should not collide: %v", errs)
	}
}

// hasPurityError reports whether any error is a purity violation. All purity
// messages share the "must not" phrasing, so matching it confirms the purity
// rule fired rather than an incidental parse/type error.
func hasPurityError(errs []*checker.TypeError) bool {
	for _, e := range errs {
		if contains(e.Error(), "must not") {
			return true
		}
	}
	return false
}

func TestPureFuncCallsImpureBuiltin(t *testing.T) {
	_, errs := check(t, `
trill meow noisy(a int) int {
  nya(a)
  bring a
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for trill function calling nya, got: %v", errs)
	}
}

func TestPureFuncCallsHiss(t *testing.T) {
	_, errs := check(t, `
trill meow boom(a int) int {
  bring hiss("bad")
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for trill function calling hiss, got: %v", errs)
	}
}

func TestPureFuncCallsGag(t *testing.T) {
	_, errs := check(t, `
trill meow recover(a int) int {
  bring gag(a)
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for trill function calling gag, got: %v", errs)
	}
}

func TestPureFuncCallsNonPureFunc(t *testing.T) {
	_, errs := check(t, `
meow helper(a int) int {
  bring a + 1
}
trill meow user(a int) int {
  bring helper(a)
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for trill function calling non-pure function, got: %v", errs)
	}
}

func TestPureFuncCallsImportedPackage(t *testing.T) {
	_, errs := check(t, `
nab "http"
trill meow grab(url string) furball {
  bring http.pounce(url)
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for trill function using imported package, got: %v", errs)
	}
}

func TestPureFuncReadsImportedPackageAsValue(t *testing.T) {
	// Reading a package member as a value (not a call) is still impure.
	_, errs := check(t, `
nab "file"
trill meow grab() furball {
  bring file.snoop
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for reading imported package member, got: %v", errs)
	}
}

func TestPureFuncCallsGroomMethod(t *testing.T) {
	// groom methods are plain meow functions that may perform I/O, so calling
	// one from a trill function must be rejected (transitive purity).
	_, errs := check(t, `
kitty Cat {
  name: string
}
groom Cat {
  meow noisy() int {
    nya(self.name)
    bring 1
  }
}
trill meow f(c Cat) int {
  bring c.noisy()
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for trill function calling groom method, got: %v", errs)
	}
}

func TestPureFuncDefinesImpureNestedFunc(t *testing.T) {
	// A function nested inside a trill body must itself be pure.
	_, errs := check(t, `
trill meow f(x int) int {
  meow noisy(y int) int {
    nya(y)
    bring y
  }
  bring x
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for impure nested function, got: %v", errs)
	}
}

func TestPureFuncDefinesImpureNestedGroom(t *testing.T) {
	// A groom block nested in a trill body must also be pure.
	_, errs := check(t, `
kitty Cat {
  name: string
}
trill meow f(x int) int {
  groom Cat {
    meow noisy() int {
      nya(x)
      bring 1
    }
  }
  bring x
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for impure nested groom, got: %v", errs)
	}
}

func TestPureFuncPassesImpureLambdaToLick(t *testing.T) {
	_, errs := check(t, `
trill meow shout(xs litter) litter {
  bring lick(xs, paw(x int) { nya(x) })
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for impure lambda passed to lick, got: %v", errs)
	}
}

func TestPureFuncCallsPureFuncAndBuiltins(t *testing.T) {
	_, errs := check(t, `
trill meow inc(a int) int {
  bring a + 1
}
trill meow twice(a int) int {
  bring inc(inc(a))
}
trill meow sized(xs litter) int {
  bring len(xs)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for pure trill functions: %v", errs)
	}
}

func TestPureFuncFullyPure(t *testing.T) {
	_, errs := check(t, `
trill meow square(a int) int {
  nyan r = a * a
  bring r
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for fully pure trill function: %v", errs)
	}
}

// --- step 2: non-pure functions referenced as bare values ---

func TestPureFuncBindsNonPureFuncAsValue(t *testing.T) {
	// Binding a non-pure function to a variable (without calling it) still lets
	// the impure function escape the pure body, so it must be rejected.
	_, errs := check(t, `
meow helper(a int) int {
  nya(a)
  bring a
}
trill meow user(a int) int {
  nyan f = helper
  bring a
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for non-pure function bound as value, got: %v", errs)
	}
}

func TestPureFuncPassesNonPureFuncAsArg(t *testing.T) {
	// Passing a non-pure function as an argument to a higher-order builtin is
	// just as impure as calling it inline.
	_, errs := check(t, `
meow dbl(a int) int {
  nya(a)
  bring a * 2
}
trill meow user(xs litter) litter {
  bring lick(xs, dbl)
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for non-pure function passed as argument, got: %v", errs)
	}
}

func TestPureFuncReturnsNonPureFuncAsValue(t *testing.T) {
	// Returning a non-pure function by name leaks impurity to the caller.
	_, errs := check(t, `
meow helper(a int) int {
  nya(a)
  bring a
}
trill meow user() int {
  bring helper
}
`)
	if !hasPurityError(errs) {
		t.Fatalf("expected purity error for non-pure function returned as value, got: %v", errs)
	}
}

func TestPureFuncReferencesPureFuncAsValue(t *testing.T) {
	// Referencing another pure function by name (as a value) is allowed — the
	// transitive purity guarantee still holds.
	_, errs := check(t, `
trill meow inc(a int) int {
  bring a + 1
}
trill meow user(a int) int {
  nyan f = inc
  bring a
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for pure function referenced as value: %v", errs)
	}
}

// `x = ...` declares a variable rather than assigning to one, so inside a block
// it shadows the outer binding and leaves it untouched. That used to pass the
// checker and produce Go that either did nothing useful or failed to compile
// with "declared and not used", so it is now reported directly.
func TestImplicitReassignmentInBlockIsRejected(t *testing.T) {
	_, errs := check(t, `
nyan total = 0
purr i (5) {
  total = total + 1
}
nya(total)
`)
	if len(errs) == 0 {
		t.Fatal("expected an error for reassigning an outer binding")
	}
	if !strings.Contains(errs[0].Message, "cannot be reassigned") {
		t.Errorf("expected a reassignment error, got %q", errs[0].Message)
	}
}

func TestImplicitReassignmentInsideFunctionIsRejected(t *testing.T) {
	_, errs := check(t, `
meow tally(n int) int {
  nyan total = 0
  purr i (n) {
    total = total + 1
  }
  bring total
}
`)
	if len(errs) == 0 {
		t.Fatal("expected an error for reassigning an outer binding")
	}
}

// Declaring a fresh name in an inner scope is not a reassignment.
func TestImplicitDeclarationOfNewNameIsAllowed(t *testing.T) {
	_, errs := check(t, `
nyan total = 0
purr i (5) {
  running = total + 1
  nya(running)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

// An explicit nyan declaration still shadows, which is a deliberate act rather
// than a mistaken assignment.
func TestExplicitShadowingIsAllowed(t *testing.T) {
	_, errs := check(t, `
nyan total = 0
purr i (5) {
  nyan total = i
  nya(total)
}
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

// Comparing against catnap asks "is this missing?", which any type may ask —
// env.hunt and an absent map key both answer with it.
func TestCompareAgainstNilIsAllowed(t *testing.T) {
	sources := []string{
		`nyan x = "cat"
nyan missing = x == catnap`,
		`nyan x = "cat"
nyan missing = catnap == x`,
		`nyan x = 42
nyan present = x != catnap`,
		`nyan x = catnap
nyan same = x == catnap`,
	}
	for _, src := range sources {
		_, errs := check(t, src)
		if len(errs) > 0 {
			t.Errorf("unexpected errors for %q: %v", src, errs)
		}
	}
}

func TestCompareMismatchedTypesStillErrors(t *testing.T) {
	_, errs := check(t, `nyan bad = 1 == "one"`)
	if len(errs) == 0 {
		t.Fatal("expected a type error comparing int and string")
	}
	if !strings.Contains(errs[0].Error(), "Cannot compare") {
		t.Errorf("expected a Cannot compare error, got %v", errs[0])
	}
}

// Left to the compiler, a bolt with no loop around it became a Go break with
// nothing to break out of — a message about generated code the reader never
// wrote.
func TestBoltAndSlinkNeedALoop(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bolt at the top level", "bolt\n", "bolt used outside a purr loop"},
		{"slink at the top level", "slink\n", "slink used outside a purr loop"},
		{
			// The lambda runs per element, wherever lick was called from, so
			// the loop outside is not one it can leave.
			"bolt inside a lambda in a loop",
			"purr i (3) {\n  nyan f = lick([1], paw(v) { bolt })\n}\n",
			"bolt used outside a purr loop",
		},
		{
			"slink inside a function declared in a loop",
			"purr i (3) {\n  meow inner(v int) int { slink }\n  nya(\"x\")\n}\n",
			"slink used outside a purr loop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := check(t, tt.input)

			if !hasError(errs, tt.want) {
				t.Errorf("got %v, want an error saying %q", errs, tt.want)
			}
		})
	}
}

func TestBoltAndSlinkInsideALoopAreFine(t *testing.T) {
	inputs := []string{
		"purr i (3) {\n  bolt\n}\n",
		"purr x ([1, 2]) {\n  slink\n}\n",
		"purr (yarn) {\n  bolt\n}\n",
		// Nested: the inner one belongs to the inner loop.
		"purr i (3) {\n  purr j (3) {\n    bolt\n  }\n}\n",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if _, errs := check(t, input); len(errs) > 0 {
				t.Errorf("got %v, want no errors", errs)
			}
		})
	}
}

// A conditional purr is checked the same way sniff is. Without this the count
// form's `purr i (n)` and the conditional form's `purr (n)` read alike but
// meant different things, and inside a typed function the second one reached
// the Go compiler as a loop over an integer.
func TestPurrConditionMustBeBool(t *testing.T) {
	_, errs := check(t, "meow f(n int) int {\n  purr (n) {\n    bolt\n  }\n  bring n\n}\nnya(f(3))\n")

	if !hasError(errs, "Condition must be bool, got int") {
		t.Errorf("got %v, want an error about a non-bool condition", errs)
	}
}

// A bring anywhere in a function needs the return type written down, and the
// conditional purr's body is no exception.
func TestBringInsideAConditionalPurrNeedsAReturnType(t *testing.T) {
	_, errs := check(t, "meow f() {\n  purr (yarn) {\n    bring 1\n  }\n}\nnya(f())\n")

	if !hasError(errs, "no return type annotation") {
		t.Errorf("got %v, want an error about the missing return type", errs)
	}
}

// A trill function stays pure all the way down, including inside a loop body.
func TestConditionalPurrBodyIsCheckedForPurity(t *testing.T) {
	_, errs := check(t, "trill meow f() int {\n  purr (yarn) {\n    nya(\"side effect\")\n    bolt\n  }\n  bring 1\n}\nnya(f())\n")

	if !hasError(errs, "nya") {
		t.Errorf("got %v, want an error about the impure call", errs)
	}
}

// hasError reports whether any error's message contains want.
func hasError(errs []*checker.TypeError, want string) bool {
	for _, e := range errs {
		if strings.Contains(e.Message, want) {
			return true
		}
	}
	return false
}

// A Go import path can end in something that is not a name a program can
// write. Rather than invent one, the program is asked to name it.
func TestGoImportThatCannotBeNamed(t *testing.T) {
	_, errs := check(t, `nab go "github.com/aws/aws-sdk-go-v2"`)

	if len(errs) == 0 {
		t.Fatal("got no errors, want one asking for a tag")
	}
	if !strings.Contains(errs[0].Error(), "tag") {
		t.Errorf("says %q, want it to say what to do about it", errs[0])
	}
}

// Named with a tag, the same path is fine.
func TestGoImportNamedWithATag(t *testing.T) {
	if _, errs := check(t, `nab go "github.com/aws/aws-sdk-go-v2" tag aws`); len(errs) > 0 {
		t.Errorf("got %v, want no errors", errs)
	}
}

// Two imports cannot share one name, however each came by it.
func TestGoImportsThatCollide(t *testing.T) {
	_, errs := check(t, "nab go \"net/url\"\nnab go \"example.com/url\"")

	if len(errs) == 0 {
		t.Fatal("got no errors, want one about the name being taken")
	}
	if !strings.Contains(errs[0].Error(), "url") {
		t.Errorf("says %q, want it to name the name", errs[0])
	}
}

// Which declaration a name reaches is settled by the scope it was written in.
// A local holding a function has the same type as the function it shadows, so
// the type cannot say which is which and the resolution is recorded instead.
func TestWhichDeclarationANameReaches(t *testing.T) {
	info, errs := check(t, `
meow double(n int) int { bring n * 2 }
meow shadowed() litter {
    nyan double = paw(x) { bring x + 100 }
    bring lick([1, 2], double)
}
nya(lick([1, 2], double))
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	var reaching, taken int
	for node := range info.ExprTypes {
		id, isIdent := node.(*ast.Ident)
		if !isIdent || id.Name != "double" {
			continue
		}
		if info.FuncRefs[id] {
			reaching++
			continue
		}
		taken++
	}
	if reaching != 1 || taken != 1 {
		t.Errorf("%d occurrences reach the function and %d were taken over, want 1 of each", reaching, taken)
	}
}

// A call through a name a local took over is checked against what the local
// is. Checking it against the function it shadows would refuse a call that is
// perfectly good — here, one that takes two arguments where the top-level
// function takes one.
func TestACallThroughALocalIsCheckedAgainstTheLocal(t *testing.T) {
	_, errs := check(t, `
meow greet(who string) string { bring "hi " + who }
meow shadowed() string {
    nyan greet = paw(a, b) { bring a + "/" + b }
    bring greet("x", "y")
}
nya(shadowed())
`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

// And a call that still reaches the top-level function is checked against it.
func TestACallThatReachesTheFunctionIsStillChecked(t *testing.T) {
	_, errs := check(t, `
meow greet(who string) string { bring "hi " + who }
nya(greet("x", "y"))
`)
	if len(errs) == 0 {
		t.Fatal("expected an error for too many arguments, got none")
	}
}
