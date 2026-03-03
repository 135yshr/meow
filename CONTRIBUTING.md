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

Meow has **zero runtime dependencies** — standard library only. Please do not introduce third-party runtime packages.
