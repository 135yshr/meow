---
title: "Meow Programming Language Standard Library"
description: "Reference for built-in functions and standard library packages in the Meow Programming Language — file I/O, HTTP, conversions, and more."
weight: 2
---

This document describes all built-in functions and standard library packages available in the Meow Programming Language.

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

### `scram([status])`

End the program with the given status.

- **status** (int, optional): 0 to 255. Omitted means 0.
- **Never returns** when given a status a process can report.
- **Returns a Furball**: If `status` is not an int, or is outside 0 to 255. There
  is nothing to end on such a status, so the program carries on and the Furball
  can be caught like any other.

A status is how a program tells the shell, cron job or CI step that started it
what it found. Without one, a check that saw an endpoint go down could only say
so in words that nothing downstream reads.

```meow
sniff (down > 0) {
  nya(to_string(down) + " endpoint(s) down")
  scram(1)
}
scram()          # the same as scram(0): finished, nothing wrong
```

A status outside 0 to 255 is refused rather than wrapped. Were it wrapped, a
shell would see only the low eight bits, and `scram(256)` would arrive as 0 — a
program asking to fail would be read as having succeeded.

On a status a process can report, nothing after `scram` runs and `gag` cannot
catch it: the program is over. A refused status is the other case — there the
program carries on, and what `scram` gave back can be caught like anything else.
As with `hiss`, a typed function that ends in `scram` still needs a `bring`
after it, since the checker reads the function as having a path with no return.

Inside a fully typed function a refused status is raised rather than returned —
such a function has no way to hand a Furball back — so `gag` catches it there
the way it catches a `hiss`.

In the playground there is no process to end, so the run simply stops where
`scram` was called and keeps whatever it printed on the way.

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

Return the length of a string (byte count), litter (element count) or basket
(entry count).

```meow
nya(len("hello"))            # => 5
nya(len([1, 2, 3]))          # => 3
nya(len({"a": 1, "b": 2}))   # => 2
```

Anything else is a Furball.

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

### `upper(s)` / `lower(s)`

Return the string with every letter in upper or lower case.

```meow
nya(upper("cat"))   # => CAT
nya(lower("CAT"))   # => cat
```

### `trim(s)`

Return the string without leading or trailing whitespace.

A line read from a file, or an environment variable, carries surrounding space,
and comparing against it without trimming answers no for a reason nothing in
the source shows.

```meow
nya("[" + trim("  cat  ") + "]")   # => [cat]
```

### `replace(s, old, new)`

Return the string with **every** occurrence of `old` replaced by `new`.

An empty `old` is a Furball: it would insert the replacement between every
character, which is never the intent.

```meow
nya(replace("a-b-c", "-", "+"))   # => a+b+c
```

### `pad(s, width)`

Return the string widened to `width` characters with spaces.

A positive width pads on the right, lining a column up on its left edge; a
negative one pads on the left, which is what a column of numbers wants. A
string already that wide is returned whole rather than cut — losing text to make
a table line up is the worse trade. Width is counted in characters, as `nibble`
and `track` are.

```meow
nya("[" + pad("ab", 5) + "]")       # => [ab   ]
nya("[" + pad("ab", 0 - 5) + "]")   # => [   ab]
```

### `sort(list)`

Return a litter with its elements in ascending order. The litter passed in is
left alone.

A litter holding more than one kind of value is a Furball rather than an
arbitrary order: there is no answer to "is `1` before `"a"`", and inventing one
would put a program's output at the mercy of which happened to come first.

```meow
nya(sort([3, 1, 2]))              # => [1, 2, 3]
nya(sort(["pear", "apple"]))      # => [apple, pear]
```

### `reverse(list)`

Return a litter with its elements in the opposite order. The litter passed in is
left alone.

```meow
nya(reverse([1, 2, 3]))   # => [3, 2, 1]
```

### `round(x, places)`

Return the number rounded to the given number of decimal places, 0 to 15.

Rounding is half away from zero, the arithmetic convention, rather than Go's
half-to-even — a reader expects `2.5` to print as `3`. An `int` is already
rounded and comes back as one, so it does not start printing as `42.0`.

```meow
nya(round(3.14159, 2))   # => 3.14
nya(round(2.5, 0))       # => 3
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

### `env.haul()`

Read the arguments the program was started with.

- **Returns**: A litter of strings, in the order they were given.
- **Returns a Furball**: If called with any arguments.

The program's own name is left out — a program wants what it was asked to do,
not the path it happens to be installed at. A program started with no arguments
gets an empty litter, so `len` answers without a special case for "none".

```meow
nab "env"

