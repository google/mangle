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

use std::collections::HashMap;

use crate::Engine;
use crate::Result;
use bumpalo::Bump;

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
                    store.add(rule.head)?;
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
                    let bump = Bump::new();
                    let mut subst: HashMap<&str, &crate::ast::BaseTerm> = HashMap::new();
                    for premise in rule.premises.iter() {
                        let ok = match premise {
                            mangle_ast::Term::Atom(query) => {
                                // TODO: eval ApplyFn terms.
                                let query = query.apply_subst(&bump, &subst);

                                // if all arguments are constants:
                                let mut found = false;
                                let _ = store.get(query, |atom| {
                                    for (i, arg) in query.args.iter().enumerate() {
                                        if let crate::ast::BaseTerm::Variable(v) = arg {
                                            subst.insert(v, atom.args[i]);
                                        }
                                    }
                                    found = true;
                                    Ok(())
                                });
                                found
                            }
                            mangle_ast::Term::NegAtom(query) => {
                                let query = query.apply_subst(&bump, &subst);

                                let mut found = false;
                                let _ = store.get(query, |_| {
                                    found = true;
                                    Ok(())
                                });
                                !found
                            }
                            mangle_ast::Term::Eq(left, right) => {
                                let left = left.apply_subst(&bump, &subst);
                                let right = right.apply_subst(&bump, &subst);
                                left == right
                            }
                            mangle_ast::Term::Ineq(left, right) => {
                                let left = left.apply_subst(&bump, &subst);
                                let right = right.apply_subst(&bump, &subst);
                                left != right
                            }
                        };
                        if ok {
                            fact_added = store.add(rule.head)?;
                        }
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
        let edge = ast::PredicateSym {
            name: "edge",
            arity: Some(2),
        };
        let reachable = ast::PredicateSym {
            name: "reachable",
            arity: Some(2),
        };
        let schema: TableStoreSchema = HashMap::from([
            (&edge, TableConfig::InMemory),
            (&reachable, TableConfig::InMemory),
        ]);
        let store = TableStoreImpl::new(&schema);

        use crate::factstore::FactStore;
        store.add(&ast::Atom {
            sym: edge,
            args: &[
                &ast::BaseTerm::Const(ast::Const::Number(1)),
                &ast::BaseTerm::Const(ast::Const::Number(2)),
            ],
        })?;
        store.add(&ast::Atom {
            sym: edge,
            args: &[
                &ast::BaseTerm::Const(ast::Const::Number(2)),
                &ast::BaseTerm::Const(ast::Const::Number(3)),
            ],
        })?;
        store.add(&ast::Atom {
            sym: edge,
            args: &[
                &ast::BaseTerm::Const(ast::Const::Number(3)),
                &ast::BaseTerm::Const(ast::Const::Number(4)),
            ],
        })?;

        let bump = Bump::new();
        let mut simple = SimpleProgram {
            bump: &bump,
            ext_preds: vec![edge],
            rules: HashMap::new(),
        };

        // Add a clause.
        simple.add_clause(&ast::Clause {
            head: &ast::Atom {
                sym: reachable,
                args: &[&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("Y")],
            },
            premises: &[&ast::Term::Atom(&ast::Atom {
                sym: edge,
                args: &[&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("Y")],
            })],
            transform: &[],
        });
        simple.add_clause(&ast::Clause {
            head: &ast::Atom {
                sym: reachable,
                args: &[&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("Z")],
            },
            premises: &[
                &ast::Term::Atom(&ast::Atom {
                    sym: edge,
                    args: &[&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("Y")],
                }),
                &ast::Term::Atom(&ast::Atom {
                    sym: reachable,
                    args: &[&ast::BaseTerm::Variable("Y"), &ast::BaseTerm::Variable("X")],
                }),
            ],
            transform: &[],
        });

        let mut single_layer = HashSet::new();
        single_layer.insert(&reachable);
        let strata = vec![single_layer.clone()];
        let stratified_program = simple.stratify(strata.iter());

        let engine = Naive {};
        engine.eval(&store, &stratified_program)?;

        use crate::factstore::ReadOnlyFactStore;
        assert!(store
            .contains(&ast::Atom {
                sym: edge,
                args: &[
                    &ast::BaseTerm::Const(ast::Const::Number(3)),
                    &ast::BaseTerm::Const(ast::Const::Number(4))
                ],
            })
            .unwrap());
        Ok(())
    }
}
