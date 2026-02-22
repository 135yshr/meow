package parser

import (
	"fmt"

	"github.com/135yshr/meow/pkg/token"
)

// ParseError represents a parser error with a cat-themed message.
type ParseError struct {
	Pos     token.Position
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: Hiss! %s, nya~", e.Pos, e.Message)
}

func newError(pos token.Position, format string, args ...any) *ParseError {
	return &ParseError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	}
}
