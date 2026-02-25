package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/135yshr/meow/compiler"
	"github.com/135yshr/meow/pkg/formatter"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/linter"
	"github.com/135yshr/meow/pkg/parser"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	verbose := false
	args := os.Args[1:]
	filtered := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--verbose" || a == "-v" {
			verbose = true
		} else {
			filtered = append(filtered, a)
		}
	}
	args = filtered

	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	c := compiler.New(logger)

	switch args[0] {
	case "help":
		if len(args) >= 2 {
			printSubcommandHelp(args[1])
		} else {
			printUsage()
		}
	case "version":
		fmt.Printf("meow version %s (commit: %s, built: %s)\n", version, commit, date)
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Hiss! Please specify a .nyan file, nya~")
			os.Exit(1)
		}
		if err := c.Run(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "build":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Hiss! Please specify a .nyan file, nya~")
			os.Exit(1)
		}
		output := ""
		if len(args) >= 4 && args[2] == "-o" {
			output = args[3]
		}
		if err := c.Build(args[1], output); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Build complete, nya~!")
	case "transpile":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Hiss! Please specify a .nyan file, nya~")
			os.Exit(1)
		}
		source, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		code, err := c.CompileToGo(string(source), args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(code)
	case "test":
		runTestCommand(c, args[1:])
	case "fmt":
		runFmtCommand(args[1:])
	case "lint":
		runLintCommand(args[1:])
	default:
		// Treat as "run" if the argument looks like a file
		if len(args) >= 1 && len(args[0]) > 0 && args[0][0] != '-' {
			if err := c.Run(args[0]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			printUsage()
			os.Exit(1)
		}
	}
}

func runTestCommand(c *compiler.Compiler, args []string) {
	var files []string
	fuzz := false
	fuzzTime := ""
	mutate := false
	cover := false
	coverProfile := ""

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-fuzz":
			fuzz = true
		case args[i] == "-fuzztime":
			if i+1 < len(args) {
				i++
				fuzzTime = args[i]
			}
		case args[i] == "-mutate":
			mutate = true
		case args[i] == "-cover":
			cover = true
		case strings.HasPrefix(args[i], "-coverprofile="):
			coverProfile = strings.TrimPrefix(args[i], "-coverprofile=")
			cover = true
		case args[i] == "-coverprofile":
			if i+1 < len(args) {
				i++
				coverProfile = args[i]
				cover = true
			}
		default:
			if len(args[i]) > 0 && args[i][0] != '-' {
				files = append(files, args[i])
			}
		}
	}

	if cover {
		c.EnableCoverage(coverProfile)
	}

	if fuzz {
		if len(files) == 0 {
			var err error
			files, err = discoverFuzzFiles(".")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			var err error
			files, err = resolveFuzzPaths(files)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		if len(files) == 0 {
			fmt.Fprintln(os.Stderr, "Hiss! No fuzz files found, nya~")
			os.Exit(1)
		}
		for _, f := range files {
			fmt.Fprintf(os.Stdout, "=== Fuzzing %s ===\n", f)
			if err := c.RunFuzz(f, fuzzTime); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return
	}

	if mutate {
		runMutateCommand(c, files)
		return
	}

	if len(files) == 0 {
		var err error
		files, err = discoverTestFiles(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		var err error
		files, err = resolveTestPaths(files)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Hiss! No test files found, nya~")
		os.Exit(1)
	}

	if coverProfile != "" {
		if err := os.WriteFile(coverProfile, []byte("mode: set\n"), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Hiss! Cannot write coverage profile header, nya~: %v\n", err)
			os.Exit(1)
		}
	}

	hasFailure := false
	for _, f := range files {
		fmt.Fprintf(os.Stdout, "=== Testing %s ===\n", f)
		if err := c.RunTest(f); err != nil {
			hasFailure = true
		}
	}
	if hasFailure {
		os.Exit(1)
	}
}

func discoverFiles(dir, pattern string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("Hiss! Cannot search for files, nya~: %w", err)
	}
	return matches, nil
}

func discoverFilesRecursive(root string, match func(string) bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && match(d.Name()) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Hiss! Cannot search for files, nya~: %w", err)
	}
	return files, nil
}

func resolvePaths(patterns []string, discover func(string) ([]string, error), discoverRecursive func(string) ([]string, error)) ([]string, error) {
	var result []string
	seen := make(map[string]struct{})
	add := func(paths []string) {
		for _, p := range paths {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			result = append(result, p)
		}
	}
	for _, p := range patterns {
		if strings.HasSuffix(p, "/...") || strings.HasSuffix(p, string(filepath.Separator)+"...") {
			root := strings.TrimSuffix(p, "/...")
			root = strings.TrimSuffix(root, string(filepath.Separator)+"...")
			if root == "." || root == "" {
				root = "."
			}
			found, err := discoverRecursive(root)
			if err != nil {
				return nil, err
			}
			add(found)
		} else {
			info, err := os.Stat(p)
			if err != nil {
				add([]string{p})
				continue
			}
			if info.IsDir() {
				found, err := discover(p)
				if err != nil {
					return nil, err
				}
				add(found)
			} else {
				add([]string{p})
			}
		}
	}
	return result, nil
}

