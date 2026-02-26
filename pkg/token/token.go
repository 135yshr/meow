package token

import "fmt"

// TokenType represents the type of a lexical token.
//
//go:generate stringer -type=TokenType
type TokenType int

const (
	// Special
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT
	INT
	FLOAT
	STRING

	// Operators
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %
	ASSIGN  // =
	EQ      // ==
	NEQ     // !=
	LT      // <
	GT      // >
	LTE     // <=
	GTE     // >=
	AND     // &&
	OR      // ||
	NOT     // !
	PIPE       // |=|
	TILDEARROW // ~>
	DOT        // .
	DOTDOT     // ..
	ARROW   // =>

	// Delimiters
	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]
	COMMA    // ,
	COLON    // :
	NEWLINE  // \n

	// Keywords
	keywordsStart
	NYAN     // nyan (let)
	MEOW     // meow (func)
	BRING    // bring (return)
	SNIFF    // sniff (if)
	SCRATCH  // scratch (else)
	PURR     // purr (while)
	PAW      // paw (lambda)
	NYA      // nya (print)
	LICK     // lick (map)
	PICKY    // picky (filter)
	CURL     // curl (reduce)
	PEEK     // peek (match)
	HISS     // hiss (error/throw)
	FETCH    // fetch (import)
	FLAUNT   // flaunt (export)
	CATNAP   // catnap (nil)
	YARN     // yarn (true)
	HAIRBALL // hairball (false)
	KITTY    // kitty (struct)
	BREED    // breed (type alias)
	COLLAR   // collar (newtype)
	TRICK    // trick (interface)
	LEARN    // learn (method impl)
	SELF     // self (self reference)

	// Type keywords
	TYPE_INT     // int
	TYPE_FLOAT   // float
	TYPE_STRING  // string
	TYPE_BOOL    // bool
	TYPE_FURBALL // furball
	TYPE_LIST    // list
	keywordsEnd
)

var keywords = map[string]TokenType{
	"nyan":     NYAN,
	"meow":     MEOW,
	"bring":    BRING,
	"sniff":    SNIFF,
	"scratch":  SCRATCH,
	"purr":     PURR,
	"paw":      PAW,
	"nya":      NYA,
	"lick":     LICK,
	"picky":    PICKY,
	"curl":     CURL,
	"peek":     PEEK,
	"hiss":     HISS,
	"fetch":    FETCH,
	"flaunt":   FLAUNT,
	"catnap":   CATNAP,
	"yarn":     YARN,
	"hairball": HAIRBALL,
	"kitty":    KITTY,
	"breed":    BREED,
	"collar":   COLLAR,
	"trick":    TRICK,
	"learn":    LEARN,
	"self":     SELF,
	"int":      TYPE_INT,
	"float":    TYPE_FLOAT,
	"string":   TYPE_STRING,
	"bool":     TYPE_BOOL,
	"furball":  TYPE_FURBALL,
	"list":     TYPE_LIST,
}

// LookupIdent returns the token type for a given identifier.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// IsKeyword reports whether the token type is a keyword.
func (t TokenType) IsKeyword() bool {
	return t > keywordsStart && t < keywordsEnd
}

// Position represents a source location.
type Position struct {
	// File is the source file name.
	File string
	// Line is the 1-based line number.
	Line int
	// Column is the 1-based column number.
	Column int
}

func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// AsToken creates a zero-value Token with this position.
func (p Position) AsToken() Token {
	return Token{Pos: p}
}

// Token represents a lexical token with its position.
type Token struct {
	// Type is the token type.
	Type TokenType
	// Literal is the raw text of the token.
	Literal string
	// Pos is the source location of the token.
	Pos Position
	// BlockComment is true when Type is COMMENT and the comment uses -~ ~- delimiters.
	BlockComment bool
}
