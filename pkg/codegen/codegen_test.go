package codegen_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

func generate(t *testing.T, input string) string {
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
	g := codegen.New()
	code, err := g.Generate(prog)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func TestHelloWorld(t *testing.T) {
	code := generate(t, `nyan name = "Nyantyu"
nya(name)`)
	if !strings.Contains(code, `meow.NewString("Nyantyu")`) {
		t.Error("expected meow.NewString")
	}
	if !strings.Contains(code, `meow.Nya(name)`) {
		t.Error("expected meow.Nya")
	}
}

func TestFuncGen(t *testing.T) {
	code := generate(t, `meow greet(who) {
  bring "Hello, " + who + "!"
}
nyan name = "Nyantyu"
nya(greet(name))`)
	if !strings.Contains(code, "func greet(who meow.Value) meow.Value") {
		t.Error("expected function declaration")
	}
	if !strings.Contains(code, "meow.Add(") {
		t.Error("expected meow.Add")
	}
}

func TestArithmetic(t *testing.T) {
	code := generate(t, `nyan x = 1 + 2 * 3`)
	if !strings.Contains(code, "meow.Add(") {
		t.Error("expected meow.Add")
	}
	if !strings.Contains(code, "meow.Mul(") {
		t.Error("expected meow.Mul")
	}
}

func TestFetchAndMemberCall(t *testing.T) {
	code := generate(t, `nab "file"
nyan content = file.snoop("data.txt")
nya(content)`)
	if !strings.Contains(code, `import meow_file "github.com/135yshr/meow/runtime/file"`) {
		t.Error("expected meow_file import")
	}
	if !strings.Contains(code, `meow_file.Snoop(meow.NewString("data.txt"))`) {
		t.Error("expected meow_file.Snoop call")
	}
	if !strings.Contains(code, `meow.Nya(content)`) {
		t.Error("expected meow.Nya")
	}
}

func TestPipeToNya(t *testing.T) {
	code := generate(t, `nab "file"
nyan lines = file.stalk("data.txt")
lines |=| lick(paw(line) { "=> " + line }) |=| nya()`)
	if !strings.Contains(code, "meow.Nya(meow.Lick(") {
		t.Error("expected piped nya call")
	}
}

func TestPipeToBareNya(t *testing.T) {
	code := generate(t, `nab "file"
nyan lines = file.stalk("data.txt")
lines |=| lick(paw(line) { "=> " + line }) |=| nya`)
	if !strings.Contains(code, "meow.Nya(meow.Lick(") {
		t.Error("expected piped bare nya call")
	}
}

func TestFetchHTTPAndPounce(t *testing.T) {
	code := generate(t, `nab "http"
nyan res = http.pounce("https://example.com")
nya(res)`)
	if !strings.Contains(code, `import meow_http "github.com/135yshr/meow/runtime/http"`) {
		t.Error("expected meow_http import")
	}
	if !strings.Contains(code, `meow_http.Pounce(meow.NewString("https://example.com"))`) {
		t.Error("expected meow_http.Pounce call")
	}
}

func TestFetchHTTPAndToss(t *testing.T) {
	code := generate(t, `nab "http"
nyan res = http.toss("https://example.com/api", "{}", "application/json")
nya(res)`)
	if !strings.Contains(code, `meow_http.Toss(meow.NewString("https://example.com/api"), meow.NewString("{}"), meow.NewString("application/json"))`) {
		t.Error("expected meow_http.Toss call with 3 args")
	}
}

func TestMapLitGen(t *testing.T) {
	code := generate(t, `nyan opts = {"maxBodyBytes": 2097152}`)
	if !strings.Contains(code, `meow.NewMap(map[string]meow.Value{"maxBodyBytes": meow.NewInt(2097152)})`) {
		t.Errorf("expected Map codegen, got:\n%s", code)
	}
}

func TestEmptyMapLitGen(t *testing.T) {
	code := generate(t, `nyan m = {}`)
	if !strings.Contains(code, `meow.NewMap(map[string]meow.Value{})`) {
		t.Errorf("expected empty Map codegen, got:\n%s", code)
	}
}

func TestMapAsArgGen(t *testing.T) {
	code := generate(t, `nab "http"
nyan res = http.pounce("https://example.com", {"maxBodyBytes": 2097152})`)
	if !strings.Contains(code, `meow_http.Pounce(meow.NewString("https://example.com"), meow.NewMap(map[string]meow.Value{"maxBodyBytes": meow.NewInt(2097152)}))`) {
		t.Errorf("expected Map arg codegen, got:\n%s", code)
	}
}

func generateTest(t *testing.T, input string) string {
	t.Helper()
	l := lexer.New(input, "test_file.nyan")
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	g := codegen.NewTest()
	code, err := g.GenerateTest(prog)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func TestGenerateTestMode(t *testing.T) {
	code := generateTest(t, `meow test_add() {
  nyan result = 1 + 2
  expect(result, 3)
}

meow test_bool() {
  judge(yarn)
}

meow helper() {
  bring 42
}`)
	if !strings.Contains(code, `import meow_testing "github.com/135yshr/meow/runtime/testing"`) {
		t.Error("expected meow_testing import")
	}
	if !strings.Contains(code, `meow_testing.Run(meow.NewString("test_add")`) {
		t.Error("expected Run call for test_add")
	}
	if !strings.Contains(code, `meow_testing.Run(meow.NewString("test_bool")`) {
		t.Error("expected Run call for test_bool")
	}
	if strings.Contains(code, `meow_testing.Run(meow.NewString("helper")`) {
		t.Error("helper should not be auto-run as test")
	}
	if !strings.Contains(code, `meow_testing.Report()`) {
		t.Error("expected Report call")
	}
	if !strings.Contains(code, `meow_testing.Expect(`) {
		t.Error("expected Expect call in generated code")
	}
	if !strings.Contains(code, `meow_testing.Judge(`) {
		t.Error("expected Judge call in generated code")
	}
}

func TestGenerateImplicitReturn(t *testing.T) {
	code := generate(t, `meow greet(who) {
  nya(who)
}`)
	if !strings.Contains(code, "return meow.NewNil()") {
		t.Error("expected implicit nil return when function does not end with bring")
	}
}

func TestGenerateNoImplicitReturnWhenExplicit(t *testing.T) {
	code := generate(t, `meow greet(who) {
  bring "Hello, " + who + "!"
}`)
	if strings.Count(code, "return ") != 1 {
		t.Error("expected only explicit return, no implicit nil return")
	}
}

func TestGenerateTestRefuse(t *testing.T) {
	code := generateTest(t, `meow test_falsy() {
  refuse(hairball)
}`)
	if !strings.Contains(code, `meow_testing.Refuse(`) {
		t.Error("expected Refuse call")
	}
}

func TestIfElse(t *testing.T) {
	code := generate(t, `sniff (x > 0) {
  nya(x)
} scratch {
  nya(0)
}`)
	if !strings.Contains(code, "if (") {
		t.Error("expected if statement")
	}
	if !strings.Contains(code, "} else {") {
		t.Error("expected else clause")
	}
}

func generateTestWithCoverage(t *testing.T, input, filename string) string {
	t.Helper()
	l := lexer.New(input, filename)
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	g := codegen.NewTest()
	g.EnableCoverage(filename)
	code, err := g.GenerateTest(prog)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func TestCoverageImports(t *testing.T) {
	code := generateTestWithCoverage(t, `meow test_add() {
  nyan result = 1 + 2
  expect(result, 3)
}`, "basic_test.nyan")
	if !strings.Contains(code, `import meow_coverage "github.com/135yshr/meow/runtime/coverage"`) {
		t.Errorf("expected meow_coverage import, got:\n%s", code)
	}
	if !strings.Contains(code, `import "os"`) {
		t.Errorf("expected os import for coverage, got:\n%s", code)
	}
}

func TestCoverageHitCalls(t *testing.T) {
	code := generateTestWithCoverage(t, `meow test_add() {
  nyan result = 1 + 2
  expect(result, 3)
}`, "basic_test.nyan")
	if !strings.Contains(code, "meow_coverage.Hit(0)") {
		t.Errorf("expected Hit(0) call, got:\n%s", code)
	}
	if !strings.Contains(code, "meow_coverage.Hit(1)") {
		t.Errorf("expected Hit(1) call, got:\n%s", code)
	}
}

func TestCoverageRegisterInInit(t *testing.T) {
	code := generateTestWithCoverage(t, `meow test_add() {
  nyan result = 1 + 2
  expect(result, 3)
}`, "basic_test.nyan")
	if !strings.Contains(code, "func init()") {
		t.Errorf("expected init function, got:\n%s", code)
	}
	if !strings.Contains(code, `meow_coverage.Register("basic_test.nyan"`) {
		t.Errorf("expected Register call with filename, got:\n%s", code)
	}
}

// A loop is one block, not one character. Recorded as a character the loop's
// coverage entry — and that of any block ending in one — highlighted the word
// `purr` alone instead of the turn it stands for.
func TestCoverageSpansAConditionalPurr(t *testing.T) {
	code := generateTestWithCoverage(t, `meow probe() {
  purr (yarn) {
    nya("looking")
    bolt
  }
}

meow test_probe() {
  probe()
}`, "probe_test.nyan")

	if !strings.Contains(code, `meow_coverage.Register("probe_test.nyan", 2, 3, 5, 1, 1)`) {
		t.Errorf("expected the purr block to span lines 2 to 5, got:\n%s", code)
	}
}

func TestCoverageReportInMain(t *testing.T) {
	code := generateTestWithCoverage(t, `meow test_add() {
  nyan result = 1 + 2
  expect(result, 3)
}`, "basic_test.nyan")
	if !strings.Contains(code, "meow_coverage.Report(os.Stdout)") {
		t.Errorf("expected coverage Report call, got:\n%s", code)
	}
	if !strings.Contains(code, `os.Getenv("MEOW_COVERPROFILE")`) {
		t.Errorf("expected MEOW_COVERPROFILE env check, got:\n%s", code)
	}
	// Report should come before meow_testing.Report
	reportIdx := strings.Index(code, "meow_coverage.Report(os.Stdout)")
	testingReportIdx := strings.Index(code, "meow_testing.Report()")
	if reportIdx > testingReportIdx {
		t.Error("coverage Report should appear before testing Report")
	}
}

func TestCoverageDisabledByDefault(t *testing.T) {
	code := generateTest(t, `meow test_add() {
  nyan result = 1 + 2
  expect(result, 3)
}`)
	if strings.Contains(code, "meow_coverage") {
		t.Error("coverage should not appear when not enabled")
	}
}

func TestFetchWithAliasAndMemberCall(t *testing.T) {
	code := generate(t, `nab "file" tag f
nyan content = f.snoop("data.txt")
nya(content)`)
	if !strings.Contains(code, `import meow_file "github.com/135yshr/meow/runtime/file"`) {
		t.Error("expected meow_file import")
	}
	if !strings.Contains(code, `meow_file.Snoop(meow.NewString("data.txt"))`) {
		t.Error("expected meow_file.Snoop call via alias")
	}
}

func TestFetchWithAliasMemberExpr(t *testing.T) {
	code := generate(t, `nab "http" tag h
nyan res = h.pounce("https://example.com")
nya(res)`)
	if !strings.Contains(code, `import meow_http "github.com/135yshr/meow/runtime/http"`) {
		t.Error("expected meow_http import")
	}
	if !strings.Contains(code, `meow_http.Pounce(meow.NewString("https://example.com"))`) {
		t.Error("expected meow_http.Pounce call via alias 'h'")
	}
}

// generateTyped generates code the way the compiler does, with the checker's
// type information in place, so that a fully typed function takes the typed
// path rather than the boxed one.
func generateTyped(t *testing.T, input string) string {
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
	code, err := g.Generate(prog)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

// A statement inside a fully typed function discards its value, so a Furball
// it answers with has to be raised where it is written — the typed function
// returns a native Go type and cannot pass one on. This holds for a statement
// that is not a call too: a pipe, an index, a piece of arithmetic.
func TestTypedStatementValueIsChecked(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{
			name: "pipe",
			stmt: `["oops"] |=| boom`,
			want: "meow.Propagate(",
		},
		{
			name: "index",
			stmt: `[1, 2][9]`,
			want: "meow.Propagate(",
		},
		{
			name: "arithmetic",
			stmt: `n / 0`,
			want: "meow.Propagate(meow.Div(",
		},
		{
			name: "call answering with a boxed value",
			stmt: `made_a_list(n)`,
			want: "meow.Propagate(made_a_list(",
		},
		{
			name: "identifier alone",
			stmt: `n`,
			want: "meow.Propagate(",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateTyped(t, `meow boom(xs litter) string {
  bring xs[0] + 1
}

meow made_a_list(n int) litter {
  bring [n]
}

meow caller(n int) int {
  `+tt.stmt+`
  bring n
}

nya(caller(5))`)
			if !strings.Contains(code, tt.want) {
				t.Errorf("expected %q in generated code, got:\n%s", tt.want, code)
			}
		})
	}
}

// A statement in a loop body is a statement like any other: the check goes
// where the statement is, not only at the top level of the function.
func TestTypedStatementInLoopBodyIsChecked(t *testing.T) {
	code := generateTyped(t, `meow boom(xs litter) string {
  bring xs[0] + 1
}

meow caller(n int) int {
  purr i (0..2) {
    ["oops"] |=| boom
  }
  bring n
}

nya(caller(5))`)
	if !strings.Contains(code, "meow.Propagate(") {
		t.Errorf("expected the loop body statement to be checked, got:\n%s", code)
	}
}

// Where the checker knows the result is a native Go type there is nothing to
// check: the unboxing raises a Furball on its way out, so wrapping it again
// would say the same thing twice.
func TestTypedNativeCallStatementIsNotWrapped(t *testing.T) {
	code := generateTyped(t, `meow doubled(xs litter) string {
  bring xs[0] + 1
}

meow caller(n int) int {
  doubled(["oops"])
  bring n
}

nya(caller(5))`)
	if !strings.Contains(code, "meow.AsString(doubled(") {
		t.Errorf("expected the native result to be unboxed, got:\n%s", code)
	}
	if strings.Contains(code, "meow.Propagate(meow.AsString(") {
		t.Error("an unboxed native result should not be wrapped again")
	}
}

// hiss raises rather than answering, so in a typed function it is emitted as a
// Go panic — a statement with no value to pass through anything.
func TestTypedHissStatementStaysAPanic(t *testing.T) {
	code := generateTyped(t, `meow caller(n int) int {
  hiss("bad")
  bring n
}

nya(caller(5))`)
	if !strings.Contains(code, "panic(meow.Hiss(") {
		t.Errorf("expected hiss to panic, got:\n%s", code)
	}
	if strings.Contains(code, "meow.Propagate(panic(") {
		t.Error("hiss must not be wrapped: a panic is not an expression")
	}
}

// A builtin named rather than called is the builtin itself. A runtime builtin
// is a plain Go function of a fixed arity, which nothing taking a meow.Value
// can be handed, so it is wrapped into the value it has to be.
func TestABuiltinNamedRatherThanCalledIsWrapped(t *testing.T) {
	code := generateTyped(t, `nyan f = upper
nya(f("hi"))
`)
	if !strings.Contains(code, `meow.BuiltinFunc("upper", 1,`) {
		t.Errorf("expected upper to be wrapped as a value, got:\n%s", code)
	}
}

// Calling one is unchanged: the wrapper is for the value position only, and a
// call still goes straight to the runtime function.
func TestACalledBuiltinIsNotWrapped(t *testing.T) {
	code := generateTyped(t, `nya(upper("hi"))`)
	if !strings.Contains(code, `meow.Upper(`) {
		t.Errorf("expected a direct call, got:\n%s", code)
	}
	if strings.Contains(code, `meow.BuiltinFunc(`) {
		t.Errorf("expected no wrapper for a call, got:\n%s", code)
	}
}

// A name a binding took over is the binding's, as it already was.
func TestABindingKeepsItsNameFromABuiltinInCodegen(t *testing.T) {
	code := generateTyped(t, `meow shadowed() litter {
  nyan upper = paw(s) { bring s + "!" }
  bring lick(["a"], upper)
}
nya(shadowed())
`)
	if strings.Contains(code, `meow.BuiltinFunc("upper"`) {
		t.Errorf("expected the binding to keep the name, got:\n%s", code)
	}
}

// The two backends drifting over which names exist is what #140 and #141 were,
// and the guard added with them covers the checker and the interpreter. Naming
// a builtin as a value needs codegen to know its arity too, so a name added to
// the checker without one here would compile to nothing callable.
func TestEveryAcceptedBuiltinCanBeNamedAsAValue(t *testing.T) {
	for _, name := range checker.BuiltinNames() {
		if !codegen.HasBuiltinValue(name) {
			t.Errorf("the checker accepts %q but codegen cannot name it as a value", name)
		}
	}
	// And the other way: an entry left behind after the checker stops
	// accepting a name would be arity bookkeeping for a builtin nothing can
	// reach, and would go on claiming a runtime function that may be gone.
	accepted := make(map[string]bool)
	for _, name := range checker.BuiltinNames() {
		accepted[name] = true
	}
	for _, name := range codegen.BuiltinValueNames() {
		if !accepted[name] {
			t.Errorf("codegen can name %q as a value but the checker does not accept it", name)
		}
	}
}
