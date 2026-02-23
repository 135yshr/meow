# Meow Language Reference

A complete reference of all keywords, operators, and syntax in the Meow language.

## Keywords

| Meow | Meaning | Example |
|------|---------|---------|
| `nyan` | Variable declaration | `nyan x = 42` |
| `meow` | Function definition | `meow add(a, b) { bring a + b }` |
| `bring` | Return a value | `bring x + 1` |
| `sniff` | Conditional branch (if) | `sniff (x > 0) { ... }` |
| `scratch` | Else branch | `} scratch { ... }` |
| `purr` | Loop (while) | `purr (i < 10) { ... }` |
| `paw` | Lambda (anonymous function) | `paw(x) { x * 2 }` |
| `nya` | Print values | `nya("Hello!")` |
| `lick` | Transform each element in a list (map) | `lick(nums, paw(x) { x * 2 })` |
| `picky` | Select elements matching a condition (filter) | `picky(nums, paw(x) { x > 0 })` |
| `curl` | Combine a list into a single value (reduce) | `curl(nums, 0, paw(a, x) { a + x })` |
| `peek` | Branch based on a value (pattern match) | `peek(v) { 0 => "zero", _ => "other" }` |
| `hiss` | Raise an error | `hiss("something went wrong")` |
| `fetch` | Import *(not yet implemented)* | --- |
| `flaunt` | Export *(not yet implemented)* | --- |
| `yarn` | True (boolean literal) | `nyan ok = yarn` |
| `hairball` | False (boolean literal) | `nyan ng = hairball` |
| `catnap` | Nil (represents no value) | `nyan nothing = catnap` |

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
| `&&` | Logical AND | `x > 0 && x < 10` |
| `\|\|` | Logical OR | `x == 0 \|\| x == 1` |
| `!` | Logical NOT | `!ok` |

### Special

| Operator | Meaning | Example |
|----------|---------|---------|
| `\|=\|` | Pipe (chain operations) | `nums \|=\| lick(double)` |
| `..` | Range | `1..10` |
| `=>` | Match arm separator | `0 => "zero"` |
| `=` | Assignment | `nyan x = 1` |

## Literals

| Type | Example | Description |
|------|---------|-------------|
| Integer | `42` | Decimal integer |
| Float | `3.14` | Floating-point number |
| String | `"Hello, world!"` | Double-quoted, `\\` for escape |

## Delimiters

| Symbol | Meaning | Example |
|--------|---------|---------|
| `(` `)` | Function call / grouping | `add(1, 2)` |
| `{` `}` | Block | `meow f() { ... }` |
| `[` `]` | List / index access | `[1, 2, 3]`, `nums[0]` |
| `,` | Separator | `add(a, b)` |

## Comments

```
# Line comment

-~ Block comment
   can span multiple lines ~-
```

## Syntax Examples

### Variable Declaration

```
nyan x = 42
nyan greeting = "Hello!"
nyan pi = 3.14
nyan cats_are_great = yarn
nyan nothing = catnap
```

### Function Definition

```
meow add(a, b) {
  bring a + b
}

nya(add(1, 2))   # => 3
```

### Conditionals

```
sniff (x > 0) {
  nya("positive")
} scratch sniff (x == 0) {
  nya("zero")
} scratch {
  nya("negative")
}
```

### Loops

```
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

### List Operations

```
nyan nums = [1, 2, 3, 4, 5]

lick(nums, paw(x) { x * 2 })           # => [2, 4, 6, 8, 10]
picky(nums, paw(x) { x % 2 == 0 })     # => [2, 4]
curl(nums, 0, paw(acc, x) { acc + x })  # => 15
```

### Pipe

```
nyan double = paw(x) { x * 2 }
nums |=| lick(double)
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
