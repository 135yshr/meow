# Meow Tutorial

A step-by-step guide to learning the Meow programming language. Each section builds on the previous one and includes runnable examples.

## Prerequisites

- Go 1.26+ installed
- Meow compiler installed (see [README](../README.md#installation))

Verify your installation:

```bash
meow version
```

## 1. Hello, World!

Create a file called `hello.nyan`:

```meow
nya("Hello, World!")
```

Run it:

```bash
meow run hello.nyan
```

Output:

```text
Hello, World!
```

`nya` is Meow's print function. It prints its arguments separated by spaces, followed by a newline.

You can also build a native binary:

```bash
meow build hello.nyan -o hello
./hello
```

## 2. Variables and Types

Declare variables with `nyan`:

```meow
nyan name = "Nyantyu"
nyan age = 3
nyan weight = 4.2
nyan is_cute = yarn        # true
nyan is_grumpy = hairball  # false
nyan nothing = catnap      # nil
```

Meow supports optional type annotations:

```meow
nyan name string = "Nyantyu"
nyan age int = 3
nyan weight float = 4.2
nyan is_cute bool = yarn
```

Type annotations help catch errors at compile time and generate faster code. Available types: `int`, `float`, `string`, `bool`, `furball`, `list`.

Print multiple values:

```meow
nyan name = "Nyantyu"
nyan age = 3
nya(name, "is", age, "years old")
# => Nyantyu is 3 years old
```

## 3. Functions

Define functions with `meow` and return values with `bring`:

```meow
meow greet(name string) string {
  bring "Hello, " + name + "!"
}

nya(greet("Nyantyu"))   # => Hello, Nyantyu!
nya(greet("Tyako"))     # => Hello, Tyako!
```

Functions can have typed parameters and return types:

```meow
meow add(a int, b int) int {
  bring a + b
}

nya(add(3, 7))   # => 10
```

Go-style grouped types — parameters without a type annotation inherit from the next one:

```meow
meow multiply(a, b int) int {
  bring a * b
}
```

Functions that don't `bring` a value return `catnap` (nil) implicitly.

## 4. Control Flow

### Conditionals

Use `sniff` for if, `scratch` for else:

```meow
meow check_age(age int) string {
  sniff (age < 1) {
    bring "kitten"
  } scratch sniff (age < 7) {
    bring "adult"
  } scratch {
    bring "senior"
  }
}

nya(check_age(0))    # => kitten
nya(check_age(3))    # => adult
nya(check_age(10))   # => senior
```

### Loops

**Count form** — iterates from `0` to `n-1`:

```meow
purr i (5) {
  nya(i)
}
# Output: 0, 1, 2, 3, 4
```

**Range form** — iterates from `a` to `b` inclusive:

```meow
purr i (1..5) {
  nya(i)
}
# Output: 1, 2, 3, 4, 5
```

**FizzBuzz example:**

```meow
meow fizzbuzz(n int) string {
  sniff (n % 15 == 0) {
    bring "FizzBuzz"
  } scratch sniff (n % 3 == 0) {
    bring "Fizz"
  } scratch sniff (n % 5 == 0) {
    bring "Buzz"
  } scratch {
    bring to_string(n)
  }
}

purr i (1..20) {
  nya(fizzbuzz(i))
}
```

## 5. Lists and Functional Operations

Create lists with square brackets:

```meow
nyan nums = [1, 2, 3, 4, 5]
nya(nums)           # => [1, 2, 3, 4, 5]
nya(len(nums))      # => 5
nya(nums[0])        # => 1
nya(head(nums))     # => 1
nya(tail(nums))     # => [2, 3, 4, 5]
```

### Map with `lick`

Apply a function to every element:

```meow
nyan nums = [1, 2, 3, 4, 5]
nyan doubled = lick(nums, paw(x) { x * 2 })
nya(doubled)   # => [2, 4, 6, 8, 10]
```

### Filter with `picky`

Keep elements matching a predicate:

```meow
nyan nums = [1, 2, 3, 4, 5]
nyan evens = picky(nums, paw(x) { x % 2 == 0 })
nya(evens)   # => [2, 4]
```

### Reduce with `curl`

Fold a list into a single value:

```meow
nyan nums = [1, 2, 3, 4, 5]
nyan sum = curl(nums, 0, paw(acc, x) { acc + x })
nya(sum)   # => 15
```

### Building lists

```meow
nyan cats = ["Nyantyu", "Tyako", "Tyomusuke"]
nyan more_cats = append(cats, "Tama")
nya(more_cats)   # => [Nyantyu, Tyako, Tyomusuke, Tama]
```

## 6. Lambdas and Pipes

### Lambdas

Anonymous functions are created with `paw`:

```meow
nyan double = paw(x int) { x * 2 }
nya(double(5))    # => 10

nyan greet = paw(name string) { "Hello, " + name }
nya(greet("Nyantyu"))   # => Hello, Nyantyu
```

Lambdas have a single expression as their body.

### Pipe Operator

The `|=|` operator chains operations by passing the left value as the first argument to the right:

```meow
nyan nums = [1, 2, 3, 4, 5]

# Without pipe
nya(lick(nums, paw(x) { x * 2 }))

# With pipe — same result, more readable
nums |=| lick(paw(x) { x * 2 }) |=| nya
```

Chain multiple operations:

```meow
[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
  |=| picky(paw(x) { x % 2 == 0 })
  |=| lick(paw(x) { x * x })
  |=| nya
# => [4, 16, 36, 64, 100]
```

## 7. Pattern Matching

Use `peek` to match a value against patterns:

```meow
meow describe(n int) string {
  bring peek(n) {
    0 => "zero",
    1..10 => "low",
    11..100 => "medium",
    _ => "high"
  }
}

nya(describe(0))     # => zero
nya(describe(5))     # => low
nya(describe(50))    # => medium
nya(describe(999))   # => high
```

Patterns:
- **Literal**: `0`, `"hello"`, `yarn`
- **Range**: `1..10` (inclusive)
- **Wildcard**: `_` (matches anything)

Always include a wildcard `_` as the last arm to ensure all cases are covered.

## 8. Error Handling

### Raising Errors

Use `hiss` to raise an error:

```meow
meow divide(a int, b int) int {
  sniff (b == 0) {
    hiss("division by zero")
  }
  bring a / b
}
```

### Catching with `gag`

```meow
nyan result = gag(paw() { divide(10, 0) })
sniff (is_furball(result)) {
  nya("Error caught:", result)
} scratch {
  nya("Result:", result)
}
# => Error caught: Hiss! division by zero
```

### Recovery with `~>`

The `~>` operator provides concise error recovery:

```meow
# Use a fallback value
nyan val = divide(10, 0) ~> 0
nya(val)   # => 0

# Use a handler function
nyan val2 = divide(10, 0) ~> paw(err) {
  nya("Handling error:", err)
  42
}
nya(val2)   # => 42
```

## 9. Structs (Kitty)

Define composite types with `kitty`:

```meow
kitty Cat {
  name: string
  age: int
}

nyan nyantyu = Cat("Nyantyu", 3)
nyan tyako = Cat("Tyako", 5)

nya(nyantyu)         # => Cat{name: Nyantyu, age: 3}
nya(nyantyu.name)    # => Nyantyu
nya(nyantyu.age)     # => 3
```

Use kitty types in functions:

```meow
meow introduce(cat) {
  nya(cat.name + " is " + to_string(cat.age) + " years old")
}

introduce(nyantyu)   # => Nyantyu is 3 years old
introduce(tyako)     # => Tyako is 5 years old
```

## 10. Standard Library

### File I/O

```meow
fetch "file"

# Read entire file
nyan content = file.snoop("data.txt")
nya(content)

# Read file line by line
nyan lines = file.stalk("data.txt")
lines |=| lick(paw(line) { "=> " + line }) |=| nya
```

### HTTP Client

```meow
fetch "http"

# GET request
http.pounce("https://httpbin.org/get") |=| nya

# POST with JSON body
http.toss("https://httpbin.org/post", {
  "name": "Nyantyu",
  "age": 3
}) |=| nya

# PUT
http.knead("https://httpbin.org/put", {"name": "Tyako"}) |=| nya

# DELETE
http.swat("https://httpbin.org/delete") |=| nya
```

With custom headers:

```meow
fetch "http"
nyan response = http.pounce("https://api.example.com/data", {
  "headers": {
    "Authorization": "Bearer my_token"
  }
})
nya(response)
```

See [stdlib.md](stdlib.md) for the full API reference.

## 11. Testing

Create a test file with the `_test.nyan` suffix:

```meow
# math_test.nyan
fetch "testing"

meow test_addition() {
  expect(1 + 1, 2, "basic addition")
  expect(10 + 20, 30, "larger addition")
}

meow test_division() {
  expect(10 / 2, 5, "basic division")
  nyan result = gag(paw() { 10 / 0 })
  judge(is_furball(result), "division by zero should error")
}
```

Run tests:

```bash
meow test math_test.nyan
```

Output:

```text
  PASS: test_addition
  PASS: test_division

All 2 tests passed, nya~!
```

### Output Verification Tests

Use the `catwalk_` prefix with `# Output:` blocks:

```meow
meow catwalk_hello() {
  nya("Hello, World!")
}
# Output:
# Hello, World!
```

### Assertions

- `judge(condition)` — assert truthy
- `expect(actual, expected)` — assert equal
- `refuse(condition)` — assert falsy

See [stdlib.md](stdlib.md) for details.

## 12. Build and Tools

### CLI Commands

```bash
meow run file.nyan              # Run a .nyan file
meow build file.nyan [-o name]  # Build a native binary
meow transpile file.nyan        # Show generated Go code
meow test [files...]            # Run test files
meow fmt [files...]             # Format .nyan files
meow lint [files...]            # Check for style issues
meow version                    # Show version info
meow help [command]             # Show help
```

### Viewing Generated Go Code

Use `transpile` to see what Go code Meow generates:

```bash
meow transpile hello.nyan
```

This is useful for understanding how Meow features map to Go.

## Next Steps

- [Language Reference](reference.md) — Complete keyword and operator reference
- [Language Specification](spec.md) — Formal grammar and semantics
- [Standard Library](stdlib.md) — Full API documentation
- [Effective Meow](effective-meow.md) — Idiomatic patterns and best practices
- [Cookbook](cookbook.md) — Task-based recipes
- [Go Comparison](go-comparison.md) — For Go developers learning Meow
