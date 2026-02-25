package linter

import (
	"fmt"
	"sort"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
)

// Severity represents the severity of a diagnostic.
type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Error {
		return "error"
	}
	return "warning"
}

// Diagnostic represents a single lint finding.
type Diagnostic struct {
	Pos      token.Position
	Severity Severity
	Rule     string
	Message  string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s: %s[%s]: %s", d.Pos, d.Severity, d.Rule, d.Message)
}

// Rule is the interface that all lint rules implement.
type Rule interface {
	Name() string
	Check(prog *ast.Program, report func(Diagnostic))
}

// Linter runs a set of rules on a program.
type Linter struct {
	rules []Rule
}

// New creates a Linter with all built-in rules.
func New() *Linter {
	return &Linter{
		rules: []Rule{
			&SnakeCaseRule{},
			&UnusedVarRule{},
			&UnreachableCodeRule{},
			&EmptyBlockRule{},
		},
	}
}

// Lint runs all rules and returns sorted diagnostics.
func (l *Linter) Lint(prog *ast.Program) []Diagnostic {
	var diags []Diagnostic
	report := func(d Diagnostic) {
		diags = append(diags, d)
	}
	for _, rule := range l.rules {
		rule.Check(prog, report)
	}
	sort.Slice(diags, func(i, j int) bool {
		if diags[i].Pos.Line != diags[j].Pos.Line {
			return diags[i].Pos.Line < diags[j].Pos.Line
		}
		return diags[i].Pos.Column < diags[j].Pos.Column
	})
	return diags
}
