package ast

import "github.com/135yshr/meow/pkg/token"

// TypeExpr represents a type annotation in the AST.
type TypeExpr interface {
	Node
	typeExprTag()
}

// BasicType represents a type keyword (int, float, string, bool, furball, list).
type BasicType struct {
	Token token.Token
	Name  string // "int", "float", "string", "bool", "furball", "list"
}

func (n *BasicType) Pos() token.Position { return n.Token.Pos }
func (n *BasicType) nodeTag()            {}
func (n *BasicType) typeExprTag()        {}

// NamedType represents a user-defined type name (e.g. UserId, Nickname).
type NamedType struct {
	Token token.Token
	Name  string
}

func (n *NamedType) Pos() token.Position { return n.Token.Pos }
func (n *NamedType) nodeTag()            {}
func (n *NamedType) typeExprTag()        {}

// Param represents a function parameter with optional type annotation.
type Param struct {
	Name    string
	TypeAnn TypeExpr // nil = no type annotation
}