func discoverTestFiles(dir string) ([]string, error) {
	return discoverFiles(dir, "*_test.nyan")
}

func discoverTestFilesRecursive(root string) ([]string, error) {
	return discoverFilesRecursive(root, func(name string) bool {
		return strings.HasSuffix(name, "_test.nyan")
	})
}

func resolveTestPaths(patterns []string) ([]string, error) {
	return resolvePaths(patterns, discoverTestFiles, discoverTestFilesRecursive)
}

func discoverFuzzFiles(dir string) ([]string, error) {
	return discoverFiles(dir, "fuzz_*.nyan")
}

func discoverFuzzFilesRecursive(root string) ([]string, error) {
	return discoverFilesRecursive(root, func(name string) bool {
		return strings.HasPrefix(name, "fuzz_") && strings.HasSuffix(name, ".nyan")
	})
}

func resolveFuzzPaths(patterns []string) ([]string, error) {
	return resolvePaths(patterns, discoverFuzzFiles, discoverFuzzFilesRecursive)
}

func runFmtCommand(args []string) {
	write := false
	var files []string
	for _, a := range args {
		if a == "-w" {
			write = true
		} else if strings.HasPrefix(a, "-") {
			fmt.Fprintf(os.Stderr, "Hiss! Unknown flag for fmt: %s, nya~\n", a)
			os.Exit(1)
		} else {
			files = append(files, a)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Hiss! Please specify .nyan files to format, nya~")
		os.Exit(1)
	}

	for _, f := range files {
		source, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Hiss! Cannot read %s, nya~: %v\n", f, err)
			os.Exit(1)
		}
		formatted := formatter.FormatSource(string(source), f)
		if write {
			info, err := os.Stat(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Hiss! Cannot stat %s, nya~: %v\n", f, err)
				os.Exit(1)
			}
			if err := os.WriteFile(f, []byte(formatted), info.Mode().Perm()); err != nil {
				fmt.Fprintf(os.Stderr, "Hiss! Cannot write %s, nya~: %v\n", f, err)
				os.Exit(1)
			}
		} else {
			fmt.Print(formatted)
		}
	}
}

func runLintCommand(args []string) {
	var patterns []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			fmt.Fprintf(os.Stderr, "Hiss! Unknown flag for lint: %s, nya~\n", a)
			os.Exit(1)
		}
		patterns = append(patterns, a)
	}

	files, err := resolveLintPaths(patterns)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Hiss! No .nyan files found, nya~")
		os.Exit(1)
	}

	l := linter.New()
	hasIssues := false

	for _, f := range files {
		source, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Hiss! Cannot read %s, nya~: %v\n", f, err)
			os.Exit(1)
		}
		lex := lexer.New(string(source), f)
		p := parser.New(lex.Tokens())
		prog, parseErrs := p.Parse()
		if len(parseErrs) > 0 {
			for _, e := range parseErrs {
				fmt.Fprintln(os.Stderr, e)
			}
			hasIssues = true
			continue
		}
		diags := l.Lint(prog)
		for _, d := range diags {
			fmt.Fprintln(os.Stderr, d)
			hasIssues = true
		}
	}

	if hasIssues {
		os.Exit(1)
	}
}

func discoverNyanFiles(dir string) ([]string, error) {
	return discoverFiles(dir, "*.nyan")
}

func discoverNyanFilesRecursive(root string) ([]string, error) {
	return discoverFilesRecursive(root, func(name string) bool {
		return strings.HasSuffix(name, ".nyan")
	})
}

func resolveLintPaths(patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return discoverNyanFiles(".")
	}
	return resolvePaths(patterns, discoverNyanFiles, discoverNyanFilesRecursive)
}

