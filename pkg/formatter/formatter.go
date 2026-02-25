package formatter

import (
	"strings"

	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/token"
)

// Config holds formatter settings.
type Config struct {
	IndentWidth   int
	MaxBlankLines int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		IndentWidth:   2,
		MaxBlankLines: 1,
	}
}

// FormatSource formats Meow source code.
func FormatSource(source, filename string) string {
	l := lexer.New(source, filename)
	return Format(l.Tokens(), DefaultConfig())
}

// Format formats a token stream into normalized source.
func Format(tokens func(func(token.Token) bool), cfg Config) string {
	// Collect all tokens first
	var toks []token.Token
	for tok := range tokens {
		toks = append(toks, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	var buf strings.Builder
	indent := 0
	blankCount := 0
	lineStart := true
	var prevMeaningful token.TokenType // prev token type ignoring NEWLINE
	afterBrace := false               // suppress newlines right after {
	firstToken := true

	writeIndent := func() {
		for range indent * cfg.IndentWidth {
			buf.WriteByte(' ')
		}
	}

	writeNewline := func() {
		buf.WriteByte('\n')
		lineStart = true
	}

	// Look ahead to find next non-NEWLINE, non-COMMENT token type
	nextMeaningful := func(from int) token.TokenType {
		for i := from; i < len(toks); i++ {
			if toks[i].Type != token.NEWLINE && toks[i].Type != token.COMMENT {
				return toks[i].Type
			}
		}
		return token.EOF
	}

	for i, tok := range toks {
		if tok.Type == token.EOF {
			break
		}

		switch tok.Type {
		case token.NEWLINE:
			if firstToken || afterBrace {
				continue
			}
			// Skip newlines between } and scratch
			if prevMeaningful == token.RBRACE && nextMeaningful(i+1) == token.SCRATCH {
				continue
			}
			if lineStart {
				blankCount++
				if blankCount > cfg.MaxBlankLines {
					continue
				}
			} else {
				blankCount = 0
			}
			writeNewline()
			continue

		case token.COMMENT:
			afterBrace = false
			if tok.Literal != "" && tok.Literal[0] == '#' {
				if lineStart {
					writeIndent()
				} else {
					buf.WriteByte(' ')
				}
				buf.WriteString(tok.Literal)
			} else {
				// The lexer strips block-comment delimiters (-~ ... ~-) and stores
				// only the inner content in tok.Literal, so we re-wrap here.
				if lineStart {
					writeIndent()
				} else {
					buf.WriteByte(' ')
				}
				buf.WriteString("-~")
				buf.WriteString(tok.Literal)
				buf.WriteString("~-")
			}
			lineStart = false
			blankCount = 0
			firstToken = false
			prevMeaningful = tok.Type
			continue
		}

		afterBrace = false

		// Handle RBRACE: decrease indent before writing
		if tok.Type == token.RBRACE {
			if indent > 0 {
				indent--
			}
			if !lineStart {
				writeNewline()
			}
			blankCount = 0
			writeIndent()
			buf.WriteByte('}')
			lineStart = false
			firstToken = false
			prevMeaningful = tok.Type
			continue
		}

		// Handle "} scratch {" pattern: scratch after RBRACE stays on the same line
		if tok.Type == token.SCRATCH && prevMeaningful == token.RBRACE {
			buf.WriteString(" scratch")
			lineStart = false
			blankCount = 0
			firstToken = false
			prevMeaningful = tok.Type
			continue
		}

		// Start of a new logical line: write indent
		if lineStart {
			blankCount = 0
			writeIndent()
			lineStart = false
		} else {
			if needsSpaceBefore(tok.Type, prevMeaningful) {
				buf.WriteByte(' ')
			}
		}

		// Write the token literal
		switch tok.Type {
		case token.STRING:
			buf.WriteByte('"')
			buf.WriteString(tok.Literal)
			buf.WriteByte('"')
		default:
			buf.WriteString(tok.Literal)
		}

		// Handle LBRACE: increase indent after writing
		if tok.Type == token.LBRACE {
			writeNewline()
			indent++
			afterBrace = true
		}

		firstToken = false
		prevMeaningful = tok.Type
	}

	result := buf.String()
	result = strings.TrimRight(result, "\n")
	if result != "" {
		result += "\n"
	}
	return result
}

func isBinaryOp(t token.TokenType) bool {
	switch t {
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
		token.ASSIGN, token.EQ, token.NEQ,
		token.LT, token.GT, token.LTE, token.GTE,
		token.AND, token.OR,
		token.PIPE, token.TILDEARROW,
		token.DOTDOT, token.ARROW:
		return true
	}
	return false
}

// isBlockKeyword returns true for keywords that take a paren-delimited condition/params
// and where a space before ( is desired.
func isBlockKeyword(t token.TokenType) bool {
	switch t {
	case token.SNIFF, token.PURR:
		return true
	}
	return false
}

func needsSpaceBefore(cur, prev token.TokenType) bool {
	// Never space after open delimiters
	if prev == token.LPAREN || prev == token.LBRACKET {
		return false
	}
	// Never space before close delimiters
	if cur == token.RPAREN || cur == token.RBRACKET {
		return false
	}
	// Never space before comma
	if cur == token.COMMA {
		return false
	}
	// Space after comma
	if prev == token.COMMA {
		return true
	}
	// Space after colon
	if prev == token.COLON {
		return true
	}
	// Never space before colon
	if cur == token.COLON {
		return false
	}
	// DOT: no space before or after
	if cur == token.DOT || prev == token.DOT {
		return false
	}
	// Space around binary operators
	if isBinaryOp(cur) || isBinaryOp(prev) {
		return true
	}
	// Space before LBRACE
	if cur == token.LBRACE {
		return true
	}
	// LPAREN: space only after block keywords (sniff, purr)
	if cur == token.LPAREN {
		if isBlockKeyword(prev) {
			return true
		}
		return false
	}
	// NOT operator: no space after
	if prev == token.NOT {
		return false
	}
	// Space after keywords (before identifiers, literals, etc.)
	if prev.IsKeyword() {
		return true
	}
	// Default: space between tokens
	return true
}
