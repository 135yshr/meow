---
title: "Meow Programming Language Release Notes"
description: "Summary of features and changes in each Meow Programming Language release, organized from newest to oldest."
weight: 1
---

A summary of features and changes in each Meow Programming Language release, organized from newest to oldest.

---

## Kitty (Struct) Types — PR #26

User-defined composite types with typed fields.

```meow
kitty Cat {
  name: string
  age: int
}

nyan nyantyu = Cat("Nyantyu", 3)
nya(nyantyu.name)   # => Nyantyu
```

**What's new:**
- `kitty` keyword for struct definitions
- Field definitions with `name: type` syntax
- Constructor functions (call type name with positional args)
- Field access with `.` notation
- Structural equality for kitty instances
- `KittyStmt` and `KittyField` AST nodes
- Runtime `Kitty` value type with `GetField` method

---

## Go-style Grouped Parameter Types — PR #25

Parameters can share type annotations using Go-style grouping:

```meow
meow add(a, b int) int {
  bring a + b
}
# a and b are both int
```

**What's new:**
- Right-to-left type propagation in parameter lists
- `(a, b int, c, d string)` groups `a,b` as `int` and `c,d` as `string`

---

## Immutable Variables and Range-based Purr — PR #23

Variables are now immutable by default. The `purr` loop gains range-based forms.

```meow
purr i (10) { nya(i) }       # count: 0 to 9
purr i (1..20) { nya(i) }    # range: 1 to 20 (inclusive)
```

**What's new:**
- `purr i (n)` — count form (0 to n-1)
- `purr i (a..b)` — range form (a to b inclusive)
- `RangeStmt` AST node with `Start`, `End`, `Inclusive` fields
- Loop variable automatically bound in each iteration

---

## Formatter and Linter — PR #21, #22

New CLI subcommands for code quality.

```bash
meow fmt my_file.nyan    # Auto-format
meow lint my_file.nyan   # Check style
```

**What's new:**
- `meow fmt` — auto-formats `.nyan` files (indentation, spacing, alignment)
- `meow lint` — checks for style issues
  - `snake-case` rule: identifiers must use snake_case
- Recursive file discovery with `./...` pattern

---

## Static Type System (Gradual Typing) — PR #20

Optional type annotations for compile-time checking and faster code generation.

```meow
meow add(a int, b int) int {
  bring a + b
}
nyan x int = 42
```

**What's new:**
- Type keywords: `int`, `float`, `string`, `bool`, `furball`, `litter`
- Variable type annotations: `nyan x int = 42`
- Function parameter types: `meow f(a int, b int)`
- Return type annotations: `meow f(a int) int`
- Type checker (`pkg/checker/`) with two-pass analysis
- Typed code generation — fully typed functions emit native Go types (`int64`, `float64`, etc.)
- Gradual typing — typed and untyped code coexist

---

## Mutation Testing — PR #18, #19

Automated mutation testing to verify test suite strength.

```bash
meow test -mutate my_test.nyan
```

**What's new:**
- `meow test -mutate` — mutation testing mode
- Auto-discovery of test files
- Schemata-based mutation (arithmetic, comparison, boolean operators)
- Kill/survive report
- Unified snake_case naming convention

---

## Statement Coverage — PR #17

Statement-level coverage tracking for `.nyan` programs.

```bash
meow test -cover my_test.nyan
```

**What's new:**
- `meow test -cover` — coverage instrumentation
- Coverage report output
- Coverage profile export (`MEOW_COVERPROFILE`)

---

## Catwalk Output Tests — PR #16

Output verification tests inspired by Go's `Example` tests.

```meow
meow catwalk_hello() {
  nya("Hello!")
}
# Output:
# Hello!
```

**What's new:**
- `catwalk_` prefix for output verification functions
- `# Output:` comment blocks to specify expected output
- `Catwalk()` runtime function — captures stdout and compares
- Recursive test file discovery with `./...` pattern

---

## Comprehensive Test Suite — PR #15

Added extensive sample tests covering all language features.

