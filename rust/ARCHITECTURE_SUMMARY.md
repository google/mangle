# Mangle Rust Implementation Architecture Summary

This document summarizes the architecture of the Rust implementation of Mangle.

## 1. Abstract Syntax Tree (AST) & Parsing
**Location:** `third_party/mangle/rust/ast`, `third_party/mangle/rust/parse`

*   **Memory Management:** Utilizes a bump-pointer allocator (`bumpalo`) for efficient allocation of AST nodes.
*   **Interning:** Uses a global or thread-local string `Interner` to deduplicate identifiers.
*   **Parser:** Recursive descent parser generating AST from source string.

## 2. Intermediate Representation (IR)
**Location:** `third_party/mangle/rust/ir`

*   **Design:** A flat, vector-based representation inspired by Carbon's SemIR.
*   **Logical IR (`Inst`):** Represents the declarative logic (Rules, Atoms, Expressions).
*   **Physical Plan IR (`physical::Op`):** Represents imperative execution logic (Iterate, Scan, Filter, Insert).

## 3. Analysis & Lowering
**Location:** `third_party/mangle/rust/analysis`

*   **Lowering (`lowering.rs`):** Converts AST into Logical IR.
*   **Planner (`planner.rs`):** Transforms Logical IR into Physical Plan IR (nested-loop joins).
*   **Type Checking (`type_check.rs`):** Validates predicates and types.

## 4. Driver & Orchestration
**Location:** `third_party/mangle/rust/driver`

*   **Role:** Orchestrates the entire compilation and execution pipeline.
*   **Stratification:** Handles program stratification (in `SimpleProgram`) to correctly evaluate negation.
*   **API:** Provides high-level `compile` and `execute` functions.

## 5. Execution Modes

### A. Server Mode (WASM)
**Location:** `third_party/mangle/rust/codegen`, `third_party/mangle/rust/vm`

*   **Codegen:** Translates Physical IR to WebAssembly.
*   **VM:** Executes WASM using `wasmtime` with host functions for data storage operations.
*   **Pluggable Storage:** The VM defines a `Host` trait that abstracts data access (`scan`, `insert`). This allows for modular, pluggable relation storage (e.g., in-memory, B-Tree, external DB) without changing the core engine or generated code.

### B. Edge Mode (Interpreter)
**Location:** `third_party/mangle/rust/interpreter`

*   **Interpreter:** Directly interprets Physical IR operations (`Op`).
*   **State:** Manages in-memory fact storage for local execution.

## 6. Data Storage & Interfaces (New)
**Location:** `third_party/mangle/rust/factstore`

*   **Role:** Acts as the central interface definition crate to prevent dependency cycles.
*   **Interfaces:**
    *   `Store` (Feature `edge`): The abstract interface for storage in Edge mode, used by `mangle-interpreter`.
    *   `Host` (Feature `server`): The abstract interface for the host environment in Server mode, used by `mangle-vm`.
*   **Implementations:**
    *   `MemStore` (in `mangle-interpreter`): In-memory implementation of `Store`.
    *   `SimpleColumnStore`/`SimpleColumnHost` (in `mangle-simplecolumn`): Implementations for the SimpleColumn file format.

## 7. Key Data Structures

*   **`InstId`**: Reference to an instruction in the IR.
*   **`NameId`**: Interned string reference.
*   **`physical::Op`**: Imperative operation (e.g., `Iterate`, `Insert`).

## 8. TODO Parity Goals (Go vs Rust)

*   **Primitives**: Check support for all Go primitives (int, float, bytes, etc.) in Rust (`i64`, `String` currently dominant).
*   **Functions**: Match built-in function availability.
*   **Parser**: Ensure syntax compliance matches the Go reference parser.