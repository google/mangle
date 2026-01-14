// Copyright 2024 Google LLC
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

use crate::Engine;
use crate::Result;
use crate::ast::{Arena, Atom, BaseTerm, Term};
use anyhow::anyhow;
use fxhash::FxHashMap;

pub struct Naive {}

impl<'e> Engine<'e> for Naive {
    fn eval<'p>(
        &'e self,
        store: &'e impl mangle_factstore::FactStore<'e>,
        program: &'p mangle_analysis::StratifiedProgram<'p>,
    ) -> Result<()> {
        // Initial facts.
        for pred in program.intensional_preds() {
            for rule in program.rules(pred) {
                if rule.premises.is_empty() {
                    store.add(program.arena(), rule.head)?;
                }
            }
        }
        loop {
            let mut fact_added = false;
            for pred in program.intensional_preds() {
                for rule in program.rules(pred) {
                    if rule.premises.is_empty() {
                        continue; // initial facts added previously
                    }
                    let arena = Arena::new_with_global_interner();
                    let subst: FxHashMap<u32, &BaseTerm> = FxHashMap::default();
                    let mut all_ok = true;
                    let subst = std::cell::RefCell::new(subst);
                    for premise in rule.premises.iter() {
                        let ok: bool = match premise {
                            Term::Atom(query) => {
                                // TODO: eval ApplyFn terms.
                                let query = query.apply_subst(&arena, &subst.borrow());
                                let found = std::cell::RefCell::new(false);
                                let _ = store.get(query.sym, query.args, &|atom: &'_ Atom<'_>| {
                                    let mut mismatch = false;
                                    for (i, arg) in query.args.iter().enumerate() {
                                        match arg {
                                            BaseTerm::Variable(v) => {
                                                let own_atom_ref = arena
                                                    .copy_base_term(store.arena(), atom.args[i]);
                                                subst.borrow_mut().insert(v.0, own_atom_ref);
                                            }
                                            c @ BaseTerm::Const(_) => {
                                                if *c == atom.args[i] {
                                                    continue;
                                                }
                                                mismatch = true;
                                                break;
                                            }
                                            _ => {
                                                return Err(anyhow!(format!(
                                                    "Unsupported term: {arg}"
                                                )));
                                            }
                                        }
                                    }
                                    *found.borrow_mut() = !mismatch;
                                    Ok(())
                                });
                                !*found.borrow()
                            }
                            Term::NegAtom(query) => {
                                let query = query.apply_subst(&arena, &subst.borrow());

                                let found = std::cell::RefCell::new(false);
                                let _ = store.get(query.sym, query.args, &|_| {
                                    *found.borrow_mut() = true;
                                    Ok(())
                                });
                                !*found.borrow()
                            }
                            Term::Eq(left, right) => {
                                let left = left.apply_subst(&arena, &subst.borrow());
                                let right = right.apply_subst(&arena, &subst.borrow());
                                left == right
                            }
                            Term::Ineq(left, right) => {
                                let left = left.apply_subst(&arena, &subst.borrow());
                                let right = right.apply_subst(&arena, &subst.borrow());
                                left != right
                            }
                        };
                        if !ok {
                            all_ok = false;
                            break;
                        }
                    }
                    if all_ok {
                        let head = rule.head.apply_subst(&arena, &subst.borrow());
                        fact_added = store.add(program.arena(), head)?;
                    }
                }
            }
            if !fact_added {
                break;
            }
        }
        Ok(())
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::ast;
    use anyhow::Result;
    use mangle_analysis::Program;
    use mangle_factstore::{TableConfig, TableStoreImpl, TableStoreSchema};

    #[test]
    pub fn test_naive() -> Result<()> {
        let arena = Arena::new_with_global_interner();
        let edge = arena.predicate_sym("edge", Some(2));
        let reachable = arena.predicate_sym("reachable", Some(2));
        let mut schema: TableStoreSchema = FxHashMap::default();
        schema.insert(edge, TableConfig::InMemory);
        schema.insert(reachable, TableConfig::InMemory);
        let store = TableStoreImpl::new(&arena, &schema);

        use crate::factstore::FactStore;
        store.add(
            &arena,
            arena.atom(
                edge,
                &[
                    &ast::BaseTerm::Const(ast::Const::Number(10)),
                    &ast::BaseTerm::Const(ast::Const::Number(20)),
                ],
            ),
        )?;
        store.add(
            &arena,
            arena.atom(
                edge,
                &[
                    &ast::BaseTerm::Const(ast::Const::Number(20)),
                    &ast::BaseTerm::Const(ast::Const::Number(30)),
                ],
            ),
        )?;
        store.add(
            &arena,
            arena.atom(
                edge,
                &[
                    &ast::BaseTerm::Const(ast::Const::Number(30)),
                    &ast::BaseTerm::Const(ast::Const::Number(40)),
                ],
            ),
        )?;

        let mut simple = Program::new(&arena);
        // Manually set ext_preds since Program::new doesn't take them anymore?
        // Wait, Program::new initializes ext_preds to empty. The struct definition shows it as public.
        simple.ext_preds = vec![edge];

        let head = arena.alloc(ast::Atom {
            sym: reachable,
            args: arena.alloc_slice_copy(&[arena.variable("X"), arena.variable("Y")]),
        });
        // Add a clause.
        simple.add_clause(
            &arena,
            arena.alloc(ast::Clause {
                head,
                premises: arena.alloc_slice_copy(&[arena.alloc(ast::Term::Atom(
                    arena.atom(edge, &[arena.variable("X"), arena.variable("Y")]),
                ))]),
                transform: &[],
            }),
        );
        simple.add_clause(
            &arena,
            arena.alloc(ast::Clause {
                head: arena.atom(reachable, &[arena.variable("X"), arena.variable("Z")]),
                premises: arena.alloc_slice_copy(&[
                    arena.alloc(ast::Term::Atom(
                        arena.atom(edge, &[arena.variable("X"), arena.variable("Y")]),
                    )),
                    arena.alloc(ast::Term::Atom(
                        arena.atom(reachable, &[arena.variable("Y"), arena.variable("X")]),
                    )),
                ]),
                transform: &[],
            }),
        );

        let stratified_program = simple.stratify().unwrap();

        let engine = Naive {};
        engine.eval(&store, &stratified_program)?;

        use crate::factstore::ReadOnlyFactStore;
        assert!(
            store
                .contains(
                    &arena,
                    arena.atom(
                        edge,
                        &[
                            &ast::BaseTerm::Const(ast::Const::Number(30)),
                            &ast::BaseTerm::Const(ast::Const::Number(40))
                        ],
                    )
                )
                .unwrap()
        );
        Ok(())
    }
}
