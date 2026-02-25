package linter

import (
	"github.com/135yshr/meow/pkg/ast"
)

// UnreachableCodeRule detects statements after a bring (return) in the same block.
type UnreachableCodeRule struct{}

func (r *UnreachableCodeRule) Name() string { return "unreachable-code" }

func (r *UnreachableCodeRule) Check(prog *ast.Program, report func(Diagnostic)) {
	for _, stmt := range prog.Stmts {
		r.checkStmt(stmt, report)
	}
}

func (r *UnreachableCodeRule) checkStmt(stmt ast.Stmt, report func(Diagnostic)) {
	switch s := stmt.(type) {
	case *ast.FuncStmt:
		r.checkBlock(s.Body, report)
	case *ast.IfStmt:
		r.checkBlock(s.Body, report)
		if len(s.ElseBody) > 0 {
			r.checkBlock(s.ElseBody, report)
		}
	case *ast.RangeStmt:
		r.checkBlock(s.Body, report)
	}
}

func (r *UnreachableCodeRule) checkBlock(stmts []ast.Stmt, report func(Diagnostic)) {
	foundReturn := false
	for _, stmt := range stmts {
		if foundReturn {
			report(Diagnostic{
				Pos:      stmt.Pos(),
				Severity: Warning,
				Rule:     r.Name(),
				Message:  "unreachable code after bring",
			})
			return // only report the first unreachable statement
		}
		if _, ok := stmt.(*ast.ReturnStmt); ok {
			foundReturn = true
		}
		// Recurse into nested blocks
		r.checkStmt(stmt, report)
	}
}
