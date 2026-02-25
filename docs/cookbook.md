# Meow Cookbook

Task-based recipes for common programming patterns in Meow. Each recipe is a complete, runnable `.nyan` program.

## 1. Read a File and Process Each Line

```meow
fetch "file"

nyan lines = file.stalk("input.txt")
lines |=| lick(paw(line) {
  ">> " + line
}) |=| nya
```

With filtering (skip empty lines):

```meow
fetch "file"

file.stalk("input.txt")
  |=| picky(paw(line) { len(line) > 0 })
  |=| lick(paw(line) { ">> " + line })
  |=| nya
```

## 2. HTTP GET Request with Error Handling

```meow
fetch "http"

nyan response = http.pounce("https://httpbin.org/get") ~> paw(err) {
  nya("Request failed:", err)
  "{}"
}
nya(response)
```

With custom headers:

```meow
fetch "http"

nyan response = http.pounce("https://api.example.com/data", {
  "headers": {
    "Authorization": "Bearer token123",
    "Accept": "application/json"
  }
})
nya(response)
```

## 3. POST JSON Data

```meow
fetch "http"

nyan data = {
  "name": "Nyantyu",
  "age": 3,
  "color": "orange"
}

nyan response = http.toss("https://httpbin.org/post", data) ~> paw(err) {
  nya("POST failed:", err)
  "{}"
}
nya(response)
```

Maps are automatically serialized to JSON with `Content-Type: application/json`.

## 4. Filter and Transform a List

Get the squares of even numbers:

```meow
nyan nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

nyan result = nums
  |=| picky(paw(x) { x % 2 == 0 })
  |=| lick(paw(x) { x * x })

nya(result)   # => [4, 16, 36, 64, 100]
```

Sum of the transformed list:

```meow
nyan nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

nyan total = nums
  |=| picky(paw(x) { x % 2 == 0 })
  |=| lick(paw(x) { x * x })
  |=| curl(0, paw(acc, x) { acc + x })

nya(total)   # => 220
```

## 5. Build Data Structures with Kitty

```meow
kitty Cat {
  name: string
  age: int
}

nyan cats = [
  Cat("Nyantyu", 3),
  Cat("Tyako", 5),
  Cat("Tyomusuke", 2)
]

# Print all cat names
purr i (len(cats)) {
  nya(cats[i].name + " is " + to_string(cats[i].age) + " years old")
}
```

Output:

```text
Nyantyu is 3 years old
Tyako is 5 years old
Tyomusuke is 2 years old
```

## 6. Recursive Fibonacci

```meow
meow fib(n int) int {
  sniff (n <= 1) {
    bring n
  }
  bring fib(n - 1) + fib(n - 2)
}

purr i (10) {
  nya(fib(i))
}
```

Output:

```text
0
1
1
2
3
5
8
13
21
34
```

## 7. Classify Values with Pattern Matching

```meow
meow classify_temp(celsius int) string {
  bring peek(celsius) {
    -50..0 => "freezing",
    1..15 => "cold",
    16..25 => "pleasant",
    26..35 => "warm",
    36..50 => "hot",
    _ => "extreme"
  }
}

nyan temps = [-10, 5, 20, 30, 42, 100]
purr i (len(temps)) {
  nya(to_string(temps[i]) + " C => " + classify_temp(temps[i]))
}
```

Output:

```text
-10 C => freezing
5 C => cold
20 C => pleasant
30 C => warm
42 C => hot
100 C => extreme
```

## 8. Chain Operations with Pipes

Process data through a pipeline:

```meow
meow double(x) { x * 2 }
meow add_one(x) { x + 1 }

# Build a pipeline
[1, 2, 3]
  |=| lick(double)
  |=| lick(add_one)
  |=| nya
# => [3, 5, 7]
```

Read a file, process it, and display:

```meow
fetch "file"

file.stalk("names.txt")
  |=| picky(paw(name) { len(name) > 0 })
  |=| lick(paw(name) { "Hello, " + name + "!" })
  |=| nya
```

## 9. Write Tests

### Basic test file (`math_test.nyan`)

```meow
fetch "testing"

meow add(a int, b int) int {
  bring a + b
}

meow test_add() {
  expect(add(1, 2), 3, "1 + 2 = 3")
  expect(add(-1, 1), 0, "-1 + 1 = 0")
  expect(add(0, 0), 0, "0 + 0 = 0")
}

meow test_negative() {
  judge(add(-5, -3) == -8, "negative addition")
}
```

Run:

```bash
meow test math_test.nyan
```

### Output verification test

```meow
meow catwalk_greeting() {
  nya("Hello, Nyantyu!")
  nya("Welcome to Meow!")
}
# Output:
# Hello, Nyantyu!
# Welcome to Meow!
```

### Testing error conditions

```meow
fetch "testing"

meow safe_divide(a int, b int) int {
  sniff (b == 0) {
    hiss("division by zero")
  }
  bring a / b
}

meow test_division_error() {
  nyan result = gag(paw() { safe_divide(10, 0) })
  judge(is_furball(result), "division by zero should be furball")
}

meow test_division_success() {
  expect(safe_divide(10, 2), 5, "10 / 2 = 5")
}
```

## 10. Find Maximum in a List (curl pattern)

```meow
meow find_max(nums list) {
  bring curl(tail(nums), head(nums), paw(max, x) {
    peek(x > max) {
      yarn => x,
      _ => max
    }
  })
}

nyan numbers = [3, 7, 1, 9, 4, 6, 2, 8, 5]
nya("Max:", find_max(numbers))   # => Max: 9
```

Find minimum using the same pattern:

```meow
meow find_min(nums list) {
  bring curl(tail(nums), head(nums), paw(min, x) {
    peek(x < min) {
      yarn => x,
      _ => min
    }
  })
}

nyan numbers = [3, 7, 1, 9, 4, 6, 2, 8, 5]
nya("Min:", find_min(numbers))   # => Min: 1
```

## 11. Error Recovery with Handler

Handle different error scenarios:

```meow
fetch "http"

meow fetch_data(url string) string {
  bring http.pounce(url) ~> paw(err) {
    nya("Failed to fetch " + url + ":", err)
    "[]"
  }
}

nyan data = fetch_data("https://api.example.com/cats")
nya(data)
```

Chain recovery with pipe:

```meow
meow risky_parse(input string) {
  sniff (len(input) == 0) {
    hiss("empty input")
  }
  bring input
}

nyan result = risky_parse("") ~> paw(err) {
  nya("Parse error handled:", err)
  "default"
}
nya(result)   # => default
```

Nested recovery:

```meow
fetch "file"

meow get_config() string {
  bring file.snoop("config.json")
    ~> paw(err) { file.snoop("config.default.json") }
    ~> paw(err) { "{}" }
}
```

This tries `config.json` first, falls back to `config.default.json`, and finally returns `"{}"` if both fail.
