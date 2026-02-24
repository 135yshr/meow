package mutation

import "github.com/135yshr/meow/pkg/ast"

// MutationEntry maps a mutant ID to the alternative expression to generate.
type MutationEntry struct {
	ID   MutantID
	Expr ast.Expr
}

// BuildSchema creates a mapping from original AST expressions to their mutation entries.
// This is used by codegen to embed all mutations in a single binary.
func BuildSchema(prog *ast.Program, mutants []Mutant) map[ast.Expr][]MutationEntry {
	schema := make(map[ast.Expr][]MutationEntry)

	for _, m := range mutants {
		// Apply the mutation temporarily to capture the mutated state
		m.Apply()

		// Find the affected expression and record it
		entry := MutationEntry{ID: m.ID}

		switch m.Kind {
		case ArithmeticSwap, ComparisonSwap, LogicalSwap:
			// For binary operator swaps, the mutant's Apply changes the Op field.
			// We need to find the binary expression and record the swapped version.
			walkExprs(prog, func(expr ast.Expr) {
				if be, ok := expr.(*ast.BinaryExpr); ok {
					if be.Pos() == m.Pos {
						// Clone the current (mutated) state
						entry.Expr = &ast.BinaryExpr{
							Token: be.Token,
							Op:    be.Op,
							Left:  be.Left,
							Right: be.Right,
						}
						schema[be] = append(schema[be], entry)
					}
				}
			})

		case BoolFlip:
			walkExprs(prog, func(expr ast.Expr) {
				if bl, ok := expr.(*ast.BoolLit); ok {
					if bl.Pos() == m.Pos {
						entry.Expr = &ast.BoolLit{Token: bl.Token, Value: bl.Value}
						schema[bl] = append(schema[bl], entry)
					}
				}
			})

		case IntBoundary:
			walkExprs(prog, func(expr ast.Expr) {
				if il, ok := expr.(*ast.IntLit); ok {
					if il.Pos() == m.Pos {
						entry.Expr = &ast.IntLit{Token: il.Token, Value: il.Value}
						schema[il] = append(schema[il], entry)
					}
				}
			})

		case StringEmpty:
			walkExprs(prog, func(expr ast.Expr) {
				if sl, ok := expr.(*ast.StringLit); ok {
					if sl.Pos() == m.Pos {
						entry.Expr = &ast.StringLit{Token: sl.Token, Value: sl.Value}
						schema[sl] = append(schema[sl], entry)
					}
				}
			})
		}

		m.Undo()
	}

	return schema
}

// walkExprs walks all expressions in the program, calling fn for each.
func walkExprs(prog *ast.Program, fn func(ast.Expr)) {
	for _, stmt := range prog.Stmts {
		walkStmtExprs(stmt, fn)
	}
}

func walkStmtExprs(stmt ast.Stmt, fn func(ast.Expr)) {
	switch s := stmt.(type) {
	case *ast.FuncStmt:
		for _, body := range s.Body {
			walkStmtExprs(body, fn)
		}
	case *ast.IfStmt:
		walkExprTree(s.Condition, fn)
		for _, body := range s.Body {
			walkStmtExprs(body, fn)
		}
		for _, body := range s.ElseBody {
			walkStmtExprs(body, fn)
		}
	case *ast.WhileStmt:
		walkExprTree(s.Condition, fn)
		for _, body := range s.Body {
			walkStmtExprs(body, fn)
		}
	case *ast.ReturnStmt:
		if s.Value != nil {
			walkExprTree(s.Value, fn)
		}
	case *ast.VarStmt:
		walkExprTree(s.Value, fn)
	case *ast.AssignStmt:
		walkExprTree(s.Value, fn)
	case *ast.ExprStmt:
		walkExprTree(s.Expr, fn)
	}
}

func walkExprTree(expr ast.Expr, fn func(ast.Expr)) {
	if expr == nil {
		return
	}
	fn(expr)
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		walkExprTree(e.Left, fn)
		walkExprTree(e.Right, fn)
	case *ast.UnaryExpr:
		walkExprTree(e.Right, fn)
	case *ast.CallExpr:
		walkExprTree(e.Fn, fn)
		for _, arg := range e.Args {
			walkExprTree(arg, fn)
		}
	case *ast.LambdaExpr:
		walkExprTree(e.Body, fn)
	case *ast.ListLit:
		for _, item := range e.Items {
			walkExprTree(item, fn)
		}
	case *ast.IndexExpr:
		walkExprTree(e.Left, fn)
		walkExprTree(e.Index, fn)
	case *ast.PipeExpr:
		walkExprTree(e.Left, fn)
		walkExprTree(e.Right, fn)
	case *ast.CatchExpr:
		walkExprTree(e.Left, fn)
		walkExprTree(e.Right, fn)
	case *ast.MapLit:
		for _, v := range e.Vals {
			walkExprTree(v, fn)
		}
	case *ast.MatchExpr:
		walkExprTree(e.Subject, fn)
		for _, arm := range e.Arms {
			walkExprTree(arm.Body, fn)
		}
	}
}
