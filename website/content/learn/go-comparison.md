---
title: "Meow vs Go"
description: "Side-by-side reference for Go developers learning Meow"
weight: 3
---

A side-by-side reference for Go developers learning Meow. Meow transpiles to Go, so many concepts map directly.

## Syntax Comparison

| Concept | Go | Meow |
|---------|----|----|
| Variable | `var x int = 42` | `nyan x int = 42` |
| Short variable | `x := 42` | `nyan x = 42` |
| Function | `func add(a, b int) int` | `meow add(a, b int) int` |
| Return | `return x` | `bring x` |
| If | `if x > 0 { }` | `sniff (x > 0) { }` |
| Else | `} else { }` | `} scratch { }` |
| Else if | `} else if x == 0 {` | `} scratch sniff (x == 0) {` |
| For (counting) | `for i := 0; i < n; i++` | `purr i (n)` |
| For (inclusive) | `for i := a; i <= b; i++` | `purr i (a..b)` |
| Lambda | `func(x int) int { return x*2 }` | `paw(x int) { x * 2 }` |
| Print | `fmt.Println(x)` | `nya(x)` |
| True | `true` | `yarn` |
| False | `false` | `hairball` |
| Nil | `nil` | `catnap` |
| Struct | `type Cat struct { ... }` | `kitty Cat { ... }` |
| Import | `import "net/http"` | `nab "http"` |
| Error | `errors.New("msg")` | `hiss("msg")` |
| Panic | `panic("msg")` | `hiss("msg")` |
| Recover | `defer func() { recover() }()` | `gag(paw() { ... })` |

## Detailed Examples

### Variable Declaration

**Go:**
```go
var name string = "Nyantyu"
age := 3
pi := 3.14
isHappy := true
```

**Meow:**
```meow
nyan name string = "Nyantyu"
nyan age = 3
nyan pi = 3.14
nyan is_happy = yarn
```

### Functions

**Go:**
```go
func greet(name string) string {
    return "Hello, " + name + "!"
}
fmt.Println(greet("Nyantyu"))
```

**Meow:**
```meow
meow greet(name string) string {
  bring "Hello, " + name + "!"
}
nya(greet("Nyantyu"))
```

### Structs

**Go:**
```go
type Cat struct {
    Name string
    Age  int
}
c := Cat{Name: "Nyantyu", Age: 3}
fmt.Println(c.Name)
```

**Meow:**
```meow
kitty Cat {
  name: string
  age: int
}
nyan c = Cat("Nyantyu", 3)
nya(c.name)
```

Note: Meow uses positional constructor arguments instead of named fields.

### Error Handling

**Go:**
```go
result, err := divide(10, 0)
if err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Println("Result:", result)
}
```

**Meow (concise):**
```meow
nyan val = divide(10, 0) ~> 0
nya(val)
```

**Meow (verbose):**
```meow
nyan result = gag(paw() { divide(10, 0) })
sniff (is_furball(result)) {
  nya("Error:", result)
} scratch {
  nya("Result:", result)
}
```

### Loops

**Go:**
```go
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

for i := 1; i <= 20; i++ {
    fmt.Println(i)
}
```

**Meow:**
```meow
purr i (10) {
  nya(i)
}

purr i (1..20) {
  nya(i)
}
```

### List/Slice Operations

**Go (imperative):**
```go
nums := []int{1, 2, 3, 4, 5}
var doubled []int
for _, n := range nums {
    doubled = append(doubled, n*2)
}
```

**Meow (functional):**
```meow
nyan nums = [1, 2, 3, 4, 5]
nyan doubled = lick(nums, paw(x) { x * 2 })
```

**Go (filter):**
```go
var evens []int
for _, n := range nums {
    if n%2 == 0 {
        evens = append(evens, n)
    }
}
```

**Meow (filter):**
```meow
nyan evens = picky(nums, paw(x) { x % 2 == 0 })
```

### Pattern Matching

Go has no built-in pattern matching. The equivalent uses switch:

**Go:**
```go
func describe(n int) string {
    switch {
    case n == 0:
        return "zero"
    case n >= 1 && n <= 10:
        return "low"
    default:
        return "other"
    }
}
```

**Meow:**
```meow
meow describe(n int) string {
  bring peek(n) {
    0 => "zero",
    1..10 => "low",
    _ => "other"
  }
}
```

### HTTP Client

**Go:**
```go
resp, err := http.Get("https://example.com")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
fmt.Println(string(body))
```

**Meow:**
```meow
nab "http"
http.pounce("https://example.com") |=| nya
```

### Testing

**Go:**
```go
func TestAdd(t *testing.T) {
    got := add(1, 2)
    if got != 3 {
        t.Errorf("add(1, 2) = %d, want 3", got)
    }
}
```

**Meow:**
```meow
meow test_add() {
  expect(add(1, 2), 3, "add(1, 2)")
}
```

## Key Differences

### Dynamic vs Static Typing

Go is statically typed. Meow uses gradual typing — type annotations are optional but recommended:

```meow
# Untyped — flexible but slower at runtime
meow add(a, b) { bring a + b }

# Typed — compile-time checks, generates native Go operations
meow add(a int, b int) int { bring a + b }
```

### No Goroutines

Meow does not support goroutines or channels. It's a single-threaded language focused on simplicity.

### Panic-based Errors

Go uses `error` return values; Meow uses panic/recover with `hiss`/`gag`. The `~>` operator provides concise syntax for what would be `if err != nil { return default }` in Go.

### Functional Operations Built-in

Go requires manual loops for map/filter/reduce. Meow has built-in `lick`, `picky`, `curl`, and the `|=|` pipe operator.

### Single-file Programs

Meow programs are single `.nyan` files. There's no module system — standard library packages are imported with `nab`, but user code lives in one file.

## Transpiled Output

Use `meow transpile` to see the generated Go code:

```bash
meow transpile hello.nyan
```

Example input:

```meow
meow add(a int, b int) int {
  bring a + b
}
nya(add(1, 2))
```

Generated Go:

```go
package main

import meow "github.com/135yshr/meow/runtime/meowrt"

func add(a int64, b int64) int64 {
  return (a + b)
}

func main() {
  meow.Nya(meow.NewInt(add(int64(1), int64(2))))
}
```

Typed Meow functions generate native Go types (`int64`, `float64`, `string`, `bool`), while untyped functions use boxed `meow.Value`.
