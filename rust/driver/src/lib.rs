// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//! The Mangle Driver.
//!
//! This crate acts as the orchestrator for the Mangle compiler pipeline.
//! It connects parsing, analysis, and execution components to provide a
//! high-level API for running Mangle programs.
//!
//! # Execution Architecture
//!
//! Mangle supports multiple execution strategies:
//!
//! 1.  **Reference Implementation (Legacy)**: A naive bottom-up evaluator that operates directly on the AST.
//!     This is implemented in the `mangle-engine` crate and serves as a correctness baseline.
//!     It is not used by this driver.
//!
//! 2.  **Interpreter (Default)**: A high-performance interpreter that executes the Mangle Intermediate Representation (IR).
//!     The driver compiles source code to IR and then executes it using `mangle-interpreter`.
//!
//! 3.  **WASM Compilation**: The IR can be compiled to WebAssembly (WASM) for execution in browsers or
//!     WASM runtimes. This is handled by `mangle-codegen`.
//!
//! # Key Responsibilities
//!
//! *   **Compilation**: Parsing source code and lowering it to the Intermediate Representation (IR).
//! *   **Stratification**: Analyzing dependencies between predicates to determine the correct
//!     evaluation order (handling negation and recursion). This is implemented in [`Program`].
//! *   **Execution**: Running the compiled plan using the [`mangle_interpreter`].
//! *   **Codegen**: Generating WASM modules from the IR.
//!
//! # Example
//!
//! ```rust
//! use mangle_ast::Arena;
//! use mangle_driver::{compile, execute};
//!
//! let arena = Arena::new_with_global_interner();
//! let source = "p(1). q(X) :- p(X).";
//!
//! // 1. Compile
//! let (mut ir, stratified) = compile(source, &arena).expect("compilation failed");
//!
//! // 2. Execute
//! let store = Box::new(mangle_interpreter::MemStore::new());
//! let interpreter = execute(&mut ir, &stratified, store).expect("execution failed");
//! ```

use anyhow::{Result, anyhow};
use ast::Arena;
use fxhash::FxHashSet;
use mangle_analysis::{LoweringContext, Planner, Program, StratifiedProgram, rewrite_unit};
use mangle_ast as ast;
use mangle_codegen::{Codegen, WasmImportsBackend};
use mangle_interpreter::{Interpreter, Store};
use mangle_ir::{Inst, InstId, Ir};
use mangle_parse::Parser;

/// Compiles source code into the Mangle Intermediate Representation (IR).
///
/// This function performs:
/// 1.  Parsing of the source string into an AST.
/// 2.  **Renaming**: Applies package rewrites to support module namespacing.
/// 3.  **Stratification**: Orders the evaluation of rules.
/// 4.  **Lowering**: Converts the AST into the flat IR.
///
/// Returns a tuple containing the IR and the stratification info (which dictates
/// the order of execution).
pub fn compile<'a>(source: &str, arena: &'a Arena) -> Result<(Ir, StratifiedProgram<'a>)> {
    let mut parser = Parser::new(arena, source.as_bytes(), "source");
    parser.next_token().map_err(|e| anyhow!(e))?;
    let unit = parser.parse_unit()?;

    // Apply package renaming
    let rewritten_unit = rewrite_unit(arena, unit);
    let unit = &rewritten_unit;

    let mut program = Program::new(arena);
    let mut all_preds = FxHashSet::default();
    let mut idb_preds = FxHashSet::default();

    for clause in unit.clauses {
        program.add_clause(arena, clause);
        idb_preds.insert(clause.head.sym);
        all_preds.insert(clause.head.sym);
        for premise in clause.premises {
            if let ast::Term::Atom(atom) = premise {
                all_preds.insert(atom.sym);
            } else if let ast::Term::NegAtom(atom) = premise {
                all_preds.insert(atom.sym);
            }
        }
    }

    for pred in all_preds {
        if !idb_preds.contains(&pred) {
            program.ext_preds.push(pred);
        }
    }

    let stratified = program.stratify().map_err(|e| anyhow!(e))?;

    let ctx = LoweringContext::new(arena);
    let ir = ctx.lower_unit(unit);

    Ok((ir, stratified))
}

/// Compiles the Intermediate Representation (IR) into a WebAssembly (WASM) module.
///
/// This uses the default `WasmImportsBackend` which expects certain host functions
/// to be available for data access.
pub fn compile_to_wasm(ir: &mut Ir, stratified: &StratifiedProgram) -> Vec<u8> {
    let mut codegen = Codegen::new_with_stratified(ir, stratified, WasmImportsBackend);
    codegen.generate()
}

