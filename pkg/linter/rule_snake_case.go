package linter

import (
	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
)

// SnakeCaseRule checks that identifiers use snake_case.
type SnakeCaseRule struct{}

func (r *SnakeCaseRule) Name() string { return "snake-case" }

func (r *SnakeCaseRule) Check(prog *ast.Program, report func(Diagnostic)) {
	for node := range ast.Preorder(prog) {
		switch n := node.(type) {
		case *ast.VarStmt:
			if !isSnakeCase(n.Name) {
				report(Diagnostic{
					Pos:      n.Token.Pos,
					Severity: Warning,
					Rule:     r.Name(),
					Message:  "variable name \"" + n.Name + "\" should be snake_case",
				})
			}
		case *ast.FuncStmt:
			if !isSnakeCase(n.Name) {
				report(Diagnostic{
					Pos:      n.Token.Pos,
					Severity: Warning,
					Rule:     r.Name(),
					Message:  "function name \"" + n.Name + "\" should be snake_case",
				})
			}
			checkParams(n.Params, n.Token.Pos, r.Name(), report)
		case *ast.LambdaExpr:
			checkParams(n.Params, n.Token.Pos, r.Name(), report)
		}
	}
}

func checkParams(params []ast.Param, pos token.Position, rule string, report func(Diagnostic)) {
	for _, p := range params {
		if p.Name == "_" {
			continue
		}
		if !isSnakeCase(p.Name) {
			report(Diagnostic{
				Pos:      pos,
				Severity: Warning,
				Rule:     rule,
				Message:  "parameter name \"" + p.Name + "\" should be snake_case",
			})
		}
	}
}

// isSnakeCase checks if a name matches [a-z_][a-z0-9_]*.
func isSnakeCase(name string) bool {
	if name == "_" {
		return true
	}
	if len(name) == 0 {
		return false
	}
	for i, c := range name {
		if c >= 'a' && c <= 'z' {
			continue
		}
		if c == '_' {
			continue
		}
		if i > 0 && c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}
