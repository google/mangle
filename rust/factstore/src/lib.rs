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
use ast::Arena;
use mangle_ast as ast;

mod tablestore;
pub use tablestore::{TableConfig, TableStoreImpl, TableStoreSchema};

/// Lifetime 'a is used for data held by this store.
pub trait ReadOnlyFactStore<'a> {
    fn contains<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom) -> Result<bool>;

    // Invokes cb for fact that matches query.
    fn get(
        &'a self,
        query: &ast::Atom,
        cb: impl FnMut(&'a ast::Atom<'a>) -> Result<()>,
    ) -> Result<()>;

    //fn get<F>(&'a self, query: &ast::Atom, cb: F) -> Result<()>
    //where
    //    F: FnMut(&'a ast::Atom<'a>) -> Result<()>;

    // Iterator over every predicate available in this store.
    fn predicates(&'a self) -> Vec<ast::PredicateIndex>;

    // Returns approximae number of facts.
    fn estimate_fact_count(&self) -> u32;
}

pub trait FactStore<'a>: ReadOnlyFactStore<'a> {
    /// Returns true if fact did not exist before.
    /// The fact is copied.
    fn add<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom<'src>) -> Result<bool>;

    /// Adds all facts from given store.
    fn merge<'src, S>(&'a self, src: &'src Arena, store: &'src S)
    where
        S: ReadOnlyFactStore<'src>;
}

pub fn get_all_facts<'a, S>(
    store: &'a S,
    mut cb: impl FnMut(&'a ast::Atom<'a>) -> Result<()>,
) -> Result<()>
where
    S: ReadOnlyFactStore<'a> + 'a,
{
    let arena = Arena::new_global();
    let preds = store.predicates();
    for pred in preds {
        store.get(&arena.new_query(pred), &mut cb)?;
    }
    Ok(())
}

#[cfg(test)]
mod test {
    use std::{cell::RefCell, collections::HashSet};

    use super::*;

    fn test_atom<'arena>(arena: &'arena Arena) -> ast::Atom<'arena> {
        ast::Atom {
            sym: arena.predicate_sym("foo", Some(1)),
            args: &[&ast::BaseTerm::Const(ast::Const::String("bar"))],
        }
    }

    struct TestStore<'a> {
        arena: &'a Arena,
        facts: RefCell<Vec<&'a ast::Atom<'a>>>,
    }

    impl<'a> ReadOnlyFactStore<'a> for TestStore<'a> {
        fn contains<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom) -> Result<bool> {
            if self.arena as *const _ as usize == src as *const _ as usize {
                return Ok(self.facts.borrow().iter().any(|x| *x == fact));
            }
            let src_predicate_name = self.arena.predicate_name(fact.sym);
            if src_predicate_name.is_none() {
                return Ok(false);
            }
            match self.arena.lookup_opt(src_predicate_name.unwrap()) {
                None => Ok(false),
                Some(n) => match self.arena.lookup_predicate_sym(n) {
                    None => Ok(false),
                    Some(sym) => Ok(self
                        .facts
                        .borrow()
                        .iter()
                        .any(|x| x.sym == sym && x.args == fact.args)),
                },
            }
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

        fn predicates(&'a self) -> Vec<ast::PredicateIndex> {
            let mut seen = HashSet::new();
            for fact in self.facts.borrow().iter() {
                let pred = &fact.sym;
                seen.insert(*pred);
            }
            seen.iter().map(|x| *x).collect()
        }

        fn estimate_fact_count(&self) -> u32 {
            self.facts.borrow().len().try_into().unwrap()
        }
    }

    impl<'a> FactStore<'a> for TestStore<'a> {
        fn add<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom) -> Result<bool> {
            // TODO: If it is from a separate arena, need to copy.
            if self.contains(src, fact)? {
                return Ok(false);
            }
            self.facts
                .borrow_mut()
                .push(self.arena.copy_atom(src, fact));
            Ok(true)
        }

        fn merge<'src, S>(&'a self, src: &'src Arena, store: &'src S)
        where
            S: ReadOnlyFactStore<'src>,
        {
            let _ = get_all_facts(store, move |fact| {
                let atom = self.arena.copy_atom(src, fact);
                self.facts.borrow_mut().push(atom);
                Ok(())
            });
        }
    }

    #[test]
    fn test_get_factsa() {
        let arena = Arena::new_global();
        let simple = TestStore {
            arena: &arena,
            facts: RefCell::new(vec![]),
        };
        let atom = test_atom(&arena);
        assert!(!simple.contains(&arena, &atom).unwrap());
        assert!(simple.add(&arena, &atom).unwrap());
        assert!(simple.contains(&arena, &atom).unwrap());
    }
}
