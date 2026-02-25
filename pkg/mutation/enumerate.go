package mutation

import (
	"fmt"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
)

// Enumerate walks the AST and returns all possible mutations.
func Enumerate(prog *ast.Program) []Mutant {
	var e enumerator
	for _, stmt := range prog.Stmts {
		e.enumStmt(stmt)
	}
	return e.mutants
}

type enumerator struct {
	mutants []Mutant
	nextID  MutantID
}

func (e *enumerator) add(desc string, pos token.Position, kind MutantKind, apply, undo func()) {
	e.mutants = append(e.mutants, Mutant{
		ID:          e.nextID,
		Description: desc,
		Pos:         pos,
		Kind:        kind,
		Apply:       apply,
		Undo:        undo,
	})
	e.nextID++
}

func (e *enumerator) enumStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.FuncStmt:
		for _, body := range s.Body {
			e.enumStmt(body)
		}
	case *ast.IfStmt:
		e.enumExpr(s.Condition)
		origCond := s.Condition
		e.add(
			fmt.Sprintf("negate if condition at %s", s.Pos()),
			s.Pos(), ConditionNegate,
			func() {
				s.Condition = &ast.UnaryExpr{
					Token: s.Token,
					Op:    token.NOT,
					Right: origCond,
				}
			},
			func() { s.Condition = origCond },
		)
		for _, body := range s.Body {
			e.enumStmt(body)
		}
		for _, body := range s.ElseBody {
			e.enumStmt(body)
		}
	case *ast.RangeStmt:
		e.enumExpr(s.Start)
		e.enumExpr(s.End)
		for _, body := range s.Body {
			e.enumStmt(body)
		}
	case *ast.ReturnStmt:
		if s.Value != nil {
			origValue := s.Value
			e.add(
				fmt.Sprintf("replace return with nil at %s", s.Pos()),
				s.Pos(), ReturnNil,
				func() { s.Value = &ast.NilLit{Token: s.Token} },
				func() { s.Value = origValue },
			)
			e.enumExpr(s.Value)
		}
	case *ast.VarStmt:
		e.enumExpr(s.Value)
	case *ast.ExprStmt:
		e.enumExpr(s.Expr)
	}
}

func (e *enumerator) enumExpr(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.BinaryExpr:
		e.enumBinary(ex)
		e.enumExpr(ex.Left)
		e.enumExpr(ex.Right)
	case *ast.UnaryExpr:
		e.enumUnary(ex)
		e.enumExpr(ex.Right)
	case *ast.BoolLit:
		origVal := ex.Value
		e.add(
			fmt.Sprintf("flip bool %t→%t at %s", ex.Value, !ex.Value, ex.Pos()),
			ex.Pos(), BoolFlip,
			func() { ex.Value = !ex.Value },
			func() { ex.Value = origVal },
		)
	case *ast.IntLit:
		origVal := ex.Value
		if ex.Value == 0 {
			e.add(
				fmt.Sprintf("int 0→1 at %s", ex.Pos()),
				ex.Pos(), IntBoundary,
				func() { ex.Value = 1 },
				func() { ex.Value = origVal },
			)
		} else {
			e.add(
				fmt.Sprintf("int %d→0 at %s", ex.Value, ex.Pos()),
				ex.Pos(), IntBoundary,
				func() { ex.Value = 0 },
				func() { ex.Value = origVal },
			)
		}
	case *ast.StringLit:
		origVal := ex.Value
		if ex.Value == "" {
			e.add(
				fmt.Sprintf("string \"\"→\"mutant\" at %s", ex.Pos()),
				ex.Pos(), StringEmpty,
				func() { ex.Value = "mutant" },
				func() { ex.Value = origVal },
			)
		} else {
			e.add(
				fmt.Sprintf("string %q→\"\" at %s", ex.Value, ex.Pos()),
				ex.Pos(), StringEmpty,
				func() { ex.Value = "" },
				func() { ex.Value = origVal },
			)
		}
	case *ast.CallExpr:
		for _, arg := range ex.Args {
			e.enumExpr(arg)
		}
	case *ast.LambdaExpr:
		e.enumExpr(ex.Body)
	case *ast.ListLit:
		for _, item := range ex.Items {
			e.enumExpr(item)
		}
	case *ast.IndexExpr:
		e.enumExpr(ex.Left)
		e.enumExpr(ex.Index)
	case *ast.PipeExpr:
		origRight := ex.Right
		e.add(
			fmt.Sprintf("remove pipe at %s", ex.Pos()),
			ex.Pos(), PipeRemove,
			func() { ex.Right = nil },
			func() { ex.Right = origRight },
		)
		e.enumExpr(ex.Left)
		e.enumExpr(ex.Right)
	case *ast.CatchExpr:
		origRight := ex.Right
		e.add(
			fmt.Sprintf("remove catch at %s", ex.Pos()),
			ex.Pos(), CatchRemove,
			func() { ex.Right = nil },
			func() { ex.Right = origRight },
		)
		e.enumExpr(ex.Left)
		e.enumExpr(ex.Right)
	case *ast.MapLit:
		for _, v := range ex.Vals {
			e.enumExpr(v)
		}
	case *ast.MatchExpr:
		e.enumExpr(ex.Subject)
		for _, arm := range ex.Arms {
			e.enumExpr(arm.Body)
		}
	}
}

var arithmeticSwaps = map[token.TokenType]token.TokenType{
	token.PLUS:  token.MINUS,
	token.MINUS: token.PLUS,
	token.STAR:  token.SLASH,
	token.SLASH: token.STAR,
}

var comparisonSwaps = map[token.TokenType]token.TokenType{
	token.EQ:  token.NEQ,
	token.NEQ: token.EQ,
	token.LT:  token.LTE,
	token.LTE: token.LT,
	token.GT:  token.GTE,
	token.GTE: token.GT,
}

var logicalSwaps = map[token.TokenType]token.TokenType{
	token.AND: token.OR,
	token.OR:  token.AND,
}

func (e *enumerator) enumBinary(ex *ast.BinaryExpr) {
	origOp := ex.Op
	if swapped, ok := arithmeticSwaps[ex.Op]; ok {
		e.add(
			fmt.Sprintf("swap %s→%s at %s", ex.Op, swapped, ex.Pos()),
			ex.Pos(), ArithmeticSwap,
			func() { ex.Op = swapped },
			func() { ex.Op = origOp },
		)
	}
	if swapped, ok := comparisonSwaps[ex.Op]; ok {
		e.add(
			fmt.Sprintf("swap %s→%s at %s", ex.Op, swapped, ex.Pos()),
			ex.Pos(), ComparisonSwap,
			func() { ex.Op = swapped },
			func() { ex.Op = origOp },
		)
	}
	if swapped, ok := logicalSwaps[ex.Op]; ok {
		e.add(
			fmt.Sprintf("swap %s→%s at %s", ex.Op, swapped, ex.Pos()),
			ex.Pos(), LogicalSwap,
			func() { ex.Op = swapped },
			func() { ex.Op = origOp },
		)
	}
}

func (e *enumerator) enumUnary(ex *ast.UnaryExpr) {
	if ex.Op == token.MINUS || ex.Op == token.NOT {
		origOp := ex.Op
		origRight := ex.Right
		e.add(
			fmt.Sprintf("remove %s at %s", ex.Op, ex.Pos()),
			ex.Pos(), NegationRemoval,
			func() {
				// Mark for removal: set Op to ILLEGAL to signal identity
				ex.Op = token.ILLEGAL
			},
			func() {
				ex.Op = origOp
				ex.Right = origRight
			},
		)
	}
}
