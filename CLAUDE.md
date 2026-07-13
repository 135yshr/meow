# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Meow is a cat-themed functional language whose `.nyan` files **transpile to Go**, then compile to native binaries via the Go toolchain. Every keyword is a cat word (`nyan`=var, `meow`=func, `sniff`=if, `purr`=while, `paw`=lambda, `nya`=print). Errors are cat-flavored: `Hiss! ... , nya~`. Zero third-party runtime dependencies — standard library only; do not introduce any.

**Requires Go 1.26+.**

## Common commands

```bash
make build            # go build -o meow ./cmd/meow   (NOTE: entry is ./cmd/meow, not root)
make test             # go test ./...
make test-update      # go test ./compiler/ -update    — regenerate golden files
make lint             # golangci-lint run ./...  (v2, config in .golangci.yml)
make vet              # go vet ./...
make generate         # go install stringer, then go generate ./...  — after changing token types
make wasm             # build playground/meow.wasm from ./cmd/playground
make cover            # coverage.out + coverage.html

# Run/inspect a .nyan program without installing the binary:
go run ./cmd/meow run examples/hello.nyan
go run ./cmd/meow transpile examples/hello.nyan   # dump generated Go, no build/run

# Single Go test:
go test ./pkg/parser/ -run TestParseLambda -v
```

Pre-commit runs go-fmt, `go vet`, `golangci-lint run --new-from-rev=origin/main`, and `go test ./...`. Commit messages use gitmoji prefixes (e.g. `✨ feat: ...`, `📝 docs: ...`).

## The `meow` CLI vs `go test`

Two separate test systems — don't confuse them:
- `go test ./...` tests the **Go implementation** of the compiler.
- `meow test` (see `runTestCommand` in `cmd/meow/main.go`) runs **`*_test.nyan`** programs written in Meow itself, plus `-fuzz`, `-mutate`, and `-cover` modes. `meow fmt` and `meow lint` operate on `.nyan` source.

## Architecture

Pipeline (CLI path): `.nyan → lexer → parser → checker → codegen → .go → go build → binary`, orchestrated by `compiler/compiler.go`.

- **`pkg/token`** — token types + keyword map. Token names are stringer-generated (`tokentype_string.go`); run `make generate` after editing the `const` block.
- **`pkg/lexer`** — tokenizer exposing `iter.Seq[Token]`.
- **`pkg/parser`** — Pratt parser; uses `iter.Pull` to adapt the lexer's push sequence to pull-based lookahead.
- **`pkg/checker`** + **`pkg/types`** — gradual type checker; produces type info that codegen consumes (typed vs `meow.Value`-boxed output).
- **`pkg/codegen`** — AST → Go source. `genCall`/`genTypedCall` map Meow builtin names to `runtime/*` Go functions.
- **`pkg/formatter`**, **`pkg/linter`**, **`pkg/mutation`** — back the `meow fmt`, `meow lint`, and `meow test -mutate` subcommands.
- **`runtime/`** — Go support code linked into generated binaries: `meowrt` (core `Value` interface, operators, builtins), `file`, `http`, `testing`, `coverage`. `nab "file"` / `nab "http"` in `.nyan` map to these.

**Two execution backends.** The CLI transpiles to Go. The **WASM playground** (`cmd/playground/main_wasm.go`) instead uses **`pkg/interpreter`**, a tree-walking evaluator over the same AST — there is no Go toolchain in the browser. A language change that touches semantics generally must be reflected in **both** codegen and the interpreter.

## Testing conventions

- **Golden tests** live in `testdata/` as `.nyan` + `.golden` pairs, driven by `compiler/compiler_test.go`. When you change compiler output intentionally, regenerate with `make test-update` and review the diff.
- Adding a keyword or builtin is a cross-cutting change. The full checklist (token → lexer → parser → AST → codegen → interpreter → tests → docs) is in **`docs/contributing.md`** — read it before such changes. Deeper design notes are in `docs/internals.md`; the language spec is `docs/spec.md`.
