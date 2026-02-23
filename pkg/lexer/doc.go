// Package lexer implements the lexical scanner for the Meow language.
// It converts source text into a stream of tokens using [iter.Seq].
//
// # Usage
//
//	l := lexer.New(source, "hello.nyan")
//	for tok := range l.Tokens() {
//	    fmt.Println(tok.Type, tok.Literal)
//	}
//
// The lexer handles all Meow token types including cat-themed keywords,
// operators (such as the |=| pipe), string/number/boolean literals,
// line comments (#) and block comments (-~ ... ~-).
package lexer
