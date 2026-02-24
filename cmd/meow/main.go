package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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

func printUsage() {
	fmt.Fprintln(os.Stderr, `Meow Language Compiler 🐱

Usage:
  meow run <file.nyan>              Run a .nyan file
  meow build <file.nyan> [-o name]  Build a binary
  meow transpile <file.nyan>        Show generated Go code
  meow test [files...]              Run _test.nyan files
  meow test -fuzz <file.nyan>      Run fuzz tests
  meow test -mutate <src> <tests>  Run mutation tests
  meow version                      Show version info
  meow <file.nyan>                  Shorthand for 'meow run'

Flags:
  --verbose, -v                     Enable debug logging
  -fuzztime <duration>              Fuzz test duration (default: 10s)`)
}
