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
use bumpalo::Bump;
use std::collections::HashMap;

/// An implementation of the `Program` trait.
pub struct SimpleProgram<'p> {
    pub bump: &'p Bump,
    pub ext_preds: Vec<ast::PredicateSym<'p>>,
    pub rules: HashMap<ast::PredicateSym<'p>, Vec<&'p ast::Clause<'p>>>,
}

/// An implementation of the `StratifiedProgram` trait.
/// This can be obtained through SimpleProgram::stratify.
pub struct SimpleStratifiedProgram<'p> {
    program: SimpleProgram<'p>,
    strata: &'p [&'p PredicateSet<'p>],
}

impl<'p> Program<'p> for SimpleProgram<'p> {
    fn extensional_preds(&'p self) -> impl Iterator<Item = &'p ast::PredicateSym<'p>> {
        self.ext_preds.iter()
    }

    fn intensional_preds(&'p self) -> impl Iterator<Item = &'p ast::PredicateSym<'p>> {
        self.rules.keys()
    }

    fn rules(
        &'p self,
        sym: &'p ast::PredicateSym<'p>,
    ) -> impl Iterator<Item = &'p ast::Clause<'p>> {
        self.rules.get(sym).unwrap().iter().copied()
    }
}

impl<'p> SimpleProgram<'p> {
    pub fn add_clause(&mut self, clause: &ast::Clause) {
        let clause = ast::copy_clause(self.bump, clause);
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
        strata: impl Iterator<Item = &'p PredicateSet<'p>>,
    ) -> SimpleStratifiedProgram<'p> {
        let mut layers = vec![];
        for layer in strata {
            layers.push(*self.bump.alloc(layer))
        }
        let strata = &*self.bump.alloc_slice_copy(&layers);
        SimpleStratifiedProgram {
            program: self,
            strata,
        }
    }
}

impl<'p> Program<'p> for SimpleStratifiedProgram<'p> {
    fn extensional_preds(&'p self) -> impl Iterator<Item = &'p ast::PredicateSym<'p>> {
        self.program.extensional_preds()
    }

    fn intensional_preds(&'p self) -> impl Iterator<Item = &'p ast::PredicateSym<'p>> {
        self.program.intensional_preds()
    }

    fn rules(
        &'p self,
        sym: &'p ast::PredicateSym<'p>,
    ) -> impl Iterator<Item = &'p ast::Clause<'p>> {
        self.program.rules(sym)
    }
}

impl<'p> StratifiedProgram<'p> for SimpleStratifiedProgram<'p> {
    fn strata(&'p self) -> impl Iterator<Item = &'p PredicateSet<'p>> {
        self.strata.iter().copied()
    }

    fn pred_to_index(&'p self, sym: &ast::PredicateSym) -> Option<usize> {
        self.strata.iter().position(|x| x.contains(sym))
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use std::collections::HashSet;

    static FOO: ast::PredicateSym = ast::PredicateSym {
        name: "foo",
        arity: Some(2),
    };
    static BAR: ast::PredicateSym = ast::PredicateSym {
        name: "bar",
        arity: Some(1),
    };

    #[test]
    fn try_eval() {
        let bump = Bump::new();
        let mut simple = SimpleProgram {
            bump: &bump,
            ext_preds: vec![FOO],
            rules: HashMap::new(),
        };

        // Add a clause.
        let atom = ast::Atom {
            sym: FOO,
            args: &[&ast::BaseTerm::Variable("X"), &ast::BaseTerm::Variable("_")],
        };
        let clause = &ast::Clause {
            head: &ast::Atom {
                sym: BAR,
                args: &[&ast::BaseTerm::Variable("X")],
            },
            premises: &[&ast::Term::Atom(&atom)],
            transform: &[],
        };
        simple.add_clause(clause);

        assert_eq!(simple.extensional_preds().collect::<Vec<_>>(), vec![&FOO]);
        assert_eq!(simple.intensional_preds().collect::<Vec<_>>(), vec![&BAR]);

        let mut single_layer = HashSet::new();
        single_layer.insert(&BAR);
        let strata = vec![single_layer.clone()];
        let stratified = simple.stratify(strata.iter());

        assert_eq!(stratified.pred_to_index(&BAR), Some(0));
        assert_eq!(stratified.strata().collect::<Vec<_>>(), vec![&single_layer])
    }
}
