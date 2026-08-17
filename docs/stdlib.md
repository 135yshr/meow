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
- `string` → reads a whole number out of it, ignoring surrounding space
- Other types → returns a Furball

```meow
nya(to_int(3.7))      # => 3
nya(to_int(yarn))     # => 1
nya(to_int("42"))     # => 42
```

Everything a program takes from outside itself — the environment, a file, an
HTTP body — arrives as text, so reading a string is how a number gets in:

```meow
nab "env"
nyan budget = to_int(env.hunt("BUDGET_MS", "15000"))
```

Text that does not spell a number is a Furball rather than a wrong answer, so
it can be recovered from like any other failure:

```meow
nya(to_int("forty two") ~> 0)   # => 0
```

### `to_float(v)`

Convert a value to a float.

- `float` → returns as-is
- `int` → widens to float
- `string` → reads a decimal number out of it, ignoring surrounding space
- Other types → returns a Furball

```meow
nya(to_float(42))       # => 42
nya(to_float("3.5"))    # => 3.5
```

As with `to_int`, unreadable text is a Furball.

### `to_string(v)`

Convert any value to its string representation.

A non-empty list whose elements are all byte values — which is what `to_bytes`
produces, and nothing else does — is reassembled into the original string, so
`to_string` inverts `to_bytes`. Any other list yields its display form, and an
empty list is ambiguous rather than empty text, so `to_string(to_bytes(""))` is
`"[]"` rather than `""`.

```meow
nya(to_string(42))          # => 42
nya(to_string([1, 2, 3]))   # => [1, 2, 3]

nya(to_string(to_bytes("hello")))   # => hello
```

### `to_bytes(s)`

Convert a string to a list of its UTF-8 bytes.

- **s** (string): The string to convert.
- **Returns**: A list of byte values, which print like integers.
- **Panics**: If the argument is not a string.

The elements are byte values, a type nothing else produces — a list you write
out yourself, such as `[65]`, holds ints instead. That is what lets `to_string`
tell a byte list apart from any other list and reassemble it, so `[65]` prints
as `[65]` while `to_bytes("A")` round-trips back to `"A"`.

```meow
nya(to_bytes("ABC"))    # => [65, 66, 67]
nya(to_bytes("あ"))     # => [227, 129, 130]
nya(to_bytes(""))       # => []
```

### `to_runes(s)`

Convert a string to a list of single-character strings, split on characters
rather than bytes.

- **s** (string): The string to convert.
- **Returns**: A list of one-character strings.
- **Returns a Furball**: If the argument is not a string.

To go back the other way, join the list with `tangle(runes, "")`.

```meow
nya(to_runes("ABC"))              # => [A, B, C]
nya(to_runes("にゃん"))            # => [に, ゃ, ん]
nya(tangle(to_runes("にゃん"), ""))  # => にゃん
```

### `whiff(s, sub)`

Report whether `s` contains `sub`.

- **Returns**: `yarn` or `hairball`.
- **Returns a Furball**: If either argument is not a string.

```meow
nya(whiff("hello,world", "world"))   # => yarn
nya(whiff("hello,world", "dog"))     # => hairball
```

### `track(s, sub)`

Find the byte offset of the first occurrence of `sub` in `s`.

- **Returns**: The offset, or `-1` when `sub` does not occur.
- **Returns a Furball**: If either argument is not a string.

```meow
nya(track("hello,world", "world"))   # => 6
nya(track("hello,world", "dog"))     # => -1
```

### `shred(s, sep)`

Split `s` around each occurrence of `sep`.

- **Returns**: A litter of strings. An empty `sep` splits `s` into its
  individual characters.
- **Returns a Furball**: If either argument is not a string.

```meow
nya(shred("a,b,c", ","))    # => [a, b, c]
nya(shred("abc", ""))       # => [a, b, c]
nya(shred("abc", ","))      # => [abc]
```

### `tangle(list, sep)`

Join a litter of strings into one, separated by `sep`. The inverse of `shred`.

- **Returns**: The joined string.
- **Returns a Furball**: If `list` is not a litter, or holds a non-string.

```meow
nya(tangle(["a", "b", "c"], ","))       # => a,b,c
nya(tangle(shred("a,b,c", ","), " / "))  # => a / b / c
```

### `nibble(s, start, end)`

Take the piece of `s` from `start` up to but not including `end`.

