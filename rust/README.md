# Mangle (Rust)

[Mangle](https://mangle.readthedocs.io/en/latest/) is a language for
deductive database programming based on Datalog.

This directory contains the Rust implementation of Mangle, featuring a modern compiler pipeline that supports two execution modes:

1.  **Server Mode**: Compiles to WebAssembly (WASM) for high-performance execution in a server environment.
2.  **Edge Mode**: Uses a pure Rust interpreter for lightweight execution on edge devices or where WASM is not available.

## Architecture

The Mangle compiler pipeline transforms declarative Datalog rules into efficient, executable logic.

### Pipeline Stages

1.  **Parsing & AST (`mangle_ast`, `mangle_parse`)**:
    *   Parses Mangle source code into a high-level Abstract Syntax Tree.
    *   Uses an Arena-based allocation strategy for efficient memory usage.

2.  **Intermediate Representation (`mangle_ir`)**:
    *   Converts the pointer-based AST into a flat, indexed Intermediate Representation (IR).
    *   Includes a **Physical Plan IR** (`mangle_ir::physical`) for imperative operations (Scan, Join, Insert).

3.  **Analysis & Lowering (`mangle_analysis`)**:
    *   **Lowering**: Translates AST to Logical IR.
    *   **Planner**: Converts declarative Rules into imperative nested-loop join plans.
    *   **Type Checking**: Validates type consistency and arity constraints.

4.  **Driver (`mangle_driver`)**:
    *   Orchestrates the pipeline (Parse -> Stratify -> Lower -> Plan -> Execute).
    *   Provides the unified entry point for compiling and running Mangle programs.
    *   Manages program stratification for handling negation.

5.  **Execution Modes**:

    *   **Server Mode (`mangle_codegen` + `mangle_vm`)**:
        *   Translates Physical IR into WebAssembly (WASM).
        *   Executes using a WASM runtime (e.g., `wasmtime`).
        *   Ideal for high-throughput server environments.

    *   **Edge Mode (`mangle_interpreter`)**:
        *   Directly interprets the Physical IR.
        *   Pure Rust implementation with minimal dependencies.
        *   Ideal for embedding and edge usage.

## Crates

*   `mangle_ast`: Abstract Syntax Tree definitions.
*   `mangle_parse`: Parser implementation.
*   `mangle_ir`: Indexed RPN Intermediate Representation.
*   `mangle_analysis`: Type checking, lowering, and query planning.
*   `mangle_driver`: Compilation and execution orchestration.
*   `mangle_codegen`: WASM compilation backend.
*   `mangle_vm`: Runtime environment for executing compiled WASM.
*   `mangle_interpreter`: Pure Rust interpreter for Mangle IR.
*   `mangle_factstore`: (Legacy/Alternative) In-memory fact storage.
*   `mangle_engine`: (Legacy) Interpreter-based engine.

## Usage

### Compiling and Running Tests

To run all tests:

```bash
cargo test
```

### Example: Running a Program (Edge Mode)

```rust
use mangle_ast::Arena;
use mangle_driver::{compile, execute};

let arena = Arena::new_with_global_interner();
let source = "p(1). q(X) :- p(X).";

// Compile (Parse -> Stratify -> Lower)
let (mut ir, stratified) = compile(source, &arena)?;

// Execute (Plan -> Interpret)
let interpreter = execute(&mut ir, &stratified)?;

// Query Results
if let Some(facts) = interpreter.get_facts("q") {
    println!("Found {} facts for q", facts.len());
}
```
