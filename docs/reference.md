# Meow Language Reference

A complete reference of all keywords, operators, and syntax in the Meow language.

## Keywords

| Meow | Meaning | Example |
|------|---------|---------|
| `nyan` | Variable declaration | `nyan x = 42` |
| `meow` | Function definition | `meow add(a int, b int) int { bring a + b }` |
| `bring` | Return a value | `bring x + 1` |
| `sniff` | Conditional branch (if) | `sniff (x > 0) { ... }` |
| `scratch` | Else branch | `} scratch { ... }` |
| `purr` | Loop (count, range, list, or condition) | `purr i (10) { ... }`, `purr (ready) { ... }` |
| `bolt` | Leave the loop | `sniff (found) { bolt }` |
| `slink` | On to the next turn | `sniff (empty) { slink }` |
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
| `breed` | Type alias (transparent) | `breed Nickname = string` |
| `collar` | Newtype (nominal wrapper) | `collar UserId = int` |
| `pose` | Interface definition | `pose Showable { meow show() string }` |
| `groom` | Add methods to a type | `groom Cat { meow show() string { ... } }` |
| `self` | Instance reference in methods | `self.name` |
| `trill` | Pure-function modifier (opt-in purity check) | `trill meow add(a int, b int) int { bring a + b }` |

## Type Keywords

Meow supports gradual static typing. Type keywords annotate variables, function
parameters, and return values.

Annotations are **optional on variables**, where the type is inferred from the
initializer, but **required on function signatures**: every parameter of a
`meow` function must have a type, and a function containing `bring` must
declare its return type. Omitting either is a compile error:

```text
Hiss! Parameter "a" of function add must have a type annotation, nya~
Hiss! Function add has bring statements but no return type annotation, nya~
```

Grouped parameters satisfy the requirement without repeating the type: in
`meow add(a, b int) int`, `a` takes the type of the next parameter that has one.

Lambdas are the exception — `paw` parameters may be left unannotated, and their
result type is inferred.

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
| `=` | Bind a name (bindings are immutable) | `nyan x = 1` |

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
| String | `"Hello, world!"` | Double-quoted. Escapes: `\"` `\\` `\n` `\t` `\r` |
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

The `nyan` keyword may be omitted, in which case `x = 42` declares `x` just as
`nyan x = 42` does.

**Bindings are immutable.** There is no assignment: a name is bound once, and
`x = ...` against a name already bound in an enclosing scope is a compile error
rather than a reassignment.

```meow
nyan total = 0
purr i (5) {
  total = total + 1   # Hiss! Variable total is already bound and cannot be
                      # reassigned — bindings are immutable, nya~
}
```

Accumulate with `curl` instead of updating a variable in a loop:

```meow
nyan total = curl([1, 2, 3, 4, 5], 0, paw(acc, x) { acc + x })
nya(total)   # => 15
```

An explicit `nyan` inside an inner scope still shadows deliberately, since that
declares a new, separate binding.

### Function Definition

```meow
meow add(a int, b int) int {
  bring a + b
}

nya(add(1, 2))   # => 3
```

### Pure Functions (Trill)

Prefix a function with `trill` to opt into a purity check. Inside a `trill`
function, the body may only call:

- other `trill` functions, and
- side-effect-free builtins: arithmetic/comparison operators, `len`, `to_int`,
  `to_float`, `to_string`, `to_bytes`, `to_runes`, `is_furball`, `head`, `tail`,
  `append`, `lick`, `picky`, `curl`, `whiff`, `track`, `shred`, `tangle`,
  `nibble`.

Any other call is a compile error: impure builtins (`nya`, `hiss`, `gag`),
imported-package members (e.g. `http.pounce(...)`), and non-`trill` user
functions are all rejected. Higher-order calls are allowed, but lambda bodies
are scanned recursively, so an impure lambda passed to `lick` is still caught.

```meow
trill meow add(a int, b int) int {
  bring a + b
}
```

A non-`trill` function referenced as a bare value — bound to a variable, passed
as an argument, or returned — is rejected too, so an impure function can't slip
out of a pure body without being called.

Both rules are about the declaration a name reaches. A pure body may hold a pure
local of its own where an impure function happens to share the name.

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

### Type Alias (Breed)

```meow
breed Nickname = string
breed Score = int

nyan name Nickname = "Nyantyu"
nya(name + " chan")   # => Nyantyu chan — fully compatible with string
```