Positions count characters, not bytes, so multi-byte text behaves the way it
reads. A negative position counts back from the end, positions are clamped to
the bounds of `s`, and an inverted range yields `""`.

- **Returns**: The substring.
- **Returns a Furball**: If `s` is not a string, or a position is not an int.

```meow
nya(nibble("hello,world", 0, 5))    # => hello
nya(nibble("hello,world", -5, 11))  # => world
nya(nibble("hello", 0, 99))         # => hello
nya(nibble("hello", 4, 2))          # =>
nya(nibble("にゃんこ", 1, 3))         # => ゃん
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

Import with `nab "http"`. Provides HTTP client operations.

The functions named after a verb — `pounce`, `toss`, `knead`, `swat`, `prowl` —
return the response body as a string, and return a **Furball naming the status**
when the response is 4xx or 5xx, so a failed request can be recovered with `~>`
like any other error rather than being mistaken for a successful one.

To read a status code instead of failing on it, use
[`http.chase`](#httpchasemethod-url--body--options), which returns the whole
response.

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

### `http.chase(method, url [, body [, options]])`

Perform a request with any method and return the whole response, rather than
just its body. This is how to inspect a status code: the verb functions answer
"give me the body, and fail if it did not work", while `chase` answers "tell me
what happened" — which is what a reachability or health check needs.

- **method** (string): HTTP method, case-insensitive.
- **url** (string): Request URL.
- **body** (map, string or `catnap`, optional): A map is sent as JSON with
  `Content-Type: application/json`; a string is sent as-is. Pass `catnap` to
  send no body. The body is positional rather than trailing so it can never be
  confused with the options map.
- **options** (map, optional): Options map.
- **Returns**: A map describing the response.

| Key | Type | Description |
|-----|------|-------------|
| `"status"` | int | HTTP status code |
| `"ok"` | bool | `yarn` when the status is 2xx |
| `"body"` | string | Response body |
| `"headers"` | map | Response headers |

A non-2xx status is reported in the map rather than as a Furball; a Furball is
still returned when the request itself could not be made.

```meow
nab "http"

nyan resp = http.chase("GET", "https://httpbin.org/status/401")
nya(resp["status"])   # => 401
nya(resp["ok"])       # => hairball

# With a JSON body
http.chase("POST", "https://httpbin.org/post", {"name": "Nyantyu"})

# With no body, but custom headers
http.chase("GET", "https://httpbin.org/get", catnap, {
  "headers": {"Authorization": "Bearer my_token"}
})
```

---

## env Package

Import with `nab "env"`. Reads the process environment, so that values which
should not be written down — tokens, endpoints that differ per deployment — can
reach a program without being staged in a plain-text file first.

### `env.hunt(name [, fallback])`

Read an environment variable.

- **name** (string): Variable name.
- **fallback** (any, optional): Returned when the variable is unset.
- **Returns**: The value as a string, or `catnap` when unset and no fallback was
  given.
- **Returns a Furball**: If `name` is not a non-empty string.

An unset variable reads as `catnap` rather than `""`, so it can be told apart
from one set to the empty string.

```meow
nab "env"

nyan token = env.hunt("API_TOKEN")
sniff (token == catnap) {
  hiss("API_TOKEN is not set")
}

