package codegen_test

import (
	"strings"
	"testing"

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
	code := generate(t, `nyan name = "Tama"
nya(name)`)
	if !strings.Contains(code, `meow.NewString("Tama")`) {
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
nyan name = "Tama"
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
	code := generate(t, `fetch "file"
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
	code := generate(t, `fetch "file"
nyan lines = file.stalk("data.txt")
lines |=| lick(paw(line) { "=> " + line }) |=| nya()`)
	if !strings.Contains(code, "meow.Nya(meow.Lick(") {
		t.Error("expected piped nya call")
	}
}

func TestPipeToBareNya(t *testing.T) {
	code := generate(t, `fetch "file"
nyan lines = file.stalk("data.txt")
lines |=| lick(paw(line) { "=> " + line }) |=| nya`)
	if !strings.Contains(code, "meow.Nya(meow.Lick(") {
		t.Error("expected piped bare nya call")
	}
}

func TestFetchHTTPAndPounce(t *testing.T) {
	code := generate(t, `fetch "http"
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
	code := generate(t, `fetch "http"
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
	code := generate(t, `fetch "http"
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
