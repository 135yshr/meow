package linter

import (
	"github.com/135yshr/meow/pkg/ast"
)

// EmptyBlockRule detects empty function, if, and while bodies.
type EmptyBlockRule struct{}

func (r *EmptyBlockRule) Name() string { return "empty-block" }

func (r *EmptyBlockRule) Check(prog *ast.Program, report func(Diagnostic)) {
	for node := range ast.Preorder(prog) {
		switch n := node.(type) {
		case *ast.FuncStmt:
			if len(n.Body) == 0 {
				report(Diagnostic{
					Pos:      n.Token.Pos,
					Severity: Warning,
					Rule:     r.Name(),
					Message:  "function \"" + n.Name + "\" has an empty body",
				})
			}
		case *ast.IfStmt:
			if len(n.Body) == 0 {
				report(Diagnostic{
					Pos:      n.Token.Pos,
					Severity: Warning,
					Rule:     r.Name(),
					Message:  "sniff block has an empty body",
				})
			}
			// ElseBody being empty is normal (else omitted)
		case *ast.RangeStmt:
			if len(n.Body) == 0 {
				report(Diagnostic{
					Pos:      n.Token.Pos,
					Severity: Warning,
					Rule:     r.Name(),
					Message:  "purr loop has an empty body",
				})
			}
		}
	}
}
