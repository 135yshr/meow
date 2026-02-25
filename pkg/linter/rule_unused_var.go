package linter

import (
	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
)

// UnusedVarRule detects variables that are declared but never referenced.
type UnusedVarRule struct{}

func (r *UnusedVarRule) Name() string { return "unused-var" }

type varEntry struct {
	pos  token.Position
	name string
	used bool
}

type unusedScope struct {
	vars map[string]*varEntry
}

func (r *UnusedVarRule) Check(prog *ast.Program, report func(Diagnostic)) {
	v := &unusedChecker{rule: r.Name(), report: report}
	v.pushScope()
	// First pass: register top-level function names so calls count as references
	for _, stmt := range prog.Stmts {
		if fn, ok := stmt.(*ast.FuncStmt); ok {
			v.define(fn.Name, fn.Token.Pos)
			v.markUsed(fn.Name) // function definitions are always "used"
		}
	}
	for _, stmt := range prog.Stmts {
		v.checkStmt(stmt)
	}
	v.reportUnused()
	v.popScope()
}

type unusedChecker struct {
	rule   string
	report func(Diagnostic)
	scopes []unusedScope
}

func (c *unusedChecker) pushScope() {
	c.scopes = append(c.scopes, unusedScope{vars: make(map[string]*varEntry)})
}

func (c *unusedChecker) popScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *unusedChecker) define(name string, pos token.Position) {
	scope := c.scopes[len(c.scopes)-1]
	if prev, ok := scope.vars[name]; ok && !prev.used && prev.name != "_" {
		c.report(Diagnostic{
			Pos:      prev.pos,
			Severity: Warning,
			Rule:     c.rule,
			Message:  "variable \"" + prev.name + "\" is declared but never used",
		})
	}
	scope.vars[name] = &varEntry{pos: pos, name: name}
}

func (c *unusedChecker) markUsed(name string) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if v, ok := c.scopes[i].vars[name]; ok {
			v.used = true
			return
		}
	}
}

func (c *unusedChecker) reportUnused() {
	scope := c.scopes[len(c.scopes)-1]
	for _, v := range scope.vars {
		if !v.used && v.name != "_" {
			c.report(Diagnostic{
				Pos:      v.pos,
				Severity: Warning,
				Rule:     c.rule,
				Message:  "variable \"" + v.name + "\" is declared but never used",
			})
		}
	}
}

func (c *unusedChecker) checkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.VarStmt:
		c.checkExpr(s.Value)
		c.define(s.Name, s.Token.Pos)
	case *ast.FuncStmt:
		c.pushScope()
		// Parameters are not checked for unused (caller provides them)
		for _, p := range s.Params {
			c.define(p.Name, s.Token.Pos)
			c.markUsed(p.Name)
		}
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		c.reportUnused()
		c.popScope()
	case *ast.IfStmt:
		c.checkExpr(s.Condition)
		c.pushScope()
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		c.reportUnused()
		c.popScope()
		if len(s.ElseBody) > 0 {
			c.pushScope()
			for _, stmt := range s.ElseBody {
				c.checkStmt(stmt)
			}
			c.reportUnused()
			c.popScope()
		}
	case *ast.RangeStmt:
		c.checkExpr(s.Start)
		c.checkExpr(s.End)
		c.pushScope()
		c.define(s.Var, s.Token.Pos)
		c.markUsed(s.Var)
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		c.reportUnused()
		c.popScope()
	case *ast.ReturnStmt:
		if s.Value != nil {
			c.checkExpr(s.Value)
		}
	case *ast.ExprStmt:
		c.checkExpr(s.Expr)
	case *ast.FetchStmt:
		// no-op
	}
}

func (c *unusedChecker) checkExpr(expr ast.Expr) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Ident:
		c.markUsed(e.Name)
	case *ast.UnaryExpr:
		c.checkExpr(e.Right)
	case *ast.BinaryExpr:
		c.checkExpr(e.Left)
		c.checkExpr(e.Right)
	case *ast.CallExpr:
		c.checkExpr(e.Fn)
		for _, arg := range e.Args {
			c.checkExpr(arg)
		}
	case *ast.LambdaExpr:
		c.pushScope()
		for _, p := range e.Params {
			c.define(p.Name, e.Token.Pos)
			c.markUsed(p.Name)
		}
		c.checkExpr(e.Body)
		c.reportUnused()
		c.popScope()
	case *ast.ListLit:
		for _, item := range e.Items {
			c.checkExpr(item)
		}
	case *ast.IndexExpr:
		c.checkExpr(e.Left)
		c.checkExpr(e.Index)
	case *ast.PipeExpr:
		c.checkExpr(e.Left)
		c.checkExpr(e.Right)
	case *ast.CatchExpr:
		c.checkExpr(e.Left)
		c.checkExpr(e.Right)
	case *ast.MatchExpr:
		c.checkExpr(e.Subject)
		for _, arm := range e.Arms {
			c.checkExpr(arm.Body)
		}
	case *ast.MemberExpr:
		c.checkExpr(e.Object)
	case *ast.MapLit:
		for _, k := range e.Keys {
			c.checkExpr(k)
		}
		for _, v := range e.Vals {
			c.checkExpr(v)
		}
	}
}
