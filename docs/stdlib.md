# Meow Standard Library Reference

This document describes all built-in functions and standard library packages available in Meow.

## Built-in Functions (no `nab` required)

These functions are available globally in every `.nyan` program.

### `nya(args...)`

Print values to standard output, separated by spaces, followed by a newline.

```meow
nya("Hello", "World")   # => Hello World
nya(42)                  # => 42
nya([1, 2, 3])           # => [1, 2, 3]
```

Returns `catnap`.

### `hiss(args...)`

Raise an error by panicking. Arguments are joined with spaces and prefixed with `"Hiss! "`.

```meow
hiss("something went wrong")
# panics with: Hiss! something went wrong

hiss("bad value:", x)
# panics with: Hiss! bad value: <value of x>
```

This function never returns.

### `gag(fn)`

Call a zero-argument function and catch any panic. If the function succeeds, its return value is returned. If it panics, the error is wrapped in a `Furball` and returned.

```meow
nyan result = gag(paw() { divide(10, 0) })
# result is a Furball with message "Hiss! division by zero"

nyan ok = gag(paw() { divide(10, 2) })
# ok is 5
```

### `is_furball(v)`

Check if a value is a `Furball` (error). Returns `yarn` or `hairball`.

```meow
nyan result = gag(paw() { hiss("oops") })
nya(is_furball(result))   # => yarn
nya(is_furball(42))       # => hairball
```

### `len(v)`

Return the length of a string (byte count) or list (element count).

```meow
nya(len("hello"))       # => 5
nya(len([1, 2, 3]))     # => 3
```

Panics if `v` is not a string or list.

### `head(list)`

Return the first element of a list.

```meow
nya(head([10, 20, 30]))   # => 10
```

Panics if the list is empty.

### `tail(list)`

Return a new list containing all elements except the first.

```meow
nya(tail([10, 20, 30]))   # => [20, 30]
```

Panics if the list is empty.

### `append(list, value)`

Return a new list with `value` appended to the end.

```meow
nyan nums = [1, 2, 3]
nya(append(nums, 4))   # => [1, 2, 3, 4]
```

### `lick(list, fn)`

Map: apply `fn` to each element and return a new list of results.

```meow
nyan doubled = lick([1, 2, 3], paw(x) { x * 2 })
nya(doubled)   # => [2, 4, 6]
```

### `picky(list, fn)`

Filter: return a new list containing only elements where `fn` returns a truthy value.

```meow
nyan evens = picky([1, 2, 3, 4, 5], paw(x) { x % 2 == 0 })
nya(evens)   # => [2, 4]
```

### `curl(list, init, fn)`

Reduce: fold the list into a single value using an accumulator.

```meow
nyan sum = curl([1, 2, 3, 4, 5], 0, paw(acc, x) { acc + x })
nya(sum)   # => 15
```

`fn` receives two arguments: the accumulator and the current element.

### `to_int(v)`

Convert a value to an integer.

- `int` → returns as-is
- `float` → truncates to int
- `bool` → `yarn` is `1`, `hairball` is `0`
- Other types → panics

```meow
nya(to_int(3.7))      # => 3
nya(to_int(yarn))     # => 1
```

### `to_float(v)`

Convert a value to a float.

- `float` → returns as-is
- `int` → widens to float
- Other types → panics

```meow
nya(to_float(42))   # => 42
```

### `to_string(v)`

Convert any value to its string representation.

```meow
nya(to_string(42))          # => 42
nya(to_string([1, 2, 3]))   # => [1, 2, 3]
```

---

## file Package

Import with `nab "file"`. Provides file I/O operations.

### `file.snoop(path)`

Read the entire contents of a file as a string. Trailing `\r\n` is stripped.

- **path** (string): File path to read.
- **Returns**: String with file contents.
- **Panics**: If the file cannot be read.

```meow
nab "file"

nyan content = file.snoop("data.txt")
nya(content)
```

### `file.stalk(path)`

Read a file line by line and return a list of strings.

- **path** (string): File path to read.
- **Returns**: List of strings, one per line.
- **Panics**: If the file cannot be read.

```meow
nab "file"

nyan lines = file.stalk("data.txt")
lines |=| lick(paw(line) { "=> " + line }) |=| nya
```

Maximum line length: 1 MiB.

---

## http Package

Import with `nab "http"`. Provides HTTP client operations. All functions return the response body as a string.

**Default settings:**
- Timeout: 10 seconds
- Max response body: 1 MiB
- User-Agent: `meow-http-client/2.0`

### Options Map

GET/DELETE/OPTIONS functions accept an optional `options` map as the last argument. POST/PUT functions accept it as the third argument.

```meow
nyan opts = {
  "maxBodyBytes": 2097152,
  "headers": {
    "Authorization": "Bearer my_token",
    "Accept": "application/json"
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `"maxBodyBytes"` | int | 1048576 (1 MiB) | Maximum response body size |
| `"headers"` | map | (none) | Custom HTTP headers |

### `http.pounce(url [, options])`

HTTP GET request.

- **url** (string): Request URL.
- **options** (map, optional): Options map.
- **Returns**: Response body as string.

```meow
nab "http"

