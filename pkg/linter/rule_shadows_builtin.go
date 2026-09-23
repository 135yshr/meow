package linter

import (
	"fmt"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/builtins"
	"github.com/135yshr/meow/pkg/token"
)

// ShadowsBuiltinRule reports a name bound over a builtin, or over a kitty or
// collar constructor the program declares.
//
// A binding takes the name for as long as it is in scope, wherever the name is
// written — read, called, or piped into. That is the ordinary rule and it is
// what the spec says, but it means an accident is possible that used not to be:
// `nyan len = len(xs)` leaves the rest of that scope with no `len` to call, and
// a binding holding a function of the same shape changes what a call does with
// no error anywhere. Nothing else can warn about that — the program is
// correct — so it is said here, at the binding, where the reader can see both
// names at once.
//
// The names at risk are the ordinary English ones. `nya`, `hiss`, `lick`,
// `picky` and `curl` are keywords, so the lexer refuses to bind them and this
// rule never sees them.
type ShadowsBuiltinRule struct{}

func (r *ShadowsBuiltinRule) Name() string { return "shadows-builtin" }

func (r *ShadowsBuiltinRule) Check(prog *ast.Program, report func(Diagnostic)) {
	s := &shadowChecker{rule: r.Name(), report: report, declared: map[string]string{}}
	// A constructor's name is declared by its kitty or collar statement, which
	// may be written after the binding that takes it, so they are collected
	// before anything is reported.
	for _, stmt := range prog.Stmts {
		switch d := stmt.(type) {
		case *ast.KittyStmt:
			s.declared[d.Name] = "kitty"
		case *ast.CollarStmt:
			s.declared[d.Name] = "collar"
		}
	}
	for _, stmt := range prog.Stmts {
		s.checkStmt(stmt)
	}
}

type shadowChecker struct {
	rule     string
	report   func(Diagnostic)
	declared map[string]string // constructor name → "kitty" or "collar"
}

// note reports a binding of `name` when the name belongs to something the
// program could otherwise have reached.
func (s *shadowChecker) note(name string, pos token.Position, what string) {
	switch {
	case builtins.Known(name):
		s.report(Diagnostic{
			Pos:      pos,
			Severity: Warning,
			Rule:     s.rule,
			Message: fmt.Sprintf("%s %q shadows the builtin %s; calls to %s in this scope reach the %s",
				what, name, name, name, what),
		})
	case s.declared[name] != "":
		s.report(Diagnostic{
			Pos:      pos,
			Severity: Warning,
			Rule:     s.rule,
			Message: fmt.Sprintf("%s %q shadows the %s %s; calls to %s in this scope reach the %s",
				what, name, s.declared[name], name, name, what),
		})
	}
}

func (s *shadowChecker) checkParams(params []ast.Param, pos token.Position) {
	for _, p := range params {
		s.note(p.Name, pos, "parameter")
	}
}

func (s *shadowChecker) checkStmt(stmt ast.Stmt) {
	switch n := stmt.(type) {
	case *ast.VarStmt:
		s.note(n.Name, n.Token.Pos, "binding")
		s.checkExpr(n.Value)
	case *ast.FuncStmt:
		s.note(n.Name, n.Token.Pos, "function")
		s.checkParams(n.Params, n.Token.Pos)
		s.checkBlock(n.Body)
	case *ast.RangeStmt:
		s.note(n.Var, n.Token.Pos, "loop variable")
		if n.IndexVar != "" {
			s.note(n.IndexVar, n.Token.Pos, "loop variable")
		}
		s.checkExpr(n.Start)
		s.checkExpr(n.End)
		s.checkBlock(n.Body)
	case *ast.WhileStmt:
		s.checkExpr(n.Cond)
		s.checkBlock(n.Body)
	case *ast.IfStmt:
		s.checkExpr(n.Condition)
		s.checkBlock(n.Body)
		s.checkBlock(n.ElseBody)
	case *ast.ReturnStmt:
		s.checkExpr(n.Value)
	case *ast.ExprStmt:
		s.checkExpr(n.Expr)
	case *ast.LearnStmt:
		// A groomed method is a function of its own: its name belongs to the
		// type rather than to the scope, but its parameters are bindings like
		// any others.
		for i := range n.Methods {
			s.checkParams(n.Methods[i].Params, n.Methods[i].Token.Pos)
			s.checkBlock(n.Methods[i].Body)
		}
	}
}

func (s *shadowChecker) checkBlock(stmts []ast.Stmt) {
	for _, stmt := range stmts {
		s.checkStmt(stmt)
	}
}

func (s *shadowChecker) checkExpr(expr ast.Expr) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.LambdaExpr:
		s.checkParams(e.Params, e.Token.Pos)
		s.checkExpr(e.Body)
		s.checkBlock(e.Block)
	case *ast.UnaryExpr:
		s.checkExpr(e.Right)
	case *ast.BinaryExpr:
		s.checkExpr(e.Left)
		s.checkExpr(e.Right)
	case *ast.CallExpr:
		s.checkExpr(e.Fn)
		for _, arg := range e.Args {
			s.checkExpr(arg)
		}
	case *ast.ListLit:
		for _, item := range e.Items {
			s.checkExpr(item)
		}
	case *ast.IndexExpr:
		s.checkExpr(e.Left)
		s.checkExpr(e.Index)
	case *ast.PipeExpr:
		s.checkExpr(e.Left)
		s.checkExpr(e.Right)
	case *ast.CatchExpr:
		s.checkExpr(e.Left)
		s.checkExpr(e.Right)
	case *ast.MatchExpr:
		s.checkExpr(e.Subject)
		for _, arm := range e.Arms {
			s.checkExpr(arm.Body)
		}
	case *ast.MemberExpr:
		s.checkExpr(e.Object)
	case *ast.MapLit:
		for _, v := range e.Vals {
			s.checkExpr(v)
		}
	}
}
