---
title: "Meow Programming Language Specification: Syntax, Types, Functions, and Pattern Matching"
description: "The official Meow Programming Language specification covering .nyan syntax, types, functions, control flow, pattern matching, and standard behavior."
weight: 1
---

This document defines the syntax and semantics of the Meow Programming Language.

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

The following 27 identifiers are reserved as keywords:

```ebnf
keyword = "nyan"   | "meow"  | "bring"    | "sniff" | "scratch"
        | "purr"   | "paw"   | "nya"      | "lick"  | "picky"
        | "curl"   | "peek"  | "hiss"     | "nab"   | "flaunt"
        | "catnap" | "yarn"  | "hairball" | "kitty" | "breed"
        | "collar" | "pose"  | "groom"    | "self"  | "trill"
        | "bolt"   | "slink" .
```

### Type Keywords

The following 7 identifiers are reserved as type keywords:

```ebnf
type_keyword = "int" | "float" | "string" | "bool" | "furball" | "litter" | "basket" .
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
| `basket` | String-keyed dictionary — keys are string literals | `{"key": value}` |
| `kitty` | User-defined struct | `kitty Name { field: type }` |
| `breed` | Type alias (transparent) | `breed Nickname = string` |
| `collar` | Newtype (nominal wrapper) | `collar UserId = int` |
| `pose` | Interface (method signatures) | `pose Showable { meow show() string }` |

A `litter` written out with one kind of thing in it is a litter of that kind,
and one written with several is a litter of anything — `[1, 2, 3]` holds ints,
`[1, "a"]` holds whatever it holds. Nothing is refused for being mixed; what
changes is only how much is known about an element before the program runs.

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
TypeExpr = type_keyword | identifier .
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
MemberExpr = Expr "." MemberName .
MemberName = identifier | keyword | type_keyword .
```

Accesses a field on a `kitty` instance, calls a method defined by `groom`, or calls a function in an imported package.

A member that is read rather than called is a value of its own. A field is what
it holds; a method — one groomed on, or one belonging to something from Go — is
that method bound to what it was reached through, which is a function like any
other:

```meow
nab go "strings"

nyan swap = strings.new_replacer("cat", "nyan")
nyan speak = swap.replace              # a function, not yet called

nya(speak("the cat"))                  # => the nyan
nya("the cat" |=| swap.replace)        # => the nyan
nya(lick(["cat", "dog"], swap.replace))
```

This is what a language of values has in place of a chain of calls. Rather than
following one call with another, a member is taken as a function and the value
is passed into it, which reads left to right and needs nothing to come after a
closing bracket. A member that is not there fails where it is written rather
than where the function it would have been is called.

