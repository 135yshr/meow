---
title: "Quick Reference"
description: "Compact keyword and operator reference card"
weight: 3
---

A complete reference of all keywords, operators, and syntax in the Meow language.

## Keywords

| Meow | Meaning | Example |
|------|---------|---------|
| `nyan` | Variable declaration | `nyan x = 42` |
| `meow` | Function definition | `meow add(a int, b int) int { bring a + b }` |
| `bring` | Return a value | `bring x + 1` |
| `sniff` | Conditional branch (if) | `sniff (x > 0) { ... }` |
| `scratch` | Else branch | `} scratch { ... }` |
| `purr` | Loop (range-based) | `purr i (10) { ... }` |
| `paw` | Lambda (anonymous function) | `paw(x int) { x * 2 }` |
| `nya` | Print values | `nya("Hello!")` |
| `lick` | Transform each element in a list (map) | `lick(nums, paw(x) { x * 2 })` |
| `picky` | Select elements matching a condition (filter) | `picky(nums, paw(x) { x > 0 })` |
| `curl` | Combine a list into a single value (reduce) | `curl(nums, 0, paw(a, x) { a + x })` |
| `peek` | Branch based on a value (pattern match) | `peek(v) { 0 => "zero", _ => "other" }` |
| `hiss` | Raise an error | `hiss("something went wrong")` |
| `gag` | Catch errors (try/recover) | `gag(paw() { risky() })` |
| `is_furball` | Check if a value is an error | `is_furball(result)` |
| `nab` | Import standard library package | `nab "http"` |
| `flaunt` | Export *(reserved)* | --- |
| `yarn` | True (boolean literal) | `nyan ok = yarn` |
| `hairball` | False (boolean literal) | `nyan ng = hairball` |
| `catnap` | Nil (represents no value) | `nyan nothing = catnap` |
| `kitty` | Struct (composite type) definition | `kitty Cat { name: string }` |

## Type Keywords

Meow supports gradual static typing. Type keywords can annotate variables, function parameters, and return values.

| Type | Meaning | Example |
|------|---------|---------|
| `int` | 64-bit signed integer | `nyan x int = 42` |
| `float` | 64-bit floating-point | `nyan pi float = 3.14` |
| `string` | UTF-8 string | `nyan name string = "Nyantyu"` |
| `bool` | Boolean | `nyan ok bool = yarn` |
| `furball` | Error value | `paw(err furball) { ... }` |
| `litter` | List of values | `nyan nums litter = [1, 2, 3]` |

### Type Annotation Syntax

Variables:

```meow
nyan x int = 42
nyan name string = "Nyantyu"
```

Function parameters and return type:

```meow
meow add(a int, b int) int {
  bring a + b
}
```

Go-style grouped parameter types — parameters without a type receive the type of the next parameter with a type:

```meow
meow add(a, b int) int {
  bring a + b
}
# a and b are both int
```

## Operators

### Arithmetic

| Operator | Meaning | Example |
|----------|---------|---------|
| `+` | Addition / string concatenation | `1 + 2`, `"a" + "b"` |
| `-` | Subtraction / unary negation | `5 - 3`, `-x` |
| `*` | Multiplication | `2 * 3` |
| `/` | Division | `10 / 2` |
| `%` | Modulo | `10 % 3` |

### Comparison

| Operator | Meaning | Example |
|----------|---------|---------|
| `==` | Equal | `x == 0` |
| `!=` | Not equal | `x != 0` |
| `<` | Less than | `x < 10` |
| `>` | Greater than | `x > 0` |
| `<=` | Less than or equal | `x <= 100` |
| `>=` | Greater than or equal | `x >= 1` |

### Logical

| Operator | Meaning | Example |
|----------|---------|---------|
| `&&` | Logical AND (short-circuit) | `x > 0 && x < 10` |
| `\|\|` | Logical OR (short-circuit) | `x == 0 \|\| x == 1` |
| `!` | Logical NOT | `!ok` |

### Special

| Operator | Meaning | Example |
|----------|---------|---------|
| `\|=\|` | Pipe (chain operations) | `nums \|=\| lick(double)` |
| `~>` | Error recovery (catch) | `divide(10, 0) ~> 0` |
| `.` | Member access | `cat.name`, `file.snoop("x")` |
| `..` | Range (inclusive) | `1..10` |
| `=>` | Match arm separator | `0 => "zero"` |
| `=` | Assignment | `nyan x = 1` |

### Operator Precedence

From lowest to highest:

| Precedence | Operators | Description |
|-----------|-----------|-------------|
| 1 (lowest) | `~>` | Error recovery |
| 2 | `\|\|` | Logical OR |
| 3 | `&&` | Logical AND |
| 4 | `==` `!=` | Equality |
| 5 | `<` `>` `<=` `>=` | Comparison |
| 6 | `\|=\|` | Pipe |
| 7 | `+` `-` | Addition, subtraction |
| 8 | `*` `/` `%` | Multiplication, division, modulo |
| 9 | `!` `-` (unary) | Unary operators |
| 10 (highest) | `()` `[]` `.` | Call, index, member access |

## Literals

| Type | Example | Description |
|------|---------|-------------|
| Integer | `42` | Decimal integer |
| Float | `3.14` | Floating-point number |
| String | `"Hello, world!"` | Double-quoted, `\\` for escape |
| List | `[1, 2, 3]` | Ordered collection |
| Map | `{"key": "value"}` | String-keyed dictionary |

## Delimiters