func runMutateCommand(c *compiler.Compiler, files []string) {
	// Explicit mode: first file is source, rest are test files.
	if len(files) >= 2 && !isPattern(files[0]) {
		if err := c.RunMutationTest(files[0], files[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// Auto-discovery mode: resolve patterns to find test files,
	// then infer source files by stripping _test suffix.
	var testFiles []string
	if len(files) == 0 {
		var err error
		testFiles, err = discoverTestFiles(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		var err error
		testFiles, err = resolveTestPaths(files)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if len(testFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Hiss! No test files found, nya~")
		os.Exit(1)
	}

	// Group test files by their inferred source file.
	type mutationPair struct {
		source string
		tests  []string
	}
	pairMap := make(map[string]*mutationPair)
	var pairOrder []string
	var skipped []string

	for _, tf := range testFiles {
		src := inferSourceFile(tf)
		if _, err := os.Stat(src); err != nil {
			skipped = append(skipped, tf)
			continue
		}
		if p, ok := pairMap[src]; ok {
			p.tests = append(p.tests, tf)
		} else {
			pairMap[src] = &mutationPair{source: src, tests: []string{tf}}
			pairOrder = append(pairOrder, src)
		}
	}

	if len(pairOrder) == 0 {
		fmt.Fprintln(os.Stderr, "Hiss! No source files found for mutation testing, nya~")
		fmt.Fprintln(os.Stderr, "  Each foo_test.nyan needs a matching foo.nyan source file.")
		if len(skipped) > 0 {
			fmt.Fprintf(os.Stderr, "  Skipped %d test file(s) with no matching source.\n", len(skipped))
		}
		os.Exit(1)
	}

	hasFailure := false
	for _, src := range pairOrder {
		pair := pairMap[src]
		fmt.Fprintf(os.Stdout, "=== Mutating %s ===\n", src)
		if err := c.RunMutationTest(pair.source, pair.tests); err != nil {
			fmt.Fprintln(os.Stderr, err)
			hasFailure = true
		}
	}
	if hasFailure {
		os.Exit(1)
	}
}

// inferSourceFile returns the source file path for a test file.
// e.g. "testdata/math_test.nyan" → "testdata/math.nyan"
func inferSourceFile(testFile string) string {
	dir := filepath.Dir(testFile)
	base := filepath.Base(testFile)
	name := strings.TrimSuffix(base, "_test.nyan")
	return filepath.Join(dir, name+".nyan")
}

// isPattern returns true if the path contains a glob/recursive pattern.
func isPattern(p string) bool {
	return strings.HasSuffix(p, "/...") ||
		strings.HasSuffix(p, string(filepath.Separator)+"...")
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Meow Language Compiler 🐱

Usage:
  meow <command> [arguments]

Commands:
  run <file.nyan>              Run a .nyan file
  build <file.nyan> [-o name]  Build a binary
  transpile <file.nyan>        Show generated Go code
  test [files...]              Run _test.nyan files
  fmt [-w] <files...>          Format .nyan source files
  lint [files/patterns...]     Run static analysis
  version                      Show version info
  help [command]               Show help for a command

  meow <file.nyan>             Shorthand for 'meow run'

Flags:
  --verbose, -v                Enable debug logging

Use "meow help <command>" for more information about a command.`)
}

func printSubcommandHelp(cmd string) {
	helps := map[string]string{
		"run": `Usage: meow run <file.nyan>

Run a .nyan program. The file is compiled to Go and executed immediately.

Examples:
  meow run hello.nyan
  meow run examples/hello.nyan`,

		"build": `Usage: meow build <file.nyan> [-o name]

Compile a .nyan file into a standalone binary.

Flags:
  -o <name>  Set the output binary name

Examples:
  meow build hello.nyan
  meow build hello.nyan -o hello`,

		"transpile": `Usage: meow transpile <file.nyan>

Show the generated Go source code without compiling or running it.

Examples:
  meow transpile hello.nyan`,

		"test": `Usage: meow test [flags] [files/patterns...]

Run test files. Without arguments, discovers and runs all *_test.nyan files
in the current directory.

Patterns:
  ./...                  Recursively find all *_test.nyan in current directory
  dir/...                Recursively find all *_test.nyan under dir/
  dir/                   Find *_test.nyan in dir/ (non-recursive)
  file_test.nyan         Run a specific test file

Flags:
  -fuzz                  Run fuzz tests
  -fuzztime <duration>   Fuzz test duration (default: 10s)
  -mutate                Run mutation tests (explicit or auto-discover pairs)
  -cover                 Enable statement coverage
  -coverprofile=<file>   Write coverage profile to file (Go-compatible format)

Examples:
  meow test
  meow test ./...
  meow test testdata/...
  meow test testdata/
  meow test math_test.nyan
  meow test -fuzz math_test.nyan
  meow test -fuzz -fuzztime 30s math_test.nyan
  meow test -mutate math.nyan math_test.nyan
  meow test -mutate ./...
  meow test -cover math_test.nyan
  meow test -coverprofile=coverage.out ./...`,

		"fmt": `Usage: meow fmt [-w] <files...>

Format .nyan source files. By default, prints the formatted output to stdout.

Flags:
  -w  Write the formatted output back to the file

Examples:
  meow fmt hello.nyan
  meow fmt -w hello.nyan
  meow fmt examples/fibonacci.nyan`,

		"lint": `Usage: meow lint [files/patterns...]

Run static analysis on .nyan files. Without arguments, checks all *.nyan files
in the current directory.

Patterns:
  ./...                  Recursively check all *.nyan files
  dir/...                Recursively check all *.nyan under dir/
  dir/                   Check *.nyan in dir/ (non-recursive)
  file.nyan              Check a specific file

Rules:
  snake-case             Identifiers must use snake_case
  unused-var             Declared variables must be used
  unreachable-code       Code after bring is unreachable
  empty-block            Function/if/while bodies must not be empty

Examples:
  meow lint hello.nyan
  meow lint ./...
  meow lint examples/`,

		"version": `Usage: meow version

Print the version, commit hash, and build date of the meow compiler.`,

		"help": `Usage: meow help [command]

Show help for the meow compiler or a specific command.

Examples:
  meow help
  meow help run
  meow help build`,
	}

	text, ok := helps[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "Hiss! Unknown command %q, nya~\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, text)
}