nyan given = env.haul()
sniff (len(given) == 0) {
  nya("usage: check <targets file>")
  scram(2)
}
nyan targets = head(given)
```

Everything typed after the `.nyan` file belongs to the program, so
`meow run check.nyan -v` hands `-v` to the program rather than reading it as
meow's own flag. Running the built binary the same way gives the same answer.

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

## Talking to AWS

There is no `aws` package. Meow reaches the AWS SDK the way it reaches any Go
library, with [`nab go`](reference.md#nab):

```meow
nab go "github.com/aws/aws-sdk-go-v2/config" tag cfg
nab go "github.com/aws/aws-sdk-go-v2/service/sts"

nyan conf = cfg.load_default_config()
nyan client = sts.new_from_config(conf)

nyan me = gag(paw() { client.get_caller_identity({}) })
sniff (is_furball(me)) {
  hiss("not authenticated:", me)
}
nya(me["account"])
nya(me["arn"])
```

Credentials and region are resolved by the SDK's own default chain —
environment variables, shared config and credentials files, SSO, container and
instance metadata — so a Meow program authenticates exactly the way the AWS CLI
does on the same machine, and never has to be handed a secret directly.

`examples/aws_sts.nyan` is that program in full.

### What became of `nab "aws"`

It was a package of Go wrappers, written by hand, one function at a time.
`nab go` makes the SDK reachable directly, so the wrappers were work that the
bridge now does — and doing it by hand meant only the three calls someone had
written were available, out of the SDK's thousands.

| Was | Now |
|-----|-----|
| `aws.whoami()` | `sts.new_from_config(conf)`, then `client.get_caller_identity({})` |
| `aws.region()` | `cfg.load_default_config()`, then `conf["region"]` |
| `aws.dig(group, opts)` | `logs.new_from_config(conf)`, then `client.filter_log_events({...})` |

Two things `aws.dig` did for you, you now write yourself, because they are the
program's business rather than the language's:

**Paging.** CloudWatch applies `"limit"` per page and may hand back a partial —
or empty — page while more results remain, so one call is not enough. Follow
`next_token` until it is empty:

```meow
nab go "github.com/aws/aws-sdk-go-v2/config" tag cfg
nab go "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs" tag logs

nyan conf = cfg.load_default_config()
nyan client = logs.new_from_config(conf)

# The API takes at most 10000 events per page, so a larger want is asked for
# across several pages rather than in one request it would refuse.
meow page_size(want int) int {
  sniff (want > 10000) {
    bring 10000
  }
  bring want
}

meow ask_for(group string, pattern string, want int, token string) basket {
  sniff (token == "") {
    bring {
      "log_group_name": group,
      "filter_pattern": pattern,
      "limit": page_size(want)
    }
  }
  bring {
    "log_group_name": group,
    "filter_pattern": pattern,
    "limit": page_size(want),
    "next_token": token
  }
}

meow as_event(e basket) basket {
  bring {
    "timestamp": e["timestamp"],
    "message": e["message"],
    "stream": e["log_stream_name"]
  }
}

# An empty page is still a page. CloudWatch can hand back nothing at all and
# still supply a token, and a Go slice that is nil arrives as catnap rather
# than as an empty litter — which curl will not read.
meow events_of(page basket) litter {
  nyan events = page["events"]
  sniff (events == catnap) {
    bring []
  }
  bring events
}

# Only as many as were asked for. "limit" is what the API takes per page, not a
# total, so a page can hold more than is left to take.
meow take_events(events litter, found litter, room int) litter {
  sniff (room <= 0 || len(events) == 0) {
    bring found
  }
  bring take_events(tail(events), append(found, as_event(head(events))), room - 1)
}