nyan body = http.pounce("https://httpbin.org/get")
nya(body)
```

With custom headers:

```meow
nyan body = http.pounce("https://api.example.com/data", {
  "headers": { "Authorization": "Bearer token123" }
})
```

### `http.toss(url, body [, options])`

HTTP POST request.

- **url** (string): Request URL.
- **body** (string or map): Request body. Maps are automatically serialized to JSON with `Content-Type: application/json`.
- **options** (map, optional): Options map.
- **Returns**: Response body as string.

```meow
nab "http"

# POST with JSON body (map → auto-JSON)
http.toss("https://httpbin.org/post", {"name": "Nyantyu", "age": 3})

# POST with raw string body
http.toss("https://httpbin.org/post", "raw data")
```

### `http.knead(url, body [, options])`

HTTP PUT request. Same arguments as `toss`.

```meow
nab "http"
http.knead("https://httpbin.org/put", {"name": "Tyako"})
```

### `http.swat(url [, options])`

HTTP DELETE request.

- **url** (string): Request URL.
- **options** (map, optional): Options map.
- **Returns**: Response body as string.

```meow
nab "http"
http.swat("https://httpbin.org/delete")
```

### `http.prowl(url [, options])`

HTTP OPTIONS request.

- **url** (string): Request URL.
- **options** (map, optional): Options map.
- **Returns**: Response body as string.

```meow
nab "http"
http.prowl("https://httpbin.org/get")
```

---

## testing Package

Import with `nab "testing"`. Provides test assertions and test execution.

### Test Function Conventions

- Functions named `test_*` are automatically wrapped with `run()` and `report()`.
- Functions named `catwalk_*` are output verification tests — they capture stdout and compare it to an expected string.
- Test functions must take no parameters.

### `testing.judge(condition [, message])`

Assert that a condition is truthy.

- **condition**: Value to check for truthiness.
- **message** (string, optional): Custom failure message.
- **Returns**: `catnap`.
- **Panics (test failure)**: If condition is falsy.

```meow
judge(1 + 1 == 2)
judge(len("hello") == 5, "string length should be 5")
```

### `testing.expect(actual, expected [, message])`

Assert that two values are equal (compared by string representation).

- **actual**: The value to check.
- **expected**: The expected value.
- **message** (string, optional): Custom failure message.
- **Returns**: `catnap`.
- **Panics (test failure)**: If values are not equal.

```meow
expect(1 + 1, 2, "basic addition")
expect(to_string(42), "42")
```

### `testing.refuse(condition [, message])`

Assert that a condition is falsy.

- **condition**: Value to check for falsiness.
- **message** (string, optional): Custom failure message.
- **Returns**: `catnap`.
- **Panics (test failure)**: If condition is truthy.

```meow
refuse(1 == 2)
refuse(is_furball(42), "42 should not be a furball")
```

### `testing.run(name, fn)`

Execute a named test function. Catches panics and records the result.

- **name** (string): Test name.
- **fn** (function): Zero-argument function to execute.
- **Returns**: `yarn` if passed, `hairball` if failed.

```meow
nab "testing"
testing.run("my test", paw() {
  judge(1 + 1 == 2)
})
```

Usually you don't call `run` directly — the `test_` prefix handles it automatically.

### `testing.catwalk(name, fn, expected)`

Execute a function, capture its stdout output, and compare with expected output. This is the Meow equivalent of Go's `Example` tests.

- **name** (string): Test name.
- **fn** (function): Zero-argument function to execute.
- **expected** (string): Expected stdout output.
- **Returns**: `yarn` if passed, `hairball` if failed.

```meow
nab "testing"
testing.catwalk("hello output", paw() {
  nya("Hello, World!")
}, "Hello, World!\n")
```

Usually you don't call `catwalk` directly — the `catwalk_` prefix handles it automatically.

### `testing.report()`

Print the test summary and exit with code 1 if any tests failed.

Output format:
```text
  PASS: test_name
  FAIL: test_name - error message

All 5 tests passed, nya~!
```

Or on failure:
```text
3 passed, 2 failed, nya~
```

### Complete Test Example

```meow
nab "testing"

meow test_arithmetic() {
  expect(1 + 1, 2, "addition")
  expect(10 - 3, 7, "subtraction")
  expect(3 * 4, 12, "multiplication")
}

meow test_string_ops() {
  judge("hello" + " " + "world" == "hello world")
  expect(len("meow"), 4)
}

meow test_error_handling() {
  nyan result = gag(paw() { hiss("test error") })
  judge(is_furball(result))
}
```

Run with:

```bash
meow test my_test.nyan
```

### Catwalk (Output Test) Example

In `_test.nyan` files, functions with the `catwalk_` prefix are paired with `# Output:` comment blocks:

```meow
meow catwalk_greeting() {
  nya("Hello, Nyantyu!")
}
# Output:
# Hello, Nyantyu!
```

The compiler extracts the expected output from the `# Output:` block and verifies that the function's actual stdout matches.
