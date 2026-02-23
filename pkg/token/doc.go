// Package token defines the lexical token types, keywords, and source positions
// used throughout the Meow language compiler.
//
// # Keywords
//
// All keywords use cat-themed names:
//
//	nyan      variable declaration      (Go: var / :=)
//	meow      function definition       (Go: func)
//	bring     return value              (Go: return)
//	sniff     if condition              (Go: if)
//	scratch   else branch               (Go: else)
//	purr      while loop                (Go: for)
//	paw       lambda expression         (Go: func(...))
//	nya       print                     (Go: fmt.Println)
//	lick      map over list
//	picky     filter list
//	curl      reduce list
//	peek      pattern match             (Go: switch)
//	hiss      error / throw             (Go: panic)
//	fetch     import (planned)          (Go: import)
//	flaunt    export (planned)
//	yarn      true literal              (Go: true)
//	hairball  false literal             (Go: false)
//	catnap    nil literal               (Go: nil)
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
