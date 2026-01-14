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
use wasmtime::{Engine, Linker, Module, Store};

#[cfg(feature = "csv_storage")]
pub mod csv_host;

pub mod composite_host;

/// Trait for the host environment that provides storage and data access.
pub trait Host {
    fn scan_start(&mut self, rel_id: i32) -> i32;
    fn scan_delta_start(&mut self, rel_id: i32) -> i32;
    fn scan_index_start(&mut self, rel_id: i32, col_idx: i32, val: i64) -> i32;
    fn scan_aggregate_start(&mut self, rel_id: i32, description: Vec<i32>) -> i32;
    fn scan_next(&mut self, iter_id: i32) -> i32;
    fn get_col(&mut self, tuple_ptr: i32, col_idx: i32) -> i64;
    fn insert(&mut self, rel_id: i32, val: i64);
    /// Merges deltas and returns 1 if changes occurred, 0 otherwise.
    fn merge_deltas(&mut self) -> i32;
    fn debuglog(&mut self, val: i64);
}

pub struct Vm {
    engine: Engine,
}

struct HostWrapper<H>(H);

impl Vm {
    pub fn new() -> Result<Self> {
        let engine = Engine::default();
        Ok(Self { engine })
    }

    pub fn execute<H: Host + Send + 'static>(&self, wasm: &[u8], host: H) -> Result<()> {
        let module = Module::new(&self.engine, wasm)?;
        let mut store = Store::new(&self.engine, HostWrapper(host));

        let mut linker = Linker::new(&self.engine);

        linker.func_wrap(
            "env",
            "scan_start",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>, rel_id: i32| -> i32 {
                caller.data_mut().0.scan_start(rel_id)
            },
        )?;

        linker.func_wrap(
            "env",
            "scan_delta_start",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>, rel_id: i32| -> i32 {
                caller.data_mut().0.scan_delta_start(rel_id)
            },
        )?;

        linker.func_wrap(
            "env",
            "scan_index_start",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>,
             rel_id: i32,
             col_idx: i32,
             val: i64|
             -> i32 { caller.data_mut().0.scan_index_start(rel_id, col_idx, val) },
        )?;

        linker.func_wrap(
            "env",
            "scan_aggregate_start",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>,
             rel_id: i32,
             ptr: i32,
             len: i32|
             -> i32 {
                let mem = caller
                    .get_export("memory")
                    .expect("memory export not found")
                    .into_memory()
                    .expect("not a memory");

                let data = mem.data(&caller);

                // Read description from memory

                // Safety: Bounds check is implicitly done by slice indexing, will panic if OOB.

                // ptr is byte offset. len is number of i32s.

                let start = ptr as usize;

                let end = start + (len as usize) * 4;

                let bytes = &data[start..end];

                let mut desc = Vec::with_capacity(len as usize);

                for chunk in bytes.chunks_exact(4) {
                    let val = i32::from_le_bytes(chunk.try_into().unwrap());

                    desc.push(val);
                }

                caller.data_mut().0.scan_aggregate_start(rel_id, desc)
            },
        )?;

        linker.func_wrap(
            "env",
            "scan_next",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>, iter_id: i32| -> i32 {
                caller.data_mut().0.scan_next(iter_id)
            },
        )?;

        linker.func_wrap(
            "env",
            "get_col",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>, ptr: i32, idx: i32| -> i64 {
                caller.data_mut().0.get_col(ptr, idx)
            },
        )?;

        linker.func_wrap(
            "env",
            "insert",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>, rel_id: i32, val: i64| {
                caller.data_mut().0.insert(rel_id, val);
            },
        )?;

        linker.func_wrap(
            "env",
            "merge_deltas",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>| -> i32 {
                caller.data_mut().0.merge_deltas()
            },
        )?;

        linker.func_wrap(
            "env",
            "debuglog",
            |mut caller: wasmtime::Caller<'_, HostWrapper<H>>, val: i64| {
                caller.data_mut().0.debuglog(val);
            },
        )?;

        let instance = linker.instantiate(&mut store, &module)?;
        let run = instance.get_typed_func::<(), ()>(&mut store, "run")?;

        run.call(&mut store, ())?;

        Ok(())
    }
}

// Minimal dummy host for tests that don't need storage
pub struct DummyHost;
impl Host for DummyHost {
    fn scan_start(&mut self, _rel_id: i32) -> i32 {
        0
    }
    fn scan_delta_start(&mut self, _rel_id: i32) -> i32 {
        0
    }
    fn scan_index_start(&mut self, _rel_id: i32, _col_idx: i32, _val: i64) -> i32 {
        0
    }
    fn scan_aggregate_start(&mut self, _rel_id: i32, _description: Vec<i32>) -> i32 {
        0
    }
    fn scan_next(&mut self, _iter_id: i32) -> i32 {
        0
    }
    fn get_col(&mut self, _ptr: i32, _idx: i32) -> i64 {
        0
    }
    fn insert(&mut self, _rel_id: i32, _val: i64) {}
    fn merge_deltas(&mut self) -> i32 {
        0
    }
    fn debuglog(&mut self, _val: i64) {}
}