**What's new:**
- Test samples for arithmetic, strings, lists, conditionals, loops, lambdas, pipes, pattern matching, error handling
- Tests moved to `testdata/` for CI

---

## Help Subcommand — PR #13, #14

```bash
meow help
meow help run
meow help build
```

**What's new:**
- `meow help [command]` — shows help text for each CLI command
- Updated CLI usage documentation

---

## Testing Framework — PR #12

Full-featured testing framework with fuzz support.

```meow
nab "testing"

meow test_basic() {
  expect(1 + 1, 2, "addition")
  judge(yarn)
  refuse(hairball)
}
```

**What's new:**
- `testing` standard library package
- `judge(condition)` — assert truthy
- `expect(actual, expected)` — assert equal
- `refuse(condition)` — assert falsy
- `run(name, fn)` — execute named test
- `report()` — print summary, exit on failure
- `test_` prefix for auto-discovered test functions
- `meow test` CLI subcommand
- Fuzz testing with `seed()` and random inputs

---

## Standard Library: HTTP Client — PR #11

```meow
nab "http"
http.pounce("https://example.com") |=| nya
```

**What's new:**
- `http` package (`nab "http"`)
- `pounce(url)` — HTTP GET
- `toss(url, body)` — HTTP POST
- `knead(url, body)` — HTTP PUT
- `swat(url)` — HTTP DELETE
- `prowl(url)` — HTTP OPTIONS
- Options map for headers and max body size
- Auto-JSON serialization for Map bodies
- 10-second timeout, 1 MiB body limit

---

## Standard Library: File I/O — PR #10

```meow
nab "file"
nyan content = file.snoop("data.txt")
nyan lines = file.stalk("data.txt")
```

**What's new:**
- `file` package (`nab "file"`)
- `snoop(path)` — read entire file as string
- `stalk(path)` — read file line by line, return list
- `nab` statement for importing standard library packages
- Member access syntax: `package.function()`

---

## Pipe Operator and Error Recovery — PR #3

```meow
[1, 2, 3] |=| lick(paw(x) { x * 2 }) |=| nya
nyan safe = divide(10, 0) ~> 0
```

**What's new:**
- `|=|` pipe operator — chains operations left-to-right
- `~>` error recovery operator — catch with fallback value or handler
- `gag(fn)` — explicit error catching
- `is_furball(v)` — check for error values
- `Furball` error value type
- `GagOr` function for catch-with-fallback

---

## CLI: GoReleaser and Homebrew — PR #4, #5

```bash
brew install 135yshr/homebrew-tap/meow
```

**What's new:**
- GoReleaser configuration for automated binary releases
- Homebrew tap for macOS installation
- `meow version` subcommand with ldflags injection
- Semantic release integration
- GitHub Actions CI workflow

---

## Initial Release — PR #1

The first release of the Meow programming language.

**Core features:**
- Cat-themed syntax with 19 keywords
- `.nyan` file extension
- Transpiles to Go source code
- Compiles to native binaries via `go build`
- Dynamic typing with boxed `meow.Value` runtime

**Language features:**
- Variables (`nyan`), functions (`meow`), return (`bring`)
- Conditionals (`sniff`/`scratch`)
- While loops (`purr`)
- Lambdas (`paw`)
- Lists with index access
- Pattern matching (`peek`) with literals, ranges, wildcards
- `lick` (map), `picky` (filter), `curl` (reduce)
- Error handling: `hiss` (raise), `Hiss! ...` error messages
- Comments: `#` line, `-~ ... ~-` block
- `nya` print function
- Built-in type conversions: `to_int`, `to_float`, `to_string`
- `len`, `head`, `tail`, `append` for lists

**Compiler pipeline:**
- Lexer with `iter.Seq[Token]`
- Pratt parser with `iter.Pull`
- AST with expression, statement, and pattern nodes
- Codegen to Go source
- CLI: `meow run`, `meow build`, `meow transpile`

**Infrastructure:**
- GitHub Actions CI
- Go 1.26 requirement
- MIT License
