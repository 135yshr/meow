# Meow Language Specification

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

The following 25 identifiers are reserved as keywords:

```text
nyan      meow      bring     sniff     scratch
purr      paw       nya       lick      picky
curl      peek      hiss      nab       flaunt
catnap    yarn      hairball  kitty     breed
collar    pose      groom     self      trill
```

### Type Keywords

The following 7 identifiers are reserved as type keywords:

```text
int    float    string    bool    furball    litter    basket
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

Meow uses a gradual type system. Values are dynamically typed at runtime (boxed as `Value`), but static type annotations enable compile-time checking and optimized code generation. They are optional on variables and `paw` parameters, and required on `meow` function signatures (see [Type Annotations](#type-annotations)).

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
| `litter` | Ordered collection of values | `[1, 2, 3]` |
| `basket` | String-keyed dictionary | `{"key": value}` |
| `kitty` | User-defined struct | `kitty Name { field: type }` |
| `breed` | Type alias (transparent) | `breed Nickname = string` |
| `collar` | Newtype (nominal wrapper) | `collar UserId = int` |
| `pose` | Interface (method signatures) | `pose Showable { meow show() string }` |

### Type Alias (breed)

A `breed` declaration creates a transparent alias for an existing type. The alias is fully interchangeable with the original type in all operations.

```ebnf
BreedStmt = "breed" identifier "=" TypeExpr newline .
```

```meow
breed Nickname = string
nyan name Nickname = "Nyantyu"   # string and Nickname are interchangeable
nya(name + " chan")              # string operations work directly
```

`breed` is a compile-time construct only — it leaves no trace in the generated code.

### Newtype (collar)

A `collar` declaration creates a distinct new type that wraps an existing type. Values must be constructed explicitly, and the inner value is accessed via `.value`.

```ebnf
CollarStmt = "collar" identifier "=" TypeExpr newline .
```

```meow
collar UserId = int
nyan id = UserId(42)       # constructor wraps the value
nya(id.value)              # .value unwraps it => 42
nya(id)                    # => UserId{value: 42}
```

Two `collar` types with the same underlying type are distinct:

```meow
collar Temperature = int
collar Humidity = int
nyan temp = Temperature(72)
nyan humid = Humidity(72)
# temp != humid — different collar types are never equal
```

### Type Annotations

Type annotations appear after identifiers. They are optional on variable
declarations and on `paw` parameters, where the type is inferred, and required
on `meow` function signatures: every parameter must have a type, and a function
containing `bring` must declare its return type. Grouped parameters satisfy the
requirement without repeating the type — in `meow add(a, b int)`, `a` takes the
type of the next parameter that has one.

```ebnf
TypeExpr = "int" | "float" | "string" | "bool" | "furball" | "litter" | "basket" | identifier .
```

Variable declaration with type:

```ebnf
VarStmt = "nyan" identifier [ TypeExpr ] "=" Expr .
```

Function with typed parameters and return type:

```ebnf
FuncStmt = [ "trill" ] "meow" identifier "(" [ ParamList ] ")" [ TypeExpr ] Block .
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

Comparison operators (`<`, `>`, `<=`, `>=`) work on `int` and `float`.

Equality operators (`==`, `!=`) require operands of the same type — comparing
two different types, as in `1 == "one"`, is an error. The one exception is
`catnap`: either side may be `catnap` whatever the other side is, because
comparing against it asks "is this missing?", which any value may be asked —
an unset environment variable and an absent map key both answer with it.
`catnap` equals only `catnap`.

```meow
nyan token = env.hunt("API_TOKEN")
sniff (token == catnap) { hiss("API_TOKEN is not set") }

nyan m = {"a": "A"}
nya(m["missing"] == catnap)   # => yarn
```

Logical operators (`&&`, `||`) use short-circuit evaluation. `&&` returns the left operand if falsy, otherwise the right. `||` returns the left operand if truthy, otherwise the right.

### Call Expression

```ebnf
CallExpr = Expr "(" [ Expr { "," Expr } ] ")" .
```

Calls a function, lambda, or built-in. Also used to construct `kitty` instances by calling the type name.

### Lambda Expression

```ebnf
LambdaExpr = "paw" "(" [ ParamList ] ")" "{" ( Expr | { Stmt } ) "}" .
```

Creates an anonymous function. The body is either a single expression, whose
value is the result, or a block of statements.

In a block body, a trailing expression statement is the result — matching the
way the single-expression form yields its value — and `bring` returns from the
lambda rather than from any enclosing function. A block that does neither
evaluates to `catnap`.

```meow
paw(x int) { x * 2 }

paw(n) { sniff (n > 10) { bring "big" } scratch { bring "small" } }

paw(w, h) {
  nyan a = w * h
  a + 1
}
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

Accesses a field on a `kitty` instance, calls a method defined by `groom`, or calls a function in an imported package.

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
FuncStmt = [ "trill" ] "meow" identifier "(" [ ParamList ] ")" [ TypeExpr ] Block .
Block    = "{" { Stmt } "}" .
```

Declares a named function. Functions that don't explicitly `bring` a value implicitly return `catnap`.

