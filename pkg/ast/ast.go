package ast

import "github.com/135yshr/meow/pkg/token"

// Node is the interface for all AST nodes.
type Node interface {
	Pos() token.Position
	nodeTag()
}

// Expr nodes produce a value.
type Expr interface {
	Node
	exprTag()
}

// Stmt nodes perform an action.
type Stmt interface {
	Node
	stmtTag()
}

// Program is the root AST node.
type Program struct {
	Stmts []Stmt
}

func (p *Program) Pos() token.Position {
	if len(p.Stmts) > 0 {
		return p.Stmts[0].Pos()
	}
	return token.Position{}
}
func (p *Program) nodeTag() {}

// --- Expressions ---

type IntLit struct {
	Token token.Token
	Value int64
}

func (n *IntLit) Pos() token.Position { return n.Token.Pos }
func (n *IntLit) nodeTag()            {}
func (n *IntLit) exprTag()            {}

type FloatLit struct {
	Token token.Token
	Value float64
}

func (n *FloatLit) Pos() token.Position { return n.Token.Pos }
func (n *FloatLit) nodeTag()            {}
func (n *FloatLit) exprTag()            {}

type StringLit struct {
	Token token.Token
	Value string
}

func (n *StringLit) Pos() token.Position { return n.Token.Pos }
func (n *StringLit) nodeTag()            {}
func (n *StringLit) exprTag()            {}

type BoolLit struct {
	Token token.Token
	Value bool
}

func (n *BoolLit) Pos() token.Position { return n.Token.Pos }
func (n *BoolLit) nodeTag()            {}
func (n *BoolLit) exprTag()            {}

type NilLit struct {
	Token token.Token
}

func (n *NilLit) Pos() token.Position { return n.Token.Pos }
func (n *NilLit) nodeTag()            {}
func (n *NilLit) exprTag()            {}

type Ident struct {
	Token token.Token
	Name  string
}

func (n *Ident) Pos() token.Position { return n.Token.Pos }
func (n *Ident) nodeTag()            {}
func (n *Ident) exprTag()            {}

type UnaryExpr struct {
	Token token.Token
	Op    token.TokenType
	Right Expr
}

func (n *UnaryExpr) Pos() token.Position { return n.Token.Pos }
func (n *UnaryExpr) nodeTag()            {}
func (n *UnaryExpr) exprTag()            {}

type BinaryExpr struct {
	Token token.Token
	Op    token.TokenType
	Left  Expr
	Right Expr
}

func (n *BinaryExpr) Pos() token.Position { return n.Token.Pos }
func (n *BinaryExpr) nodeTag()            {}
func (n *BinaryExpr) exprTag()            {}

type CallExpr struct {
	Token token.Token
	Fn    Expr
	Args  []Expr
}

func (n *CallExpr) Pos() token.Position { return n.Token.Pos }
func (n *CallExpr) nodeTag()            {}
func (n *CallExpr) exprTag()            {}

type LambdaExpr struct {
	Token  token.Token
	Params []string
	Body   Expr
}

func (n *LambdaExpr) Pos() token.Position { return n.Token.Pos }
func (n *LambdaExpr) nodeTag()            {}
func (n *LambdaExpr) exprTag()            {}

type ListLit struct {
	Token token.Token
	Items []Expr
}

func (n *ListLit) Pos() token.Position { return n.Token.Pos }
func (n *ListLit) nodeTag()            {}
func (n *ListLit) exprTag()            {}

type IndexExpr struct {
	Token token.Token
	Left  Expr
	Index Expr
}

func (n *IndexExpr) Pos() token.Position { return n.Token.Pos }
func (n *IndexExpr) nodeTag()            {}
func (n *IndexExpr) exprTag()            {}

type PipeExpr struct {
	Token token.Token
	Left  Expr
	Right Expr
}

func (n *PipeExpr) Pos() token.Position { return n.Token.Pos }
func (n *PipeExpr) nodeTag()            {}
func (n *PipeExpr) exprTag()            {}

// MatchExpr represents a peek (pattern match) expression.
type MatchExpr struct {
	Token   token.Token
	Subject Expr
	Arms    []MatchArm
}

func (n *MatchExpr) Pos() token.Position { return n.Token.Pos }
func (n *MatchExpr) nodeTag()            {}
func (n *MatchExpr) exprTag()            {}

type MatchArm struct {
	Pattern Pattern
	Body    Expr
}

// Pattern is the interface for match patterns.
type Pattern interface {
	Node
	patternTag()
}

type LiteralPattern struct {
	Token token.Token
	Value Expr
}

func (n *LiteralPattern) Pos() token.Position { return n.Token.Pos }
func (n *LiteralPattern) nodeTag()            {}
func (n *LiteralPattern) patternTag()         {}

type RangePattern struct {
	Token token.Token
	Low   Expr
	High  Expr
}

func (n *RangePattern) Pos() token.Position { return n.Token.Pos }
func (n *RangePattern) nodeTag()            {}
func (n *RangePattern) patternTag()         {}

type WildcardPattern struct {
	Token token.Token
}

func (n *WildcardPattern) Pos() token.Position { return n.Token.Pos }
func (n *WildcardPattern) nodeTag()            {}
func (n *WildcardPattern) patternTag()         {}

// --- Statements ---

type VarStmt struct {
	Token token.Token
	Name  string
	Value Expr
}

func (n *VarStmt) Pos() token.Position { return n.Token.Pos }
func (n *VarStmt) nodeTag()            {}
func (n *VarStmt) stmtTag()            {}

type AssignStmt struct {
	Token token.Token
	Name  string
	Value Expr
}

func (n *AssignStmt) Pos() token.Position { return n.Token.Pos }
func (n *AssignStmt) nodeTag()            {}
func (n *AssignStmt) stmtTag()            {}

type FuncStmt struct {
	Token  token.Token
	Name   string
	Params []string
	Body   []Stmt
}

func (n *FuncStmt) Pos() token.Position { return n.Token.Pos }
func (n *FuncStmt) nodeTag()            {}
func (n *FuncStmt) stmtTag()            {}

type ReturnStmt struct {
	Token token.Token
	Value Expr
}

func (n *ReturnStmt) Pos() token.Position { return n.Token.Pos }
func (n *ReturnStmt) nodeTag()            {}
func (n *ReturnStmt) stmtTag()            {}

type IfStmt struct {
	Token     token.Token
	Condition Expr
	Body      []Stmt
	ElseBody  []Stmt
}

func (n *IfStmt) Pos() token.Position { return n.Token.Pos }
func (n *IfStmt) nodeTag()            {}
func (n *IfStmt) stmtTag()            {}

type WhileStmt struct {
	Token     token.Token
	Condition Expr
	Body      []Stmt
}

func (n *WhileStmt) Pos() token.Position { return n.Token.Pos }
func (n *WhileStmt) nodeTag()            {}
func (n *WhileStmt) stmtTag()            {}

type ExprStmt struct {
	Token token.Token
	Expr  Expr
}

func (n *ExprStmt) Pos() token.Position { return n.Token.Pos }
func (n *ExprStmt) nodeTag()            {}
func (n *ExprStmt) stmtTag()            {}
