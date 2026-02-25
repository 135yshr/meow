---
title: "Language Specification"
description: "Formal definition of Meow syntax and semantics"
weight: 1
---

This document defines the syntax and semantics of the Meow programming language.

Meow is a cat-themed programming language that transpiles to Go. Source files use the `.nyan` extension and are encoded in UTF-8.

## Notation

This specification uses Extended Backus-Naur Form (EBNF) for grammar productions:

```ebnf
Production  = name "=" Expression "." .
Expression  = Term { "|" Term } .
Term        = Factor { Factor } .
Factor      = name | literal | "(" Expression ")" | "[" Expression "]" | "{" Expression "}" .
```

- `[ ... ]` denotes optional (0 or 1).
- `{ ... }` denotes repetition (0 or more).
- `" ... "` denotes terminal symbols.

## Source Code Representation

Source code is UTF-8 encoded text in `.nyan` files. Newlines serve as statement terminators (semicolons are not used). The compiler processes a single `.nyan` file at a time.

## Lexical Elements

### Comments

Two forms of comments:

```ebnf
LineComment  = "#" { any_char } newline .
BlockComment = "-~" { any_char } "~-" .
```

Line comments start with `#` and extend to the end of the line. Block comments start with `-~` and end with `~-`, and may span multiple lines. Comments are treated as whitespace by the parser.

### Keywords

The following 19 identifiers are reserved as keywords:

```text
nyan      meow      bring     sniff     scratch
purr      paw       nya       lick      picky
curl      peek      hiss      fetch     flaunt
catnap    yarn      hairball  kitty
```

### Type Keywords

The following 6 identifiers are reserved as type keywords:

```text
int    float    string    bool    furball    list
```

### Identifiers

```ebnf
identifier = letter { letter | digit | "_" } .
letter     = "a" ... "z" | "A" ... "Z" | "_" .
digit      = "0" ... "9" .
```

Identifiers name variables, functions, types, and struct fields. By convention, all user-facing identifiers use `snake_case`.

### Integer Literals

```ebnf
int_lit = digit { digit } .
```

Integer literals are sequences of decimal digits representing 64-bit signed integers.

### Float Literals

```ebnf
float_lit = digit { digit } "." digit { digit } .
```

Float literals contain a decimal point and represent 64-bit IEEE 754 floating-point numbers.

### String Literals

```ebnf
string_lit = '"' { char | escape } '"' .
escape     = "\" ( '"' | "\" | "n" | "t" | "r" ) .
```

String literals are enclosed in double quotes. Supported escape sequences: `\"`, `\\`, `\n`, `\t`, `\r`.

### Operators and Delimiters

```text
Operators:
  +    -    *    /    %
  =    ==   !=   <    >    <=   >=
  &&   ||   !
  |=|  ~>   .    ..   =>

Delimiters:
  (    )    {    }    [    ]    ,    :
```

## Types

Meow uses a gradual type system. Values are dynamically typed at runtime (boxed as `Value`), but optional static type annotations enable compile-time checking and optimized code generation.

### Primitive Types

| Type | Description | Examples |
|------|-------------|---------|
| `int` | 64-bit signed integer | `42`, `-7`, `0` |
| `float` | 64-bit floating-point | `3.14`, `-0.5` |
| `string` | UTF-8 text | `"hello"`, `""` |
| `bool` | Boolean | `yarn` (true), `hairball` (false) |

### Special Types

| Type | Description |
|------|-------------|
| `furball` | Error value carrying a message string |
| `catnap` | The nil/null value (singleton) |

### Composite Types

| Type | Description | Syntax |
|------|-------------|--------|
| `list` | Ordered collection of values | `[1, 2, 3]` |
| Map | String-keyed dictionary | `{"key": value}` |
| `kitty` | User-defined struct | `kitty Name { field: type }` |

### Type Annotations

Type annotations are optional but recommended. They appear after identifiers:

```ebnf
TypeExpr = "int" | "float" | "string" | "bool" | "furball" | "list" .
```

Variable declaration with type:

```ebnf
VarStmt = "nyan" identifier [ TypeExpr ] "=" Expr .
```

Function with typed parameters and return type:

```ebnf
FuncStmt = "meow" identifier "(" [ ParamList ] ")" [ TypeExpr ] Block .
ParamList = Param { "," Param } .
Param = identifier [ TypeExpr ] .
```

Go-style grouped types propagate right-to-left: in `(a, b int)`, both `a` and `b` receive type `int`.

## Expressions

