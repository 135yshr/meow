package compiler

import (
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/mutation"
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
	modRoot := c.findModuleRoot()
	goVersion := readGoVersion(filepath.Join(modRoot, "go.mod"))
	modContent := fmt.Sprintf("module meow_build\n\ngo %s\n\nrequire github.com/135yshr/meow v0.0.0\n\nreplace github.com/135yshr/meow => %s\n", goVersion, modRoot)
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

// CompileTestToGo compiles a .nyan file to Go source in test mode.
func (c *Compiler) CompileTestToGo(source, filename string) (string, error) {
	// First pass: extract catwalk output expectations from comments.
	c.logger.Debug("extracting catwalk outputs", "file", filename)
	l1 := lexer.New(source, filename)
	catwalkOutputs := codegen.ExtractCatwalkOutputs(l1.Tokens())

	// Second pass: normal lex + parse.
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

	c.logger.Debug("generating test Go code", "file", filename)
	gen := codegen.NewTest()
	if len(catwalkOutputs) > 0 {
		gen.SetCatwalkOutput(catwalkOutputs)
	}
	raw, err := gen.GenerateTest(prog)
	if err != nil {
		return "", err
	}
	formatted, err := format.Source([]byte(raw))
	if err != nil {
		return raw, nil
	}
	return string(formatted), nil
}

// BuildTest compiles a _test.nyan file to an executable binary.
func (c *Compiler) BuildTest(nyanPath, outputPath string) error {
	source, err := os.ReadFile(nyanPath)
	if err != nil {
		return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", nyanPath, err)
	}

	goCode, err := c.CompileTestToGo(string(source), filepath.Base(nyanPath))
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "meow-test-build-*")
	if err != nil {
		return fmt.Errorf("Hiss! Cannot create temp dir, nya~: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write Go source, nya~: %w", err)
	}

	modRoot := c.findModuleRoot()
	goVersion := readGoVersion(filepath.Join(modRoot, "go.mod"))
	modContent := fmt.Sprintf("module meow_build\n\ngo %s\n\nrequire github.com/135yshr/meow v0.0.0\n\nreplace github.com/135yshr/meow => %s\n", goVersion, modRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

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

	c.logger.Debug("building test", "output", absOutput)
	cmd := exec.Command("go", "build", "-o", absOutput, ".")
	cmd.Dir = tmpDir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Hiss! go build failed, nya~: %w", err)
	}

	return nil
}

// CompileFuzzToGo compiles a .nyan file to fuzz test Go source.
// Returns helper code and fuzz test code separately.
func (c *Compiler) CompileFuzzToGo(source, filename string) (helpers, fuzzTests string, err error) {
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
		return "", "", fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	c.logger.Debug("generating fuzz Go code", "file", filename)
	gen := codegen.New()
	helpers, fuzzTests, err = gen.GenerateFuzz(prog)
	if err != nil {
		return "", "", err
	}

	if formatted, fmtErr := format.Source([]byte(helpers)); fmtErr == nil {
		helpers = string(formatted)
	}
	if formatted, fmtErr := format.Source([]byte(fuzzTests)); fmtErr == nil {
		fuzzTests = string(formatted)
	}
	return helpers, fuzzTests, nil
}

