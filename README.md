<p align="center">
<pre align="center">
    /\_____/\
   /  o   o  \
  ( ==  ^  == )
   )         (
  (           )
 ( (  )   (  ) )
(__(__)___(__)__)

 ███╗   ███╗███████╗ ██████╗ ██╗    ██╗
 ████╗ ████║██╔════╝██╔═══██╗██║    ██║
 ██╔████╔██║█████╗  ██║   ██║██║ █╗ ██║
 ██║╚██╔╝██║██╔══╝  ██║   ██║██║███╗██║
 ██║ ╚═╝ ██║███████╗╚██████╔╝╚███╔███╔╝
 ╚═╝     ╚═╝╚══════╝ ╚═════╝  ╚══╝╚══╝
</pre>
</p>

<p align="center">
  <b>The purrfect functional programming language 🐱</b>
</p>

<p align="center">
  <a href="https://github.com/135yshr/meow/actions"><img src="https://github.com/135yshr/meow/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build Status"></a>
  <a href="https://pkg.go.dev/github.com/135yshr/meow"><img src="https://pkg.go.dev/badge/github.com/135yshr/meow.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/135yshr/meow"><img src="https://goreportcard.com/badge/github.com/135yshr/meow" alt="Go Report Card"></a>
  <a href="https://github.com/135yshr/meow"><img src="https://img.shields.io/github/go-mod/go-version/135yshr/meow" alt="Go Version"></a>
  <a href="https://github.com/135yshr/meow/blob/main/LICENSE"><img src="https://img.shields.io/github/license/135yshr/meow" alt="License"></a>
</p>

<p align="center">
  <a href="https://github.com/135yshr/meow/issues"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="Contributions Welcome"></a>
  <a href="https://github.com/135yshr/meow/stargazers"><img src="https://img.shields.io/github/stars/135yshr/meow?style=flat" alt="Stars"></a>
  <a href="https://github.com/135yshr/meow/issues"><img src="https://img.shields.io/github/issues/135yshr/meow" alt="Issues"></a>
  <a href="https://github.com/135yshr/meow/pulls"><img src="https://img.shields.io/badge/PRs-welcome-blue.svg" alt="PRs Welcome"></a>
</p>

---

**Meow** is a cat-themed functional programming language that transpiles `.nyan` files into Go source code. It's a joke language — but one that actually works. Write real programs with cat words, compile them to native binaries, and run them at full speed.

```
nyan name = "Tama"
meow greet(who) {
  bring "Hello, " + who + "!"
}
nya(greet(name))
```

```
$ meow run hello.nyan
Hello, Tama!
```

## Features

- **Cat-themed syntax** — Every keyword is a cat word (`nyan`, `meow`, `sniff`, `purr`, ...)
- **Transpiles to Go** — Generates clean, readable Go code
- **Native binaries** — Compiled output runs at full Go speed
- **Gradual typing** — Optional type annotations with `meow.Value` boxing
- **First-class functions** — Lambdas with `paw(x) { x * 2 }`
- **Structs & newtypes** — `kitty` (struct), `collar` (newtype), `breed` (alias)
- **Interfaces & methods** — `pose` (interface), `groom` (method impl)
- **Standard library** — `nab "file"`, `nab "http"` for file I/O and HTTP
- **List operations** — `lick` (map), `picky` (filter), `curl` (reduce)
- **Pattern matching** — `peek` expression with ranges and wildcards
- **Pipe operator** — Chain operations with `|=|`
- **Cat error messages** — `Hiss! "x" is not defined, nya~`
- **Modern Go internals** — Built with `iter.Seq`, `iter.Pull`, generics, `go:embed`

## Installation

### Homebrew

```bash
brew install 135yshr/homebrew-tap/meow
```

### From Source

```bash
git clone https://github.com/135yshr/meow.git
cd meow
go install .
```

### Go Install

```bash
go install github.com/135yshr/meow@latest
```

**Requires Go 1.26+**

## Quick Start

Create `hello.nyan`:

```
nyan name = "Tama"
nya("Hello, " + name + "!")
```

Run it:

```bash
meow run hello.nyan
# => Hello, Tama!
```

Or build a binary:

```bash
meow build hello.nyan -o hello
./hello
```

## Language Reference

### Variables

```
nyan x = 42
nyan greeting = "Hello!"
nyan pi = 3.14
nyan cats_are_great = yarn      # true
nyan dogs_rule = hairball       # false
nyan nothing = catnap           # nil
```

### Functions

```
meow add(a, b) {
  bring a + b
}

nya(add(1, 2))   # => 3
```

### Control Flow

```
# if / else
sniff (x > 0) {
  nya("positive")
} scratch sniff (x == 0) {
  nya("zero")
} scratch {
  nya("negative")
}

# while loop
nyan i = 0
purr (i < 10) {
  nya(i)
  i = i + 1
}
```

### Lambdas

```
nyan double = paw(x) { x * 2 }
nya(double(5))   # => 10
```

### Lists

```
nyan nums = [1, 2, 3, 4, 5]

# map
lick(nums, paw(x) { x * 2 })         # => [2, 4, 6, 8, 10]

# filter
picky(nums, paw(x) { x % 2 == 0 })   # => [2, 4]

# reduce
curl(nums, 0, paw(acc, x) { acc + x }) # => 15
```

### Error Handling

```
meow divide(a, b) {
  sniff (b == 0) { hiss("division by zero") }
  bring a / b
}

# Concise error recovery with ~>
nyan val = divide(10, 0) ~> 0
nya(val)   # => 0

# With a handler function
nyan val2 = divide(10, 0) ~> paw(err) { 42 }
nya(val2)  # => 42

# Verbose style with gag/is_furball
nyan result = gag(paw() { divide(10, 0) })
sniff (is_furball(result)) {
  nya("caught:", result)
} scratch {
  nya("ok:", result)
}
# => caught: Hiss! division by zero
```

### Pattern Matching

```
nyan result = peek(score) {
  0 => "zero",
  1..10 => "low",
  11..100 => "high",
  _ => "off the charts"
}
```

### Comments

```
# This is a line comment

-~ This is a
   block comment ~-
```

## Language Cheat Sheet

> Full reference: [docs/reference.md](docs/reference.md)

### Keywords

| Meow | Meaning | Example |
|------|---------|---------|
| `nyan` | Variable declaration | `nyan x = 42` |
| `meow` | Function definition | `meow add(a, b) { bring a + b }` |
| `bring` | Return value | `bring x + 1` |
| `sniff` | If condition | `sniff (x > 0) { ... }` |
| `scratch` | Else branch | `} scratch { ... }` |
| `purr` | While loop | `purr (i < 10) { ... }` |
| `paw` | Lambda (anonymous function) | `paw(x) { x * 2 }` |
| `nya` | Print | `nya("Hello!")` |
| `lick` | Map over list | `lick(nums, paw(x) { x * 2 })` |
| `picky` | Filter list | `picky(nums, paw(x) { x > 0 })` |
| `curl` | Reduce list | `curl(nums, 0, paw(a, x) { a + x })` |
| `peek` | Pattern match | `peek(v) { 0 => "zero", _ => "other" }` |
| `hiss` | Raise error | `hiss("something went wrong")` |
| `gag` | Catch errors | `gag(paw() { risky() })` |
| `is_furball` | Check if error | `is_furball(result)` |
| `kitty` | Struct definition | `kitty Cat { name: string }` |
| `breed` | Type alias | `breed Name = string` |
| `collar` | Newtype wrapper | `collar UserId = int` |
| `pose` | Interface definition | `pose Showable { meow show() string }` |
| `groom` | Method implementation | `groom Cat { meow show() string { ... } }` |
| `self` | Self reference in methods | `bring self.name` |
| `nab` | Import stdlib package | `nab "file"` |
| `flaunt` | Export *(planned)* | — |
| `yarn` | True (boolean literal) | `nyan ok = yarn` |
| `hairball` | False (boolean literal) | `nyan ng = hairball` |
| `catnap` | Nil (empty value) | `nyan nothing = catnap` |

### Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `+` `-` `*` `/` `%` | Arithmetic | `1 + 2`, `10 % 3` |
| `==` `!=` | Equality | `x == 0` |
| `<` `>` `<=` `>=` | Comparison | `x < 10` |
| `&&` `\|\|` `!` | Logical | `x > 0 && !done` |
| `\|=\|` | Pipe | `nums \|=\| lick(double)` |
| `~>` | Error recovery | `divide(10, 0) ~> 0` |
| `..` | Range | `1..10` |
| `=>` | Match arm | `0 => "zero"` |
| `=` | Assignment | `nyan x = 1` |

### Literals & Delimiters

| Token | Meaning |
|-------|---------|
| `42`, `3.14` | Integer / Float |
| `"hello"` | String |
| `( )` `{ }` `[ ]` | Parens / Braces / Brackets |
| `,` | Separator |
| `#` | Line comment |
| `-~ ... ~-` | Block comment |

## CLI Usage

```
Meow Language Compiler 🐱

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

Use "meow help <command>" for more information about a command.
```

## How It Works

```
.nyan → [Lexer] → iter.Seq[Token] → [Parser] → AST → [Codegen] → .go → go build → Binary
```

The compiler pipeline:

1. **Lexer** (`pkg/lexer`) — Tokenizes source using `iter.Seq[Token]`
2. **Parser** (`pkg/parser`) — Pratt parser with `iter.Pull` for push-to-pull conversion
3. **Codegen** (`pkg/codegen`) — Transforms AST into Go source code
4. **Compiler** (`compiler`) — Orchestrates the pipeline, runs `go build`
5. **Runtime** (`runtime/meowrt`) — Provides `Value` interface, operators, and builtins

## Examples

Check the [`examples/`](examples/) directory:

| File | Description |
|------|-------------|
| [`hello.nyan`](examples/hello.nyan) | Hello World with functions |
| [`fibonacci.nyan`](examples/fibonacci.nyan) | Recursive Fibonacci sequence |
| [`fizzbuzz.nyan`](examples/fizzbuzz.nyan) | Classic FizzBuzz with `sniff`/`scratch` chains |
| [`list_ops.nyan`](examples/list_ops.nyan) | `lick`, `picky`, `curl` demo |
| [`error_handling.nyan`](examples/error_handling.nyan) | `hiss`, `gag`, `is_furball`, `~>` demo |
| [`trick_example.nyan`](examples/trick_example.nyan) | `pose`, `groom`, `self` demo |
| [`file_io.nyan`](examples/file_io.nyan) | File I/O with `nab "file"` |
| [`http_example.nyan`](examples/http_example.nyan) | HTTP client with `nab "http"` |

## Project Structure

```
meow/
├── main.go                  # CLI entry point
├── cmd/meow/                # Alternative CLI entry
├── compiler/                # Pipeline orchestration + E2E tests
├── pkg/
│   ├── token/               # Token types, keywords, positions
│   ├── lexer/               # iter.Seq-based tokenizer
│   ├── ast/                 # AST node definitions + tree walker
│   ├── parser/              # Pratt parser (iter.Pull)
│   ├── checker/             # Type checker (gradual typing)
│   └── codegen/             # AST → Go source generation
├── runtime/
│   ├── meowrt/              # Core: Value, operators, builtins
│   ├── file/                # File I/O (nab "file")
│   ├── http/                # HTTP client (nab "http")
│   └── testing/             # Test framework (nab "testing")
├── examples/                # Sample .nyan programs
└── testdata/                # Golden file tests
```

## Contributing

Contributions are welcome! Whether it's a bug fix, new feature, or just a better cat pun — we'd love your help.

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes
6. Push to the branch and open a Pull Request

### Ideas for Contributions

- [ ] More cat-themed error messages
- [ ] String interpolation (`"Hello, {name}!"`)
- [ ] `flaunt` (export) for multi-file support
- [ ] REPL mode (`meow repl`)
- [ ] Syntax highlighting for popular editors
- [x] Homebrew formula
- [ ] Playground website

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Update golden files
go test ./compiler/ -update
```

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<p align="center">
<pre align="center">
  /\_/\
 ( o.o )  Made with 🐱 and Go
  > ^ <
</pre>
</p>

<p align="center">
  <sub>If this project made you smile, consider giving it a ⭐</sub>
</p>
