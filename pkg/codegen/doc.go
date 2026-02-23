// Package codegen translates the Meow AST into Go source code.
// The generated code depends on the meowrt runtime package for dynamic typing.
//
// # Generated Code Structure
//
// The output is a self-contained Go main package with the following layout:
//
//	package main
//	import meow "github.com/135yshr/meow/runtime/meowrt"
//	// user-defined functions
//	func main() { /* top-level statements */ }
//
// All Meow values are represented as meow.Value at runtime.
//
// # Runtime Dependency
//
// Generated code calls functions from the meowrt package for arithmetic,
// comparisons, built-in operations (nya, lick, picky, curl), and
// pattern matching.
//
// # Supported Constructs
//
//   - Variable declarations and assignments
//   - Function definitions and calls (including lambdas)
//   - Arithmetic, comparison, and logical operators
//   - If/else, while loops
//   - List literals and index access
//   - Pipe expressions (|=|)
//   - Error recovery expressions (~>)
//   - Pattern matching (peek)
//   - Error raising (hiss) and recovery (gag/isFurball)
//
// # Usage
//
//	gen := codegen.New()
//	goSource, err := gen.Generate(prog)
package codegen
