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
	// Stmts is the list of top-level statements.
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
	// Token is the source token.
	Token token.Token
	// Value is the parsed integer value.
	Value int64
}

func (n *IntLit) Pos() token.Position { return n.Token.Pos }
func (n *IntLit) nodeTag()            {}
func (n *IntLit) exprTag()            {}

// FloatLit represents a floating-point literal.
type FloatLit struct {
	// Token is the source token.
	Token token.Token
	// Value is the parsed float value.
	Value float64
}

func (n *FloatLit) Pos() token.Position { return n.Token.Pos }
func (n *FloatLit) nodeTag()            {}
func (n *FloatLit) exprTag()            {}

// StringLit represents a string literal.
type StringLit struct {
	// Token is the source token.
	Token token.Token
	// Value is the string content without quotes.
	Value string
}

func (n *StringLit) Pos() token.Position { return n.Token.Pos }
func (n *StringLit) nodeTag()            {}
func (n *StringLit) exprTag()            {}

// BoolLit represents a boolean literal (yarn/hairball).
type BoolLit struct {
	// Token is the source token.
	Token token.Token
	// Value is true for yarn, false for hairball.
	Value bool
}

func (n *BoolLit) Pos() token.Position { return n.Token.Pos }
func (n *BoolLit) nodeTag()            {}
func (n *BoolLit) exprTag()            {}

// NilLit represents a nil literal (catnap).
type NilLit struct {
	// Token is the source token.
	Token token.Token
}

func (n *NilLit) Pos() token.Position { return n.Token.Pos }
func (n *NilLit) nodeTag()            {}
func (n *NilLit) exprTag()            {}

// Ident represents an identifier.
type Ident struct {
	// Token is the source token.
	Token token.Token
	// Name is the identifier name.
	Name string
}

func (n *Ident) Pos() token.Position { return n.Token.Pos }
func (n *Ident) nodeTag()            {}
func (n *Ident) exprTag()            {}

// UnaryExpr represents a unary operation (e.g. -x, !x).
type UnaryExpr struct {
	// Token is the operator token.
	Token token.Token
	// Op is the operator type (MINUS or NOT).
	Op token.TokenType
	// Right is the operand expression.
	Right Expr
}

func (n *UnaryExpr) Pos() token.Position { return n.Token.Pos }
func (n *UnaryExpr) nodeTag()            {}
func (n *UnaryExpr) exprTag()            {}

// BinaryExpr represents a binary operation (e.g. a + b).
type BinaryExpr struct {
	// Token is the operator token.
	Token token.Token
	// Op is the operator type.
	Op token.TokenType
	// Left is the left-hand operand.
	Left Expr
	// Right is the right-hand operand.
	Right Expr
}

func (n *BinaryExpr) Pos() token.Position { return n.Token.Pos }
func (n *BinaryExpr) nodeTag()            {}
func (n *BinaryExpr) exprTag()            {}

// CallExpr represents a function call (e.g. greet(name)).
type CallExpr struct {
	// Token is the opening parenthesis token.
	Token token.Token
	// Fn is the function being called.
	Fn Expr
	// Args is the list of arguments.
	Args []Expr
}

func (n *CallExpr) Pos() token.Position { return n.Token.Pos }
func (n *CallExpr) nodeTag()            {}
func (n *CallExpr) exprTag()            {}

// LambdaExpr represents a lambda expression (e.g. paw(x) { x * 2 }).
type LambdaExpr struct {
	// Token is the paw keyword token.
	Token token.Token
	// Params is the list of parameter names.
	Params []string
	// Body is the lambda body expression.
	Body Expr
}

func (n *LambdaExpr) Pos() token.Position { return n.Token.Pos }
func (n *LambdaExpr) nodeTag()            {}
func (n *LambdaExpr) exprTag()            {}

// ListLit represents a list literal (e.g. [1, 2, 3]).
type ListLit struct {
	// Token is the opening bracket token.
	Token token.Token
	// Items is the list of element expressions.
	Items []Expr
}

func (n *ListLit) Pos() token.Position { return n.Token.Pos }
func (n *ListLit) nodeTag()            {}
func (n *ListLit) exprTag()            {}

// IndexExpr represents an index access (e.g. list[0]).
type IndexExpr struct {
	// Token is the opening bracket token.
	Token token.Token
	// Left is the expression being indexed.
	Left Expr
	// Index is the index expression.
	Index Expr
}

func (n *IndexExpr) Pos() token.Position { return n.Token.Pos }
func (n *IndexExpr) nodeTag()            {}
func (n *IndexExpr) exprTag()            {}

// PipeExpr represents a pipe operation (e.g. xs |> lick(f)).
type PipeExpr struct {
	// Token is the pipe operator token.
	Token token.Token
	// Left is the input expression.
	Left Expr
	// Right is the function expression to pipe into.
	Right Expr
}

func (n *PipeExpr) Pos() token.Position { return n.Token.Pos }
func (n *PipeExpr) nodeTag()            {}
func (n *PipeExpr) exprTag()            {}

