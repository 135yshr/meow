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
	afterBrace := false                // suppress newlines right after {
	afterUnaryMinus := false           // suppress space after unary minus
	inlineBlock := false               // inside an inline lambda body
	firstToken := true
	// One entry per open brace, true where it opened a basket literal rather
	// than a block, so its closing brace is treated the same way.
	var literalBrace []bool

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
			if n := len(literalBrace); n > 0 && literalBrace[n-1] {
				// Closing a basket literal. It ends the line it sits on, or
				// takes the indent of its own line where the source spread the
				// basket out — either way it forces neither shape.
				literalBrace = literalBrace[:n-1]
				if indent > 0 {
					indent--
				}
				if lineStart {
					writeIndent()
				}
				buf.WriteByte('}')
				lineStart = false
				firstToken = false
				prevMeaningful = tok.Type
				continue
			}
			if len(literalBrace) > 0 {
				literalBrace = literalBrace[:len(literalBrace)-1]
			}
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
			} else if needsSpaceBefore(toks, i, prevMeaningful) {
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
			switch {
			case opensABasket(prevMeaningful):
				// A basket keeps whatever shape it was written in: no newline
				// is forced after the brace, but the indent is there for one
				// the source put in itself.
				literalBrace = append(literalBrace, true)
				indent++
			case isLambdaBrace(toks, i) && canInlineBlock(toks, i):
				literalBrace = append(literalBrace, false)
				inlineBlock = true
			default:
				literalBrace = append(literalBrace, false)
				writeNewline()
				indent++
				afterBrace = true
			}
		}

		// Track unary minus: MINUS is unary when previous non-trivia token
		// is not an expression-completing token.
		prevForUnary := prevMeaningful
		if prevForUnary == token.COMMENT {
			prevForUnary = previousNonTriviaType(toks, i-1)
		}
		if tok.Type == token.MINUS && !isExpressionEnd(prevForUnary) {
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

func previousNonTriviaType(toks []token.Token, idx int) token.TokenType {
	for i := idx; i >= 0; i-- {
		if toks[i].Type != token.NEWLINE && toks[i].Type != token.COMMENT {
			return toks[i].Type
		}
	}
	return token.EOF
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
		token.ARROW:
		return true
	}
	return false
}

// opensABasket reports whether a brace following prev opens a basket literal
// rather than a block.
//
// A block's brace closes a header — `meow f() {`, `sniff (c) {`, `kitty Cat {`
// — so it follows a `)`, a name or a type. A basket's brace stands where a
// value does: after `=`, an open bracket, a comma, a colon, a `bring`, an
// operator. Treated alike, `{"body": "hi"}` was broken open across three lines
// as though it were a body, and `{}` across two.
func opensABasket(prev token.TokenType) bool {
	switch prev {
	case token.LPAREN, token.LBRACKET, token.LBRACE,
		token.COMMA, token.COLON, token.BRING, token.NOT:
		return true
	}
	return isBinaryOp(prev)
}

// opensALoopSubject reports whether the LPAREN at toks[idx] is the one holding
// what a purr walks — `purr i (5)`, `purr i, x (xs)`.
//
// It reads as part of the loop's opening rather than as a call, so it keeps
// the space the whole language is written with. Judged by the token just
// before it, that token is the loop variable and the space was dropped, which
// turned every `purr i (10)` in the repository into something that looked like
// a call to `i`.
func opensALoopSubject(toks []token.Token, idx int) bool {
	// Back over the loop variable, and the index variable if there is one.
	for i := idx - 1; i >= 0; i-- {
		switch toks[i].Type {
		case token.IDENT, token.COMMA:
			continue
		case token.PURR:
			return true
		default:
			return false
		}
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

// needsSpaceBefore reports whether toks[idx] wants a space in front of it,
// given the meaningful token before it.
//
// A few rules need to see further back than one token, so the whole stream is
// passed rather than the two types alone.
func needsSpaceBefore(toks []token.Token, idx int, prev token.TokenType) bool {
	cur := toks[idx].Type
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
	// DOTDOT: a range is one thing, written `1..5`. Spaced out as an operator it
	// came back as `1 .. 5`, which is in no document and in no source file here.
	if cur == token.DOTDOT || prev == token.DOTDOT {
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
	// A basket literal packs its contents against its braces — `{"a": 1}`, the
	// way every source file here writes one. A block's brace is followed by a
	// newline, so this only ever decides the literal case.
	if prev == token.LBRACE {
		brace := idx - 1
		for brace >= 0 && (toks[brace].Type == token.NEWLINE || toks[brace].Type == token.COMMENT) {
			brace--
		}
		return !opensABasket(previousNonTriviaType(toks, brace-1))
	}
	// LPAREN: part of an opening rather than a call — `sniff (c)`, `purr (c)`,
	// and the `purr i (5)` whose loop variable sits between the two.
	if cur == token.LPAREN {
		return isBlockKeyword(prev) || opensALoopSubject(toks, idx)
	}
	// LBRACKET: an index reaches back into whatever it follows, so it sits
	// tight against it — `resp["body"]`, not `resp ["body"]`. Opening a litter
	// it is a value like any other and takes the spacing of what came before.
	if cur == token.LBRACKET {
		return !isExpressionEnd(prev)
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
