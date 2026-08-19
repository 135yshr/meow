# Contributing to Meow

Thank you for your interest in contributing to the Meow programming language!

## Quick Start

```bash
git clone https://github.com/135yshr/meow.git
cd meow
go install golang.org/x/tools/cmd/stringer@v0.42.0
go build ./cmd/meow
go test ./...
```

**Requires Go 1.26+**

## How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes with tests
4. Ensure all tests pass (`go test ./...`) and no vet warnings (`go vet ./...`)
5. Commit with [gitmoji](https://gitmoji.dev/) prefix (e.g., `✨ feat: Add new feature`)
6. Push and open a Pull Request

## Detailed Guide

For the full contributor guide — including how to add keywords, built-in functions, stdlib packages, and testing conventions — see **[docs/contributing.md](docs/contributing.md)**.

## Dependencies

Meow is **standard-library only**. The compiler, the interpreter and every
`runtime/` package stay on the standard library, and the module requires
nothing else.

A program that needs a third-party library reaches it with `nab go "path"`,
which imports the Go package into the program being built rather than into
meow itself. That is where a dependency belongs: the program that wanted it
carries it, and meow stays a compiler.

Please do not introduce third-party runtime packages. If one seems unavoidable
— something `nab go` genuinely cannot reach — raise it in an issue first.
