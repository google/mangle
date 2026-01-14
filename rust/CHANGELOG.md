# Changelog

All notable changes in mangle/rust will be documented in this file.

## upcoming - [0.4.0]

### 🚀 Features

- change of architecture, WASM execution
- seminaive evaluation, aggregation, indexed lookup

### 🐛 Bug Fixes

- add missing parser code

### ⚙️ Miscellaneous Tasks

- change AST, use interning (changes API)

## [0.1.1] - 2024-07-24

### 🚀 Features

- Add forgotten Display implementation for `mangle_ast::Term`

### 🐛 Bug Fixes

- Fix `repository` field in Cargo.toml (main reason for this release),
  pointed out in #34 (thanks!).

### ⚙️ Miscellaneous Tasks

- Add dependency on 'googletest' crate.

## [0.1.0] - 2024-06-11

- Initial set up.
