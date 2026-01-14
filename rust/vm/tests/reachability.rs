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

use anyhow::Result;
use mangle_analysis::{LoweringContext, Program};
use mangle_ast as ast;
use mangle_codegen::{Codegen, WasmImportsBackend};
use mangle_vm::{Host, Vm};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

// A simple in-memory host for testing
struct MemHost {
    // stable: rel_id -> List of Tuples
    stable: HashMap<i32, Vec<Vec<i64>>>,
    // delta: rel_id -> List of Tuples
    delta: HashMap<i32, Vec<Vec<i64>>>,
    // next_delta: rel_id -> List of Tuples
    next_delta: HashMap<i32, Vec<Vec<i64>>>,

    iters: HashMap<i32, (i32, usize, bool)>, // (rel_id, idx, is_delta)
    index_filters: HashMap<i32, (usize, i64)>, // iter_id -> (col, val)
    next_iter_id: i32,
    facts_added: bool,
}

impl MemHost {
    fn new() -> Self {
        Self {
            stable: HashMap::new(),
            delta: HashMap::new(),
            next_delta: HashMap::new(),
            iters: HashMap::new(),
            index_filters: HashMap::new(),
            next_iter_id: 1,
            facts_added: false,
        }
    }

    fn hash_name(name: &str) -> i32 {
        let mut hash: u32 = 5381;
        for c in name.bytes() {
            hash = ((hash << 5).wrapping_add(hash)).wrapping_add(c as u32);
        }
        hash as i32
    }

    fn add_fact(&mut self, rel: &str, args: Vec<i64>) {
        let id = Self::hash_name(rel);
        // Initial facts go to stable AND delta (so they are seen in first iteration)
        self.stable.entry(id).or_default().push(args.clone());
        self.delta.entry(id).or_default().push(args);
    }

    fn get_facts(&self, rel: &str) -> Vec<Vec<i64>> {
        let id = Self::hash_name(rel);
        let mut all = self.stable.get(&id).cloned().unwrap_or_default();
        if let Some(d) = self.delta.get(&id) {
            // In current model, delta is subset of stable after merge?
            // Actually, merge moves delta to stable.
            // But facts in `delta` might not be in `stable` yet if we check mid-loop?
            // For final result checking, everything should be in stable.
            // But let's check.
            for t in d {
                if !all.contains(t) {
                    all.push(t.clone());
                }
            }
        }
        if let Some(nd) = self.next_delta.get(&id) {
            for t in nd {
                if !all.contains(t) {
                    all.push(t.clone());
                }
            }
        }
        all
    }
}

impl Host for MemHost {
    fn scan_start(&mut self, rel_id: i32) -> i32 {
        let id = self.next_iter_id;
        self.next_iter_id += 1;
        self.iters.insert(id, (rel_id, 0, false));
        id
    }

    fn scan_delta_start(&mut self, rel_id: i32) -> i32 {
        let id = self.next_iter_id;
        self.next_iter_id += 1;
        self.iters.insert(id, (rel_id, 0, true));
        id
    }

    fn scan_index_start(&mut self, rel_id: i32, col_idx: i32, val: i64) -> i32 {
        // Fallback implementation: Scan stable, but only return matching rows.

        // Actually, for it to work with existing scan_next, we should store the filter in iters.

        // Let's update iters structure.

        let id = self.next_iter_id;

        self.next_iter_id += 1;

        self.iters.insert(id, (rel_id, 0, false));

        self.index_filters.insert(id, (col_idx as usize, val));

        id
    }

    fn scan_aggregate_start(&mut self, _rel_id: i32, _description: Vec<i32>) -> i32 {
        0
    }

    fn scan_next(&mut self, iter_id: i32) -> i32 {
        let filter = self.index_filters.get(&iter_id).copied();
        if let Some((rel_id, idx, is_delta)) = self.iters.get_mut(&iter_id) {
            let rel_id = *rel_id;

            let tuples_opt = if *is_delta {
                self.delta.get(&rel_id)
            } else {
                self.stable.get(&rel_id)
            };

            if let Some(tuples) = tuples_opt {
                while *idx < tuples.len() {
                    let tuple = &tuples[*idx];
                    let matches = if let Some((col, val)) = filter {
                        tuple[col] == val
                    } else {
                        true
                    };

                    if matches {
                        let ptr = (iter_id << 16) | (*idx as i32 + 1);
                        *idx += 1;
                        return ptr;
                    }
                    *idx += 1;
                }
            }
        }
        0
    }

    fn get_col(&mut self, ptr: i32, col_idx: i32) -> i64 {
        let iter_id = ptr >> 16;
        let tuple_idx = (ptr & 0xFFFF) - 1;
        if let Some((rel_id, _, is_delta)) = self.iters.get(&iter_id) {
            let tuples = if *is_delta {
                self.delta.get(rel_id)
            } else {
                self.stable.get(rel_id)
            };

            if let Some(tuples) = tuples
                && let Some(row) = tuples.get(tuple_idx as usize)
            {
                return row[col_idx as usize];
            }
        }
        0
    }