A member may be named after a keyword. Nothing but a member can follow a dot,
so there is nothing for `x.string` to be ambiguous with, and a Go method named
`String` is spelled `.string()` — see [Importing a Go package](#importing-a-go-package).
Everywhere a keyword can stand, it is still a keyword: `nyan string = 1` remains
an error.

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

A function named rather than called is the function itself, and can be kept,
piped into, or mapped over a list like any other value. Its arity is known, so
naming one is what a call with too few arguments already is, with none of them
supplied:

```meow
meow double(n int) int { bring n * 2 }

nyan twice = double
nya(twice(21))                 # => 42
nya(3 |=| double)              # => 6
nya(lick([1, 2, 3], double))   # => [2, 4, 6]
```

A binding takes the name over for as long as it is in scope, whatever it holds
and whether the name is read or called. Which declaration a name reaches is
settled where it is written, so a local holding a function of its own shape is
that function, and is checked as one:

```meow
meow double(n int) int { bring n * 2 }

meow rename() string {
  nyan double = paw(a, b) { bring a + "/" + b }
  bring double("x", "y")       # => x/y, and takes two arguments
}
```

#### Nested Functions

A `meow` may be declared inside another. It is visible only within the block it
was written in — a `sniff` or `purr` body has its own scope, as it does for
ordinary bindings — it may read the enclosing scope, and it shadows a top-level
function of the same name.

A nested function may call itself, and may call a sibling written after it, since
those calls happen once both declarations have run. Calling one *before* its
declaration has run is a failure, and reports that the function is undefined.

```meow
meow outer(x int) int {
  meow inner(y int) int { bring x + y }
  bring inner(10)
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

Both rules are about the declaration a name reaches, not the name itself. A
binding takes a name over for as long as it is in scope, so a pure body may hold
one of its own where an impure function happens to share the name:

```meow
meow helper(x int) int {
  nya("side effect")
  bring x + 1
}

trill meow adds(n int) int {
  nyan helper = paw(x) { bring x + 1 }
  bring helper(n)             # the local, and so allowed
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
RangeStmt = "purr" identifier [ "," identifier ] "(" RangeExpr ")" Block .
RangeExpr = Expr [ ".." Expr ] .
WhileStmt = "purr" "(" Expr ")" Block .
BoltStmt  = "bolt" .
SlinkStmt = "slink" .
```

Four forms:

- **Count form**: `purr i (n)` — iterates `i` from `0` to `n-1`.
- **Range form**: `purr i (a..b)` — iterates `i` from `a` to `b` (inclusive).
- **Element form**: `purr x (litter)` — iterates over a litter's elements.
  `purr i, x (litter)` also binds the index. Over a `basket`, `purr k (basket)`
  binds each key and `purr k, v (basket)` binds key and value.
- **Conditional form**: `purr (cond)` — repeats while `cond` holds, tested
  before each turn. It has no loop variable, which is what tells it apart from
  the forms above. As with `sniff`, `cond` must be a `bool`.

```meow
purr i (5) { nya(i) }         # 0, 1, 2, 3, 4
purr i (1..5) { nya(i) }     # 1, 2, 3, 4, 5
purr w (["a", "b"]) { nya(w) }        # a, b
purr i, w (["a", "b"]) { nya(i) }     # 0, 1
purr k ({"a": 1, "b": 2}) { nya(k) }         # a, b
purr k, v ({"a": 1}) { nya(k + to_string(v)) }   # a1
```

`bolt` leaves the loop it is written in; `slink` ends that turn and starts the
next. Both belong to the nearest enclosing `purr`, and neither may be written
outside one — including inside a `paw` or a `meow` declared in a loop body, since
those run when they are called rather than as part of the turn.

```meow
purr x ([1, 2, 3, 4]) {
  sniff (x == 4) { bolt }     # stop here
  sniff (x == 2) { slink }    # skip this one
  nya(x)                       # 1, 3
}
```

Because bindings are immutable, a conditional `purr` cannot count its own way to
a stopping point: its condition is about something the loop does not change — a
reply that has not arrived, a file that is not there yet — and it usually ends
with `bolt` when the body has what it came for. A condition that is already
false means the body never runs, and a failure while working the condition out
ends the program rather than reading as false.

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
NabStmt = "nab" [ "go" ] string_lit [ "tag" identifier ] newline .
```

Imports a standard library package. Available packages: `"clock"`, `"env"`,
`"file"`, `"http"`, `"json"`, `"random"`, `"testing"`.

```meow
nab "http"
```

#### Importing a Go package

With `go`, the string is a Go import path rather than one of Meow's own
packages, and any Go package can be named. Nothing has to be written for it
first: the call is made through the bridge, which reads what Meow has a shape
for and holds what it has not.

```meow
nab go "strings"
nab go "net/url" tag u

nya(strings.to_upper("nyan"))          # => NYAN

nyan parsed = u.parse("https://example.com/a/b")
nya(parsed["host"])                    # => example.com
```

The package is called by the last element of its path — `url` for `"net/url"` —
except that a major-version element belongs to the module rather than the
package, so `"github.com/x/y/v2"` is `y`. Go has such a suffix only from `v2`
up, so `"k8s.io/api/core/v1"` really is called `v1`.

When that leaves a name a program cannot write — `"github.com/aws/aws-sdk-go-v2"`
is not a name, and a package called `nyan` could never be said, since `nyan`
begins a binding wherever it appears — `tag` names it instead:

```meow
nab go "github.com/aws/aws-sdk-go-v2/config" tag cfg
```

Go names are spelled the way Meow writes names: `strings.to_upper` is
`strings.ToUpper`, and `sts.new_from_config` is `sts.NewFromConfig`. A keyword
is a name after a dot like any other — nothing but a member can follow one — so
Go's `String()` is written `.string()`. A name holding an initialism cannot be
spelled this way — `to_valid_utf8` reaches for
`ToValidUtf8`, which is not what Go calls it — so such a name is written as Go
writes it, `strings.ToValidUTF8`. Getting it wrong is a build error naming the
spelling that exists, not a surprise at runtime.

What comes back is read if Meow has a shape for it and held if not. A record
becomes a basket, under the names a Meow program writes; a `time.Time` becomes
its text; a trailing `error` becomes a furball. A client, a connection, a
handle — anything with methods and nothing to read — is held whole, and the
next call is made on it:

```meow
nab go "regexp"

nyan re = regexp.must_compile("[0-9]+")
nya(re.find_string("abc 123 def"))     # => 123
```

What is read also remembers what it was read out of, so reading a value is not
what stops it being one. It is still called on, and still handed to the next
call as itself:

```meow
nab go "net/url" tag u
nab go "time"

nyan p = u.parse("https://example.com/a/b?x=1")
nya(p["host"])                         # => example.com
nya(p.hostname())                      # => example.com

nyan d = time.ParseDuration("90m")
nya(d)                                 # => 5400000000000
nya(d.minutes())                       # => 90
```

Some records are a basket by their shape and a handle by their use —
`aws.Config` has fields worth reading and interface fields that no basket could
be built back into — and this is what lets them be both. Being handed on as
itself rather than as what it read as is also what keeps what the reading does
not say: a `time.Time` goes to the next call down to the nanosecond, not down to
the second its text gives.

Only what was read remembers. A basket a program wrote itself remembers nothing
and is built into the record as before, and a value that is all there — a plain
string, a plain number — remembers nothing either, since there is nothing more
of it to reach.

What was read is a reading, taken when it was taken. A Meow value cannot be
changed once made, and the Go value behind it can be — by a call that is handed
it, or by a method called on it — so after such a call the two say different
things: the Go value has moved on and the basket still holds what was read. The
basket is not wrong when that happens; it is what was there to read at the time,
which is the only thing an unchanging value can be.

A method with nothing of its own to say gives back a fresh reading of what it
was called on, so the doing has somewhere to show. Go writes a great many such
methods — `Set`, `Add`, `Reset` — and what they do is to the thing they are
called on:

```meow
nab go "net/url" tag u

nyan q = u.ParseQuery("x=1&y=2")
nya(q.set("z", "9"))                   # => {x: [1], y: [2], z: [9]}
nya(q)                                 # => {x: [1], y: [2]}
```

The change arrives as a new value beside the old one, which is how a language
whose values do not change says that something happened. A handle is not read,
so there is nothing to read again: the same handle comes back, and one call can
follow another.

A method with nothing to say is one with no answer, which is not the same as one
whose answer is nothing: a search that found nothing still answers `catnap`, and
a method that failed answers with the furball. A trailing `error` is the failure
rather than an answer, so a method returning only an `error` has nothing to say
when it does not fail.

The version is the toolchain's choice unless the program makes it, which it
does with `@`:

```meow
nab go "github.com/aws/aws-sdk-go-v2/aws/arn@v1.32.0"
```

A function taking an empty interface — `fmt.Sprintf`, `json.Marshal` — gets the
plain Go value behind the Meow one, a litter arriving as a slice and a basket as
a map. An interface with methods is another matter: only something held from Go
can satisfy one.

```meow
nab go "fmt"
nab go "encoding/json" tag j

nya(fmt.sprintf("%s has %d", "nyan", 4))
nya(to_string(j.marshal({"name": "nyan"})))   # => {"name":"nyan"}
```

Generics, channels, and functions taking functions are not reached this way. A
Go package is also out of reach in the playground, which has no Go toolchain —
as every `nab` already is.

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

A groomed method read rather than called is that method bound to the instance,
and has the method's own type — see [Member Expression](#member-expression):

```meow
nyan tell = c.show
nya(tell())         # => Nyantyu (age 3)
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
| `scram` | `scram([status])` | End the program with `status` (0–255, default 0); `Furball` outside that range |

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

5. **Reporting errors**: A failure nothing catches ends the program, and is written to standard error prefixed with the position of the statement that was running, in the same `file:line:column` form the compiler's own errors use:

   ```text
   probe.nyan:12:3: Hiss! Cannot read "3 " as an Int, nya~
   ```

   The prefix is on the report only. A `Furball` caught with `gag` or `~>` carries the message alone, so a program that prints or matches one sees what it always did.

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
