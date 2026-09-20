package codegen_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

// generateFuzz runs the fuzz generator the way the compiler does, with the
// checker's type information in place.
func generateFuzz(t *testing.T, input string) (helpers, tests string) {
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
	ti, checkErrs := c.Check(prog)
	if len(checkErrs) > 0 {
		for _, e := range checkErrs {
			t.Errorf("checker error: %s", e)
		}
		t.FailNow()
	}
	g := codegen.New()
	g.SetTypeInfo(ti)
	helpers, tests, _, err := g.GenerateFuzz(prog)
	if err != nil {
		t.Fatal(err)
	}
	return helpers, tests
}

const fuzzSource = `meow half(x int) int {
  bring x / 2
}

meow fuzz_half_round_trip(x int) {
  seed(1)
  expect(half(x) * 2, x)
}
`

// The body of a fuzz target ends its statements with `return __f` for a
// Furball, so it cannot be generated into the f.Fuzz closure, which returns
// nothing. It gets a function of its own instead.
func TestFuzzBodyIsAFunctionOfItsOwn(t *testing.T) {
	helpers, tests := generateFuzz(t, fuzzSource)

	if !strings.Contains(helpers, "func fuzz_half_round_trip(x meow.Value) meow.Value {") {
		t.Errorf("expected a boxed body function, got:\n%s", helpers)
	}
	if !strings.Contains(helpers, "return __f") {
		t.Errorf("expected the body to keep its Furball short-circuit, got:\n%s", helpers)
	}
	if strings.Contains(tests, "return __f") {
		t.Errorf("a Furball returned from inside the f.Fuzz closure would not compile, got:\n%s", tests)
	}
}

// A Furball reaching the closure is what the fuzzer was run to find, so it
// fails the case, naming both the assertion and the input that provoked it.
func TestAFuzzFailureIsRaisedOnTheTestingT(t *testing.T) {
	_, tests := generateFuzz(t, fuzzSource)

	if !strings.Contains(tests, "meow.AsFurball(fuzz_half_round_trip(x))") {
		t.Errorf("expected the closure to call the body and inspect its answer, got:\n%s", tests)
	}
	if !strings.Contains(tests, `__t.Fatalf("Hiss! fuzz_half_round_trip(x=%#v) failed, nya~: %s", x_raw, __f.Message)`) {
		t.Errorf("expected a failure naming the target, the input and the message, got:\n%s", tests)
	}
}

// The testing.T is named __t, so that a fuzz parameter called `t` shadows
// nothing the failure report needs.
func TestAFuzzParameterMayBeCalledT(t *testing.T) {
	_, tests := generateFuzz(t, `meow fuzz_shadow(t int) {
  seed(1)
  judge(t == t)
}
`)

	if !strings.Contains(tests, "func(__t *testing.T, t_raw int64)") {
		t.Errorf("expected the testing.T to be named __t, got:\n%s", tests)
	}
	if !strings.Contains(tests, "__t.Fatalf(") {
		t.Errorf("expected the failure to be raised on __t, got:\n%s", tests)
	}
}

// Go's testing package answers a fuzz target it cannot feed with a panic and
// a stack, which says nothing about the .nyan file that caused it.
func TestAFuzzTargetWithoutParametersIsRefused(t *testing.T) {
	l := lexer.New(`meow fuzz_nothing() {
  judge(1 == 1)
}
`, "test.nyan")
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	_, _, _, err := codegen.New().GenerateFuzz(prog)
	if err == nil {
		t.Fatal("expected a fuzz target with no parameters to be refused")
	}
	if !strings.Contains(err.Error(), "fuzz_nothing takes no parameters") {
		t.Errorf("expected the error to name the target, got: %v", err)
	}
}

// A package nabbed for a fuzz target and used nowhere else is imported by the
// file that holds the bodies, and by no other: an import the test file never
// mentions would not compile.
func TestAPackageUsedOnlyByAFuzzBodyIsImportedWhereTheBodyIs(t *testing.T) {
	helpers, tests := generateFuzz(t, `nab "file"

meow fuzz_reads(name string) {
  seed("nope")
  nyan text = file.snoop(name) ~> "fallback"
  judge(len(text) >= 0)
}
`)

	if !strings.Contains(helpers, `meow_file "github.com/135yshr/meow/runtime/file"`) {
		t.Errorf("expected the file package to be imported beside the body, got:\n%s", helpers)
	}
	if strings.Contains(tests, "meow_file") {
		t.Errorf("the test file neither uses nor should import the file package, got:\n%s", tests)
	}
}
