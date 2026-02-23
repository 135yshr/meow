package lexer

import (
	"iter"
	"unicode"
	"unicode/utf8"

	"github.com/135yshr/meow/pkg/token"
)

// Lexer tokenizes Meow source code.
type Lexer struct {
	input string
	file  string
	pos   int
	line  int
	col   int
}

// New creates a new Lexer for the given source.
func New(input, file string) *Lexer {
	return &Lexer{
		input: input,
		file:  file,
		pos:   0,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *Lexer) peekAt(offset int) rune {
	p := l.pos + offset
	if p >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[p:])
	return r
}

func (l *Lexer) makeToken(typ token.TokenType, lit string, pos token.Position) token.Token {
	return token.Token{Type: typ, Literal: lit, Pos: pos}
}

func (l *Lexer) currentPos() token.Position {
	return token.Position{File: l.file, Line: l.line, Column: l.col}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) readString() token.Token {
	pos := l.currentPos()
	l.advance() // skip opening quote
	start := l.pos
	for l.pos < len(l.input) {
		r := l.peek()
		if r == '"' {
			lit := l.input[start:l.pos]
			l.advance() // skip closing quote
			return l.makeToken(token.STRING, lit, pos)
		}
		if r == '\\' {
			l.advance() // skip backslash
		}
		l.advance()
	}
	return l.makeToken(token.ILLEGAL, l.input[start:l.pos], pos)
}

func (l *Lexer) readNumber() token.Token {
	pos := l.currentPos()
	start := l.pos
	isFloat := false
	for l.pos < len(l.input) {
		r := l.peek()
		if r == '.' && l.peekAt(1) != '.' {
			if isFloat {
				break
			}
			isFloat = true
			l.advance()
			continue
		}
		if !unicode.IsDigit(r) {
			break
		}
		l.advance()
	}
	lit := l.input[start:l.pos]
	if isFloat {
		return l.makeToken(token.FLOAT, lit, pos)
	}
	return l.makeToken(token.INT, lit, pos)
}

func (l *Lexer) readIdent() token.Token {
	pos := l.currentPos()
	start := l.pos
	for l.pos < len(l.input) {
		r := l.peek()
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		l.advance()
	}
	lit := l.input[start:l.pos]
	return l.makeToken(token.LookupIdent(lit), lit, pos)
}

func (l *Lexer) readLineComment() token.Token {
	pos := l.currentPos()
	start := l.pos
	for l.pos < len(l.input) && l.peek() != '\n' {
		l.advance()
	}
	return l.makeToken(token.COMMENT, l.input[start:l.pos], pos)
}

func (l *Lexer) readBlockComment() token.Token {
	pos := l.currentPos()
	l.advance() // skip ~
	start := l.pos
	for l.pos < len(l.input) {
		if l.peek() == '~' && l.peekAt(1) == '-' {
			lit := l.input[start:l.pos]
			l.advance() // ~
			l.advance() // -
			return l.makeToken(token.COMMENT, lit, pos)
		}
		l.advance()
	}
	return l.makeToken(token.ILLEGAL, l.input[start:l.pos], pos)
}

