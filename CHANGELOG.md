# [v0.5.2](https://github.com/135yshr/meow/compare/v0.5.1...v0.5.2) (2026-03-19)

# [v0.5.1](https://github.com/135yshr/meow/compare/v0.5.0...v0.5.1) (2026-03-16)

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-03-03

### Added
- macOS binary signing and notarization
- Automatic currying (partial application)
- Qualified import with tag alias
- `pose` (interface definition) and `groom` (method implementation)
- `collar` (newtype wrapper) and `breed` (type alias)
- Pattern matching with `peek`
- Pipe operator `|=|`
- Error recovery operator `~>`
- Standard library: `nab "file"`, `nab "http"`
- Playground (WASM-based)
- Homebrew formula

### Fixed
- Overlapping nav icons spacing
- X card preview with twitter:image:src
- OGP Twitter Card not rendering on X

## [0.1.1] - 2026-02-23

### Fixed
- Minor bug fixes and stability improvements

## [0.1.0] - 2026-02-22

### Added
- Initial release
- Core language: `nyan`, `meow`, `sniff`, `scratch`, `purr`, `paw`, `nya`
- Literals: `yarn` (true), `hairball` (false), `catnap` (nil)
- List operations: `lick` (map), `picky` (filter), `curl` (reduce)
- Error handling: `hiss`, `gag`, `is_furball`
- `kitty` (struct definition)
- Compiler pipeline: Lexer → Parser → Checker → Codegen → Go build
- CLI: `run`, `build`, `transpile`, `test`, `version`, `help`
- Golden file test infrastructure

[Unreleased]: https://github.com/135yshr/meow/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/135yshr/meow/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/135yshr/meow/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/135yshr/meow/releases/tag/v0.1.0
