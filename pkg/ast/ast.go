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

// IntLit represents an integer literal.
type IntLit struct {
	Token token.Token
	Value int64
}

func (n *IntLit) Pos() token.Position { return n.Token.Pos }
func (n *IntLit) nodeTag()            {}
func (n *IntLit) exprTag()            {}

// FloatLit represents a floating-point literal.
type FloatLit struct {
	Token token.Token
	Value float64
}

func (n *FloatLit) Pos() token.Position { return n.Token.Pos }
func (n *FloatLit) nodeTag()            {}
func (n *FloatLit) exprTag()            {}

// StringLit represents a string literal.
type StringLit struct {
	Token token.Token
	Value string
}

func (n *StringLit) Pos() token.Position { return n.Token.Pos }
func (n *StringLit) nodeTag()            {}
func (n *StringLit) exprTag()            {}

// BoolLit represents a boolean literal (yarn/hairball).
type BoolLit struct {
	Token token.Token
	Value bool
}

func (n *BoolLit) Pos() token.Position { return n.Token.Pos }
func (n *BoolLit) nodeTag()            {}
func (n *BoolLit) exprTag()            {}

// NilLit represents a nil literal (catnap).
type NilLit struct {
	Token token.Token
}

func (n *NilLit) Pos() token.Position { return n.Token.Pos }
func (n *NilLit) nodeTag()            {}
func (n *NilLit) exprTag()            {}

// Ident represents an identifier.
type Ident struct {
	Token token.Token
	Name  string
}

func (n *Ident) Pos() token.Position { return n.Token.Pos }
func (n *Ident) nodeTag()            {}
func (n *Ident) exprTag()            {}

// UnaryExpr represents a unary operation (e.g. -x, !x).
type UnaryExpr struct {
	Token token.Token
	Op    token.TokenType
	Right Expr
}

func (n *UnaryExpr) Pos() token.Position { return n.Token.Pos }
func (n *UnaryExpr) nodeTag()            {}
func (n *UnaryExpr) exprTag()            {}

// BinaryExpr represents a binary operation (e.g. a + b).
type BinaryExpr struct {
	Token token.Token
	Op    token.TokenType
	Left  Expr
	Right Expr
}

func (n *BinaryExpr) Pos() token.Position { return n.Token.Pos }
func (n *BinaryExpr) nodeTag()            {}
func (n *BinaryExpr) exprTag()            {}

// CallExpr represents a function call (e.g. greet(name)).
type CallExpr struct {
	Token token.Token
	Fn    Expr
	Args  []Expr
}

func (n *CallExpr) Pos() token.Position { return n.Token.Pos }
func (n *CallExpr) nodeTag()            {}
func (n *CallExpr) exprTag()            {}

// LambdaExpr represents a lambda expression (e.g. paw(x) { x * 2 }).
type LambdaExpr struct {
	Token  token.Token
	Params []string
	Body   Expr
}

func (n *LambdaExpr) Pos() token.Position { return n.Token.Pos }
func (n *LambdaExpr) nodeTag()            {}
func (n *LambdaExpr) exprTag()            {}

// ListLit represents a list literal (e.g. [1, 2, 3]).
type ListLit struct {
	Token token.Token
	Items []Expr
}

func (n *ListLit) Pos() token.Position { return n.Token.Pos }
func (n *ListLit) nodeTag()            {}
func (n *ListLit) exprTag()            {}

// IndexExpr represents an index access (e.g. list[0]).
type IndexExpr struct {
	Token token.Token
	Left  Expr
	Index Expr
}

func (n *IndexExpr) Pos() token.Position { return n.Token.Pos }
func (n *IndexExpr) nodeTag()            {}
func (n *IndexExpr) exprTag()            {}

// PipeExpr represents a pipe operation (e.g. xs |> lick(f)).
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

// MatchArm represents a single arm in a peek expression.
type MatchArm struct {
	Pattern Pattern
	Body    Expr
}

// Pattern is the interface for match patterns.
type Pattern interface {
	Node
	patternTag()
}

// LiteralPattern matches a specific value.
type LiteralPattern struct {
	Token token.Token
	Value Expr
}

func (n *LiteralPattern) Pos() token.Position { return n.Token.Pos }
func (n *LiteralPattern) nodeTag()            {}
func (n *LiteralPattern) patternTag()         {}

// RangePattern matches a range of values (e.g. 1..10).
type RangePattern struct {
	Token token.Token
	Low   Expr
	High  Expr
}

func (n *RangePattern) Pos() token.Position { return n.Token.Pos }
func (n *RangePattern) nodeTag()            {}
func (n *RangePattern) patternTag()         {}

// WildcardPattern matches any value (_).
type WildcardPattern struct {
	Token token.Token
}

func (n *WildcardPattern) Pos() token.Position { return n.Token.Pos }
func (n *WildcardPattern) nodeTag()            {}
func (n *WildcardPattern) patternTag()         {}

// --- Statements ---

// VarStmt represents a variable declaration (nyan x = ...).
type VarStmt struct {
	Token token.Token
	Name  string
	Value Expr
}

func (n *VarStmt) Pos() token.Position { return n.Token.Pos }
func (n *VarStmt) nodeTag()            {}
func (n *VarStmt) stmtTag()            {}

// AssignStmt represents a variable reassignment (x = ...).
type AssignStmt struct {
	Token token.Token
	Name  string
	Value Expr
}

func (n *AssignStmt) Pos() token.Position { return n.Token.Pos }
func (n *AssignStmt) nodeTag()            {}
func (n *AssignStmt) stmtTag()            {}

// FuncStmt represents a function definition (meow f(x) { ... }).
type FuncStmt struct {
	Token  token.Token
	Name   string
	Params []string
	Body   []Stmt
}

func (n *FuncStmt) Pos() token.Position { return n.Token.Pos }
func (n *FuncStmt) nodeTag()            {}
func (n *FuncStmt) stmtTag()            {}

// ReturnStmt represents a return statement (bring ...).
type ReturnStmt struct {
	Token token.Token
	Value Expr
}

func (n *ReturnStmt) Pos() token.Position { return n.Token.Pos }
func (n *ReturnStmt) nodeTag()            {}
func (n *ReturnStmt) stmtTag()            {}

// IfStmt represents a conditional statement (sniff/scratch).
type IfStmt struct {
	Token     token.Token
	Condition Expr
	Body      []Stmt
	ElseBody  []Stmt
}

func (n *IfStmt) Pos() token.Position { return n.Token.Pos }
func (n *IfStmt) nodeTag()            {}
func (n *IfStmt) stmtTag()            {}

// WhileStmt represents a while loop (purr).
type WhileStmt struct {
	Token     token.Token
	Condition Expr
	Body      []Stmt
}

func (n *WhileStmt) Pos() token.Position { return n.Token.Pos }
func (n *WhileStmt) nodeTag()            {}
func (n *WhileStmt) stmtTag()            {}

// ExprStmt represents an expression used as a statement.
type ExprStmt struct {
	Token token.Token
	Expr  Expr
}

func (n *ExprStmt) Pos() token.Position { return n.Token.Pos }
func (n *ExprStmt) nodeTag()            {}
func (n *ExprStmt) stmtTag()            {}
