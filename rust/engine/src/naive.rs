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

use fxhash::FxHashMap;

use crate::ast::{Arena, BaseTerm, Term};
use crate::Engine;
use crate::Result;
use anyhow::anyhow;

pub struct Naive {}

impl<'e> Engine<'e> for Naive {
    fn eval<'p>(
        &'e self,
        store: &'e impl mangle_factstore::FactStore<'e>,
        program: &'p impl mangle_analysis::StratifiedProgram<'p>,
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
                    let arena = Arena::new_global();
                    let mut subst: FxHashMap<u32, &crate::ast::BaseTerm> = FxHashMap::default();
                    let mut all_ok = true;
                    for premise in rule.premises.iter() {
                        let ok = match premise {
                            Term::Atom(query) => {
                                // TODO: eval ApplyFn terms.
                                let query = query.apply_subst(&arena, &subst);
                                let mut found = false;
                                let _ = store.get(query, |atom| {
                                    let mut mismatch = false;
                                    for (i, arg) in query.args.iter().enumerate() {
                                        match arg {
                                            BaseTerm::Variable(v) => {
                                                subst.insert(v.0, atom.args[i]);
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
                                                )))
                                            }
                                        }
                                    }
                                    found = !mismatch;
                                    Ok(())
                                });
                                found
                            }
                            Term::NegAtom(query) => {
                                let query = query.apply_subst(&arena, &subst);

                                let mut found = false;
                                let _ = store.get(query, |_| {
                                    found = true;
                                    Ok(())
                                });
                                !found
                            }
                            Term::Eq(left, right) => {
                                let left = left.apply_subst(&arena, &subst);
                                let right = right.apply_subst(&arena, &subst);
                                left == right
                            }
                            Term::Ineq(left, right) => {
                                let left = left.apply_subst(&arena, &subst);
                                let right = right.apply_subst(&arena, &subst);
                                left != right
                            }
                        };
                        if !ok {
                            all_ok = false;
                            break;
                        }
                    }
                    if all_ok {
                        let head = rule.head.apply_subst(&arena, &subst);
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
    use std::collections::HashSet;

    use super::*;
    use crate::ast;
    use anyhow::Result;
    use mangle_analysis::SimpleProgram;
    use mangle_factstore::{TableConfig, TableStoreImpl, TableStoreSchema};

    #[test]
    pub fn test_naive() -> Result<()> {
        let arena = Arena::new_global();
        let edge = arena.predicate_sym("edge", Some(2));
        let reachable = arena.predicate_sym("reachable", Some(2));
        let mut schema: TableStoreSchema = FxHashMap::default();
        schema.insert(edge, TableConfig::InMemory);
        schema.insert(reachable, TableConfig::InMemory);
        let store = TableStoreImpl::new(&schema);

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

        let mut simple = SimpleProgram {
            arena: &arena,
            ext_preds: vec![edge],
            rules: FxHashMap::default(),
        };

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

        let mut single_layer = HashSet::new();
        single_layer.insert(reachable);
        let strata = vec![single_layer];
        let stratified_program = simple.stratify(strata);

        let engine = Naive {};
        engine.eval(&store, &stratified_program)?;

        use crate::factstore::ReadOnlyFactStore;
        assert!(store
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
            .unwrap());
        Ok(())
    }
}