#[cfg(test)]
mod tests {
    use super::*;
    use mangle_analysis::LoweringContext;
    use mangle_ast as ast;
    use mangle_codegen::{Codegen, WasmImportsBackend};
    use std::collections::HashMap;

    #[test]
    fn test_e2e_execution() -> Result<()> {
        let arena = ast::Arena::new_with_global_interner();
        let foo = arena.predicate_sym("foo", Some(1));
        let bar = arena.predicate_sym("bar", Some(1));
        let x = arena.variable("X");

        let clause = ast::Clause {
            head: arena.atom(foo, &[x]),
            premises: arena
                .alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(bar, &[x])))]),
            transform: &[],
        };
        let unit = ast::Unit {
            decls: &[],
            clauses: arena.alloc_slice_copy(&[&clause]),
        };

        let ctx = LoweringContext::new(&arena);
        let mut ir = ctx.lower_unit(&unit);

        let mut codegen = Codegen::new(&mut ir, WasmImportsBackend);
        let wasm = codegen.generate();

        let vm = Vm::new()?;
        vm.execute(&wasm, DummyHost)?;

        Ok(())
    }

    #[test]
    fn test_e2e_function() -> Result<()> {
        let arena = ast::Arena::new_with_global_interner();
        let foo = arena.predicate_sym("foo", Some(1));
        let plus = arena.function_sym("fn:plus", Some(2));

        let c1 = arena.const_(ast::Const::Number(1));
        let c2 = arena.const_(ast::Const::Number(2));

        let head_arg = arena.apply_fn(plus, &[c1, c2]);
        let clause = ast::Clause {
            head: arena.atom(foo, &[head_arg]),
            premises: &[],
            transform: &[],
        };

        let unit = ast::Unit {
            decls: &[],
            clauses: arena.alloc_slice_copy(&[&clause]),
        };

        let ctx = LoweringContext::new(&arena);
        let mut ir = ctx.lower_unit(&unit);

        let mut codegen = Codegen::new(&mut ir, WasmImportsBackend);
        let wasm = codegen.generate();

        let vm = Vm::new()?;
        vm.execute(&wasm, DummyHost)?;
        Ok(())
    }

    // --- Real Implementation Test ---

    struct MemHost {
        // Map rel_id -> List of Tuples (Vec<i64>)
        data: HashMap<i32, Vec<Vec<i64>>>,
        // Iterator state: iter_id -> (rel_id, current_index)
        iters: HashMap<i32, (i32, usize)>,
        next_iter_id: i32,
    }

    impl MemHost {
        fn new() -> Self {
            Self {
                data: HashMap::new(),
                iters: HashMap::new(),
                next_iter_id: 1,
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
            self.data.entry(id).or_default().push(args);
        }

        fn get_facts(&self, rel: &str) -> Vec<Vec<i64>> {
            let id = Self::hash_name(rel);
            self.data.get(&id).cloned().unwrap_or_default()
        }
    }

    impl Host for MemHost {
        fn scan_start(&mut self, rel_id: i32) -> i32 {
            let id = self.next_iter_id;
            self.next_iter_id += 1;
            self.iters.insert(id, (rel_id, 0));
            id
        }

        fn scan_delta_start(&mut self, rel_id: i32) -> i32 {
            self.scan_start(rel_id)
        }

        fn scan_index_start(&mut self, _rel_id: i32, _col_idx: i32, _val: i64) -> i32 {
            0 // TODO: Actual index implementation
        }

        fn scan_aggregate_start(&mut self, _rel_id: i32, _description: Vec<i32>) -> i32 {
            0 // TODO: Actual aggregate implementation
        }

        fn scan_next(&mut self, iter_id: i32) -> i32 {
            if let Some((rel_id, idx)) = self.iters.get_mut(&iter_id)
                && let Some(tuples) = self.data.get(rel_id)
                && *idx < tuples.len()
            {
                // Return (iter_id << 16) | (idx + 1)
                let ptr = (iter_id << 16) | (*idx as i32 + 1);
                *idx += 1;
                return ptr;
            }
            0 // Null
        }

        fn get_col(&mut self, ptr: i32, col_idx: i32) -> i64 {
            let iter_id = ptr >> 16;
            let tuple_idx = (ptr & 0xFFFF) - 1;

            if let Some((rel_id, _)) = self.iters.get(&iter_id)
                && let Some(tuples) = self.data.get(rel_id)
            {
                return tuples[tuple_idx as usize][col_idx as usize];
            }
            0
        }

        fn insert(&mut self, rel_id: i32, val: i64) {
            self.data.entry(rel_id).or_default().push(vec![val]);
        }

        fn merge_deltas(&mut self) -> i32 {
            0
        }
        fn debuglog(&mut self, val: i64) {
            eprintln!("WASM LOG: {}", val);
        }
    }

    #[test]
    fn test_e2e_mem_store() -> Result<()> {
        let arena = ast::Arena::new_with_global_interner();
        // p(X) :- q(X).
        // q is extensional.
        let p = arena.predicate_sym("p", Some(1));
        let q = arena.predicate_sym("q", Some(1));
        let x = arena.variable("X");

        let clause = ast::Clause {
            head: arena.atom(p, &[x]),
            premises: arena.alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(q, &[x])))]),
            transform: &[],
        };

        let unit = ast::Unit {
            decls: &[],
            clauses: arena.alloc_slice_copy(&[&clause]),
        };

        let ctx = LoweringContext::new(&arena);
        let mut ir = ctx.lower_unit(&unit);

        let mut codegen = Codegen::new(&mut ir, WasmImportsBackend);
        let wasm = codegen.generate();

        // Setup Host
        let mut host = MemHost::new();
        host.add_fact("q", vec![10]);
        host.add_fact("q", vec![20]);

        let vm = Vm::new()?;

        use std::sync::{Arc, Mutex};

        #[derive(Clone)]
        struct SharedMemHost {
            inner: Arc<Mutex<MemHost>>,
        }

        impl Host for SharedMemHost {
            fn scan_start(&mut self, rel_id: i32) -> i32 {
                self.inner.lock().unwrap().scan_start(rel_id)
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
            fn scan_next(&mut self, iter_id: i32) -> i32 {
                self.inner.lock().unwrap().scan_next(iter_id)
            }
            fn get_col(&mut self, ptr: i32, idx: i32) -> i64 {
                self.inner.lock().unwrap().get_col(ptr, idx)
            }
            fn insert(&mut self, rel_id: i32, val: i64) {
                self.inner.lock().unwrap().insert(rel_id, val);
            }
            fn merge_deltas(&mut self) -> i32 {
                self.inner.lock().unwrap().merge_deltas()
            }
            fn debuglog(&mut self, val: i64) {
                self.inner.lock().unwrap().debuglog(val);
            }
        }

        let shared_host = SharedMemHost {
            inner: Arc::new(Mutex::new(host)),
        };

        vm.execute(&wasm, shared_host.clone())?; // Clone increments Arc ref

        let final_host = shared_host.inner.lock().unwrap();
        let results = final_host.get_facts("p");

        // Check contains 10 and 20
        assert!(results.iter().any(|t| t[0] == 10));
        assert!(results.iter().any(|t| t[0] == 20));

        Ok(())
    }

    #[cfg(feature = "csv_storage")]
    #[test]
    fn test_e2e_csv_host() -> Result<()> {
        use crate::csv_host::CsvHost;
        use std::io::Write;
        use std::sync::{Arc, Mutex};
        use tempfile::NamedTempFile;

        // 1. Create a CSV file
        let mut file = NamedTempFile::new()?;
        writeln!(file, "10")?;
        writeln!(file, "20")?;
        let path = file.path().to_path_buf();

        // 2. Setup CsvHost
        let mut host = CsvHost::new();
        host.add_file("q", path);

        // 3. Compile Program: p(X) :- q(X).
        let arena = ast::Arena::new_with_global_interner();
        let p = arena.predicate_sym("p", Some(1));
        let q = arena.predicate_sym("q", Some(1));
        let x = arena.variable("X");

        let clause = ast::Clause {
            head: arena.atom(p, &[x]),
            premises: arena.alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(q, &[x])))]),
            transform: &[],
        };
        let unit = ast::Unit {
            decls: &[],
            clauses: arena.alloc_slice_copy(&[&clause]),
        };

        let ctx = LoweringContext::new(&arena);
        let mut ir = ctx.lower_unit(&unit);

        let mut codegen = Codegen::new(&mut ir, WasmImportsBackend);
        let wasm = codegen.generate();

        // 4. Execute
        // Wrapper for shared host
        #[derive(Clone)]
        struct SharedCsvHost {
            inner: Arc<Mutex<CsvHost>>,
        }
        impl Host for SharedCsvHost {
            fn scan_start(&mut self, rel_id: i32) -> i32 {
                self.inner.lock().unwrap().scan_start(rel_id)
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
            fn scan_next(&mut self, iter_id: i32) -> i32 {
                self.inner.lock().unwrap().scan_next(iter_id)
            }
            fn get_col(&mut self, ptr: i32, idx: i32) -> i64 {
                self.inner.lock().unwrap().get_col(ptr, idx)
            }
            fn insert(&mut self, rel_id: i32, val: i64) {
                self.inner.lock().unwrap().insert(rel_id, val);
            }
            fn merge_deltas(&mut self) -> i32 {
                self.inner.lock().unwrap().merge_deltas()
            }
            fn debuglog(&mut self, val: i64) {
                self.inner.lock().unwrap().debuglog(val);
            }
        }

        let shared_host = SharedCsvHost {
            inner: Arc::new(Mutex::new(host)),
        };
        let vm = Vm::new()?;

        vm.execute(&wasm, shared_host)?;

        Ok(())
    }
}