// Tokens returns an iterator over all tokens in the source.
func (l *Lexer) Tokens() iter.Seq[token.Token] {
	return func(yield func(token.Token) bool) {
		for {
			l.skipWhitespace()
			if l.pos >= len(l.input) {
				yield(l.makeToken(token.EOF, "", l.currentPos()))
				return
			}
			pos := l.currentPos()
			r := l.peek()

			switch {
			case r == '\n':
				l.advance()
				if !yield(l.makeToken(token.NEWLINE, "\n", pos)) {
					return
				}
			case r == '"':
				if !yield(l.readString()) {
					return
				}
			case unicode.IsDigit(r):
				if !yield(l.readNumber()) {
					return
				}
			case unicode.IsLetter(r) || r == '_':
				if !yield(l.readIdent()) {
					return
				}
			case r == '+':
				l.advance()
				if !yield(l.makeToken(token.PLUS, "+", pos)) {
					return
				}
			case r == '*':
				l.advance()
				if !yield(l.makeToken(token.STAR, "*", pos)) {
					return
				}
			case r == '/':
				l.advance()
				if !yield(l.makeToken(token.SLASH, "/", pos)) {
					return
				}
			case r == '%':
				l.advance()
				if !yield(l.makeToken(token.PERCENT, "%", pos)) {
					return
				}
			case r == '(':
				l.advance()
				if !yield(l.makeToken(token.LPAREN, "(", pos)) {
					return
				}
			case r == ')':
				l.advance()
				if !yield(l.makeToken(token.RPAREN, ")", pos)) {
					return
				}
			case r == '{':
				l.advance()
				if !yield(l.makeToken(token.LBRACE, "{", pos)) {
					return
				}
			case r == '}':
				l.advance()
				if !yield(l.makeToken(token.RBRACE, "}", pos)) {
					return
				}
			case r == '[':
				l.advance()
				if !yield(l.makeToken(token.LBRACKET, "[", pos)) {
					return
				}
			case r == ']':
				l.advance()
				if !yield(l.makeToken(token.RBRACKET, "]", pos)) {
					return
				}
			case r == ',':
				l.advance()
				if !yield(l.makeToken(token.COMMA, ",", pos)) {
					return
				}
			case r == '-':
				l.advance()
				if l.peek() == '~' {
					// block comment -~ ... ~-
					if !yield(l.readBlockComment()) {
						return
					}
				} else {
					if !yield(l.makeToken(token.MINUS, "-", pos)) {
						return
					}
				}
			case r == '=':
				l.advance()
				if l.peek() == '=' {
					l.advance()
					if !yield(l.makeToken(token.EQ, "==", pos)) {
						return
					}
				} else if l.peek() == '>' {
					l.advance()
					if !yield(l.makeToken(token.ARROW, "=>", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.ASSIGN, "=", pos)) {
						return
					}
				}
			case r == '!':
				l.advance()
				if l.peek() == '=' {
					l.advance()
					if !yield(l.makeToken(token.NEQ, "!=", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.NOT, "!", pos)) {
						return
					}
				}
			case r == '<':
				l.advance()
				if l.peek() == '=' {
					l.advance()
					if !yield(l.makeToken(token.LTE, "<=", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.LT, "<", pos)) {
						return
					}
				}
			case r == '>':
				l.advance()
				if l.peek() == '=' {
					l.advance()
					if !yield(l.makeToken(token.GTE, ">=", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.GT, ">", pos)) {
						return
					}
				}
			case r == '&':
				l.advance()
				if l.peek() == '&' {
					l.advance()
					if !yield(l.makeToken(token.AND, "&&", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.ILLEGAL, "&", pos)) {
						return
					}
				}
			case r == '|':
				l.advance()
				if l.peek() == '=' && l.peekAt(1) == '|' {
					l.advance() // =
					l.advance() // |
					if !yield(l.makeToken(token.PIPE, "|=|", pos)) {
						return
					}
				} else if l.peek() == '|' {
					l.advance()
					if !yield(l.makeToken(token.OR, "||", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.ILLEGAL, "|", pos)) {
						return
					}
				}
			case r == '.':
				l.advance()
				if l.peek() == '.' {
					l.advance()
					if !yield(l.makeToken(token.DOTDOT, "..", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.ILLEGAL, ".", pos)) {
						return
					}
				}
			case r == '~':
				l.advance()
				if l.peek() == '>' {
					l.advance()
					if !yield(l.makeToken(token.TILDEARROW, "~>", pos)) {
						return
					}
				} else {
					if !yield(l.makeToken(token.ILLEGAL, "~", pos)) {
						return
					}
				}
			case r == '#':
				// line comment
				if !yield(l.readLineComment()) {
					return
				}
			default:
				l.advance()
				if !yield(l.makeToken(token.ILLEGAL, string(r), pos)) {
					return
				}
			}
		}
	}
}
