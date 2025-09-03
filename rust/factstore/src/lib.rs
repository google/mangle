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

pub trait Receiver<'a> {
    fn next(&self, item: &'a ast::Atom<'a>) -> Result<()>;
}

impl<'a, Closure: Fn(&'a ast::Atom<'a>) -> Result<()>> Receiver<'a> for Closure {
    fn next(&self, item: &'a ast::Atom<'a>) -> Result<()> {
        (*self)(item)
    }
}

/// Lifetime 'a is used for data held by this store.
pub trait ReadOnlyFactStore<'a> {
    fn arena(&'a self) -> &'a Arena;

    fn contains<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom<'src>) -> Result<bool>;

    // Sends atoms that matches query `Atom{ sym: query_sym, args: query_args}`.
    // pub sym: PredicateIndex,
    fn get<'query, R: Receiver<'a>>(
        &'a self,
        query_sym: ast::PredicateIndex,
        query_args: &'query [&'query ast::BaseTerm<'query>],
        cb: &R,
    ) -> Result<()>;

    // Invokes cb for every predicate available in this store.
    // It would be nice to use `impl Iterator` here.
    fn predicates(&'a self) -> Vec<ast::PredicateIndex>;

    // Returns approximae number of facts.
    fn estimate_fact_count(&self) -> u32;
}

/// A fact store that can be mutated.
/// Implementations must make use of interior mutability.
pub trait FactStore<'a>: ReadOnlyFactStore<'a> {
    /// Returns true if fact did not exist before.
    /// The fact is copied.
    fn add<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom<'src>) -> Result<bool>;

    /// Adds all facts from given store.
    fn merge<'src, S>(&'a self, src: &'src Arena, store: &'src S)
    where
        S: ReadOnlyFactStore<'src>;
}

/// Invokes cb for every fact in the store.
pub fn get_all_facts<'a, S, R: Receiver<'a>>(store: &'a S, cb: &R) -> Result<()>
where
    S: ReadOnlyFactStore<'a> + 'a,
{
    let arena = Arena::new_with_global_interner();
    let preds = store.predicates();

    for pred in preds {
        arena.copy_predicate_sym(store.arena(), pred);
        store.get(pred, &arena.new_query(pred).args, cb)?;
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
        fn arena(&'a self) -> &'a Arena {
            self.arena
        }

        fn contains<'src>(&'a self, src: &'src Arena, fact: &'src ast::Atom<'src>) -> Result<bool> {
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
                    Some(sym) => {
                        Ok(self.facts.borrow().iter().any(|x| x.sym == sym && x.args == fact.args))
                    }
                },
            }
        }

        fn get<'query, R: Receiver<'a>>(
            &'a self,
            query_sym: ast::PredicateIndex,
            query_args: &'query [&'query ast::BaseTerm<'query>],
            cb: &R,
        ) -> Result<()> {
            for fact in self.facts.borrow().iter() {
                if fact.sym == query_sym {
                    if fact.matches(query_args) {
                        cb.next(&fact)?;
                    }
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
            // If the fact is from a different arena, it must be copied.
            if self.contains(src, fact)? {
                return Ok(false);
            }
            self.facts.borrow_mut().push(self.arena.copy_atom(src, fact));
            Ok(true)
        }

        fn merge<'src, S>(&'a self, src: &'src Arena, store: &'src S)
        where
            S: ReadOnlyFactStore<'src>,
        {
            let _ = get_all_facts(store, &move |fact| {
                let atom = self.arena.copy_atom(src, fact);
                self.facts.borrow_mut().push(atom);
                Ok(())
            });
        }
    }

    #[test]
    fn test_get_factsa() {
        let arena = Arena::new_with_global_interner();
        let simple = TestStore { arena: &arena, facts: RefCell::new(vec![]) };
        let atom = test_atom(&arena);

        assert!(!simple.contains(&arena, &atom).unwrap());
        assert!(simple.add(&arena, &atom).unwrap());
        assert!(simple.contains(&arena, &atom).unwrap());
    }

    #[test]
    fn test_multi_arena() {
        let arena1 = Arena::new_with_global_interner();
        let arena2 = Arena::new_with_global_interner();
        let store = TestStore { arena: &arena1, facts: RefCell::new(vec![]) };

        let atom_in_arena2 = test_atom(&arena2);

        // Register the predicate symbol in the store's arena as well.
        let index1 = arena1.predicate_sym("foo", Some(1));

        println!("predicate_sym: {:?}", index1);

        // Add atom from arena2 to store with arena1
        assert!(store.add(&arena2, &atom_in_arena2).unwrap());

        // Check if the atom is now in the store
        assert!(store.contains(&arena2, &atom_in_arena2).unwrap());

        // Verify that the stored atom is in arena1
        let found = RefCell::new(false);
        let _ = get_all_facts(&store, &|fact: &ast::Atom| {
            // This is a bit of a hack to check if the atom is in arena1.
            // We can't directly compare arenas, but we can check if the symbols
            // are the same.
            assert_eq!(fact.sym, arena1.predicate_sym("foo", Some(1)));
            *found.borrow_mut() = true;
            Ok(())
        });
        assert!(*found.borrow());
    }
}
