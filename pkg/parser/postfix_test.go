package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/ast"
)

// shape writes an expression out as the nesting of its postfix forms, so that a
// test can say what a chain parsed to in one line rather than a ladder of type
// assertions.
func shape(e ast.Expr) string {
	switch n := e.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.IntLit:
		return fmt.Sprint(n.Value)
	case *ast.StringLit:
		return fmt.Sprintf("%q", n.Value)
	case *ast.SelfExpr:
		return "self"
	case *ast.MemberExpr:
		return fmt.Sprintf("Member(%s, %s)", shape(n.Object), n.Member)
	case *ast.CallExpr:
		args := make([]string, len(n.Args))
		for i, a := range n.Args {
			args[i] = shape(a)
		}
		return fmt.Sprintf("Call(%s; %s)", shape(n.Fn), strings.Join(args, ", "))
	case *ast.IndexExpr:
		return fmt.Sprintf("Index(%s, %s)", shape(n.Left), shape(n.Index))
	case *ast.PipeExpr:
		return fmt.Sprintf("Pipe(%s, %s)", shape(n.Left), shape(n.Right))
	case *ast.BinaryExpr:
		return fmt.Sprintf("Binary(%s %s %s)", shape(n.Left), n.Token.Literal, shape(n.Right))
	case *ast.LambdaExpr:
		return "Lambda"
	default:
		return fmt.Sprintf("%T", e)
	}
}

func valueOf(t *testing.T, src string) ast.Expr {
	t.Helper()
	prog := parse(t, "nyan v = "+src)
	return prog.Stmts[0].(*ast.VarStmt).Value
}

// A member read, a call and a subscript can follow any operand, in any order and
// any number of times. They used to be three different things: a subscript was
// a postfix applied by parsePostfix, while a member and a call were read only
// straight after a name — so `(c).name`, `(f)(4)`, `cats[0].name` and
// `make().name` were all refused, and a chain could not get past its first
// non-name link (#153).
func TestAPostfixFollowsAnyOperand(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		// The forms #153 names.
		{"(c).name", "Member(c, name)"},
		{"(f)(4)", "Call(f; 4)"},
		{"(5 |=| Point).x", "Member(Pipe(5, Point), x)"},
		// A chain that starts somewhere other than a name.
		{"cats[0].name", "Member(Index(cats, 0), name)"},
		{"cats[0].shout()", "Call(Member(Index(cats, 0), shout); )"},
		{"make().name", "Member(Call(make; ), name)"},
		{"make().shout()", "Call(Member(Call(make; ), shout); )"},
		{"f(1)(2)", "Call(Call(f; 1); 2)"},
		{"a.b.c", "Member(Member(a, b), c)"},
		{"a.b().c()", "Call(Member(Call(Member(a, b); ), c); )"},
		{"a.b[0].c", "Member(Index(Member(a, b), 0), c)"},
		{"handlers[0](1)", "Call(Index(handlers, 0); 1)"},
		{`data["k"].v`, `Member(Index(data, "k"), v)`},
		{"self.pos.x", "Member(Member(self, pos), x)"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			if got := shape(valueOf(t, tt.src)); got != tt.want {
				t.Errorf("parsed as %s, want %s", got, tt.want)
			}
		})
	}
}

// Every chain that parsed before parses to the same tree. The backends key a
// great deal off these shapes — a builtin is dispatched by the name a call is
// made through, a nab'd package's member is recognized by its object being that
// package's name, a groom method by the object's type — so the refactor is only
// safe if none of them moves.
func TestAPostfixThatParsedBeforeParsesTheSame(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"f(1, 2)", "Call(f; 1, 2)"},
		{"upper(s)", "Call(upper; s)"},
		{"file.snoop(p)", "Call(Member(file, snoop); p)"},
		{"c.name", "Member(c, name)"},
		{"c.shout()", "Call(Member(c, shout); )"},
		{"self.name", "Member(self, name)"},
		{"grid[1][0]", "Index(Index(grid, 1), 0)"},
		{"f()[0]", "Index(Call(f; ), 0)"},
		{"[1, 2][0]", "Index(*ast.ListLit, 0)"},
		{"(xs)[0]", "Index(xs, 0)"},
		{"x.string", "Member(x, string)"},
		{"xs |=| lick(f)", "Pipe(xs, Call(lick; f))"},
		{"a + b.c", "Binary(a + Member(b, c))"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			if got := shape(valueOf(t, tt.src)); got != tt.want {
				t.Errorf("parsed as %s, want %s", got, tt.want)
			}
		})
	}
}

