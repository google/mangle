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

use crate::{ast, PredicateSet, Program, StratifiedProgram};
use ast::Arena;
use fxhash::{FxHashMap, FxHashSet};

/// An implementation of the `Program` trait.
pub struct SimpleProgram<'p> {
    pub arena: &'p Arena,
    pub ext_preds: Vec<ast::PredicateIndex>,
    pub rules: FxHashMap<ast::PredicateIndex, Vec<&'p ast::Clause<'p>>>,
}

/// An implementation of the `StratifiedProgram` trait.
/// This can be obtained through SimpleProgram::stratify.
pub struct SimpleStratifiedProgram<'p> {
    program: SimpleProgram<'p>,
    strata: Vec<PredicateSet>,
}

impl<'p> Program<'p> for SimpleProgram<'p> {
    fn arena(&'p self) -> &'p Arena {
        self.arena
    }

    fn extensional_preds(&'p self) -> PredicateSet {
        let mut set = FxHashSet::default();
        set.extend(self.ext_preds.iter());
        set
    }

    fn intensional_preds(&'p self) -> PredicateSet {
        let mut set = FxHashSet::default();
        set.extend(self.rules.keys());
        set
    }

    fn rules(&'p self, sym: ast::PredicateIndex) -> impl Iterator<Item = &'p ast::Clause<'p>> {
        self.rules.get(&sym).unwrap().iter().copied()
    }
}

impl<'p> SimpleProgram<'p> {
    pub fn add_clause<'src>(&mut self, src: &'src Arena, clause: &'src ast::Clause) {
        let clause = self.arena.copy_clause(src, clause);
        let sym = clause.head.sym;
        use std::collections::hash_map::Entry;
        match self.rules.entry(sym) {
            Entry::Occupied(mut v) => v.get_mut().push(clause),
            Entry::Vacant(v) => {
                v.insert(vec![clause]);
            }
        }
    }

    /// Produces a StratifiedProgram with given set of layers.
    /// TODO: write analysis that computes the set of layers.
    pub fn stratify(
        self,
        strata: impl IntoIterator<Item = PredicateSet>,
    ) -> SimpleStratifiedProgram<'p> {
        SimpleStratifiedProgram { program: self, strata: strata.into_iter().collect() }
    }
}

impl<'p> Program<'p> for SimpleStratifiedProgram<'p> {
    fn arena(&'p self) -> &'p Arena {
        self.program.arena()
    }

    fn extensional_preds(&'p self) -> PredicateSet {
        self.program.extensional_preds()
    }

    fn intensional_preds(&'p self) -> PredicateSet {
        self.program.intensional_preds()
    }

    fn rules(&'p self, sym: ast::PredicateIndex) -> impl Iterator<Item = &'p ast::Clause<'p>> {
        self.program.rules(sym)
    }
}

impl<'p> StratifiedProgram<'p> for SimpleStratifiedProgram<'p> {
    fn strata(&'p self) -> Vec<PredicateSet> {
        self.strata.to_vec()
    }

    fn pred_to_index(&'p self, sym: ast::PredicateIndex) -> Option<usize> {
        self.strata.iter().position(|x| x.contains(&sym))
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use googletest::matchers::{elements_are, eq};
    use googletest::verify_that;

    #[test]
    fn try_eval() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let foo = arena.predicate_sym("foo", Some(2));
        let bar = arena.predicate_sym("bar", Some(1));
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![foo], rules: FxHashMap::default() };

        // Add a clause.
        let clause = ast::Clause {
            head: arena.atom(bar, &[arena.variable("X")]),
            premises: &[&ast::Term::Atom(
                arena.atom(foo, &[arena.variable("X"), arena.variable("_")]),
            )],
            transform: &[],
        };
        simple.add_clause(&arena, &clause);

        verify_that!(simple.extensional_preds(), elements_are![&foo])?;
        verify_that!(simple.intensional_preds(), elements_are![&bar])?;

        let mut single_layer = FxHashSet::default();
        single_layer.insert(bar);
        let strata = vec![single_layer.clone()];
        let stratified = simple.stratify(strata.into_iter());

        verify_that!(stratified.pred_to_index(bar), eq(Some(0)))?;
        verify_that!(stratified.strata(), elements_are![&single_layer])
    }
}