### Literal Expressions

```ebnf
IntLit     = int_lit .
FloatLit   = float_lit .
StringLit  = string_lit .
BoolLit    = "yarn" | "hairball" .
NilLit     = "catnap" .
ListLit    = "[" [ Expr { "," Expr } [ "," ] ] "]" .
MapLit     = "{" [ MapEntry { "," MapEntry } [ "," ] ] "}" .
MapEntry   = StringLit ":" Expr .
```

### Identifier Expression

```ebnf
Ident = identifier .
```

Evaluates to the value bound to the identifier in the current scope.

### Unary Expressions

```ebnf
UnaryExpr = ( "-" | "!" ) Expr .
```

- `-` negates an `int` or `float`.
- `!` inverts truthiness.

### Binary Expressions

```ebnf
BinaryExpr = Expr op Expr .
```

Arithmetic operators (`+`, `-`, `*`, `/`, `%`) require operands of the same type. `+` also concatenates strings.

Comparison operators (`<`, `>`, `<=`, `>=`) work on `int` and `float`. Equality operators (`==`, `!=`) work on all types.

Logical operators (`&&`, `||`) use short-circuit evaluation. `&&` returns the left operand if falsy, otherwise the right. `||` returns the left operand if truthy, otherwise the right.

### Call Expression

```ebnf
CallExpr = Expr "(" [ Expr { "," Expr } ] ")" .
```

Calls a function, lambda, or builtin. Also used to construct `kitty` instances by calling the type name.

### Lambda Expression

```ebnf
LambdaExpr = "paw" "(" [ ParamList ] ")" "{" Expr "}" .
```

Creates an anonymous function. The body is a single expression (not a block of statements).

```meow
paw(x int) { x * 2 }
```

### Index Expression

```ebnf
IndexExpr = Expr "[" Expr "]" .
```

Accesses a list element by zero-based index.

### Member Expression

```ebnf
MemberExpr = Expr "." identifier .
```

Accesses a field on a `kitty` instance, or calls a function in an imported package.

### Pipe Expression

```ebnf
PipeExpr = Expr "|=|" Expr .
```

Passes the left expression as the first argument to the right expression. If the right side is a function call, the left value is prepended to its arguments:

```meow
x |=| f(y)    # equivalent to f(x, y)
x |=| f       # equivalent to f(x)
```

### Catch Expression

```ebnf
CatchExpr = Expr "~>" Expr .
```

If the left expression panics, the right side is used as a fallback. If the right side is a function, it receives the `Furball` error as its argument:

```meow
risky() ~> 0                    # fallback value
risky() ~> paw(err) { handle(err) }  # handler function
```

### Match Expression

```ebnf
MatchExpr = "peek" "(" Expr ")" "{" { MatchArm } "}" .
MatchArm  = Pattern "=>" Expr [ "," ] .
Pattern   = LitPattern | RangePattern | WildcardPattern .
LitPattern      = Expr .
RangePattern    = Expr ".." Expr .
WildcardPattern = "_" .
```

Evaluates the subject and tests it against each pattern in order. Returns the body of the first matching arm.

```meow
peek(n) {
  0 => "zero",
  1..10 => "low",
  _ => "other"
}
```

## Statements

### Variable Declaration

```ebnf
VarStmt = "nyan" identifier [ TypeExpr ] "=" Expr newline .
```

Declares a variable and binds it to a value.

```meow
nyan x int = 42
nyan name = "Nyantyu"
```

### Reassignment

```ebnf
AssignStmt = identifier "=" Expr newline .
```

Rebinds an existing variable to a new value.

### Function Declaration

```ebnf
FuncStmt = "meow" identifier "(" [ ParamList ] ")" [ TypeExpr ] Block .
Block    = "{" { Stmt } "}" .
```

Declares a named function. Functions that don't explicitly `bring` a value implicitly return `catnap`.

```meow
meow greet(name string) string {
  bring "Hello, " + name + "!"
}
```

### Return Statement

```ebnf
ReturnStmt = "bring" [ Expr ] newline .
```

Returns a value from the enclosing function.

### Conditional Statement

```ebnf
IfStmt = "sniff" "(" Expr ")" Block [ "scratch" ( IfStmt | Block ) ] .
```

Evaluates the condition. If truthy, executes the body. Optional `scratch` provides else/else-if branches.

```meow
sniff (x > 0) {
  nya("positive")
} scratch sniff (x == 0) {
  nya("zero")
} scratch {
  nya("negative")
}
```