// MatchExpr represents a peek (pattern match) expression.
type MatchExpr struct {
	// Token is the peek keyword token.
	Token token.Token
	// Subject is the expression being matched.
	Subject Expr
	// Arms is the list of match arms.
	Arms []MatchArm
}

func (n *MatchExpr) Pos() token.Position { return n.Token.Pos }
func (n *MatchExpr) nodeTag()            {}
func (n *MatchExpr) exprTag()            {}

// MatchArm represents a single arm in a peek expression.
type MatchArm struct {
	// Pattern is the pattern to match against.
	Pattern Pattern
	// Body is the expression to evaluate when matched.
	Body Expr
}

// Pattern is the interface for match patterns.
type Pattern interface {
	Node
	patternTag()
}

// LiteralPattern matches a specific value.
type LiteralPattern struct {
	// Token is the source token.
	Token token.Token
	// Value is the literal value to match.
	Value Expr
}

func (n *LiteralPattern) Pos() token.Position { return n.Token.Pos }
func (n *LiteralPattern) nodeTag()            {}
func (n *LiteralPattern) patternTag()         {}

// RangePattern matches an inclusive range of values (e.g. 1..10).
type RangePattern struct {
	// Token is the dotdot operator token.
	Token token.Token
	// Low is the lower bound (inclusive).
	Low Expr
	// High is the upper bound (inclusive).
	High Expr
}

func (n *RangePattern) Pos() token.Position { return n.Token.Pos }
func (n *RangePattern) nodeTag()            {}
func (n *RangePattern) patternTag()         {}

// WildcardPattern matches any value (_).
type WildcardPattern struct {
	// Token is the underscore token.
	Token token.Token
}

func (n *WildcardPattern) Pos() token.Position { return n.Token.Pos }
func (n *WildcardPattern) nodeTag()            {}
func (n *WildcardPattern) patternTag()         {}

// --- Statements ---

// VarStmt represents a variable declaration (nyan x = ...).
type VarStmt struct {
	// Token is the nyan keyword token.
	Token token.Token
	// Name is the variable name.
	Name string
	// Value is the initial value expression.
	Value Expr
}

func (n *VarStmt) Pos() token.Position { return n.Token.Pos }
func (n *VarStmt) nodeTag()            {}
func (n *VarStmt) stmtTag()            {}

// AssignStmt represents a variable reassignment (x = ...).
type AssignStmt struct {
	// Token is the assignment operator token.
	Token token.Token
	// Name is the variable name.
	Name string
	// Value is the new value expression.
	Value Expr
}

func (n *AssignStmt) Pos() token.Position { return n.Token.Pos }
func (n *AssignStmt) nodeTag()            {}
func (n *AssignStmt) stmtTag()            {}

// FuncStmt represents a function definition (meow f(x) { ... }).
type FuncStmt struct {
	// Token is the meow keyword token.
	Token token.Token
	// Name is the function name.
	Name string
	// Params is the list of parameter names.
	Params []string
	// Body is the list of statements in the function body.
	Body []Stmt
}

func (n *FuncStmt) Pos() token.Position { return n.Token.Pos }
func (n *FuncStmt) nodeTag()            {}
func (n *FuncStmt) stmtTag()            {}

// ReturnStmt represents a return statement (bring ...).
type ReturnStmt struct {
	// Token is the bring keyword token.
	Token token.Token
	// Value is the return value expression, or nil for bare return.
	Value Expr
}

func (n *ReturnStmt) Pos() token.Position { return n.Token.Pos }
func (n *ReturnStmt) nodeTag()            {}
func (n *ReturnStmt) stmtTag()            {}

// IfStmt represents a conditional statement (sniff/scratch).
type IfStmt struct {
	// Token is the sniff keyword token.
	Token token.Token
	// Condition is the condition expression.
	Condition Expr
	// Body is the list of statements in the then branch.
	Body []Stmt
	// ElseBody is the list of statements in the else branch.
	ElseBody []Stmt
}

func (n *IfStmt) Pos() token.Position { return n.Token.Pos }
func (n *IfStmt) nodeTag()            {}
func (n *IfStmt) stmtTag()            {}

// WhileStmt represents a while loop (purr).
type WhileStmt struct {
	// Token is the purr keyword token.
	Token token.Token
	// Condition is the loop condition expression.
	Condition Expr
	// Body is the list of statements in the loop body.
	Body []Stmt
}

func (n *WhileStmt) Pos() token.Position { return n.Token.Pos }
func (n *WhileStmt) nodeTag()            {}
func (n *WhileStmt) stmtTag()            {}

// ExprStmt represents an expression used as a statement.
type ExprStmt struct {
	// Token is the first token of the expression.
	Token token.Token
	// Expr is the expression.
	Expr Expr
}

func (n *ExprStmt) Pos() token.Position { return n.Token.Pos }
func (n *ExprStmt) nodeTag()            {}
func (n *ExprStmt) stmtTag()            {}