// RunFuzz compiles a .nyan file and runs Go fuzz testing.
func (c *Compiler) RunFuzz(nyanPath, fuzzTime string) error {
	source, err := os.ReadFile(nyanPath)
	if err != nil {
		return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", nyanPath, err)
	}

	helpers, fuzzTests, err := c.CompileFuzzToGo(string(source), filepath.Base(nyanPath))
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "meow-fuzz-*")
	if err != nil {
		return fmt.Errorf("Hiss! Cannot create temp dir, nya~: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(helpers), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write main.go, nya~: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(fuzzTests), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write main_test.go, nya~: %w", err)
	}

	modRoot := c.findModuleRoot()
	goVersion := readGoVersion(filepath.Join(modRoot, "go.mod"))
	modContent := fmt.Sprintf("module meow_build\n\ngo %s\n\nrequire github.com/135yshr/meow v0.0.0\n\nreplace github.com/135yshr/meow => %s\n", goVersion, modRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("Hiss! go mod tidy failed, nya~: %w", err)
	}

	if fuzzTime == "" {
		fuzzTime = "10s"
	}

	c.logger.Debug("running fuzz", "fuzztime", fuzzTime)
	cmd := exec.Command("go", "test", "-fuzz=.", fmt.Sprintf("-fuzztime=%s", fuzzTime))
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunTest compiles and runs a _test.nyan file.
func (c *Compiler) RunTest(nyanPath string) error {
	tmpBin, err := os.CreateTemp("", "meow-test-run-*")
	if err != nil {
		return err
	}
	tmpBin.Close()
	defer os.Remove(tmpBin.Name())

	if err := c.BuildTest(nyanPath, tmpBin.Name()); err != nil {
		return err
	}

	cmd := exec.Command(tmpBin.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunMutationTest runs mutation testing on a source file using the given test files.
func (c *Compiler) RunMutationTest(sourcePath string, testPaths []string) error {
	// Read and parse the source file
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", sourcePath, err)
	}

	l := lexer.New(string(source), filepath.Base(sourcePath))
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		return fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	// Enumerate mutations
	mutants := mutation.Enumerate(prog)
	if len(mutants) == 0 {
		fmt.Println("No mutations found, nya~")
		return nil
	}
	fmt.Printf("Found %d mutations, nya~\n", len(mutants))

	// Parse test files and combine ASTs (source AST nodes are shared so mutant closures remain valid)
	combinedProg := &ast.Program{Stmts: append([]ast.Stmt{}, prog.Stmts...)}
	for _, tp := range testPaths {
		data, err := os.ReadFile(tp)
		if err != nil {
			return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", tp, err)
		}
		tl := lexer.New(string(data), filepath.Base(tp))
		tparser := parser.New(tl.Tokens())
		testProg, testErrs := tparser.Parse()
		if len(testErrs) > 0 {
			var msgs []string
			for _, e := range testErrs {
				msgs = append(msgs, e.Error())
			}
			return fmt.Errorf("%s", strings.Join(msgs, "\n"))
		}
		combinedProg.Stmts = append(combinedProg.Stmts, testProg.Stmts...)
	}

	// Build schema using source-only mutants to avoid mutating test code
	schema := mutation.BuildSchema(combinedProg, mutants)

	// Generate mutated test binary
	gen := codegen.NewTest()
	gen.SetMutations(schema)
	raw, err := gen.GenerateTest(combinedProg)
	if err != nil {
		return err
	}

	formatted, fmtErr := format.Source([]byte(raw))
	if fmtErr != nil {
		formatted = []byte(raw)
	}

	// Build the mutated binary
	tmpDir, err := os.MkdirTemp("", "meow-mutate-*")
	if err != nil {
		return fmt.Errorf("Hiss! Cannot create temp dir, nya~: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), formatted, 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write Go source, nya~: %w", err)
	}

	modRoot := c.findModuleRoot()
	goVersion := readGoVersion(filepath.Join(modRoot, "go.mod"))
	modContent := fmt.Sprintf("module meow_build\n\ngo %s\n\nrequire github.com/135yshr/meow v0.0.0\n\nreplace github.com/135yshr/meow => %s\n", goVersion, modRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("Hiss! go mod tidy failed, nya~: %w", err)
	}

	binPath := filepath.Join(tmpDir, "mutant_test")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = tmpDir
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("Hiss! go build failed, nya~: %w", err)
	}

	// Run mutation tests
	runner := mutation.NewRunner(binPath, 10*time.Second)
	results := runner.RunAll(mutants)

	// Report
	mutation.Report(os.Stdout, mutants, results)
	return nil
}

// readGoVersion parses a go.mod file and returns the Go version directive.
// Falls back to "1.26" if the file cannot be read or parsed.
func readGoVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "1.26"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go ")
		}
	}
	return "1.26"
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
