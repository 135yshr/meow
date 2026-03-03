# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 0.2.x   | :white_check_mark: |
| < 0.2   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Meow, please report it responsibly.

**Do NOT open a public GitHub issue.**

Instead, please email: **isago@oreha.dev**

Include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Response Timeline

- **Acknowledgment**: within 48 hours
- **Initial assessment**: within 1 week
- **Fix or mitigation**: as soon as possible, depending on severity

## Scope

Since Meow transpiles to Go and compiles via `go build`, security concerns may include:

- Code injection through `.nyan` source files
- Unsafe code generation in the transpiler
- Vulnerabilities in the runtime (`runtime/meowrt`)
- Supply chain issues in the build pipeline

### Runtime File I/O and HTTP Access

The Meow standard library (`nab "file"` and `nab "http"`) provides direct access to the filesystem and network, respectively. Compiled Meow programs run with the same OS-level privileges as any native executable, similar to programs written in Go, Python, or Ruby. This means:

- **File I/O** (`file.snoop`, `file.stalk`): No path restrictions are enforced by the runtime. Programs can read any file accessible to the user running the binary.
- **HTTP client** (`http.pounce`, `http.toss`, etc.): No URL restrictions are enforced. Programs can make requests to any reachable host, including internal/private networks.

These are **by design** — Meow is a compiled language, not a sandboxed environment. If you need to restrict file or network access, use OS-level mechanisms (e.g., containers, seccomp, firewall rules).

Thank you for helping keep Meow safe!
