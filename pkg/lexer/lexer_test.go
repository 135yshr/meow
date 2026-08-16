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
	input := `nyan meow bring sniff scratch purr paw nya lick picky curl peek hiss nab flaunt catnap yarn hairball`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	expected := []token.TokenType{
		token.NYAN, token.MEOW, token.BRING, token.SNIFF, token.SCRATCH,
		token.PURR, token.PAW, token.NYA, token.LICK, token.PICKY,
		token.CURL, token.PEEK, token.HISS, token.NAB, token.FLAUNT,
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
	input := `nyan name = "Nyantyu"
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

func TestDotToken(t *testing.T) {
	input := `file.snoop`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.IDENT, "file"},
		{token.DOT, "."},
		{token.IDENT, "snoop"},
		{token.EOF, ""},
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

func TestDotDotStillWorks(t *testing.T) {
	input := `1..10`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.INT, "1"},
		{token.DOTDOT, ".."},
		{token.INT, "10"},
		{token.EOF, ""},
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

func TestColonToken(t *testing.T) {
	input := `{"key": 42}`
	l := lexer.New(input, "test.nyan")
	tokens := collect(l)
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.LBRACE, "{"},
		{token.STRING, "key"},
		{token.COLON, ":"},
		{token.INT, "42"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
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

func TestStringEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"double quote", `"say \"hi\""`, `say "hi"`},
		{"backslash", `"a\\b"`, `a\b`},
		{"newline", `"a\nb"`, "a\nb"},
		{"tab", `"a\tb"`, "a\tb"},
		{"carriage return", `"a\rb"`, "a\rb"},
		{"quote ends the literal after an escaped backslash", `"a\\"`, `a\`},
		// An unrecognized escape keeps both characters so patterns and
		// Windows paths survive instead of silently losing the backslash.
		{"unknown escape is kept whole", `"\d+"`, `\d+`},
		{"no escapes", `"plain"`, "plain"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := collect(lexer.New(tt.input, "test.nyan"))
			if tokens[0].Type != token.STRING {
				t.Fatalf("expected STRING, got %v (%q)", tokens[0].Type, tokens[0].Literal)
			}
			if tokens[0].Literal != tt.want {
				t.Errorf("expected %q, got %q", tt.want, tokens[0].Literal)
			}
		})
	}
}

func TestUnterminatedStringIsIllegal(t *testing.T) {
	tokens := collect(lexer.New(`"no closing quote`, "test.nyan"))
	if tokens[0].Type != token.ILLEGAL {
		t.Errorf("expected ILLEGAL, got %v (%q)", tokens[0].Type, tokens[0].Literal)
	}
}

func TestEscapedQuoteDoesNotEndString(t *testing.T) {
	tokens := collect(lexer.New(`"a\"b" nyan`, "test.nyan"))
	if tokens[0].Type != token.STRING || tokens[0].Literal != `a"b` {
		t.Fatalf("expected STRING %q, got %v (%q)", `a"b`, tokens[0].Type, tokens[0].Literal)
	}
	if tokens[1].Type != token.NYAN {
		t.Errorf("expected the literal to end before nyan, got %v", tokens[1].Type)
	}
}