    fn insert(&mut self, rel_id: i32, val: i64) {
        let row = vec![val];
        if self.stable.get(&rel_id).is_some_and(|v| v.contains(&row))
            || self.delta.get(&rel_id).is_some_and(|v| v.contains(&row))
            || self
                .next_delta
                .get(&rel_id)
                .is_some_and(|v| v.contains(&row))
        {
            return;
        }

        self.next_delta.entry(rel_id).or_default().push(row);
        self.facts_added = true;
    }

    fn merge_deltas(&mut self) -> i32 {
        let changed = if !self.next_delta.is_empty() { 1 } else { 0 };
        for (rel, tuples) in self.delta.drain() {
            self.stable.entry(rel).or_default().extend(tuples);
        }
        self.delta = std::mem::take(&mut self.next_delta);
        self.facts_added = changed == 1;
        changed
    }

    fn debuglog(&mut self, _val: i64) {}
}

// Wrapper for thread-safety (Arc<Mutex>)
#[derive(Clone)]
struct SharedMemHost {
    inner: Arc<Mutex<MemHost>>,
}

impl Host for SharedMemHost {
    fn scan_start(&mut self, rel_id: i32) -> i32 {
        self.inner.lock().unwrap().scan_start(rel_id)
    }
    fn scan_next(&mut self, iter_id: i32) -> i32 {
        self.inner.lock().unwrap().scan_next(iter_id)
    }
    fn get_col(&mut self, ptr: i32, idx: i32) -> i64 {
        self.inner.lock().unwrap().get_col(ptr, idx)
    }
    fn insert(&mut self, rel_id: i32, val: i64) {
        self.inner.lock().unwrap().insert(rel_id, val);
    }

    fn scan_delta_start(&mut self, rel_id: i32) -> i32 {
        self.inner.lock().unwrap().scan_delta_start(rel_id)
    }

    fn scan_index_start(&mut self, rel_id: i32, col_idx: i32, val: i64) -> i32 {
        self.inner
            .lock()
            .unwrap()
            .scan_index_start(rel_id, col_idx, val)
    }

    fn scan_aggregate_start(&mut self, rel_id: i32, description: Vec<i32>) -> i32 {
        self.inner
            .lock()
            .unwrap()
            .scan_aggregate_start(rel_id, description)
    }

    fn merge_deltas(&mut self) -> i32 {
        self.inner.lock().unwrap().merge_deltas()
    }

    fn debuglog(&mut self, val: i64) {
        self.inner.lock().unwrap().debuglog(val);
    }
}

#[ignore]
#[test]
fn test_reachability_arity1() -> Result<()> {
    // Problem: Reachable nodes from node 1.
    // edge(1, 2). edge(2, 3). edge(3, 4).
    // reachable(Y) :- edge(1, Y).
    // reachable(Z) :- reachable(Y), edge(Y, Z).

    let arena = ast::Arena::new_with_global_interner();
    let edge = arena.predicate_sym("edge", Some(2));
    let reachable = arena.predicate_sym("reachable", Some(1));

    let _x = arena.variable("X");
    let y = arena.variable("Y");
    let z = arena.variable("Z");

    let c1 = arena.const_(ast::Const::Number(1));

    // Rule 1: reachable(Y) :- edge(1, Y).
    let rule1 = ast::Clause {
        head: arena.atom(reachable, &[y]),
        premises: arena
            .alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(edge, &[c1, y])))]),
        transform: &[],
    };

    // Rule 2: reachable(Z) :- reachable(Y), edge(Y, Z).
    let rule2 = ast::Clause {
        head: arena.atom(reachable, &[z]),
        premises: arena.alloc_slice_copy(&[
            arena.alloc(ast::Term::Atom(arena.atom(reachable, &[y]))),
            arena.alloc(ast::Term::Atom(arena.atom(edge, &[y, z]))),
        ]),
        transform: &[],
    };

    let unit = ast::Unit {
        decls: &[],
        clauses: arena.alloc_slice_copy(&[&rule1, &rule2]),
    };

    // Stratify
    let mut program = Program::new(&arena);
    for clause in unit.clauses {
        program.add_clause(&arena, clause);
    }
    let stratified = program.stratify().expect("stratification failed");

    let ctx = LoweringContext::new(&arena);
    let mut ir = ctx.lower_unit(&unit);

    let mut codegen = Codegen::new_with_stratified(&mut ir, &stratified, WasmImportsBackend);
    let wasm = codegen.generate();

    let mut host = MemHost::new();
    host.add_fact("edge", vec![1, 2]);
    host.add_fact("edge", vec![2, 3]);
    host.add_fact("edge", vec![3, 4]);
    host.add_fact("edge", vec![4, 5]);

    let shared_host = SharedMemHost {
        inner: Arc::new(Mutex::new(host)),
    };
    let vm = Vm::new()?;

    // The WASM now contains the fixpoint loop internally!
    vm.execute(&wasm, shared_host.clone())?;

    let final_host = shared_host.inner.lock().unwrap();
    let facts = final_host.get_facts("reachable");

    let mut values: Vec<i64> = facts.iter().map(|v| v[0]).collect();
    values.sort();

    assert_eq!(values, vec![2, 3, 4, 5]);

    Ok(())
}