# pages bounds the walk. The token is the server's to hand out, and a run of
# empty pages leaves want where it was, so nothing else here has to end.
meow read_page(group string, pattern string, want int, token string, found litter, pages int) litter {
  sniff (want <= 0) {
    bring found
  }
  # Running out of pages is not the same as running out of events. There is a
  # token still outstanding here, so handing back what arrived would be the
  # partial answer the failure below refuses to give.
  sniff (pages <= 0) {
    hiss("gave up with pages still unread in", group)
  }
  nyan page = gag(paw() { client.filter_log_events(ask_for(group, pattern, want, token)) })
  sniff (is_furball(page)) {
    # The failure is the answer. Handing back the pages that did arrive would
    # say "these are the events" when the truth is "some of them, maybe" — and
    # for a check asking whether a marker arrived, that reads as a confident no.
    #
    # It is raised rather than returned: gag has already caught this one, so
    # handing it back would print the failure and still leave with 0, which is
    # the same lie told to whatever reads the exit status.
    hiss("could not read the log group:", page)
  }
  nyan more = take_events(events_of(page), found, want)
  nyan next = to_string(page["next_token"])
  sniff (next == "catnap" || next == "") {
    bring more
  }
  bring read_page(group, pattern, want - (len(more) - len(found)), next, more, pages - 1)
}
```

Bindings are immutable, so this is written as recursion rather than as a loop
with an accumulator. Three things in it are worth keeping in a program of your
own, all of which `aws.dig` did for you:

- **A failure is raised, not returned.** `gag` marks what it caught as handled,
  so handing that value back would print the failure and still leave with 0 —
  the same lie told to whatever reads the exit status.
- **An empty page is still a page.** CloudWatch can hand back nothing and still
  supply a token, and a nil Go slice arrives as `catnap` rather than an empty
  litter, which `curl` will not read.
- **The walk is bounded, and running out of pages is a failure.** The token is
  the server's to hand out, and a run of empty pages leaves `want` where it
  was, so nothing else here has to end. Reaching that bound means a token is
  still outstanding, so it raises rather than handing back what arrived —
  giving up is not the same as finishing.
- **Only as many as were asked for.** `"limit"` is per page rather than a
  total, so a page can hold more than is left to take; asking for 2 has to
  give 2 even when the page holds 3.

**Names.** `aws.dig` renamed the SDK's fields; the bridge gives them to you as
the SDK writes them, in Meow's spelling. Where the old shape is wanted, a small
`meow` function converts it.

| `aws.dig` | the SDK |
|-----------|---------|
| `group` (argument) | `log_group_name` |
| `"pattern"` | `"filter_pattern"` |
| `"start"` | `"start_time"` |
| `"end"` | `"end_time"` |
| `"limit"` | `"limit"`, per page rather than in total |
| `"stream"` (result) | `"log_stream_name"` |

`"start_time"` and `"end_time"` are milliseconds since the Unix epoch, as they
were, so `clock.now() * 1000` still lines up with them. `"limit"` is what the
API takes per page — at most 10000 — rather than the total `aws.dig` gathered,
which is why the program above asks for it a page at a time.

**A timeout.** Each bridged call is bounded by 30 seconds, as `nab "aws"` was.

---

---

## json Package

Import with `nab "json"`. Reads JSON text into Meow values and writes it back.

A program that talks to an HTTP API can otherwise only match replies as text,
and deciding "did the record arrive" by searching a body for a word answers yes
when the word appears somewhere else in the payload.

### `json.unravel(text)`

Read JSON text.

- **text** (string): The document to read.
- **Returns**: An object as a `basket`, an array as a `litter`, `null` as
  `catnap`, and strings, numbers and booleans as themselves.
- Text that is not JSON is a **Furball**, so it can be recovered from with `~>`.

```meow
nab "json"

nyan doc = json.unravel("{\"hits\": [{\"marker\": \"m1\"}], \"count\": 1}")
nya(doc["count"])            # => 1
nya(len(doc["hits"]))        # => 1
purr hit (doc["hits"]) {
  nya(hit["marker"])         # => m1
}

nya(json.unravel("<html>") ~> "not json")   # => not json
```

JSON has one number type and Meow has two, so a whole value comes back as an
`int` and anything else as a `float` — an id or a count does not arrive
reading `42.0`. Whole values are read exactly, including ones past 2^53 that a
float could not hold; past `int64` there is nothing exact left to offer, so
those read as floats rather than being refused.

### `json.wind(value)`

Write a value as JSON text.

- **value**: A `basket`, `litter`, string, number, boolean or `catnap`.
- **Returns**: The JSON text as a string.
- A value JSON has no shape for — a Furball, a `kitty`, a function — is a
  **Furball**, rather than text that would read back as something else.

```meow
nya(json.wind({"a": 1, "b": [1, 2, 3]}))   # => {"a":1,"b":[1,2,3]}
```

A round trip keeps what JSON has a shape for, but two things do not survive it.
Keys lose the order they were written in — a `basket` has no order, so they come
back sorted, while a `litter` keeps its order, being a sequence. And a whole
`float` comes back an `int`: JSON has no way to write "the float one", so `1.0`
is written `1` and read as a whole value, by the rule above.

```meow
nya(json.wind(1.0))                  # => 1
nya(json.unravel(json.wind(1.0)))    # => 1, an int now
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
