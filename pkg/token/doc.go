// Package token defines the lexical token types, keywords, and source positions
// used throughout the Meow language compiler.
//
// # Keywords
//
// All keywords use cat-themed names:
//
//	nyan      variable declaration
//	meow      function definition
//	bring     return value
//	sniff     if condition
//	scratch   else branch
//	purr      while loop
//	paw       lambda expression
//	nya       print
//	lick      map over list
//	picky     filter list
//	curl      reduce list
//	peek      pattern match
//	hiss      error / throw
//	fetch     import (planned)
//	flaunt    export (planned)
//	yarn      true literal
//	hairball  false literal
//	catnap    nil literal
//
// # Operators
//
//	+   -   *   /   %          arithmetic
//	==  !=                     equality
//	<   >   <=  >=             comparison
//	&&  ||  !                  logical
//	|=|                        pipe (chain operations)
//	..                         range (used in peek arms)
//	=>                         match arm separator
//	=                          assignment
//
// # Delimiters
//
//	( )    parentheses   — function calls, grouping
//	{ }    braces        — blocks
//	[ ]    brackets      — list literals, index access
//	,      comma         — separator
//
// # Comments
//
//	#            line comment
//	-~ ... ~-    block comment
//
// # Literals
//
//	42           integer
//	3.14         float
//	"hello"      string
package token
