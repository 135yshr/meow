// Package compiler orchestrates the Meow compilation pipeline:
// lexing, parsing, code generation, and invoking go build to produce a binary.
//
// # Pipeline
//
// The compilation process follows these stages:
//
//  1. Lex — source text is tokenized by [github.com/135yshr/meow/pkg/lexer]
//  2. Parse — tokens are parsed into an AST by [github.com/135yshr/meow/pkg/parser]
//  3. Codegen — the AST is translated to Go source by [github.com/135yshr/meow/pkg/codegen]
//  4. Format — the generated Go code is formatted with go/format
//  5. Build — go build compiles the Go source into an executable (optional)
//
// # Usage
//
//	// Transpile only
//	c := compiler.New(logger)
//	goCode, err := c.CompileToGo(source, "hello.nyan")
//	_ = goCode
//
//	// Build a binary
//	c := compiler.New(logger)
//	err := c.Build("hello.nyan", "hello")
//
//	// Compile and run
//	c := compiler.New(logger)
//	err := c.Run("hello.nyan")
package compiler
