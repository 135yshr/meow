// Package compiler orchestrates the Meow compilation pipeline:
// lexing, parsing, code generation, and invoking go build to produce a binary.
package compiler

import (
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

// Compiler orchestrates the compilation pipeline.
type Compiler struct {
	logger *slog.Logger
}

// New creates a new Compiler.
func New(logger *slog.Logger) *Compiler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Compiler{logger: logger}
}

// CompileToGo compiles a .nyan file to Go source code.
func (c *Compiler) CompileToGo(source, filename string) (string, error) {
	c.logger.Debug("lexing", "file", filename)
	l := lexer.New(source, filename)

	c.logger.Debug("parsing", "file", filename)
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		return "", fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	c.logger.Debug("generating Go code", "file", filename)
	gen := codegen.New()
	raw, err := gen.Generate(prog)
	if err != nil {
		return "", err
	}
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		// If formatting fails, return raw code for debugging
		return raw, nil
	}
	return string(formatted), nil
}

// Build compiles a .nyan file to an executable binary.
func (c *Compiler) Build(nyanPath, outputPath string) error {
	source, err := os.ReadFile(nyanPath)
	if err != nil {
		return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", nyanPath, err)
	}

	goCode, err := c.CompileToGo(string(source), filepath.Base(nyanPath))
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "meow-build-*")
	if err != nil {
		return fmt.Errorf("Hiss! Cannot create temp dir, nya~: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write Go source, nya~: %w", err)
	}

	// Create go.mod in temp dir
	modContent := fmt.Sprintf("module meow_build\n\ngo 1.26\n\nrequire github.com/135yshr/meow v0.0.0\n\nreplace github.com/135yshr/meow => %s\n", c.findModuleRoot())
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	// Run go mod tidy to generate go.sum
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("Hiss! go mod tidy failed, nya~: %w", err)
	}

	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(nyanPath), ".nyan")
		outputPath = base
	}

	absOutput, _ := filepath.Abs(outputPath)

	c.logger.Debug("building", "output", absOutput)
	cmd := exec.Command("go", "build", "-o", absOutput, ".")
	cmd.Dir = tmpDir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Hiss! go build failed, nya~: %w", err)
	}

	return nil
}

// Run compiles and runs a .nyan file.
func (c *Compiler) Run(nyanPath string) error {
	tmpBin, err := os.CreateTemp("", "meow-run-*")
	if err != nil {
		return err
	}
	tmpBin.Close()
	defer os.Remove(tmpBin.Name())

	if err := c.Build(nyanPath, tmpBin.Name()); err != nil {
		return err
	}

	cmd := exec.Command(tmpBin.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *Compiler) findModuleRoot() string {
	// Walk up from executable to find go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// fallback
	return "."
}