| Symbol | Meaning | Example |
|--------|---------|---------|
| `(` `)` | Function call / grouping | `add(1, 2)` |
| `{` `}` | Block / map literal | `meow f() { ... }` |
| `[` `]` | List / index access | `[1, 2, 3]`, `nums[0]` |
| `,` | Separator | `add(a, b)` |
| `:` | Type annotation / map key-value | `name: string`, `{"k": v}` |

## Comments

```meow
# Line comment

-~ Block comment
   can span multiple lines ~-
```

## Syntax Examples

### Variable Declaration

```meow
nyan x = 42
nyan greeting = "Hello!"
nyan pi float = 3.14
nyan cats_are_great = yarn
nyan nothing = catnap
```

### Function Definition

```meow
meow add(a int, b int) int {
  bring a + b
}

nya(add(1, 2))   # => 3
```

### Struct (Kitty) Definition

```meow
kitty Cat {
  name: string
  age: int
}

nyan nyantyu = Cat("Nyantyu", 3)
nya(nyantyu)            # => Cat{name: Nyantyu, age: 3}
nya(nyantyu.name)       # => Nyantyu
nya(nyantyu.age)        # => 3
```

Fields are defined with `name: type` syntax. Instances are created by calling the type name as a constructor. Fields are accessed with `.` notation.

### Conditionals

```meow
sniff (x > 0) {
  nya("positive")
} scratch sniff (x == 0) {
  nya("zero")
} scratch {
  nya("negative")
}
```

### Loops

Count form — iterates from `0` to `n-1`:

```meow
purr i (10) {
  nya(i)
}
# prints 0, 1, 2, ..., 9
```

Range form — iterates from `a` to `b` inclusive:

```meow
purr i (1..20) {
  nya(i)
}
# prints 1, 2, 3, ..., 20
```

### Error Handling

Use `hiss` to raise an error and stop execution. The error message
is prefixed with `Hiss!` automatically.

```meow
meow divide(a int, b int) int {
  sniff (b == 0) {
    hiss("division by zero")
  }
  bring a / b
}

nya(divide(10, 0))   # => Hiss! division by zero
```

Multiple arguments are joined with spaces:

```meow
hiss("unexpected value:", x)
```

Use `gag` to catch errors. Wrap risky code in a lambda and pass it
to `gag`. If the code panics, `gag` returns a `Furball` (error value)
instead of crashing. Use `is_furball` to check if a value is an error.

```meow
nyan result = gag(paw() { divide(10, 0) })

sniff (is_furball(result)) {
  nya("caught:", result)
} scratch {
  nya("ok:", result)
}
# => caught: Hiss! division by zero
```

If the code succeeds, `gag` returns its result normally:

```meow
nyan ok = gag(paw() { divide(10, 2) })
nya(ok)   # => 5
```

### Error Recovery with `~>`

The `~>` (cat tail arrow) operator provides concise error recovery.
If the left-hand expression panics, the fallback on the right is used instead.
The `~` resembles a cat's tail, and `>` points to the fallback.

```meow
nyan val = divide(10, 0) ~> 0
nya(val)   # => 0
```

When no error occurs, the original result is returned:

```meow
nyan val = divide(10, 2) ~> 0
nya(val)   # => 5
```

The fallback can also be a handler function that receives the error:

```meow
nyan val = divide(10, 0) ~> paw(err) { 42 }
nya(val)   # => 42
```

### Lambdas

```meow
nyan double = paw(x int) { x * 2 }
nya(double(5))   # => 10
```

### List Operations

```meow
nyan nums = [1, 2, 3, 4, 5]

lick(nums, paw(x) { x * 2 })           # => [2, 4, 6, 8, 10]
picky(nums, paw(x) { x % 2 == 0 })     # => [2, 4]
curl(nums, 0, paw(acc, x) { acc + x })  # => 15
```

### Map Literals

```meow
nyan headers = {
  "Content-Type": "application/json",
  "Authorization": "Bearer token123"
}
```

### Pipe

```meow
nyan double = paw(x) { x * 2 }
nums |=| lick(double)
```

The pipe operator passes the left value as the first argument of the right expression:

```meow
[1, 2, 3] |=| lick(paw(x) { x * 2 }) |=| nya
# => [2, 4, 6]
```

### Pattern Matching

```meow
nyan result = peek(score) {
  0 => "zero",
  1..10 => "low",
  11..100 => "high",
  _ => "off the charts"
}
```

Patterns can be:
- **Literal** — match a specific value (`0`, `"hello"`, `yarn`)
- **Range** — match an inclusive range (`1..10`)
- **Wildcard** — match anything (`_`)

### Import (Nab)

Use `nab` to import a standard library package:

```meow
nab "file"
nab "http"
nab "testing"
```

After importing, call package functions with `package.function()` syntax:

```meow
nab "file"
nyan content = file.snoop("data.txt")
nya(content)
```

Available packages: `file`, `http`, `testing`. See [stdlib.md](stdlib.md) for details.

### Member Access

The `.` operator accesses fields on `kitty` instances and functions on imported packages:

```meow
# Kitty field access
nyan nyantyu = Cat("Nyantyu", 3)
nya(nyantyu.name)   # => Nyantyu

# Package function call
nab "http"
http.pounce("https://example.com") |=| nya
```

### Testing

Test functions use the `test_` prefix and `catwalk_` prefix:

```meow
nab "testing"

meow test_addition() {
  expect(1 + 1, 2, "basic addition")
}

meow test_string() {
  judge("hello" == "hello")
}
```

Run tests with `meow test`:

```bash
meow test my_test.nyan
```

See [stdlib.md](stdlib.md) for `judge`, `expect`, `refuse`, and other testing functions.
