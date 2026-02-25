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
	if cfg.IndentWidth <= 0 {
		cfg.IndentWidth = DefaultConfig().IndentWidth
	}
	if cfg.MaxBlankLines < 0 {
		cfg.MaxBlankLines = 0
	}

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
	afterUnaryMinus := false           // suppress space after unary minus
	inlineBlock := false               // inside an inline lambda body
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
			if firstToken || afterBrace || inlineBlock {
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
			if lineStart {
				writeIndent()
			} else {
				buf.WriteByte(' ')
			}
			if tok.BlockComment {
				// The lexer strips block-comment delimiters (-~ ... ~-) and stores
				// only the inner content in tok.Literal, so we re-wrap here.
				buf.WriteString("-~")
				buf.WriteString(tok.Literal)
				buf.WriteString("~-")
			} else {
				buf.WriteString(tok.Literal)
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
			if inlineBlock {
				buf.WriteString(" }")
				inlineBlock = false
				lineStart = false
				firstToken = false
				prevMeaningful = tok.Type
				continue
			}
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
			if afterUnaryMinus {
				// No space after unary minus
			} else if needsSpaceBefore(tok.Type, prevMeaningful) {
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
			if isLambdaBrace(toks, i) && canInlineBlock(toks, i) {
				inlineBlock = true
			} else {
				writeNewline()
				indent++
				afterBrace = true
			}
		}

		// Track unary minus: MINUS is unary when previous token is not an
		// expression-completing token (e.g. at line start, after operator, etc.)
		if tok.Type == token.MINUS && !isExpressionEnd(prevMeaningful) {
			afterUnaryMinus = true
		} else {
			afterUnaryMinus = false
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

// isLambdaBrace checks if the LBRACE at toks[idx] belongs to a paw lambda.
func isLambdaBrace(toks []token.Token, idx int) bool {
	// Find previous meaningful token (should be RPAREN)
	j := idx - 1
	for j >= 0 && (toks[j].Type == token.NEWLINE || toks[j].Type == token.COMMENT) {
		j--
	}
	if j < 0 || toks[j].Type != token.RPAREN {
		return false
	}
	// Find matching LPAREN
	depth := 0
	for k := j; k >= 0; k-- {
		switch toks[k].Type {
		case token.RPAREN:
			depth++
		case token.LPAREN:
			depth--
			if depth == 0 {
				// Check if PAW precedes this LPAREN
				m := k - 1
				for m >= 0 && (toks[m].Type == token.NEWLINE || toks[m].Type == token.COMMENT) {
					m--
				}
				return m >= 0 && toks[m].Type == token.PAW
			}
		}
	}
	return false
}

// canInlineBlock checks if the brace block at toks[idx] (LBRACE) has no
// nested braces, no comments, and no newlines, so it can be safely rendered
// on a single line.
func canInlineBlock(toks []token.Token, idx int) bool {
	depth := 0
	for i := idx; i < len(toks); i++ {
		switch toks[i].Type {
		case token.LBRACE:
			depth++
			if depth > 1 {
				return false
			}
		case token.RBRACE:
			depth--
			if depth == 0 {
				return true
			}
		case token.NEWLINE, token.COMMENT:
			if depth == 1 {
				return false
			}
		}
	}
	return false
}

func isExpressionEnd(t token.TokenType) bool {
	switch t {
	case token.IDENT, token.INT, token.FLOAT, token.STRING,
		token.RPAREN, token.RBRACKET, token.RBRACE,
		token.YARN, token.HAIRBALL, token.CATNAP:
		return true
	}
	return false
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
