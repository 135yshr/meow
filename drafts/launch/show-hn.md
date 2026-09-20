# Launch prep: Show HN + r/ProgrammingLanguages

Prep work for [issue #105](https://github.com/135yshr/meow/issues/105). Nothing here has been
posted. Everything below is for the owner to review, edit and post manually.

Prepared against `main` at `v0.21.4`.

**Scope note.** The playground audit is a **source-level review plus live HTTP header checks**.
I have no browser, so nothing below was observed rendering. Sizes, compression, MIME types and
interpreter behaviour were measured for real; layout and first-paint claims are reasoned from
the CSS and templates and are marked as such.

---

## Table of contents

1. [Verdict: can we post?](#1-verdict-can-we-post)
2. [Playground audit](#2-playground-audit)
3. [English review of the landing page and tutorial](#3-english-review-of-the-landing-page-and-tutorial)
4. [Ready-to-paste: Show HN](#4-ready-to-paste-show-hn)
5. [Ready-to-paste: r/ProgrammingLanguages](#5-ready-to-paste-rprogramminglanguages)
6. [Prepared answers to the hard questions](#6-prepared-answers-to-the-hard-questions)
7. [Timing and day-of checklist](#7-timing-and-day-of-checklist)

---

## 1. Verdict: can we post?

Not yet. Four blockers, all small and all in files nobody else is likely to be touching.

| # | Blocker | File |
|---|---|---|
| **B1** | Playground nav links are unstyled — default browser blue on a dark navy header, ~1.7:1 contrast, effectively invisible, and they run together with no spacing | `playground/style.css` (rule absent) |
| **B2** | On a 375px phone the Output panel sits below the fold; pressing Run appears to do nothing | `playground/style.css:88`, `:168-180` |
| **B3** | A 1.5 MB WASM download with no progress indication — 8-30s of a greyed-out button on 3G | `playground/app.js:11-27`, `playground/index.html:83` |
| **B4** | `judge` / `expect` / `refuse` / `seed` pass the type checker, then die in the playground with `Hiss! undefined function judge, nya~`. The tutorial teaches these in §11 | `pkg/interpreter/interpreter.go` (cases absent) |

The good news, verified rather than assumed: the wasm **is** already served gzip-compressed with the
correct MIME type, the homepage does **not** eagerly download it, and every single example shipped
in the playground and on the homepage runs correctly.

---

## 2. Playground audit

### 2.1 Artifact size — measured

`make wasm` was run. `playground/meow.wasm` **is gitignored** (`.gitignore:4`), so building it leaves
no tracked change in the working tree.

| | bytes | |
|---|---:|---|
| `make wasm` output, uncompressed | 5,571,188 | **5.57 MB** |
| same, gzip -9 | 1,433,159 | **1.43 MB** |
| built with `-ldflags="-s -w"` | 5,435,123 | 5.44 MB |
| same, gzip -9 | 1,400,780 | 1.40 MB |
| **currently deployed, over the wire** | **1,485,877** | **1.49 MB** |

Stripping symbols saves **2.4% of the gzipped size**. Not worth doing.

### 2.2 Compression — already correct, no action needed

Live headers for `https://meow.oreha.dev/playground/meow.wasm`:

```
content-type: application/wasm
content-encoding: gzip
content-length: 1485877
vary: Accept-Encoding
cache-control: max-age=600
server: GitHub.com   (via Fastly)
```

Three things follow:

- **`WebAssembly.instantiateStreaming` works.** It requires `application/wasm`, and GitHub Pages
  sends it. `playground/app.js:14` is safe today.
- **No Brotli.** GitHub Pages does not offer `br`. Brotli would plausibly reach ~1.1 MB, a ~25% win,
  but getting it means moving the site to Cloudflare Pages or Netlify. Out of scope for this launch;
  worth an issue.
- **`cache-control: max-age=600`** is GitHub Pages' fixed 10-minute TTL and is not configurable.
  Revalidation uses the weak ETag and returns 304, so a returning visitor re-validates rather than
  re-downloads. Acceptable.

### 2.3 What a 3G visitor actually sees — source-level reasoning

Transfer time for 1.49 MB, using Chrome DevTools' presets:

| profile | throughput | download |
|---|---|---|
| Slow 3G | ~400 kbps | **~30 s** |
| Fast 3G | ~1.6 Mbps | **~8-9 s** |

Plus ~400ms of added RTT per round trip and ~0.3-1s of wasm compilation on a mid-range phone.

During all of that, here is the entire feedback the page gives:

- `playground/index.html:83` — `<span id="status">Loading WASM...</span>`, styled at
  `playground/style.css:78-81` as `#888` at `0.85rem` (13.6px). Small grey text in the toolbar.
- `playground/index.html:82` — `<button id="run-btn" disabled>`, greyed out.
- `playground/app.js:125-129` — the editor is **pre-filled** with the Hello World example.

So the page looks finished and idle. There is no spinner, no progress bar, no byte counter, and
nothing that says a 1.5 MB download is in flight. The most likely reading, from a phone, is that
the Run button is broken.

It is also slower to start than it needs to be: `loadWasm()` is the last line of `app.js`
(`playground/app.js:133`), and `app.js` is the third of three classic scripts at
`playground/index.html:103-105`. The 1.5 MB fetch does not begin until `wasm_exec.js`,
`examples.js` and `app.js` have all been fetched and executed.

### 2.4 Mobile at 375px — source-level reasoning

`playground/style.css` contains exactly one media query, `@media (max-width: 768px)` at lines
168-180. It does three things: stacks `.editor-panel` to a column, drops `main` padding to
`0.5rem`, and lets `.toolbar` wrap. That leaves the following.

**`.editor-panel { min-height: 400px }` at line 88 is never relaxed.** In column mode the two
`flex: 1` panels split that 400px floor, so:

- Code textarea: ~170px, about 9 lines at `14px/1.5`.
- Output panel: entirely below the fold on a 375x667 viewport.

Rough vertical budget at 375x667: header ~170px (see below) + wrapped toolbar ~80px + gaps ~20px
+ footer ~40px = ~310px of chrome, leaving ~355px — under the 400px floor, so the page scrolls and
Output is the part that falls off. **Pressing Run produces no visible change.** This is the worst
mobile defect.

**`header` keeps `padding: 1rem 2rem` (line 18) on mobile** — 64px of horizontal padding out of
375px. The `h1` "Meow Programming Language Playground" at `1.5rem` wraps to 2-3 lines and the
subtitle to 2, before the nav even starts.

**The nav is entirely unstyled.** The only anchor rule in the whole file is `footer a` at line 159.
There is no rule for `.playground-nav`, `header nav a`, or anything equivalent. So the six links at
`playground/index.html:67-74` render with the UA default `#0000EE` (visited `#551A8B`) on the
`#16213e` header. That is roughly **1.7:1** contrast against a 4.5:1 requirement — functionally
invisible. They also get no padding or separators, so they read as one run-on line: *"Home Tutorial
Specification Standard Library Cookbook GitHub"*. **This is visible on desktop too**, and it is the
single most embarrassing defect for a launch.

### 2.5 Robustness and content

- **No non-streaming fallback.** `playground/app.js:14-17` calls `instantiateStreaming` and catches
  into a dead-end error message. Correct today, but any host change, proxy, or content-type rewrite
  turns the playground permanently into `Failed to load WASM: ...`. A five-line
  `WebAssembly.instantiate(await response.arrayBuffer())` fallback removes the whole class of risk.
- **No `<noscript>`** anywhere in `playground/index.html`. With JS off: an empty editor and a dead
  button.
- **The default example is the weakest one.** `playground/app.js:127` loads `MEOW_EXAMPLES[0]`,
  which is `nya("Hello, World!")` (`playground/examples.js:5`). After 8-30 seconds of waiting, the
  payoff is one line of output. "List Operations" (index 3) or "Pattern Matching" (index 5) show
  the language off far better.
- **Nothing on the page says what the playground cannot do.** See §2.6 — `nab` of any kind is
  unsupported, but the tutorial's §10 teaches `nab "http"`.
- **Duplicated deploy path.** `website/hugo.toml` mounts `../playground` to `static/playground`,
  *and* `.github/workflows/hugo.yml` copies the same eight files explicitly in its "Copy Playground
  into Hugo output" step. Harmless, but two mechanisms for one job invites drift.

### 2.6 Interpreter behaviour — verified by execution

I ran the shipped examples and the suspected gaps directly through the same
`lexer -> parser -> checker -> interpreter` pipeline the WASM binary uses
(`cmd/playground/main_wasm.go:29-61`).

**All 9 examples in `playground/examples.js` run correctly.** Hello World, Fibonacci, FizzBuzz,
List Operations, Kitty & Groom, Pattern Matching, Error Handling, Collar (Newtype), Pure Functions
(Trill). **All 5 code blocks on the homepage also run correctly** — the hero example and the four
behind the "▶ Run" buttons.

**But nothing in CI checks this.** `MEOW_EXAMPLES` appears only in `playground/examples.js` and
`playground/app.js`; no Go test references it. A future language change can silently break every
example the playground ships, and the first person to notice would be an HN reader.

Gaps a visitor can hit, all reproduced:

| input | result |
|---|---|
| `judge(1 == 1)` | `Hiss! undefined function judge, nya~` — **checker accepts it, interpreter does not implement it** |
| `expect(2, 2)` | same. `refuse` and `seed` are the same case |
| `nab "http"` | `Hiss! nab "http" is not supported in the playground, nya~` — honest message, fine |
| `nab go "strings"` | same clear message, fine |
| a 10M-step loop | `Hiss! step limit exceeded (10000000 steps), nya~`, with partial output preserved. Good behaviour |

The `nab` messages are genuinely good — they name the limitation. `judge`/`expect`/`refuse`/`seed`
are the problem: a program the type checker accepted dies claiming a documented builtin does not
exist, and `website/content/learn/tutorial.md` §11 teaches exactly those four. Someone reading the
tutorial in one tab and pasting into the playground in the other will hit this.

### 2.7 Prioritised fix list

#### Blockers — fix before posting

- **B1. Style the playground nav.** `playground/style.css` — add a `.playground-nav a` rule near
  line 158 (colour, spacing, hover). Currently no rule exists; the links are default blue on dark
  navy at ~1.7:1 contrast with no separation. Affects every viewport.
- **B2. Make Output reachable on mobile.** `playground/style.css:168-180` — inside the existing
  `max-width: 768px` block add `.editor-panel { min-height: 0 }` (overriding line 88), give
  `#editor` and `#output` explicit heights such as `min-height: 40vh` / `25vh`, and shrink
  `header { padding: 0.75rem 1rem }` and `header h1 { font-size: 1.15rem }`.
- **B3. Report download progress.** `playground/app.js:11-27` — read `response.body` through a
  `ReadableStream` reader so you can show *"Loading compiler… 640 KB / 1.5 MB"*, and surface that
  in the Output panel rather than the 13.6px grey `#status` span at `playground/index.html:83`.
  State the cost up front: *"First run downloads a ~1.5 MB WebAssembly build of the compiler."*
- **B4. Fix the `judge`/`expect`/`refuse`/`seed` trap.** `pkg/interpreter/interpreter.go` — either
  implement the four in `dispatchBuiltin`, or have the checker reject them under the interpreter
  backend with a message that says *why*. Shipping "undefined function" for a documented builtin
  is the kind of thing that becomes the top comment.

#### Should fix — cheap, meaningfully better

- **S1.** Add a non-streaming `instantiate` fallback — `playground/app.js:14-17`.
- **S2.** Change the default example away from Hello World — `playground/app.js:127`.
- **S3.** Add a Go test that runs every `MEOW_EXAMPLES` entry through `pkg/interpreter`. Today
  `playground/examples.js` has no Go-side consumer at all.
- **S4.** Add a "what the playground can't do" note to `playground/index.html` — one line under the
  subtitle at line 66, covering `nab` and the step limit, linking to the CLI for the rest.
- **S5.** Add a playground CTA to the hero. `website/layouts/index.html:19-22` currently offers
  only "Get Started" and "Download"; both ask the reader to leave or install. See §3.3.
- **S6.** Preload the wasm. Add
  `<link rel="preload" as="fetch" type="application/wasm" crossorigin href="meow.wasm">` to
  `playground/index.html` `<head>` near line 50, so the download overlaps the three scripts at
  lines 103-105 rather than queuing behind them.

#### Nice to have

- **N1.** `<noscript>` in `playground/index.html`.
- **N2.** Escape the raw `<` at `website/layouts/index.html:155` (`(n <= `). Hugo's minifier
  normalises it to `&lt;=` in production — verified on the live page — so it works, but it is
  invalid source relying on minifier behaviour, in the most-clicked example on the site.
- **N3.** Do **not** bother with `-ldflags="-s -w"`: measured at 2.4% off the gzipped size.
- **N4.** Deduplicate the two playground copy mechanisms (§2.5).
- **N5.** File an issue for Brotli, which needs a host move (§2.2).

#### Verified fine — no action

- wasm is gzipped and served as `application/wasm` (§2.2).
- The homepage lazy-loads the wasm on first Run click, not on page load
  (`website/static/js/main.js:49-73`). Only `wasm_exec.js` (~17 KB) loads eagerly
  (`website/layouts/index.html:288`).
- `getBaseURL()` in `website/static/js/main.js:35-47` falls back to a hardcoded `/meow/`, but
  `website/layouts/_default/baseof.html:10` always emits a canonical link, so it resolves to `/`
  on the apex domain. Verified correct against the live page. It is a landmine only if that
  canonical tag is ever removed.
- `pre { overflow-x: auto }` (`website/static/css/style.css:88`) keeps long code lines from
  breaking the mobile layout.

---

## 3. English review of the landing page and tutorial

**Headline:** the tutorial is genuinely good — clean, idiomatic, well-sequenced, no awkwardness
worth flagging. The problems are all on the landing page, and the biggest one is that the best
sentence you have written about this project is in the README and not on the site.

### 3.1 The SEO stuffing is the first thing an HN reader will notice

The exact phrase "Meow Programming Language" appears in the `<h1>`, the hero paragraph, and three
of five section headings:

| line | current | suggested |
|---|---|---|
| `website/layouts/index.html:56` | `What is Meow Programming Language?` | `What is Meow?` |
| `:68` | `Why Meow Programming Language?` | `Why you might actually use it` |
| `:231` | `Frequently Asked Questions about Meow Programming Language` | `FAQ` |
| `:284` | `Read the Meow Programming Language Tutorial` | `Read the tutorial` |

Nobody writes like this, and HN is specifically allergic to it. It reads as machine-optimised, and
that colours how the reader reads everything else on the page — including the parts that are true
and impressive. The SEO cost of these four edits is near zero; the page title, meta description and
FAQ answer bodies still carry the phrase.

### 3.2 Put the README's line on the site

`README.md:44` says:

> It's a joke language — but one that actually works.

That is the whole value proposition and it disarms the "why another toy language" reflex before
anyone types it. The hero says something flat instead.

**Current** (`website/layouts/index.html:15-18`):

> Meow Programming Language is a cat-themed functional programming language that transpiles
> `.nyan` files to Go and compiles to native binaries.

**Suggested:**

> Meow is a joke language that actually works. Every keyword is a cat word — `nyan`, `meow`,
> `purr`, `hiss` — and your `.nyan` file transpiles to real Go, then compiles to a native binary.
> Static types, pattern matching, a formatter, a linter, and a test runner with coverage, fuzzing
> and mutation testing. A joke is funnier when it's load-bearing.

### 3.3 Nothing in the first screenful says "try it now"

`website/layouts/index.html:19-22` offers **Get Started** (→ tutorial) and **Download** (→ GitHub
releases). Both ask the reader to leave or to install something. The playground — the strongest
asset, and the reason this submission is worth clicking — is reachable only from the nav, from the
FAQ at line 246, and from per-example buttons far down the page.

Add a primary **Try it in your browser →** button pointing at `/playground/` and demote Download to
tertiary. This is listed as **S5** above because the Show HN URL recommendation in §4 depends on it.

### 3.4 The performance claim will get picked apart

`website/layouts/index.html:60-61`:

> …so your `.nyan` programs run at the same speed as hand-written Go code.

and line 78:

> Transpiles to Go and compiles to native binaries. Your cat code runs at full speed.

This is not what the generated code shows, and `meow transpile` is a subcommand you advertise, so
someone will check within five minutes. What actually happens:

A function whose parameters and return type are **all** scalars gets a genuinely native body —

```
meow calc(n int) int {
  nyan x int = n * 2
  nyan y = n + 1
  bring x + y
}
```
```go
func calc(n int64) int64 {
	var x int64 = (n * int64(2))
	var y int64 = (n + int64(1))
	return meow.Returning(__caller, (x + y))
}
```

— but one `list[int]` parameter, or a struct return, and the *whole* function falls back to the
boxed path. Top-level code is always boxed: `nyan a int = 3` becomes `var a meow.Value` even with
the annotation, and `a + b` becomes `meow.Add(a, b)`. Lambdas are always boxed. And every statement
emits a `meow.Here("file:line:col")` call so runtime errors can point at Meow source.

Being specific here is both more honest *and* more interesting than the current claim:

| line | current | suggested |
|---|---|---|
| `:60-61` | `…so your .nyan programs run at the same speed as hand-written Go code.` | `Annotate a function's parameters and return type and the whole body compiles to plain, unboxed Go — int really is int64, + really is +. Top-level code stays dynamically typed. Run meow transpile and read exactly what you got.` |
| `:78` | `Transpiles to Go and compiles to native binaries. Your cat code runs at full speed.` | `Transpiles to Go and compiles to a native binary. Fully annotated functions come out as ordinary Go with no boxing — run meow transpile and check.` |

### 3.5 Smaller landing-page notes

- `:92` — the "Pipe Operator" card uses 🚧, the construction-sign emoji. It reads as *"not finished
  yet"*. Swap it.
- `:88` — "Powerful `peek` expressions with ranges, wildcards, and multi-case support." *Powerful*
  is filler; every language claims it. → "`peek` expressions with ranges, wildcards and multi-case
  arms."
- `website/content/learn/_index.md:6-7` — "Start your journey with the Meow Programming Language,
  the purrfect cat-themed language that transpiles `.nyan` files to Go." *Start your journey with*
  is stock marketing filler. → "Meow transpiles `.nyan` files to Go. Start here."
- The tagline "The purrfect functional programming language" (`website/hugo.toml`) is on-brand and
  worth keeping, but it carries no information — which is exactly why §3.2 matters.

### 3.6 The tutorial opens with friction

`website/content/learn/tutorial.md:9-18` leads with **Prerequisites: Go 1.26+ installed, Meow
compiler installed**, and the install link points at `/community/contributing#clone-and-build` —
i.e. clone the repo and build from source. Meanwhile the homepage says
`brew install 135yshr/homebrew-tap/meow`. Inconsistent, and it puts a toolchain install between an
HN reader and their first line of Meow.

**Suggested replacement for lines 9-18:**

> ## Before you start
>
> Nothing to install — every example on this page runs in the [Playground](/playground/).
>
> To run Meow locally you also need the Go toolchain (Go 1.26+), because `meow build` hands the
> generated Go to `go build`:
>
> ```bash
> brew install 135yshr/homebrew-tap/meow
> meow version
> ```

Also: **"Next Steps" at line 499 lists six links and none of them is the Playground.** Add one.

---

## 4. Ready-to-paste: Show HN

### 4.1 Title

1. `Show HN: Meow – a cat-themed functional language that transpiles to Go` ← **recommended**
2. `Show HN: Meow – a joke language that compiles to real native binaries`
3. `Show HN: I made a language where every keyword is a cat noise, and it transpiles to Go`

**Take #1**, the issue's own suggestion. HN titles reward concreteness over cleverness, and
"transpiles to Go" is the load-bearing fact that tells a skimmer this is not a 200-line weekend toy
— it is the single word that earns the click from the Go crowd. #2 is punchier, but putting "joke
language" in the *title* hands people permission to dismiss it without clicking; that line lands far
better in the first comment, where it reads as self-awareness rather than as a disclaimer. #3 is
first-person and slightly try-hard, which HN punishes.

Keep the en dash after "Meow" — it is the HN house style for `Show HN: Name – description`.

### 4.2 URL

**`https://meow.oreha.dev/`** — not the playground direct link.

The homepage has runnable examples inline, explains what the thing is, and links onward. A cold
playground link means a 1.5 MB download with no context. **This recommendation assumes S5 is done**
— if the hero still has no "Try it in your browser" button when you post, switch the submission URL
to `https://meow.oreha.dev/playground/` and accept the download, because getting people *running
code* matters more than getting them reading.

### 4.3 First comment

> Author here.
>
> Meow started as a joke: could I write a language where every keyword is a cat noise and still end
> up with something I'd actually use? `nyan` is var, `meow` is func, `sniff`/`scratch` are if/else,
> `purr` loops, `paw` is a lambda, `nya` prints. Errors come back as `Hiss! ..., nya~`. There are 34
> keywords and every one of them is a cat word.
>
> The part I didn't expect to be interesting is what happens underneath. Three things, in case
> anyone wants to argue with me about them.
>
> **1. The type checker decides how much of your program stays Go.**
>
> Meow is gradually typed, but the boundary isn't where I first assumed it would be. Function
> signatures are mandatory — parameters *and* return type. In exchange, if every one of those types
> is a scalar (`int`, `byte`, `float`, `string`, `bool`), the entire body compiles to ordinary
> unboxed Go, with locals inferred from the parameters:
>
>     meow calc(n int) int {
>       nyan x int = n * 2
>       nyan y = n + 1        # no annotation, inferred
>       bring x + y
>     }
>
> becomes
>
>     func calc(n int64) int64 {
>         var x int64 = (n * int64(2))
>         var y int64 = (n + int64(1))
>         return (x + y)
>     }
>
> `+` is a Go `+`. No interface, no dynamic dispatch, no allocation.
>
> It's all-or-nothing per function, though, and that surprised me when I built it: one `list[int]`
> parameter, or a struct return, and the *whole* body falls back to the boxed path. Lists, maps,
> structs, lambdas and all top-level code are always `meow.Value`, an interface, where `a + b`
> becomes `meow.Add(a, b)` and dispatches at runtime. So a Meow program is a dynamically typed shell
> wrapped around statically typed islands, and how much of it is fast is a direct function of how
> many annotations you wrote.
>
> There's a `meow transpile` subcommand that prints the generated Go, so you can check my work
> instead of taking my word for it. I'd rather you did. It'll also show you a
> `meow.Here("file:12:3")` call in front of every statement — that's how a runtime error points at
> your `.nyan` source rather than at generated Go. It's a single global string assignment that the
> Go compiler inlines, but it's there, and I'd rather mention it than have someone find it.
>
> **2. The lexer is an `iter.Seq` and the parser has to pull on it.**
>
> `lexer.Tokens()` returns `iter.Seq[token.Token]` — a *push* sequence. You hand it a `yield` and it
> calls you. A Pratt parser wants precisely the opposite: it wants to ask for the next token and
> peek one past it. Go 1.23's `iter.Pull` inverts exactly that, handing back a `next()` backed by a
> coroutine:
>
>     next, stop := iter.Pull(tokens)
>
> The parser keeps `cur` and `peek`, fills `peek` from `next()`, and synthesises an EOF token when
> the sequence runs dry. One token of lookahead turned out to be enough for the whole grammar. It's
> a small thing, but it's the cleanest push-to-pull adapter I've written in Go, and it meant the
> lexer never had to know a parser existed.
>
> **3. There are two backends over the same front end.**
>
> The CLI transpiles to Go and shells out to `go build`. That's the real path, and it's how you get
> a native binary.
>
> The browser playground can't do that — there's no Go toolchain inside WASM. So there's a second
> backend: `pkg/interpreter`, a tree-walking evaluator, about 1,000 lines, compiled to WASM
> alongside the lexer, parser and checker. The playground runs
> `lexer → parser → checker → interpreter`; the CLI runs
> `lexer → parser → checker → codegen → go build`. They share the first three stages *and* the
> runtime library, which is why the semantics and the error messages match. The tax is that every
> semantic change has to land twice, and I pay it on most language features.
>
> Playground caveats, since you'll hit them: no `nab` imports (the stdlib and Go interop are
> compile-and-link features), and there's a 10M-step limit so a runaway loop can't hang your tab.
> It tells you when you hit either.
>
> ---
>
> Beyond the compiler it has a formatter, a linter, and `meow test`, which runs tests written *in
> Meow* with coverage, fuzzing and mutation testing. About 16k lines of Go, 73 end-to-end golden
> tests that compile and run real binaries, and zero third-party dependencies — `go.sum` is empty.
> If your program needs a Go package, `nab go "net/http"` imports it into the program being built
> rather than into Meow itself.
>
> Try it: https://meow.oreha.dev/playground/ (heads up, first load pulls ~1.5 MB of WebAssembly)
> Source: https://github.com/135yshr/meow
>
> Happy to answer anything.

**Before pasting, check:**

- If B4 is not fixed, add `judge`/`expect`/`refuse`/`seed` to the playground caveats paragraph.
- If B3 is not fixed, keep the "~1.5 MB" warning in — it is doing real work there.
- Every technical claim above was verified against the source. The `calc` example is genuine
  `meow transpile` output with the `meow.Here`/`meow.Returning` lines elided — which the comment
  then discloses two paragraphs later, deliberately. Do not remove that disclosure and keep the
  elision; that combination is dishonest.

---

## 5. Ready-to-paste: r/ProgrammingLanguages

> ⚠️ **Confirm the subreddit rules yourself before posting.** r/ProgrammingLanguages has an active
> self-promotion policy and historically restricts "look at my language" posts to specific formats
> or days. I have not checked it — read the sidebar and the pinned rules post at the time you post,
> and adjust. Reddit rules change and getting removed as spam wastes the launch.
>
> Practical guidance that is usually true there regardless: the subreddit rewards *design
> discussion* and punishes announcements. Lead with a design decision you can be argued with about,
> not with a link. Reply to every comment. Do not post the same day as the Show HN — space them by
> a day or two so you can give each thread real attention.

### 5.1 Title

`Meow: a cat-themed language that transpiles to Go — notes on where I put the gradual typing boundary`

Alternative, if the sidebar prefers a plainer format:
`Gradual typing where the boundary is the function signature: what I learned building Meow`

The second is better subreddit-fit but worse at conveying what the project *is*. If rules permit
only one post, use the first.

### 5.2 Body

> I've been building [Meow](https://github.com/135yshr/meow), a language where every keyword is a
> cat word (`nyan` = var, `meow` = func, `purr` = loop, `paw` = lambda, `hiss` = error). It
> transpiles to Go and then compiles to a native binary via the Go toolchain. The theme is a joke;
> the implementation isn't, and a couple of the design decisions turned out more interesting than I
> expected. I'd like to be argued with about the first one in particular.
>
> **Where the gradual typing boundary sits.**
>
> Most gradually typed languages let you annotate anything, anywhere, and box whatever isn't
> annotated. I ended up somewhere narrower, mostly by accident, and then kept it on purpose.
>
> Function signatures in Meow are *mandatory* — both parameters and return type. If all of those
> types are scalars (`int`, `byte`, `float`, `string`, `bool`), codegen emits a plain Go function
> over native Go types, and the checker's per-expression type info then decides which locals stay
> native inside the body. Locals get inferred; you don't annotate them. Everything outside such a
> function — top-level code, lambdas, anything touching a list, map or struct — is boxed into a
> `meow.Value` interface where operators dispatch at runtime.
>
> The consequence I didn't anticipate is that it's **all-or-nothing per function**. A single
> `list[int]` parameter, and the entire body drops to the boxed path — not just the expressions
> involving that list. It makes the performance model very easy to explain ("a fully-scalar
> signature buys you a fast body") and very coarse. The alternative — per-expression native/boxed
> decisions with unboxing at the boundaries — is obviously more precise, and I have a partial
> version of that machinery inside typed bodies already. I'm genuinely unsure whether extending it
> across the function boundary is worth the complexity for a language like this, and I'd be
> interested in how other people have drawn that line.
>
> **Push-to-pull between lexer and parser.**
>
> The lexer is a Go 1.23 `iter.Seq[token.Token]` — a push sequence: you hand it a `yield` and it
> drives. The Pratt parser wants pull: give me the next token, let me peek one past it. `iter.Pull`
> inverts it coroutine-style, so the parser is `next, stop := iter.Pull(tokens)` plus `cur`/`peek`
> fields, and synthesises EOF when the sequence drains. One token of lookahead covered the whole
> grammar. Nothing revolutionary, but it's the nicest version of that adapter I've written, and the
> lexer stays completely unaware that a parser exists.
>
> **Two backends over one front end.**
>
> The CLI path is `lexer → parser → checker → codegen → go build`. The browser playground can't
> shell out to a Go toolchain, so there's a second backend — a ~1,000-line tree-walking interpreter
> — and the playground path is `lexer → parser → checker → interpreter`. Both share the front end
> *and* the runtime value library, so semantics and error messages stay identical. The cost is that
> every semantic change lands twice. I keep expecting to regret this and so far I don't: having a
> second implementation of the semantics has caught real bugs in the first one.
>
> Also has a formatter, a linter, and a test runner that runs tests written in Meow, with coverage,
> fuzzing and mutation testing. ~16k lines of Go, no third-party dependencies.
>
> Playground (runs in-browser, ~1.5 MB wasm on first load): https://meow.oreha.dev/playground/
> Source: https://github.com/135yshr/meow
>
> Happy to go deeper on any of it.

---

## 6. Prepared answers to the hard questions

Written to be adapted, not pasted verbatim — canned-sounding replies read worse than slightly rough
honest ones. The common thread: **agree with the criticism first, then say the specific true thing.**
Never get defensive about the cat theme; it is the only unwinnable fight in the thread.

**"Why another toy language? What's the point?"**

> There isn't one, really — I built it because I wanted to. But "toy" and "unfinished" aren't the
> same thing, and the interesting part for me was refusing to let the joke be an excuse. It has a
> type checker, a formatter, a linter, mutation testing, 73 end-to-end tests that compile and run
> real binaries. The constraint "every keyword must be a cat noise" turned out to be a surprisingly
> good forcing function for actually finishing things, because the theme stops being funny the
> moment the tool doesn't work.

**"How is this different from [Brainfuck-style esolang / LOLCODE / MeowLang]?"**

> Most cat- or joke-themed languages are esoteric by design — the point is that they're hard or
> absurd to program in. Meow isn't esoteric. It's a fairly conventional statically typed functional
> language wearing a cat costume: real types, pattern matching, first-class functions, a module
> system, a stdlib. The syntax is silly; the semantics are boring on purpose. It also compiles to a
> native binary rather than being interpreted, which most of the joke languages don't.

**"Is the cat theme a joke, or do you expect people to use this?"**

> The theme is a joke. The compiler isn't. I use it for small scripts and I wouldn't tell anyone to
> put it in production — not because it's broken, but because a one-maintainer language with 34
> cat-noise keywords is a bad bet for code somebody else has to read in three years. If you want the
> serious version of what's underneath, that's just Go.

**"Transpiling to Go means you inherit Go's semantics — isn't this just Go with silly names?"**

> Partly, and I'd rather own that than dodge it: the memory model, the GC and the numeric semantics
> are Go's, and that's deliberate. What isn't Go: the gradual typing, the pipe operator, pattern
> matching, `furball` error values as first-class data, and a runtime-boxed dynamic layer that Go
> has no equivalent of. `meow transpile` will show you exactly how much is a thin rename and how
> much isn't. Some of it is a thin rename.

**"How fast is it really?"**

> Depends entirely on annotations, and the honest answer is more interesting than "fast". Inside a
> function whose signature is all scalars you get plain Go — `int64` arithmetic, no boxing, no
> dispatch. Anywhere else you're going through an interface and paying for it. So a tight numeric
> function is Go-speed and a top-level script is more like a dynamic language. Run `meow transpile`
> and you'll see precisely which one you wrote. (There's also a global string assignment before each
> statement for error positions; it inlines, but it exists.)

**"5MB of WebAssembly for a playground?"**

> 1.5 MB gzipped over the wire, and yes, it's a lot. It's a full Go binary: lexer, parser, type
> checker and a tree-walking interpreter, plus the Go runtime, which is most of it. TinyGo would cut
> it substantially and is the obvious next thing to try. [If B3 is fixed:] It streams with a progress
> indicator now, but on a slow connection the first load is still a real wait.

**"Why not just write a real language?"**

> This is a real language, it's just one with a stupid vocabulary. The question I'd actually answer:
> why didn't I do something more original? Because I wanted to learn how the pieces fit together end
> to end — lexer through native binary — and the cat theme kept it enjoyable enough that I finished
> instead of abandoning it at the parser like the previous two attempts.

**"Does it handle concurrency / generics / [X]?"**

> Check before answering — don't guess in the thread. There are no concurrency keywords in the
> language today (verified: nothing goroutine-, channel- or spawn-like in the keyword table). Say so
> plainly; "not yet, and here's why it's awkward given the transpilation model" is a good answer,
> a vague one is not.

**If someone finds a bug in the playground during the thread:** fix it, reply with the commit, and
say the deploy takes a few minutes. Visibly fixing something inside the thread is the single best
thing that can happen to a Show HN.

---

## 7. Timing and day-of checklist

**When.** US weekday morning — roughly **08:00–10:00 US Eastern, Tuesday through Thursday**. Avoid
Monday (backlog), Friday and weekends (thin traffic and thin moderation). For JST that is late
evening, around **21:00–23:00 JST**, which is workable but means committing to staying up.

**The non-negotiable part:** post only when you can be at a keyboard for the next **4–6 hours**.
On Show HN the author's responsiveness in the first two hours is most of what determines whether the
thread lives. A post with no author replies dies regardless of the project.

**Order of operations:**

1. Fix B1–B4. Deploy. Confirm the deploy finished — `.github/workflows/hugo.yml` rebuilds the wasm
   on every push touching `playground/**`, `pkg/interpreter/**`, `pkg/lexer/**`, `pkg/parser/**`,
   `pkg/ast/**` or `pkg/checker/**`.
2. Apply §3 copy edits if you want them. They are not blockers, but §3.4 is the one most likely to
   draw a dismissive comment.
3. Hard-refresh the playground on a real phone and a real desktop. Confirm the nav is legible,
   Output is visible on the phone, and the loading state says something.
4. Submit to HN. Immediately post the §4.3 comment as the first comment — don't wait.
5. Stay in the thread.
6. Post to r/ProgrammingLanguages a day or two later, after re-reading the rules (§5).
7. Record both URLs on issue #105, which is its stated completion condition.

**Do not** ask anyone to upvote, and do not post the link in Slacks or group chats asking for
support. HN detects voting rings and penalises them, and it is the one mistake that cannot be
undone.

**Also in the issue but not covered here:** the issue lists creating an esolangs.org page as prior
work that "would improve credibility if done first." That is a separate task and is not blocking —
though note that esolangs.org is arguably the wrong venue given that §6's "how is this different
from an esolang" answer is *"it isn't one."*
