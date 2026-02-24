package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/135yshr/meow/compiler"
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

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-fuzz":
			fuzz = true
		case "-fuzztime":
			if i+1 < len(args) {
				i++
				fuzzTime = args[i]
			}
		case "-mutate":
			mutate = true
		default:
			if len(args[i]) > 0 && args[i][0] != '-' {
				files = append(files, args[i])
			}
		}
	}

	if fuzz {
		if len(files) == 0 {
			fmt.Fprintln(os.Stderr, "Hiss! Please specify a .nyan file for fuzzing, nya~")
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
		if len(files) < 2 {
			fmt.Fprintln(os.Stderr, "Hiss! mutate requires source and test files, nya~")
			os.Exit(1)
		}
		sourcePath := files[0]
		testPaths := files[1:]
		if err := c.RunMutationTest(sourcePath, testPaths); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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

func discoverTestFiles(dir string) ([]string, error) {
	pattern := filepath.Join(dir, "*_test.nyan")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("Hiss! Cannot search for test files, nya~: %w", err)
	}
	return matches, nil
}

func discoverTestFilesRecursive(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.nyan") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Hiss! Cannot search for test files, nya~: %w", err)
	}
	return files, nil
}

func resolveTestPaths(patterns []string) ([]string, error) {
	var result []string
	for _, p := range patterns {
		if strings.HasSuffix(p, "/...") || strings.HasSuffix(p, string(filepath.Separator)+"...") {
			root := strings.TrimSuffix(p, "/...")
			root = strings.TrimSuffix(root, string(filepath.Separator)+"...")
			if root == "." || root == "" {
				root = "."
			}
			found, err := discoverTestFilesRecursive(root)
			if err != nil {
				return nil, err
			}
			result = append(result, found...)
		} else {
			info, err := os.Stat(p)
			if err != nil {
				// Not a file/dir — treat as literal path
				result = append(result, p)
				continue
			}
			if info.IsDir() {
				found, err := discoverTestFiles(p)
				if err != nil {
					return nil, err
				}
				result = append(result, found...)
			} else {
				result = append(result, p)
			}
		}
	}
	return result, nil
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
  -mutate                Run mutation tests (requires source and test files)

Examples:
  meow test
  meow test ./...
  meow test testdata/...
  meow test testdata/
  meow test math_test.nyan
  meow test -fuzz math_test.nyan
  meow test -fuzz -fuzztime 30s math_test.nyan
  meow test -mutate math.nyan math_test.nyan`,

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
