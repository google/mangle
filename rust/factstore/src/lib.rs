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

use anyhow::{anyhow, Result};
use bumpalo::Bump;
use mangle_ast as ast;

mod tablestore;
pub use tablestore::{TableConfig, TableStoreImpl, TableStoreSchema};

/// Lifetime 'a is used for data held by this store.
pub trait ReadOnlyFactStore<'a> {
    fn contains(&'a self, fact: &ast::Atom) -> Result<bool>;

    // Invokes cb for fact that matches query.
    fn get(
        &'a self,
        query: &ast::Atom,
        cb: impl FnMut(&'a ast::Atom<'a>) -> Result<()>,
    ) -> Result<()>;

    //fn get<F>(&'a self, query: &ast::Atom, cb: F) -> Result<()>
    //where
    //    F: FnMut(&'a ast::Atom<'a>) -> Result<()>;

    // Invokes cb for every predicate available in this store.
    // It would be nice to use `impl Iterator` here.
    fn list_predicates(&'a self, cb: impl FnMut(&'a ast::PredicateSym));

    // Returns approximae number of facts.
    fn estimate_fact_count(&self) -> u32;
}

pub trait FactStore<'a>: ReadOnlyFactStore<'a> {
    /// Returns true if fact did not exist before.
    /// The fact is copied.
    fn add(&'a self, fact: &ast::Atom) -> Result<bool>;

    /// Adds all facts from given store.
    fn merge<'other, S>(&'a self, store: &'other S)
    where
        S: ReadOnlyFactStore<'other>;
}

/// Constructs a query (allocated in bump).
/// `pred.arity` must be present.
/// TODO: move to ast
/// TODO: make an allocator interface that has slice with all _ arguments.
fn new_query<'b>(bump: &'b Bump, pred: &'b ast::PredicateSym) -> &'b ast::Atom<'b> {
    let var = &*bump.alloc(ast::BaseTerm::Variable("_"));
    let mut args = vec![];
    for _ in 0..pred.arity.unwrap() {
        args.push(var);
    }
    let args = &*bump.alloc_slice_copy(&args);

    bump.alloc(ast::Atom { sym: *pred, args })
}

pub fn get_all_facts<'a, S>(
    store: &'a S,
    mut cb: impl FnMut(&'a ast::Atom<'a>) -> Result<()>,
) -> Result<()>
where
    S: ReadOnlyFactStore<'a> + 'a,
{
    let bump = Bump::new();
    let mut preds = vec![];
    store.list_predicates(|pred: &ast::PredicateSym| {
        preds.push(pred);
    });
    for pred in preds {
        store.get(new_query(&bump, pred), &mut cb)?;
    }
    Ok(())
}

#[cfg(test)]
mod test {
    use std::{cell::RefCell, collections::HashSet};

    use super::*;

    static TEST_ATOM: ast::Atom = ast::Atom {
        sym: ast::PredicateSym {
            name: "foo",
            arity: Some(1),
        },
        args: &[&ast::BaseTerm::Const(ast::Const::String("bar"))],
    };

    struct TestStore<'a> {
        bump: &'a Bump,
        facts: RefCell<Vec<&'a ast::Atom<'a>>>,
    }

    impl<'a> ReadOnlyFactStore<'a> for TestStore<'a> {
        fn contains<'store>(&'store self, fact: &ast::Atom) -> Result<bool> {
            Ok(self.facts.borrow().iter().any(|x| *x == fact))
        }

        fn get(
            &'a self,
            query: &ast::Atom,
            mut cb: impl FnMut(&'a ast::Atom<'a>) -> Result<()>,
        ) -> Result<()> {
            for fact in self.facts.borrow().iter() {
                // TODO matches
                if fact.sym == query.sym {
                    cb(fact)?;
                }
            }
            Ok(())
        }

        fn list_predicates(&'a self, mut cb: impl FnMut(&'a ast::PredicateSym)) {
            let mut seen = HashSet::new();
            for fact in self.facts.borrow().iter() {
                let pred = &fact.sym;
                if !seen.contains(pred) {
                    seen.insert(pred);
                    cb(pred)
                }
            }
        }

        fn estimate_fact_count(&self) -> u32 {
            self.facts.borrow().len().try_into().unwrap()
        }
    }

    impl<'a> FactStore<'a> for TestStore<'a> {
        fn add(&'a self, fact: &ast::Atom) -> Result<bool> {
            if self.contains(fact)? {
                return Ok(false);
            }
            self.facts
                .borrow_mut()
                .push(ast::copy_atom(&self.bump, fact));
            Ok(true)
        }

        fn merge<'other, S>(&'a self, store: &'other S)
        where
            S: ReadOnlyFactStore<'other>,
        {
            let _ = get_all_facts(store, move |fact| {
                let atom = ast::copy_atom(&self.bump, fact);
                self.facts.borrow_mut().push(atom);
                Ok(())
            });
        }
    }

    #[test]
    fn test_get_factsa() {
        let bump = Bump::new();
        let simple = TestStore {
            bump: &bump,
            facts: RefCell::new(vec![]),
        };
        assert!(!simple.contains(&TEST_ATOM).unwrap());
        assert!(simple.add(&TEST_ATOM).unwrap());
    }
}