```meow
meow greet(name string) string {
  bring "Hello, " + name + "!"
}
```

#### Pure Functions (trill)

Prefixing a declaration with `trill` opts the function into a compile-time purity check. Inside a `trill` function the body may only call other `trill` functions and side-effect-free builtins (arithmetic/comparison operators, `len`, `to_int`, `to_float`, `to_string`, `to_bytes`, `to_runes`, `is_furball`, `head`, `tail`, `append`, `lick`, `picky`, `curl`, `whiff`,
`track`, `shred`, `tangle`, `nibble`). Calling `nya`, `hiss`, `gag`, an imported-package member, or a non-`trill` user function is a compile error. Lambda bodies are scanned recursively, so an impure lambda passed to `lick`/`picky`/`curl` is also rejected.

```meow
trill meow add(a int, b int) int {
  bring a + b
}
```

Referencing a non-`trill` function as a bare value — binding it to a variable, passing it as an argument, or returning it — is also a compile error, so impurity cannot escape a pure body without being called.

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
RangeStmt = "purr" identifier [ "," identifier ] "(" RangeExpr ")" Block .
RangeExpr = Expr [ ".." Expr ] .
```

Three forms:

- **Count form**: `purr i (n)` — iterates `i` from `0` to `n-1`.
- **Range form**: `purr i (a..b)` — iterates `i` from `a` to `b` (inclusive).
- **Element form**: `purr x (litter)` — iterates over a litter's elements.
  `purr i, x (litter)` also binds the index. Over a `basket`, `purr k (basket)`
  binds each key and `purr k, v (basket)` binds key and value.

```meow
purr i (5) { nya(i) }         # 0, 1, 2, 3, 4
purr i (1..5) { nya(i) }     # 1, 2, 3, 4, 5
purr w (["a", "b"]) { nya(w) }        # a, b
purr i, w (["a", "b"]) { nya(i) }     # 0, 1
purr k ({"a": 1, "b": 2}) { nya(k) }         # a, b
purr k, v ({"a": 1}) { nya(k + to_string(v)) }   # a1
```

A `basket` is walked in sorted key order. Go, which the compiler targets,
randomizes map iteration, so walking one in its own order would give a program
output that differed from run to run.

The count and element forms are written the same way, so which one a `purr`
means depends on what its subject turns out to be: a litter is walked element by
element, and anything else is read as a count. When the subject's type is known
ahead of time — a literal, a `litter`-annotated binding — the choice is settled
while compiling; otherwise it is settled when the loop runs. Either way the
answer is the same, so a `purr` over a call's result or a map lookup behaves
like a `purr` over a litter written out in full.

### Nab Statement

```ebnf
NabStmt = "nab" string_lit newline .
```

Imports a standard library package. Available packages: `"aws"`, `"clock"`,
`"env"`, `"file"`, `"http"`, `"json"`, `"random"`, `"testing"`.

```meow
nab "http"
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

### Breed Statement

```ebnf
BreedStmt = "breed" identifier "=" TypeExpr newline .
```

Declares a type alias. See [Type Alias (breed)](#type-alias-breed) above.

### Collar Statement

```ebnf
CollarStmt = "collar" identifier "=" TypeExpr newline .
```

Declares a newtype wrapper. See [Newtype (collar)](#newtype-collar) above.

### Pose Statement

```ebnf
PoseStmt   = "pose" identifier "{" { PoseMethod } "}" .
PoseMethod = "meow" identifier "(" [ ParamList ] ")" [ TypeExpr ] newline .
```

Defines an interface — a named set of method signatures. Types structurally satisfy a pose if they have all required methods with matching signatures.

```meow
pose Showable {
    meow show() string
}
```

A `pose` is a compile-time construct used for structural type checking. It does not generate runtime code.

### Groom Statement

```ebnf
GroomStmt = "groom" identifier "{" { FuncStmt } "}" .
```

Adds methods to an existing `kitty` or `collar` type. Each method is a `meow` function that receives the instance as `self` implicitly.

```meow
kitty Cat { name: string, age: int }

groom Cat {
    meow show() string {
        bring self.name + " (age " + to_string(self.age) + ")"
    }
    meow is_kitten() bool {
        bring self.age < 1
    }
}

nyan c = Cat("Nyantyu", 3)
nya(c.show())       # => Nyantyu (age 3)
```

The `self` keyword refers to the instance the method is called on. For `kitty` types, `self.field` accesses fields. For `collar` types, `self.value` accesses the wrapped value.

### Self Expression

```ebnf
SelfExpr = "self" .
```

Refers to the current instance within a `groom` block. Only valid inside method bodies defined by `groom`.

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
| `to_int` | `to_int(v)` → int | Convert float, bool, or int to int |
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
import meow_file "github.com/135yshr/meow/runtime/file"    // from nab "file"
import meow_http "github.com/135yshr/meow/runtime/http"    // from nab "http"
import meow_testing "github.com/135yshr/meow/runtime/testing" // from nab "testing"

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
| empty list `[]` | no |
| non-empty list | yes |
| empty map `{}` | no |
| non-empty map | yes |
| kitty | yes |
| furball | no |
| func | yes |