nyan level = env.hunt("LOG_LEVEL", "info")
```

### `env.sniffed(name)`

Report whether an environment variable is set, including when it is set to the
empty string.

- **Returns**: `yarn` or `hairball`.

```meow
nab "env"
nya(env.sniffed("HOME"))   # => yarn
```

### `env.prowl()`

List the names of every environment variable, sorted.

- **Returns**: A litter of strings.

Only names are returned — listing values would make it far too easy to print a
secret by accident. Use `env.hunt` to read one deliberately.

```meow
nab "env"
nya(len(env.prowl()))
```

---

## clock Package

Import with `nab "clock"`. Reads the wall clock and pauses execution.

Times are reported as plain integers and strings rather than as an opaque
timestamp type, so they can be printed, compared and written out with the
operators the language already has.

### `clock.now()`

The current time as whole seconds since the Unix epoch.

- **Returns**: An int.

```meow
nab "clock"
nya(clock.now())   # => 1755266400
```

### `clock.nanos()`

The current time as nanoseconds since the Unix epoch. Wire formats that carry
timestamps — OpenTelemetry among them — ask for nanoseconds.

- **Returns**: An int.

```meow
nab "clock"
nya(clock.nanos())   # => 1755266400500000000
```

### `clock.stamp()`

The current UTC time as an RFC 3339 string. Always UTC, so two machines in
different zones agree.

- **Returns**: A string.

```meow
nab "clock"
nya(clock.stamp())   # => 2025-08-15T14:00:00Z
```

### `clock.nap(milliseconds)`

Pause for the given number of milliseconds.

- **milliseconds** (int): How long to pause. Zero is allowed.
- **Returns**: `catnap`.
- **Returns a Furball**: If the argument is not a non-negative int. A negative
  delay is reported rather than treated as zero, since it almost always means
  the caller computed it wrongly.

```meow
nab "clock"
clock.nap(250)
```

---

## random Package

Import with `nab "random"`. Produces random values.

`roll`, `drift` and `pick` draw from `math/rand/v2`, which suits sampling and
jitter. `tuft` draws from a cryptographic source instead, because it exists to
label things — a marker that collides, or that an observer can predict, defeats
the purpose of having one.

### `random.roll(n)`

A random int in `[0, n)`.

- **n** (int): Exclusive upper bound; must be positive.
- **Returns a Furball**: If `n` is not a positive int.

```meow
nab "random"
nya(random.roll(6))   # => 0..5
```

### `random.drift()`

A random float in `[0, 1)`.

```meow
nab "random"
nya(random.drift())   # => 0.5772156649
```

### `random.pick(list)`

A random element of a litter.

- **Returns a Furball**: If the argument is not a litter, or is empty.

```meow
nab "random"
nya(random.pick(["Nyantyu", "Tyako", "Mikan"]))
```

### `random.tuft(n)`

`n` random bytes as a lowercase hex string, so the result is `2n` characters
long. Drawn from a cryptographic source, which makes it suitable for the
markers and correlation IDs it is meant for.

- **n** (int): Number of bytes, 1 to 1024.
- **Returns a Furball**: If `n` is outside that range, or the source fails.

```meow
nab "random"
nya(random.tuft(8))   # => 3f9a1c04b7e25d68
```

---

## aws Package

Import with `nab "aws"`. Talks to Amazon Web Services through the official
[aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2).

Credentials and region are resolved by the SDK's own default chain —
environment variables, shared config and credentials files, SSO, container and
instance metadata — so a Meow program authenticates exactly the way the AWS CLI
does on the same machine, and never has to be handed a secret directly.

Results come back as ordinary Maps and litters, so they can be indexed and
printed with what the language already has. Every call is bounded by a 30-second
timeout, and every failure is a Furball, so `~>` recovers from it.

This is the only part of Meow that depends on a third-party package; everything
else is standard library only.

### `aws.whoami()`

The identity the program is authenticated as.

- **Returns**: A map with `"account"`, `"arn"` and `"user_id"`.

The cheapest way to answer "are my credentials working, and am I in the account
I think I am" before doing anything that matters.

```meow
nab "aws"

nyan me = aws.whoami() ~> paw(err) { hiss("not authenticated:", err) }
nya(me["account"])
nya(me["arn"])
```

### `aws.region()`

The region the SDK resolved, so a program can confirm where it is about to act.

- **Returns**: A string.

```meow
nab "aws"
nya(aws.region())   # => ap-northeast-1
```

### `aws.dig(group [, options])`

Search a CloudWatch Logs group and return the matching events.

- **group** (string): Log group name.
- **options** (map, optional): See below.
- **Returns**: A litter of maps, each with `"timestamp"`, `"message"` and
  `"stream"`.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `"pattern"` | string | (all events) | CloudWatch Logs filter pattern |
| `"start"` | int | (none) | Earliest event, in milliseconds since the epoch |
| `"end"` | int | (none) | Latest event, in milliseconds since the epoch |
| `"limit"` | int | 10000 | Maximum events to return, 1 to 10000 |

Times are milliseconds since the Unix epoch, the unit the API uses, so
`clock.now() * 1000` lines up with them.

Paging is handled for you. CloudWatch applies `"limit"` per page and may hand
back a partial — or empty — page while more results remain, so `dig` follows the
continuation tokens until it has `"limit"` events or the log group is exhausted.

```meow
nab "aws"
nab "clock"

nyan since = (clock.now() - 300) * 1000
nyan hits = aws.dig("/aws/lambda/canary", {
  "pattern": "nyan-marker-001", "start": since
})

sniff (len(hits) > 0) {
  nya("OK stored:", head(hits)["message"])
} scratch {
  nya("NG not stored")
}
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
