package ast

import "github.com/135yshr/meow/pkg/token"

// TypeExpr represents a type annotation in the AST.
type TypeExpr interface {
	Node
	typeExprTag()
}

// BasicType represents a primitive type (int, float, string, bool).
type BasicType struct {
	Token token.Token
	Name  string // "int", "float", "string", "bool"
}

func (n *BasicType) Pos() token.Position { return n.Token.Pos }
func (n *BasicType) nodeTag()            {}
func (n *BasicType) typeExprTag()        {}

// Param represents a function parameter with optional type annotation.
type Param struct {
	Name    string
	TypeAnn TypeExpr // nil = no type annotation
}
