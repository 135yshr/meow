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

		case NegationRemoval:
			// After Apply, Op is ILLEGAL. Record Right as the mutation result.
			walkExprs(prog, func(expr ast.Expr) {
				if ue, ok := expr.(*ast.UnaryExpr); ok {
					if ue.Pos() == m.Pos {
						entry.Expr = ue.Right
						schema[ue] = append(schema[ue], entry)
					}
				}
			})

		case CatchRemove:
			// After Apply, Right is nil. Record Left as the mutation result.
			walkExprs(prog, func(expr ast.Expr) {
				if ce, ok := expr.(*ast.CatchExpr); ok {
					if ce.Pos() == m.Pos {
						entry.Expr = ce.Left
						schema[ce] = append(schema[ce], entry)
					}
				}
			})

		case PipeRemove:
			// After Apply, Right is nil. Record Left as the mutation result.
			walkExprs(prog, func(expr ast.Expr) {
				if pe, ok := expr.(*ast.PipeExpr); ok {
					if pe.Pos() == m.Pos {
						entry.Expr = pe.Left
						schema[pe] = append(schema[pe], entry)
					}
				}
			})

		case ConditionNegate:
			// Undo to capture original condition pointer as schema key,
			// then Apply to capture the negated condition as entry.Expr.
			m.Undo()
			walkStmts(prog, func(stmt ast.Stmt) {
				if ifStmt, ok := stmt.(*ast.IfStmt); ok && ifStmt.Pos() == m.Pos {
					key := ifStmt.Condition
					m.Apply()
					entry.Expr = ifStmt.Condition
					schema[key] = append(schema[key], entry)
				}
			})

		case ReturnNil:
			// Undo to capture original value pointer as schema key,
			// then Apply to capture the nil literal as entry.Expr.
			m.Undo()
			walkStmts(prog, func(stmt ast.Stmt) {
				if retStmt, ok := stmt.(*ast.ReturnStmt); ok && retStmt.Pos() == m.Pos && retStmt.Value != nil {
					key := retStmt.Value
					m.Apply()
					entry.Expr = &ast.NilLit{Token: retStmt.Token}
					schema[key] = append(schema[key], entry)
				}
			})
		}

		m.Undo()
	}

	return schema
}

// walkStmts walks all statements in the program, calling fn for each.
func walkStmts(prog *ast.Program, fn func(ast.Stmt)) {
	for _, stmt := range prog.Stmts {
		walkStmtTree(stmt, fn)
	}
}

func walkStmtTree(stmt ast.Stmt, fn func(ast.Stmt)) {
	fn(stmt)
	switch s := stmt.(type) {
	case *ast.FuncStmt:
		for _, body := range s.Body {
			walkStmtTree(body, fn)
		}
	case *ast.IfStmt:
		for _, body := range s.Body {
			walkStmtTree(body, fn)
		}
		for _, body := range s.ElseBody {
			walkStmtTree(body, fn)
		}
	case *ast.RangeStmt:
		for _, body := range s.Body {
			walkStmtTree(body, fn)
		}
	}
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
	case *ast.RangeStmt:
		walkExprTree(s.Start, fn)
		walkExprTree(s.End, fn)
		for _, body := range s.Body {
			walkStmtExprs(body, fn)
		}
	case *ast.ReturnStmt:
		if s.Value != nil {
			walkExprTree(s.Value, fn)
		}
	case *ast.VarStmt:
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
