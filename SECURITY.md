# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 0.2.x   | :white_check_mark: |
| < 0.2   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Meow, please report it responsibly.

**Do NOT open a public GitHub issue.**

Instead, please email: **135yshr@gmail.com**

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

Thank you for helping keep Meow safe!
