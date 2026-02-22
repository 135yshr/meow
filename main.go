package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/135yshr/meow/compiler"
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
	default:
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

func printUsage() {
	fmt.Fprintln(os.Stderr, `Meow Language Compiler

Usage:
  meow run <file.nyan>              Run a .nyan file
  meow build <file.nyan> [-o name]  Build a binary
  meow transpile <file.nyan>        Show generated Go code
  meow <file.nyan>                  Shorthand for 'meow run'

Flags:
  --verbose, -v                     Enable debug logging`)
}
