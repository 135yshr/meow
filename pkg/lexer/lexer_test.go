package lexer_test

import (
	"testing"

	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/token"
)

func collect(l *lexer.Lexer) []token.Token {
	var tokens []token.Token
	for tok := range l.Tokens() {
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

func TestKeywords(t *testing.T) {
	input := `nyan meow bring sniff scratch purr paw nya lick picky curl peek hiss fetch flaunt catnap yarn hairball`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	expected := []token.TokenType{
		token.NYAN, token.MEOW, token.BRING, token.SNIFF, token.SCRATCH,
		token.PURR, token.PAW, token.NYA, token.LICK, token.PICKY,
		token.CURL, token.PEEK, token.HISS, token.FETCH, token.FLAUNT,
		token.CATNAP, token.YARN, token.HAIRBALL, token.EOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: expected %v, got %v (%q)", i, expected[i], tok.Type, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `+ - * / % = == != < > <= >= && || ! |=| ~> .. =>`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.PLUS, "+"}, {token.MINUS, "-"}, {token.STAR, "*"},
		{token.SLASH, "/"}, {token.PERCENT, "%"}, {token.ASSIGN, "="},
		{token.EQ, "=="}, {token.NEQ, "!="}, {token.LT, "<"},
		{token.GT, ">"}, {token.LTE, "<="}, {token.GTE, ">="},
		{token.AND, "&&"}, {token.OR, "||"}, {token.NOT, "!"},
		{token.PIPE, "|=|"}, {token.TILDEARROW, "~>"},
		{token.DOTDOT, ".."}, {token.ARROW, "=>"}, {token.EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tt := range expected {
		if tokens[i].Type != tt.typ || tokens[i].Literal != tt.lit {
			t.Errorf("token[%d]: expected (%v, %q), got (%v, %q)", i, tt.typ, tt.lit, tokens[i].Type, tokens[i].Literal)
		}
	}
}

func TestLiterals(t *testing.T) {
	input := `42 3.14 "hello world" myVar _under`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	tests := []struct {
		typ token.TokenType
		lit string
	}{
		{token.INT, "42"},
		{token.FLOAT, "3.14"},
		{token.STRING, "hello world"},
		{token.IDENT, "myVar"},
		{token.IDENT, "_under"},
		{token.EOF, ""},
	}
	if len(tokens) != len(tests) {
		t.Fatalf("expected %d tokens, got %d", len(tests), len(tokens))
	}
	for i, tt := range tests {
		if tokens[i].Type != tt.typ || tokens[i].Literal != tt.lit {
			t.Errorf("token[%d]: expected (%v, %q), got (%v, %q)", i, tt.typ, tt.lit, tokens[i].Type, tokens[i].Literal)
		}
	}
}

func TestHelloWorld(t *testing.T) {
	input := `nyan name = "Tama"
nya(name)`
	l := lexer.New(input, "hello.nyan")
	tokens := collect(l)
	expected := []token.TokenType{
		token.NYAN, token.IDENT, token.ASSIGN, token.STRING, token.NEWLINE,
		token.NYA, token.LPAREN, token.IDENT, token.RPAREN, token.EOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: expected %v, got %v (%q)", i, expected[i], tok.Type, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `# line comment
nyan x = 1
-~ block
comment ~-
nyan y = 2`
	l := lexer.New(input, "test.nyan")
	var nonComment []token.Token
	for tok := range l.Tokens() {
		if tok.Type != token.COMMENT {
			nonComment = append(nonComment, tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}
	expected := []token.TokenType{
		token.NEWLINE, token.NYAN, token.IDENT, token.ASSIGN, token.INT, token.NEWLINE,
		token.NEWLINE, token.NYAN, token.IDENT, token.ASSIGN, token.INT, token.EOF,
	}
	if len(nonComment) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(nonComment))
	}
	for i, tok := range nonComment {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: expected %v, got %v", i, expected[i], tok.Type)
		}
	}
}
