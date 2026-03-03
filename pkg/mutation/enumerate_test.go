package mutation_test

import (
	"testing"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/mutation"
	"github.com/135yshr/meow/pkg/parser"
)

func parse(t *testing.T, input string) *ast.Program {
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
	return prog
}

func TestEnumerateArithmetic(t *testing.T) {
	prog := parse(t, `nyan x = 1 + 2`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.ArithmeticSwap {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ArithmeticSwap mutation for +")
	}
}

func TestEnumerateComparison(t *testing.T) {
	prog := parse(t, `nyan x = 1 == 2`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.ComparisonSwap {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ComparisonSwap mutation for ==")
	}
}

func TestEnumerateLogical(t *testing.T) {
	prog := parse(t, `nyan x = yarn && hairball`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.LogicalSwap {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected LogicalSwap mutation for &&")
	}
}

func TestEnumerateBoolFlip(t *testing.T) {
	prog := parse(t, `nyan x = yarn`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.BoolFlip {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected BoolFlip mutation for yarn")
	}
}

func TestEnumerateIntBoundary(t *testing.T) {
	prog := parse(t, `nyan x = 42`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.IntBoundary {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected IntBoundary mutation for 42")
	}
}

func TestEnumerateIntZero(t *testing.T) {
	prog := parse(t, `nyan x = 0`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.IntBoundary {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected IntBoundary mutation for 0")
	}
}

func TestEnumerateStringEmpty(t *testing.T) {
	prog := parse(t, `nyan x = "hello"`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.StringEmpty {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected StringEmpty mutation for non-empty string")
	}
}

func TestEnumerateConditionNegate(t *testing.T) {
	prog := parse(t, `sniff (x > 0) { nya(x) }`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.ConditionNegate {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ConditionNegate mutation for if statement")
	}
}

func TestEnumerateReturnNil(t *testing.T) {
	prog := parse(t, `meow f() { bring 42 }`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.ReturnNil {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ReturnNil mutation for return statement")
	}
}

func TestApplyUndo(t *testing.T) {
	prog := parse(t, `nyan x = 1 + 2`)
	mutants := mutation.Enumerate(prog)

	for _, m := range mutants {
		if m.Kind == mutation.ArithmeticSwap {
			m.Apply()
			m.Undo()
			break
		}
	}

	// Verify the AST is back to normal by re-enumerating
	mutants2 := mutation.Enumerate(prog)
	if len(mutants) != len(mutants2) {
		t.Errorf("expected same number of mutants after undo, got %d vs %d", len(mutants), len(mutants2))
	}
}

func TestApplyUndoBoolFlip(t *testing.T) {
	prog := parse(t, `nyan x = yarn`)
	mutants := mutation.Enumerate(prog)

	for _, m := range mutants {
		if m.Kind == mutation.BoolFlip {
			// Get the bool literal
			varStmt := prog.Stmts[0].(*ast.VarStmt)
			boolLit := varStmt.Value.(*ast.BoolLit)
			if !boolLit.Value {
				t.Error("expected true before apply")
			}
			m.Apply()
			if boolLit.Value {
				t.Error("expected false after apply")
			}
			m.Undo()
			if !boolLit.Value {
				t.Error("expected true after undo")
			}
			break
		}
	}
}

func TestEnumerateNegation(t *testing.T) {
	prog := parse(t, `nyan x = -5`)
	mutants := mutation.Enumerate(prog)

	found := false
	for _, m := range mutants {
		if m.Kind == mutation.NegationRemoval {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected NegationRemoval mutation for -5")
	}
}

func TestEnumerateMultipleMutations(t *testing.T) {
	prog := parse(t, `meow f(a, b) {
  sniff (a > 0 && b != 0) {
    bring a + b
  }
  bring 0
}`)
	mutants := mutation.Enumerate(prog)
	if len(mutants) == 0 {
		t.Fatal("expected mutations")
	}

	kinds := make(map[mutation.MutantKind]int)
	for _, m := range mutants {
		kinds[m.Kind]++
	}

	if kinds[mutation.ArithmeticSwap] == 0 {
		t.Error("expected ArithmeticSwap mutation")
	}
	if kinds[mutation.ComparisonSwap] == 0 {
		t.Error("expected ComparisonSwap mutation")
	}
	if kinds[mutation.LogicalSwap] == 0 {
		t.Error("expected LogicalSwap mutation")
	}
	if kinds[mutation.ConditionNegate] == 0 {
		t.Error("expected ConditionNegate mutation")
	}
	if kinds[mutation.ReturnNil] == 0 {
		t.Error("expected ReturnNil mutation")
	}
}
