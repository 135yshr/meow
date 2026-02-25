package linter

import (
	"testing"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

func parse(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input, "test.nyan")
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Log(e)
		}
		t.Fatalf("parse errors")
	}
	return prog
}

func lint(t *testing.T, input string) []Diagnostic {
	t.Helper()
	prog := parse(t, input)
	l := New()
	return l.Lint(prog)
}

func findByRule(diags []Diagnostic, rule string) []Diagnostic {
	var result []Diagnostic
	for _, d := range diags {
		if d.Rule == rule {
			result = append(result, d)
		}
	}
	return result
}

// --- snake_case rule ---

func TestSnakeCaseRule_BadVar(t *testing.T) {
	diags := lint(t, `nyan myVar = 1`)
	found := findByRule(diags, "snake-case")
	if len(found) == 0 {
		t.Fatal("expected snake-case warning for myVar")
	}
}

func TestSnakeCaseRule_GoodVar(t *testing.T) {
	diags := lint(t, `nyan my_var = 1`)
	found := findByRule(diags, "snake-case")
	if len(found) != 0 {
		t.Fatalf("unexpected snake-case warning: %v", found)
	}
}

func TestSnakeCaseRule_BadFunc(t *testing.T) {
	diags := lint(t, `meow myFunc() { }`)
	found := findByRule(diags, "snake-case")
	if len(found) == 0 {
		t.Fatal("expected snake-case warning for myFunc")
	}
}

func TestSnakeCaseRule_GoodFunc(t *testing.T) {
	diags := lint(t, `meow my_func() {
  nya(1)
}`)
	found := findByRule(diags, "snake-case")
	if len(found) != 0 {
		t.Fatalf("unexpected snake-case warning: %v", found)
	}
}

func TestSnakeCaseRule_Underscore(t *testing.T) {
	diags := lint(t, `nyan _ = 1`)
	found := findByRule(diags, "snake-case")
	if len(found) != 0 {
		t.Fatalf("unexpected snake-case warning for _: %v", found)
	}
}

// --- unused-var rule ---

func TestUnusedVarRule_Unused(t *testing.T) {
	diags := lint(t, `nyan x = 1`)
	found := findByRule(diags, "unused-var")
	if len(found) == 0 {
		t.Fatal("expected unused-var warning for x")
	}
}

func TestUnusedVarRule_Used(t *testing.T) {
	diags := lint(t, `nyan x = 1
nya(x)`)
	found := findByRule(diags, "unused-var")
	if len(found) != 0 {
		t.Fatalf("unexpected unused-var warning: %v", found)
	}
}

func TestUnusedVarRule_UnderscoreIgnored(t *testing.T) {
	diags := lint(t, `nyan _ = 1`)
	found := findByRule(diags, "unused-var")
	if len(found) != 0 {
		t.Fatalf("unexpected unused-var warning for _: %v", found)
	}
}

func TestUnusedVarRule_FuncParamIgnored(t *testing.T) {
	diags := lint(t, `meow f(x int) {
  nya(1)
}`)
	found := findByRule(diags, "unused-var")
	if len(found) != 0 {
		t.Fatalf("unexpected unused-var warning for function param: %v", found)
	}
}

// --- unreachable-code rule ---

func TestUnreachableCodeRule_AfterBring(t *testing.T) {
	diags := lint(t, `meow f() {
  bring 1
  nya(2)
}`)
	found := findByRule(diags, "unreachable-code")
	if len(found) == 0 {
		t.Fatal("expected unreachable-code warning after bring")
	}
}

func TestUnreachableCodeRule_NoBring(t *testing.T) {
	diags := lint(t, `meow f() {
  nya(1)
  nya(2)
}`)
	found := findByRule(diags, "unreachable-code")
	if len(found) != 0 {
		t.Fatalf("unexpected unreachable-code warning: %v", found)
	}
}

func TestUnreachableCodeRule_BringAtEnd(t *testing.T) {
	diags := lint(t, `meow f() {
  nya(1)
  bring 2
}`)
	found := findByRule(diags, "unreachable-code")
	if len(found) != 0 {
		t.Fatalf("unexpected unreachable-code warning: %v", found)
	}
}

// --- empty-block rule ---

func TestEmptyBlockRule_EmptyFunc(t *testing.T) {
	diags := lint(t, `meow f() { }`)
	found := findByRule(diags, "empty-block")
	if len(found) == 0 {
		t.Fatal("expected empty-block warning for empty function")
	}
}

func TestEmptyBlockRule_NonEmptyFunc(t *testing.T) {
	diags := lint(t, `meow f() {
  nya(1)
}`)
	found := findByRule(diags, "empty-block")
	if len(found) != 0 {
		t.Fatalf("unexpected empty-block warning: %v", found)
	}
}

func TestEmptyBlockRule_EmptyElseOk(t *testing.T) {
	diags := lint(t, `sniff (yarn) {
  nya(1)
}`)
	found := findByRule(diags, "empty-block")
	if len(found) != 0 {
		t.Fatalf("unexpected empty-block warning for missing else: %v", found)
	}
}

// --- integration ---

func TestLintCleanCode(t *testing.T) {
	diags := lint(t, `meow fib(n int) int {
  sniff (n <= 1) {
    bring n
  }
  bring fib(n - 1) + fib(n - 2)
}

nyan i = 0
purr (i < 10) {
  nya(fib(i))
  i = i + 1
}`)
	if len(diags) != 0 {
		for _, d := range diags {
			t.Log(d)
		}
		t.Fatalf("expected no diagnostics for clean code, got %d", len(diags))
	}
}
