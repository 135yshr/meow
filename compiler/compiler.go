package compiler

import (
	"errors"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/mutation"
	"github.com/135yshr/meow/pkg/parser"
)

// Compiler orchestrates the compilation pipeline.
type Compiler struct {
	logger       *slog.Logger
	coverEnabled bool
	coverProfile string
	// goPins holds the versions the program pinned its Go imports to, by
	// import path. It is read where the program is, and used where the build's
	// go.mod is written. An import with no pin is left for the toolchain to
	// resolve like any other.
	goPins map[string]string
}

// New creates a new Compiler.
func New(logger *slog.Logger) *Compiler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Compiler{logger: logger}
}

// EnableCoverage activates statement coverage for test runs.
func (c *Compiler) EnableCoverage(profile string) {
	c.coverEnabled = true
	c.coverProfile = profile
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

	c.logger.Debug("type checking", "file", filename)
	ch := checker.New()
	typeInfo, typeErrs := ch.Check(prog)
	if len(typeErrs) > 0 {
		var msgs []string
		for _, e := range typeErrs {
			msgs = append(msgs, e.Error())
		}
		return "", fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	c.recordGoPins(prog)

	c.logger.Debug("generating Go code", "file", filename)
	gen := codegen.New()
	gen.SetTypeInfo(typeInfo)
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

// recordGoPins reads the versions the program pinned its Go imports to.
func (c *Compiler) recordGoPins(prog *ast.Program) {
	pins := make(map[string]string)
	for _, stmt := range prog.Stmts {
		if fs, ok := stmt.(*ast.FetchStmt); ok && fs.Go && fs.Version != "" {
			pins[fs.Path] = fs.Version
		}
	}
	c.goPins = pins
}

// fetchGoPins asks for the exact versions the program named, before the rest
// is left to the toolchain.
//
// It is `go get` rather than a require line written by hand because a pin
// names an import path, and only the toolchain knows which module provides
// one — the module is often a prefix of the path, and sometimes the whole of
// it.
func (c *Compiler) fetchGoPins(dir string) error {
	for path, version := range c.goPins {
		spec := path + "@" + version
		c.logger.Debug("fetching pinned package", "spec", spec)
		cmd := exec.Command("go", "get", spec)
		cmd.Dir = dir
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Hiss! Cannot fetch %s, nya~: %w", spec, err)
		}
	}
	return nil
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
	goVersion := goVersionFor(modRoot)
	modContent, err := buildModContent(goVersion, modRoot)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	if err := c.fetchGoPins(tmpDir); err != nil {
		return err
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

// Run compiles and runs a .nyan file, passing args on to the program.
//
// The arguments are what env.haul reads, so `meow run prog.nyan --verbose` and
// running the built binary the same way agree. The program's exit status comes
// back as an *exec.ExitError, which the caller reports as its own — a program
// that scrams with 3 is no use if the tool that ran it answers 1.
func (c *Compiler) Run(nyanPath string, args ...string) error {
	tmpBin, err := os.CreateTemp("", "meow-run-*")
	if err != nil {
		return err
	}
	tmpBin.Close()
	defer os.Remove(tmpBin.Name())

	if err := c.Build(nyanPath, tmpBin.Name()); err != nil {
		return err
	}

	cmd := exec.Command(tmpBin.Name(), args...)
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

	c.logger.Debug("type checking", "file", filename)
	ch := checker.New()
	typeInfo, typeErrs := ch.Check(prog)
	if len(typeErrs) > 0 {
		var msgs []string
		for _, e := range typeErrs {
			msgs = append(msgs, e.Error())
		}
		return "", fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	c.recordGoPins(prog)

	c.logger.Debug("generating test Go code", "file", filename)
	gen := codegen.NewTest()
	gen.SetTypeInfo(typeInfo)
	if c.coverEnabled {
		gen.EnableCoverage(filename)
	}
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
// If a companion source file exists (e.g. math.nyan for math_test.nyan),
// it is automatically prepended so the test can call its functions.
func (c *Compiler) BuildTest(nyanPath, outputPath string) error {
	source, err := os.ReadFile(nyanPath)
	if err != nil {
		return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", nyanPath, err)
	}

	combined := string(source)
	if companionPath := companionSourcePath(nyanPath); companionPath != "" {
		companionData, err := os.ReadFile(companionPath)
		if err == nil {
			c.logger.Debug("including companion source", "file", companionPath)
			combined = string(companionData) + "\n" + combined
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("Hiss! Cannot read companion %s, nya~: %w", companionPath, err)
		}
	}

	goCode, err := c.CompileTestToGo(combined, filepath.Base(nyanPath))
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
	goVersion := goVersionFor(modRoot)
	modContent, err := buildModContent(goVersion, modRoot)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	if err := c.fetchGoPins(tmpDir); err != nil {
		return err
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
func (c *Compiler) CompileFuzzToGo(source, filename string) (helpers, fuzzTests string, fuzzNames []string, err error) {
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
		return "", "", nil, fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	c.logger.Debug("type checking", "file", filename)
	ch := checker.New()
	typeInfo, typeErrs := ch.Check(prog)
	if len(typeErrs) > 0 {
		var msgs []string
		for _, e := range typeErrs {
			msgs = append(msgs, e.Error())
		}
		return "", "", nil, fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	c.recordGoPins(prog)

	c.logger.Debug("generating fuzz Go code", "file", filename)
	gen := codegen.New()
	gen.SetTypeInfo(typeInfo)
	helpers, fuzzTests, fuzzNames, err = gen.GenerateFuzz(prog)
	if err != nil {
		return "", "", nil, err
	}

	if formatted, fmtErr := format.Source([]byte(helpers)); fmtErr == nil {
		helpers = string(formatted)
	}
	if formatted, fmtErr := format.Source([]byte(fuzzTests)); fmtErr == nil {
		fuzzTests = string(formatted)
	}
	return helpers, fuzzTests, fuzzNames, nil
}

// RunFuzz compiles a .nyan file and runs Go fuzz testing.
// Each fuzz_ function in the file is executed individually.
func (c *Compiler) RunFuzz(nyanPath, fuzzTime string) error {
	source, err := os.ReadFile(nyanPath)
	if err != nil {
		return fmt.Errorf("Hiss! Cannot read %s, nya~: %w", nyanPath, err)
	}

	helpers, fuzzTests, fuzzNames, err := c.CompileFuzzToGo(string(source), filepath.Base(nyanPath))
	if err != nil {
		return err
	}

	if len(fuzzNames) == 0 {
		return fmt.Errorf("Hiss! No fuzz_ functions found in %s, nya~", nyanPath)
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
	goVersion := goVersionFor(modRoot)
	modContent, err := buildModContent(goVersion, modRoot)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	if err := c.fetchGoPins(tmpDir); err != nil {
		return err
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

	// Run each fuzz function individually (Go requires -fuzz to match exactly one target)
	for _, name := range fuzzNames {
		c.logger.Debug("running fuzz", "target", name, "fuzztime", fuzzTime)
		fmt.Fprintf(os.Stdout, "  --- %s ---\n", name)
		cmd := exec.Command("go", "test", fmt.Sprintf("-fuzz=^%s$", regexp.QuoteMeta(name)), fmt.Sprintf("-fuzztime=%s", fuzzTime))
		cmd.Dir = tmpDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Hiss! fuzz %s failed, nya~: %w", name, err)
		}
	}
	return nil
}

// TestsFailed is how RunTest says the test binary ran and came back unhappy,
// as opposed to never having got that far.
//
// The binary names the tests it failed and counts them on its way out, so a
// caller has nothing to add. Everything else that can go wrong does need
// saying, and there are two ways to arrive at the same silence from here: a
// `go build` fails with an exit status of its own, and a binary that never
// starts — a TMPDIR mounted noexec, an executable bit that did not survive —
// fails with something that is not an exit status at all. Only a process that
// really ran is quiet.
type TestsFailed struct{ Err error }

func (e *TestsFailed) Error() string { return e.Err.Error() }
func (e *TestsFailed) Unwrap() error { return e.Err }

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
	if c.coverProfile != "" {
		cmd.Env = append(os.Environ(), "MEOW_COVERPROFILE="+c.coverProfile)
	}
	if err := cmd.Run(); err != nil {
		// An exit status means the binary ran and spoke for itself. Anything
		// else means it never started, and only this says so.
		var exited *exec.ExitError
		if errors.As(err, &exited) {
			return &TestsFailed{Err: err}
		}
		return err
	}
	return nil
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
	c.recordGoPins(combinedProg)

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
	goVersion := goVersionFor(modRoot)
	modContent, err := buildModContent(goVersion, modRoot)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("Hiss! Cannot write go.mod, nya~: %w", err)
	}

	if err := c.fetchGoPins(tmpDir); err != nil {
		return err
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

var validGoVersion = regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`)

// defaultGoVersion is the go directive used when no meow source tree is in
// scope to read one from.
const defaultGoVersion = "1.26"

// goVersionFor returns the go directive for the generated module. Inside the
// meow source tree it mirrors that tree's go.mod; outside it there is no go.mod
// to consult, so the compiler's own baseline is used.
func goVersionFor(modRoot string) string {
	if modRoot == "" {
		return defaultGoVersion
	}
	return readGoVersion(filepath.Join(modRoot, "go.mod"))
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
			fields := strings.Fields(strings.TrimPrefix(line, "go "))
			if len(fields) > 0 && validGoVersion.MatchString(fields[0]) {
				return fields[0]
			}
			return "1.26"
		}
	}
	return "1.26"
}

// meowModulePath is the import path of the module that hosts runtime/*, which
// generated programs link against.
const meowModulePath = "github.com/135yshr/meow"

// Version is the version of the meow compiler, used to pin the runtime module
// when compiling from outside the meow source tree. cmd/meow overwrites it with
// the value baked in at release time; when it is left at "dev" (a plain
// `go build`) the version is recovered from the embedded build info instead.
var Version = "dev"

// runtimeRequirement returns the module version generated programs should
// require, and whether one is known at all. It prefers the release version
// baked into the binary and falls back to the version recorded by `go install`.
func runtimeRequirement() (string, bool) {
	if v, ok := asModuleVersion(Version); ok {
		return v, true
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v, ok := asModuleVersion(bi.Main.Version); ok {
			return v, true
		}
		for _, dep := range bi.Deps {
			if dep.Path != meowModulePath {
				continue
			}
			if v, ok := asModuleVersion(dep.Version); ok {
				return v, true
			}
		}
	}
	return "", false
}

// asModuleVersion normalizes a version string into a go.mod requirement.
// Release tooling stamps the tag without its "v" prefix, so it is restored
// here; anything that is not a semantic version ("dev", "(devel)", "") is
// rejected so the caller can fall back.
func asModuleVersion(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !semverLike.MatchString(v) {
		return "", false
	}
	return v, true
}

var semverLike = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[0-9A-Za-z.\-]+)?(\+[0-9A-Za-z.\-]+)?$`)

// buildModContent generates go.mod content for a temporary build directory.
//
// modRoot is the meow source tree to link against, or "" when the compiler is
// running outside it. With a source tree we point at it with a `replace` so the
// working copy of runtime/* is used; without one we pin the published module by
// version, which is what lets a .nyan file anywhere on disk compile.
func buildModContent(goVersion, modRoot string) (string, error) {
	if strings.ContainsAny(modRoot, "\n\r") {
		return "", fmt.Errorf("hiss! module root path contains invalid characters, nya~")
	}
	if modRoot == "" {
		version, ok := runtimeRequirement()
		if !ok {
			// No version is known, so there is nothing meaningful to pin —
			// "latest" is not a version and does not belong in a go.mod. Leave
			// the requirement out entirely and let `go mod tidy` resolve the
			// import it finds in the generated source.
			return fmt.Sprintf("module meow_build\n\ngo %s\n", goVersion), nil
		}
		return fmt.Sprintf("module meow_build\n\ngo %s\n\nrequire %s %s\n",
			goVersion, meowModulePath, version), nil
	}
	// go.mod tokenises on whitespace, so a path containing a space has to be
	// quoted to stay parseable.
	return fmt.Sprintf("module meow_build\n\ngo %s\n\nrequire %s v0.0.0\n\nreplace %s => %s\n",
		goVersion, meowModulePath, meowModulePath, strconv.Quote(modRoot)), nil
}

// companionSourcePath returns the inferred source file path for a test file.
// e.g. "testdata/math_test.nyan" → "testdata/math.nyan"
func companionSourcePath(testPath string) string {
	dir := filepath.Dir(testPath)
	base := filepath.Base(testPath)
	if !strings.HasSuffix(base, "_test.nyan") {
		return ""
	}
	name := strings.TrimSuffix(base, "_test.nyan")
	return filepath.Join(dir, name+".nyan")
}

// findModuleRoot locates the meow source tree to compile against, or returns ""
// when there is none in scope.
//
// It searches upward from the working directory and from the compiler binary's
// own location, so both `meow run` inside a checkout and `go run ./cmd/meow`
// are covered. Crucially it only accepts a directory that really is the meow
// module: any other go.mod (the user's own project, or a parent directory that
// merely happens to have one) is ignored, because pointing the generated
// `replace` at it produces a go.mod that cannot resolve runtime/meowrt.
func (c *Compiler) findModuleRoot() string {
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		starts = append(starts, filepath.Dir(exe))
	}
	for _, start := range starts {
		if root := findMeowModuleFrom(start); root != "" {
			return root
		}
	}
	return ""
}

// findMeowModuleFrom walks up from dir looking for the root of the meow module.
func findMeowModuleFrom(dir string) string {
	for {
		if isMeowModuleRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// isMeowModuleRoot reports whether dir is the root of the meow module: its
// go.mod must declare the meow module path, and it must actually carry the
// runtime package that generated code imports.
func isMeowModuleRoot(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	if modulePath(string(data)) != meowModulePath {
		return false
	}
	info, err := os.Stat(filepath.Join(dir, "runtime", "meowrt"))
	return err == nil && info.IsDir()
}

// modulePath extracts the module path from go.mod content.
func modulePath(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "module "); ok {
			path := strings.TrimSpace(after)
			// The module path may be quoted; go.mod allows either form.
			if unquoted, err := strconv.Unquote(path); err == nil {
				return unquoted
			}
			return path
		}
	}
	return ""
}
