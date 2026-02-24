package codegen

import (
	"iter"
	"strings"

	"github.com/135yshr/meow/pkg/token"
)

// CatwalkOutput maps catwalk function names to their expected output strings.
type CatwalkOutput map[string]string

// ExtractCatwalkOutputs scans a token stream for catwalk_ functions and
// extracts their # Output: blocks. The token stream is consumed into a slice
// so it can be scanned with random access.
func ExtractCatwalkOutputs(tokens iter.Seq[token.Token]) CatwalkOutput {
	// Collect tokens into a slice.
	var toks []token.Token
	for t := range tokens {
		toks = append(toks, t)
	}

	result := make(CatwalkOutput)

	for i := 0; i < len(toks)-1; i++ {
		// Look for MEOW followed by IDENT starting with "catwalk_".
		if toks[i].Type != token.MEOW {
			continue
		}
		j := i + 1
		// Skip newlines between MEOW and IDENT.
		for j < len(toks) && toks[j].Type == token.NEWLINE {
			j++
		}
		if j >= len(toks) || toks[j].Type != token.IDENT {
			continue
		}
		name := toks[j].Literal
		if !strings.HasPrefix(name, "catwalk_") {
			continue
		}

		// Find the matching LBRACE.
		braceStart := -1
		for k := j + 1; k < len(toks); k++ {
			if toks[k].Type == token.LBRACE {
				braceStart = k
				break
			}
		}
		if braceStart < 0 {
			continue
		}

		// Track braces to find the matching RBRACE.
		depth := 1
		braceEnd := -1
		for k := braceStart + 1; k < len(toks); k++ {
			switch toks[k].Type {
			case token.LBRACE:
				depth++
			case token.RBRACE:
				depth--
				if depth == 0 {
					braceEnd = k
				}
			}
			if braceEnd >= 0 {
				break
			}
		}
		if braceEnd < 0 {
			continue
		}

		// Scan backwards from RBRACE to find the Output block.
		// Walk back over NEWLINE and COMMENT tokens to find "# Output:".
		outputIdx := -1
		for k := braceEnd - 1; k > braceStart; k-- {
			switch toks[k].Type {
			case token.NEWLINE:
				continue
			case token.COMMENT:
				trimmed := strings.TrimSpace(toks[k].Literal)
				if trimmed == "# Output:" {
					outputIdx = k
				}
				continue
			default:
				// Hit a non-comment/newline token; stop scanning.
				goto doneScanning
			}
		}
	doneScanning:

		if outputIdx < 0 {
			continue
		}

		// Collect output lines: comments after "# Output:" up to RBRACE.
		var lines []string
		for k := outputIdx + 1; k < braceEnd; k++ {
			if toks[k].Type == token.NEWLINE {
				continue
			}
			if toks[k].Type != token.COMMENT {
				break
			}
			line := toks[k].Literal
			// Strip "# " prefix.
			line = strings.TrimPrefix(line, "# ")
			lines = append(lines, line)
		}

		if len(lines) > 0 {
			result[name] = strings.Join(lines, "\n") + "\n"
		} else {
			result[name] = "\n"
		}
	}

	return result
}
