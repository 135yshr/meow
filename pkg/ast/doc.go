// Package ast defines the abstract syntax tree nodes for the Meow language.
// It contains expression, statement, and pattern node types produced by the
// parser.
//
// # Interfaces
//
// All AST nodes implement [Node]. The three main categories are:
//
//   - [Expr]    — nodes that produce a value
//   - [Stmt]    — nodes that perform an action
//   - [Pattern] — nodes used in peek (match) arms
//
// # Expressions
//
//   - [IntLit]      integer literal (42)
//   - [FloatLit]    floating-point literal (3.14)
//   - [StringLit]   string literal ("hello")
//   - [BoolLit]     boolean literal (yarn / hairball)
//   - [NilLit]      nil literal (catnap)
//   - [Ident]       identifier
//   - [UnaryExpr]   unary operation (! -)
//   - [BinaryExpr]  binary operation (+ - * / % == != < > <= >= && ||)
//   - [CallExpr]    function call
//   - [LambdaExpr]  lambda expression (paw)
//   - [ListLit]     list literal ([1, 2, 3])
//   - [IndexExpr]   index access (list[0])
//   - [PipeExpr]    pipe operation (|=|)
//   - [MatchExpr]   pattern match (peek)
//
// # Statements
//
//   - [VarStmt]     variable declaration (nyan x = ... or x = ...)
//   - [FuncStmt]    function definition (meow f(x) { ... })
//   - [ReturnStmt]  return statement (bring ...)
//   - [IfStmt]      conditional (sniff / scratch)
//   - [RangeStmt]   range-based loop (purr i (n) or purr i (a..b))
//   - [ExprStmt]    expression used as a statement
//
// # Patterns
//
//   - [LiteralPattern]   matches a specific value
//   - [RangePattern]     matches an inclusive range (1..10)
//   - [WildcardPattern]  matches any value (_)
//
// # Tree Walking
//
// Callers can walk the tree by type-switching on [Expr], [Stmt], or [Pattern].
// The [Program] node is the root and holds a slice of top-level [Stmt] nodes.
package ast