`breed` creates a transparent alias. The alias and the original type are interchangeable in all contexts including arithmetic, comparisons, and function calls. `breed` is erased at compile time — no runtime cost.

### Newtype (Collar)

```meow
collar UserId = int
collar Email = string

nyan id = UserId(42)
nyan email = Email("nyantyu@meow.cat")

nya(id)           # => UserId{value: 42}
nya(id.value)     # => 42
nya(email.value)  # => nyantyu@meow.cat
```

`collar` creates a distinct wrapper type. Values are constructed with `TypeName(value)` and unwrapped with `.value`. Different collar types with the same underlying type are **not** interchangeable:

```meow
collar Celsius = int
collar Fahrenheit = int

nyan c = Celsius(100)
nyan f = Fahrenheit(212)
# c != f — different collar types are never equal
```

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

A lambda body may also be a block of statements, so control flow is available
inside `paw`. A trailing expression is the result, just as in the
single-expression form, and `bring` returns from the lambda:

```meow
nyan classify = paw(n) {
  sniff (n > 10) { bring "big" } scratch { bring "small" }
}
nya(classify(50))   # => big

nyan area = paw(w, h) {
  nyan a = w * h
  a + 1
}
nya(area(3, 4))     # => 13
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
nab "env"
nab "clock"
nab "json"
nab "random"
nab "testing"
```

After importing, call package functions with `package.function()` syntax:

```meow
nab "file"
nyan content = file.snoop("data.txt")
nya(content)
```

Available packages: `clock`, `env`, `file`, `http`, `json`, `random`, `testing`. See [stdlib.md](stdlib.md) for details.

With `go`, the string is a Go import path rather than one of Meow's own
packages, and any Go package can be reached without a wrapper written for it:

```meow
nab go "strings"
nab go "net/url" tag u                              # name it yourself
nab go "github.com/Masterminds/semver/v3@v3.3.1"    # pin a version

nya(strings.to_upper("nyan"))
nyan parsed = u.parse("https://example.com/a/b")
nya(parsed["host"])
```

The package is called by the last element of its path, or by `tag` when that
is not a name a program can write. Go names are spelled the way Meow writes
names — `strings.to_upper` is `strings.ToUpper` — and a name holding an
initialism is written as Go writes it, `strings.ToValidUTF8`. Without `@` the
version is the toolchain's choice. See [spec.md](spec.md#importing-a-go-package)
for what comes back and how it is read.

### Member Access

The `.` operator accesses fields on `kitty` instances, calls methods defined by `groom`, and calls functions on imported packages:

```meow
# Kitty field access
nyan nyantyu = Cat("Nyantyu", 3)
nya(nyantyu.name)   # => Nyantyu

# Method call (defined by groom)
nya(nyantyu.show())  # => Nyantyu (age 3)

# Package function call
nab "http"
http.pounce("https://example.com") |=| nya
```

A member may be named after a keyword, since nothing but a member can follow a
dot — a Go method named `String` is written `.string()`.

A member read rather than called is a value of its own. A field is what it
holds; a method is that method bound to what it was reached through, which is a
function like any other:

```meow
nyan tell = nyantyu.show      # not called: the method, bound to nyantyu
nya(tell())                   # => Nyantyu (age 3)
nya(lick([nyantyu], tell))
```

The same holds for a named function, which can be kept, piped into, or mapped
over a list:

```meow
meow double(n int) int { bring n * 2 }

nyan twice = double
nya(twice(21))                # => 42
nya(3 |=| double)             # => 6
nya(lick([1, 2, 3], double))  # => [2, 4, 6]
```

A binding takes a name over for as long as it is in scope, whether the name is
read or called, so a local of the same name is the one that is reached.

### Interface (Pose)

Define an interface with `pose` — a named set of method signatures:

```meow
pose Showable {
    meow show() string
}
```

A type structurally satisfies a `pose` if it has all required methods with matching signatures. No explicit declaration is needed (structural typing).

### Methods (Groom)

Add methods to a `kitty` or `collar` type with `groom`:

```meow
kitty Cat {
  name: string,
  age: int
}

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
nya(c.is_kitten())  # => hairball
```

`self` refers to the instance the method is called on. For `kitty` types, use `self.field` to access fields. For `collar` types, use `self.value` to access the wrapped value:

```meow
collar Label = string

groom Label {
    meow display() string {
        bring "[ " + self.value + " ]"
    }
}

nyan tag = Label("important")
nya(tag.display())   # => [ important ]
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
