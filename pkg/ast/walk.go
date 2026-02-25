package ast

import "iter"

// Preorder returns a depth-first pre-order iterator over the AST.
func Preorder(node Node) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		walk(node, yield)
	}
}

func walk(node Node, yield func(Node) bool) bool {
	if node == nil {
		return true
	}
	if !yield(node) {
		return false
	}
	switch n := node.(type) {
	case *Program:
		for _, s := range n.Stmts {
			if !walk(s, yield) {
				return false
			}
		}
	case *VarStmt:
		if !walk(n.Value, yield) {
			return false
		}
	case *FuncStmt:
		for _, s := range n.Body {
			if !walk(s, yield) {
				return false
			}
		}
	case *ReturnStmt:
		if n.Value != nil {
			if !walk(n.Value, yield) {
				return false
			}
		}
	case *IfStmt:
		if !walk(n.Condition, yield) {
			return false
		}
		for _, s := range n.Body {
			if !walk(s, yield) {
				return false
			}
		}
		for _, s := range n.ElseBody {
			if !walk(s, yield) {
				return false
			}
		}
	case *RangeStmt:
		if n.Start != nil {
			if !walk(n.Start, yield) {
				return false
			}
		}
		if !walk(n.End, yield) {
			return false
		}
		for _, s := range n.Body {
			if !walk(s, yield) {
				return false
			}
		}
	case *ExprStmt:
		if !walk(n.Expr, yield) {
			return false
		}
	case *BinaryExpr:
		if !walk(n.Left, yield) {
			return false
		}
		if !walk(n.Right, yield) {
			return false
		}
	case *UnaryExpr:
		if !walk(n.Right, yield) {
			return false
		}
	case *CallExpr:
		if !walk(n.Fn, yield) {
			return false
		}
		for _, a := range n.Args {
			if !walk(a, yield) {
				return false
			}
		}
	case *LambdaExpr:
		if !walk(n.Body, yield) {
			return false
		}
	case *ListLit:
		for _, item := range n.Items {
			if !walk(item, yield) {
				return false
			}
		}
	case *IndexExpr:
		if !walk(n.Left, yield) {
			return false
		}
		if !walk(n.Index, yield) {
			return false
		}
	case *PipeExpr:
		if !walk(n.Left, yield) {
			return false
		}
		if !walk(n.Right, yield) {
			return false
		}
	case *MatchExpr:
		if !walk(n.Subject, yield) {
			return false
		}
		for _, arm := range n.Arms {
			if !walk(arm.Pattern, yield) {
				return false
			}
			if !walk(arm.Body, yield) {
				return false
			}
		}
	case *RangePattern:
		if !walk(n.Low, yield) {
			return false
		}
		if !walk(n.High, yield) {
			return false
		}
	case *LiteralPattern:
		if !walk(n.Value, yield) {
			return false
		}
	case *FetchStmt:
		// leaf node, no children
	case *MemberExpr:
		if !walk(n.Object, yield) {
			return false
		}
	}
	return true
}