// A postfix binds tighter than any operator, so it attaches to the operand next
// to it and never to a whole binary expression written without parentheses.
func TestAPostfixBindsTighterThanAnOperator(t *testing.T) {
	if got := shape(valueOf(t, "a - b.c")); got != "Binary(a - Member(b, c))" {
		t.Errorf("parsed as %s", got)
	}
	if got := shape(valueOf(t, "-f(1)")); got != "*ast.UnaryExpr" {
		t.Errorf("parsed as %s", got)
	}
	u := valueOf(t, "-f(1)").(*ast.UnaryExpr)
	if got := shape(u.Right); got != "Call(f; 1)" {
		t.Errorf("negated %s, want the call", got)
	}
}

// A newline still ends the expression, so a line that opens with `(`, `.` or
// `[` is not taken as continuing the line above. `(` is the one that matters:
// read as a call, the line below would swallow a parenthesised expression.
func TestAPostfixDoesNotReachAcrossANewline(t *testing.T) {
	prog := parse(t, "nyan f = g\n(1 + 2)")
	if len(prog.Stmts) != 2 {
		t.Fatalf("expected two statements, got %d", len(prog.Stmts))
	}
	if got := shape(prog.Stmts[0].(*ast.VarStmt).Value); got != "g" {
		t.Errorf("the binding is %s, want g alone", got)
	}
}

// `purr i (3)` names its loop variable and then its range: the `(` is the
// statement's own, not a call of `i`.
func TestAPurrRangeIsNotReadAsACall(t *testing.T) {
	prog := parse(t, "purr i (3) {\n  nya(i)\n}")
	r, ok := prog.Stmts[0].(*ast.RangeStmt)
	if !ok {
		t.Fatalf("expected a RangeStmt, got %T", prog.Stmts[0])
	}
	if r.Var != "i" {
		t.Errorf("loop variable is %q, want i", r.Var)
	}
}

// A lambda can be called where it is written, since it is an operand like any
// other.
func TestALambdaCanBeCalledWhereItIsWritten(t *testing.T) {
	if got := shape(valueOf(t, "paw(x) { x * 2 }(5)")); got != "Call(Lambda; 5)" {
		t.Errorf("parsed as %s", got)
	}
}

// A keyword is still a member name after any operand, as it is after a name.
func TestAKeywordMemberFollowsAnyOperand(t *testing.T) {
	if got := shape(valueOf(t, "xs[0].string")); got != "Member(Index(xs, 0), string)" {
		t.Errorf("parsed as %s", got)
	}
}

// A match pattern could always be written with a member or a call after a name,
// because the name's own parser read them; the postfix forms are read for a
// pattern as they are anywhere else, at either end of a range.
func TestAPatternKeepsItsPostfixForms(t *testing.T) {
	prog := parse(t, "nyan v = peek(x) {\n  limits.low..limits.high => 1,\n  bound(2) => 2,\n  _ => 3\n}")
	m := prog.Stmts[0].(*ast.VarStmt).Value.(*ast.MatchExpr)
	r, ok := m.Arms[0].Pattern.(*ast.RangePattern)
	if !ok {
		t.Fatalf("first arm is %T, want a range", m.Arms[0].Pattern)
	}
	if got := shape(r.Low) + " .. " + shape(r.High); got != "Member(limits, low) .. Member(limits, high)" {
		t.Errorf("range is %s", got)
	}
	lit, ok := m.Arms[1].Pattern.(*ast.LiteralPattern)
	if !ok {
		t.Fatalf("second arm is %T, want a literal", m.Arms[1].Pattern)
	}
	if got := shape(lit.Value); got != "Call(bound; 2)" {
		t.Errorf("pattern is %s", got)
	}
}
