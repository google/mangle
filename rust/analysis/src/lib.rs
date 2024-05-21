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

use mangle_ast as ast;

mod simple_program;

pub use simple_program::{SimpleProgram, SimpleStratifiedProgram};

type PredicateSet<'a> = std::collections::HashSet<&'a ast::PredicateSym<'a>>;

/// Represents a program.
/// `extensional_preds` and `intensional_preds` are disjoint.
pub trait Program<'p> {
    /// Returns predicates for extensional DB.
    /// May return an empty iterator.
    fn extensional_preds(&'p self) -> impl Iterator<Item = &'p ast::PredicateSym<'p>>;

    /// Returns predicates for intensional DB.
    /// May return an empty iterator.
    fn intensional_preds(&'p self) -> impl Iterator<Item = &'p ast::PredicateSym<'p>>;

    /// Maps predicates of intensional DB to rules.
    /// May return an empty iterator.
    fn rules(&'p self, sym: &'p ast::PredicateSym<'p>)
        -> impl Iterator<Item = &'p ast::Clause<'p>>;
}

// A stratified program is a program that can be separated in
// dependency layers (strata) such that if index(p) < index(q)
// then p does not depend on q. 
pub trait StratifiedProgram<'p> : Program<'p> {
    /// Returns an iterator of strata, in dependency order.
    /// TODO: consider Iterator<Iterator<PredicateSet>>.
    fn strata(&'p self) -> impl Iterator<Item = &'p PredicateSet<'p>>;

    /// Returns the stratum (index into strata list).
    fn pred_to_index(&'p self, sym: &ast::PredicateSym) -> Option<usize>;
}