/// Executes a compiled Mangle program using the pure Rust interpreter.
///
/// This function:
/// 1.  Iterates through the strata defined in `StratifiedProgram`.
/// 2.  Identifies recursive predicates within each stratum.
/// 3.  Executes non-recursive strata once.
/// 4.  Executes recursive strata using a semi-naive evaluation loop.
///
/// Returns the `Interpreter` instance, which holds the final state (facts) of the execution.
pub fn execute<'a>(
    ir: &'a mut Ir,
    stratified: &StratifiedProgram<'a>,
    store: Box<dyn Store + 'a>,
) -> Result<Interpreter<'a>> {
    let arena = stratified.arena();

    // 1. Pre-plan everything that needs mutable access to IR
    let mut strata_plans = Vec::new();

    for stratum in stratified.strata() {
        let mut stratum_pred_names = FxHashSet::default();
        for pred in &stratum {
            if let Some(name) = arena.predicate_name(*pred) {
                stratum_pred_names.insert(name);
            }
        }

        // Identify rules for this stratum
        let mut rule_ids = Vec::new();
        for (i, inst) in ir.insts.iter().enumerate() {
            if let Inst::Rule { head, .. } = inst
                && let Inst::Atom { predicate, .. } = ir.get(*head)
            {
                let head_name = ir.resolve_name(*predicate);
                if stratum_pred_names.contains(head_name) {
                    rule_ids.push(InstId::new(i));
                }
            }
        }

        if rule_ids.is_empty() {
            strata_plans.push(None);
            continue;
        }

        // Check if any rule in the stratum is recursive
        let mut is_recursive = false;
        for &rule_id in &rule_ids {
            if let Inst::Rule { premises, .. } = ir.get(rule_id) {
                for &premise in premises {
                    if let Inst::Atom { predicate, .. } = ir.get(premise) {
                        let pred_name = ir.resolve_name(*predicate);
                        if stratum_pred_names.contains(pred_name) {
                            is_recursive = true;
                            break;
                        }
                    }
                }
            }
            if is_recursive {
                break;
            }
        }

        if !is_recursive {
            let mut ops = Vec::new();
            for rule_id in rule_ids {
                let planner = Planner::new(ir);
                ops.push(planner.plan_rule(rule_id)?);
            }
            strata_plans.push(Some(StratumPlan::NonRecursive(ops)));
        } else {
            let mut initial_ops = Vec::new();
            for &rule_id in &rule_ids {
                let planner = Planner::new(ir);
                initial_ops.push(planner.plan_rule(rule_id)?);
            }

            let mut delta_plans = Vec::new();
            for &rule_id in &rule_ids {
                let premises = if let Inst::Rule { premises, .. } = ir.get(rule_id) {
                    premises.clone()
                } else {
                    continue;
                };

                for &premise in &premises {
                    let (predicate, pred_name) =
                        if let Inst::Atom { predicate, .. } = ir.get(premise) {
                            (*predicate, ir.resolve_name(*predicate).to_string())
                        } else {
                            continue;
                        };

                    if stratum_pred_names.contains(pred_name.as_str()) {
                        let planner = Planner::new(ir).with_delta(predicate);
                        delta_plans.push(planner.plan_rule(rule_id)?);
                    }
                }
            }
            strata_plans.push(Some(StratumPlan::Recursive {
                initial_ops,
                delta_plans,
            }));
        }
    }

    // 2. Now execute using the interpreter
    let mut interpreter = Interpreter::new(ir, store);

    // Initialize EDB relations
    for pred in stratified.extensional_preds() {
        if let Some(name) = arena.predicate_name(pred) {
            interpreter.store_mut().create_relation(name);
        }
    }

    for plan in strata_plans {
        match plan {
            Some(StratumPlan::NonRecursive(ops)) => {
                for op in ops {
                    interpreter.execute(&op)?;
                }
            }
            Some(StratumPlan::Recursive {
                initial_ops,
                delta_plans,
            }) => {
                for op in initial_ops {
                    interpreter.execute(&op)?;
                }
                interpreter.store_mut().merge_deltas();

                loop {
                    let mut changes = 0;
                    for op in &delta_plans {
                        changes += interpreter.execute(op)?;
                    }
                    if changes == 0 {
                        break;
                    }
                    interpreter.store_mut().merge_deltas();
                }
            }
            None => {}
        }
        interpreter.store_mut().merge_deltas();
    }

    Ok(interpreter)
}

enum StratumPlan {
    NonRecursive(Vec<mangle_ir::physical::Op>),
    Recursive {
        initial_ops: Vec<mangle_ir::physical::Op>,
        delta_plans: Vec<mangle_ir::physical::Op>,
    },
}

#[cfg(test)]
mod tests {
    use super::*;
    use mangle_interpreter::{MemStore, Value};

    #[test]
    fn test_driver_e2e() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            p(1).
            p(2).
            q(X) :- p(X).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let store = Box::new(MemStore::new());
        let interpreter = execute(&mut ir, &stratified, store)?;

        // Check results
        let facts: Vec<_> = interpreter
            .store()
            .scan("q")
            .expect("relation q not found")
            .collect();
        assert!(!facts.is_empty(), "relation q not found");

