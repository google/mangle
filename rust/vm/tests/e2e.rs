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

#[cfg(feature = "csv_storage")]
use mangle_simplecolumn::host::SimpleColumnHost;
#[cfg(feature = "csv_storage")]
use mangle_vm::csv_host::CsvHost;

#[cfg(feature = "csv_storage")]
#[test]
fn test_e2e_composite_storage() -> Result<()> {
    // 1. CSV for 'p' (10, 30)
    let mut file_p = NamedTempFile::new()?;
    writeln!(file_p, "10")?;
    writeln!(file_p, "30")?;
    let path_p = file_p.path().to_path_buf();

    // 2. SimpleColumn for 'q' (10, 20)
    let mut file_q = NamedTempFile::new()?;
    writeln!(file_q, "1")?;
    writeln!(file_q, "q 1 2")?;
    writeln!(file_q, "10")?;
    writeln!(file_q, "20")?;
    let path_q = file_q.path().to_path_buf();

    // 3. Setup Sub-Hosts
    let mut csv_host = CsvHost::new();
    csv_host.add_file("p", path_p);

    let mut sc_host = SimpleColumnHost::new();
    sc_host.load_file("q", &path_q)?;

    // 4. Setup Composite Host
    let mut comp_host = CompositeHost::new();
    let h_csv = comp_host.add_host(Box::new(csv_host));
    let h_sc = comp_host.add_host(Box::new(sc_host));

    comp_host.route_relation("p", h_csv);
    comp_host.route_relation("q", h_sc);

    // 5. Compile Program: r(X) :- p(X), q(X).
    let arena = ast::Arena::new_with_global_interner();
    let p = arena.predicate_sym("p", Some(1));
    let q = arena.predicate_sym("q", Some(1));
    let r = arena.predicate_sym("r", Some(1));
    let x = arena.variable("X");

    let clause = ast::Clause {
        head: arena.atom(r, &[x]),
        premises: arena.alloc_slice_copy(&[
            arena.alloc(ast::Term::Atom(arena.atom(p, &[x]))),
            arena.alloc(ast::Term::Atom(arena.atom(q, &[x]))),
        ]),
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

    // 6. Execute via Shared Wrapper
    #[derive(Clone)]
    struct SharedHost<H>(Arc<Mutex<H>>);
    impl<H: Host> Host for SharedHost<H> {
        fn scan_start(&mut self, id: i32) -> i32 {
            self.0.lock().unwrap().scan_start(id)
        }
        fn scan_next(&mut self, id: i32) -> i32 {
            self.0.lock().unwrap().scan_next(id)
        }
        fn get_col(&mut self, p: i32, i: i32) -> i64 {
            self.0.lock().unwrap().get_col(p, i)
        }
        fn insert(&mut self, id: i32, v: i64) {
            self.0.lock().unwrap().insert(id, v)
        }
    }

    let shared_host = SharedHost(Arc::new(Mutex::new(comp_host)));
    let vm = Vm::new()?;
    vm.execute(&wasm, shared_host)?;

    Ok(())
}