### Loop Statement

```ebnf
RangeStmt = "purr" identifier "(" RangeExpr ")" Block .
RangeExpr = Expr [ ".." Expr ] .
```

Two forms:

- **Count form**: `purr i (n)` — iterates `i` from `0` to `n-1`.
- **Range form**: `purr i (a..b)` — iterates `i` from `a` to `b` (inclusive).

```meow
purr i (5) { nya(i) }         # 0, 1, 2, 3, 4
purr i (1..5) { nya(i) }     # 1, 2, 3, 4, 5
```

### Fetch Statement

```ebnf
FetchStmt = "fetch" string_lit newline .
```

Imports a standard library package. Available packages: `"file"`, `"http"`, `"testing"`.

```meow
fetch "http"
```

### Kitty Statement

```ebnf
KittyStmt  = "kitty" identifier "{" { KittyField } "}" .
KittyField = identifier ":" TypeExpr [ "," ] newline .
```

Defines a struct type with named, typed fields. A constructor function with the same name is automatically created.

```meow
kitty Point {
  x: int
  y: int
}

nyan p = Point(3, 7)
nya(p.x)   # => 3
```

### Expression Statement

```ebnf
ExprStmt = Expr newline .
```

Any expression can appear as a statement. The result is discarded.

## Built-in Functions

### I/O

| Function | Signature | Description |
|----------|-----------|-------------|
| `nya` | `nya(args...)` | Print values (space-separated) with trailing newline |

### Error Handling

| Function | Signature | Description |
|----------|-----------|-------------|
| `hiss` | `hiss(args...)` | Raise error — panics with `"Hiss! ..."` |
| `gag` | `gag(fn)` → value \| furball | Call `fn()`; recover from panic, return `Furball` on error |
| `is_furball` | `is_furball(v)` → bool | Check if `v` is a `Furball` error value |

### Type Conversion

| Function | Signature | Description |
|----------|-----------|-------------|
| `to_int` | `to_int(v)` → int | Convert float or bool to int |
| `to_float` | `to_float(v)` → float | Convert int to float |
| `to_string` | `to_string(v)` → string | Convert any value to its string representation |

### Collections

| Function | Signature | Description |
|----------|-----------|-------------|
| `len` | `len(v)` → int | Length of string or list |
| `head` | `head(list)` → value | First element of a list |
| `tail` | `tail(list)` → list | All elements except the first |
| `append` | `append(list, value)` → list | New list with value appended |

### Functional Operations

| Function | Signature | Description |
|----------|-----------|-------------|
| `lick` | `lick(list, fn)` → list | Map: apply `fn` to each element |
| `picky` | `picky(list, fn)` → list | Filter: keep elements where `fn` returns truthy |
| `curl` | `curl(list, init, fn)` → value | Reduce: fold list with accumulator |

## Error Model

Errors in Meow use a panic/recover model:

1. **Raising errors**: `hiss("message")` panics with the message `"Hiss! message"`. Error messages are suffixed with `", nya~"` when raised from runtime functions.

2. **Error values**: When a panic is caught, it becomes a `Furball` — a value that carries the error message string.

3. **Catching errors**: Three mechanisms:
   - `gag(fn)` — calls `fn()` and catches panics, returning a `Furball` on error.
   - `expr ~> fallback` — evaluates `expr`; if it panics, uses `fallback` instead.
   - `expr ~> paw(err) { ... }` — evaluates `expr`; if it panics, calls the handler with the `Furball`.

4. **Checking errors**: `is_furball(v)` returns `yarn` if `v` is a `Furball`, `hairball` otherwise.

## Program Structure

A Meow program is a single `.nyan` file containing a sequence of top-level statements. The generated Go code follows this structure:

```go
package main

import meow "github.com/135yshr/meow/runtime/meowrt"
import meow_file "github.com/135yshr/meow/runtime/file"    // from fetch "file"
import meow_http "github.com/135yshr/meow/runtime/http"    // from fetch "http"

// user-defined functions

func main() {
    // top-level statements
}
```

## Truthiness

All values have a truthiness used by `sniff` conditions and logical operators:

| Value | Truthy? |
|-------|---------|
| `yarn` | yes |
| `hairball` | no |
| `catnap` | no |
| `0` (int) | no |
| `0.0` (float) | no |
| `""` (empty string) | no |
| non-zero int/float | yes |
| non-empty string | yes |
| list | yes |
| map | yes |
| kitty | yes |
| furball | yes |
| func | yes |