        let mut values: Vec<i64> = facts
            .iter()
            .map(|t| match t[0] {
                Value::Number(n) => n,
                _ => panic!("expected number"),
            })
            .collect();
        values.sort();

        assert_eq!(values, vec![1, 2]);

        Ok(())
    }

    #[test]
    fn test_driver_e2e_with_package() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            Package pkg!
            p(1).
            q(X) :- p(X).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let store = Box::new(MemStore::new());
        let interpreter = execute(&mut ir, &stratified, store)?;

        // Check results - predicates should be prefixed with "pkg."
        let facts: Vec<_> = interpreter
            .store()
            .scan("pkg.q")
            .expect("relation pkg.q not found")
            .collect();
        assert!(!facts.is_empty(), "relation pkg.q not found");

        let values: Vec<i64> = facts
            .iter()
            .map(|t| match t[0] {
                Value::Number(n) => n,
                _ => panic!("expected number"),
            })
            .collect();
        assert_eq!(values, vec![1]);

        Ok(())
    }

    #[test]
    fn test_driver_let_transform() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            p(1).
            p(2).
            q(Y) :- p(X) |> let Y = fn:plus(X, 10).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let store = Box::new(MemStore::new());
        let interpreter = execute(&mut ir, &stratified, store)?;

        let facts: Vec<_> = interpreter
            .store()
            .scan("q")
            .expect("relation q not found")
            .collect();
        let mut values: Vec<i64> = facts
            .iter()
            .map(|t| match t[0] {
                Value::Number(n) => n,
                _ => panic!("expected number"),
            })
            .collect();
        values.sort();

        assert_eq!(values, vec![11, 12]);
        Ok(())
    }

    #[test]
    fn test_driver_aggregation() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            p(1, 10).
            p(1, 20).
            p(2, 30).
            q(K, S) :- p(K, V) |> do fn:group_by(K); let S = fn:sum(V).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let store = Box::new(MemStore::new());
        let interpreter = execute(&mut ir, &stratified, store)?;

        let facts: Vec<_> = interpreter
            .store()
            .scan("q")
            .expect("relation q not found")
            .collect();
        let mut results: Vec<(i64, i64)> = facts
            .iter()
            .map(|t| {
                if let (Value::Number(k), Value::Number(s)) = (&t[0], &t[1]) {
                    (*k, *s)
                } else {
                    panic!("expected numbers");
                }
            })
            .collect();
        results.sort();

        assert_eq!(results, vec![(1, 30), (2, 30)]);
        Ok(())
    }

    #[test]
    fn test_driver_aggregation_count() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            p(1, 10).
            p(1, 20).
            p(2, 30).
            q(K, C) :- p(K, V) |> do fn:group_by(K); let C = fn:count(V).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let store = Box::new(MemStore::new());
        let interpreter = execute(&mut ir, &stratified, store)?;

        let facts: Vec<_> = interpreter
            .store()
            .scan("q")
            .expect("relation q not found")
            .collect();
        let mut results: Vec<(i64, i64)> = facts
            .iter()
            .map(|t| {
                if let (Value::Number(k), Value::Number(c)) = (&t[0], &t[1]) {
                    (*k, *c)
                } else {
                    panic!("expected numbers");
                }
            })
            .collect();
        results.sort();

        assert_eq!(results, vec![(1, 2), (2, 1)]);
        Ok(())
    }

    #[test]
    fn test_driver_reachability() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            edge(1, 2).
            edge(2, 3).
            edge(3, 4).
            edge(4, 5).
            reachable(X, Y) :- edge(X, Y).
            reachable(X, Z) :- reachable(X, Y), edge(Y, Z).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let store = Box::new(MemStore::new());
        let interpreter = execute(&mut ir, &stratified, store)?;

        let facts: Vec<_> = interpreter
            .store()
            .scan("reachable")
            .expect("reachable relation not found")
            .collect();
        assert_eq!(facts.len(), 10); // (1,2),(1,3),(1,4),(1,5), (2,3),(2,4),(2,5), (3,4),(3,5), (4,5)

        let mut reachable_from_1: Vec<i64> = facts
            .iter()
            .filter(|t| t[0] == Value::Number(1))
            .map(|t| match t[1] {
                Value::Number(n) => n,
                _ => panic!("expected number"),
            })
            .collect();
        reachable_from_1.sort();
        assert_eq!(reachable_from_1, vec![2, 3, 4, 5]);

        Ok(())
    }

    #[test]
    fn test_compile_to_wasm() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            p(1).
            q(X) :- p(X).
        "#;

        let (mut ir, stratified) = compile(source, &arena)?;
        let wasm_bytes = compile_to_wasm(&mut ir, &stratified);

        // Basic check that we generated something that looks like WASM
        assert!(!wasm_bytes.is_empty());
        assert_eq!(&wasm_bytes[0..4], b"\0asm"); // WASM magic header

        Ok(())
    }
}
