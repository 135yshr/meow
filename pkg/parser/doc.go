// Package parser implements a Pratt (precedence-climbing) recursive descent
// parser for the Meow language. It consumes tokens via [iter.Pull] and
// produces an [ast.Program] AST.
//
// # Parsing Strategy
//
// Expressions are parsed using a Pratt parser with the following precedence
// levels (lowest to highest):
//
//	||           logical OR
//	&&           logical AND
//	== !=        equality
//	< > <= >=    comparison
//	|=|          pipe
//	+ -          additive
//	* / %        multiplicative
//	! - (unary)  prefix operators
//	() []        call / index
//
// Statements are dispatched by leading keyword:
//
//	nyan      variable declaration
//	meow      function definition
//	bring     return
//	sniff     if / else if / else
//	purr      while loop
//
// # Error Handling
//
// Parse errors are collected as [*ParseError] values. When errors occur
// the parser attempts to continue, so multiple errors may be reported
// from a single parse.
//
// # Usage
//
//	p := parser.New(lexer.New(src, file).Tokens())
//	prog, errs := p.Parse()
//	if errs != nil {
//	    for _, e := range errs {
//	        fmt.Println(e)
//	    }
//	}
package parser
