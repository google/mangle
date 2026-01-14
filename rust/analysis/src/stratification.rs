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

use crate::PredicateSet;
use fxhash::{FxHashMap, FxHashSet};
use mangle_ast as ast;
use mangle_ast::Arena;
use std::fmt;

/// Represents a Mangle program consisting of logic rules and declarations.
///
/// `Program` wraps the AST and provides the necessary methods to identify
/// predicates and their dependencies. It is the primary input for the stratification algorithm.
///
/// It distinguishes between *extensional* predicates (stored facts) and *intensional*
/// predicates (derived rules).
#[derive(Clone)]
pub struct Program<'p> {
    pub arena: &'p Arena,
    pub ext_preds: Vec<ast::PredicateIndex>,
    pub rules: FxHashMap<ast::PredicateIndex, Vec<&'p ast::Clause<'p>>>,
}

impl<'p> fmt::Debug for Program<'p> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("Program")
            .field("ext_preds", &self.ext_preds)
            .field("rules", &self.rules)
            .finish()
    }
}

/// A program that has been successfully stratified.
///
/// Contains the original program and the computed strata.
/// This structure is used to guide the execution order of the IR.
///
/// Stratification ensures that if a predicate `p` depends negatively on `q`,
/// then `q` is evaluated in an earlier stratum than `p`. This allows for
/// the correct evaluation of negation and aggregation (semi-naive evaluation).
#[derive(Clone)]
pub struct StratifiedProgram<'p> {
    program: Program<'p>,
    strata: Vec<PredicateSet>,
}

impl<'p> fmt::Debug for StratifiedProgram<'p> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("StratifiedProgram")
            .field("program", &self.program)
            .field("strata", &self.strata)
            .finish()
    }
}

type EdgeMap = FxHashMap<ast::PredicateIndex, bool>;
type DepGraph = FxHashMap<ast::PredicateIndex, EdgeMap>;
type Nodeset = FxHashSet<ast::PredicateIndex>;

impl<'p> Program<'p> {
    pub fn new(arena: &'p Arena) -> Self {
        Self {
            arena,
            ext_preds: Vec::new(),
            rules: FxHashMap::default(),
        }
    }

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

    /// Returns the AST Arena containing the program data.
    pub fn arena(&'p self) -> &'p ast::Arena {
        self.arena
    }

    /// Returns predicates for extensional DB (stored facts).
    pub fn extensional_preds(&'p self) -> PredicateSet {
        let mut set = FxHashSet::default();
        set.extend(&self.ext_preds);
        set
    }

    /// Returns predicates for intensional DB (derived rules).
    pub fn intensional_preds(&'p self) -> PredicateSet {
        let mut set = FxHashSet::default();
        set.extend(self.rules.keys());
        set
    }

    /// Maps predicates of intensional DB to their defining rules.
    pub fn rules(&'p self, sym: ast::PredicateIndex) -> impl Iterator<Item = &'p ast::Clause<'p>> {
        self.rules.get(&sym).unwrap().iter().copied()
    }

    /// Analyzes the program's dependency graph and attempts to stratify it.
    ///
    /// Stratification partitions the predicates into ordered layers (strata).
    /// This is essential for evaluating programs with negation or aggregation,
    /// ensuring that dependencies are fully evaluated before they are used.
    ///
    /// Returns a `StratifiedProgram` on success, or an error if the program
    /// contains unstratifiable cycles (e.g., negation cycles).
    pub fn stratify(self) -> Result<StratifiedProgram<'p>, String> {
        let dep = make_dep_graph(&self);
        let mut strata = dep.sccs();

        let mut pred_to_stratum: FxHashMap<ast::PredicateIndex, usize> = FxHashMap::default();

        for (i, c) in strata.iter().enumerate() {
            for sym in c {
                pred_to_stratum.insert(*sym, i);
            }
            for sym in c {
                if let Some(edges) = dep.get(sym) {
                    for (dest, negated) in edges {
                        if !*negated {
                            continue;
                        }
                        let dest_stratum = pred_to_stratum.get(dest);
                        if let Some(dest_stratum) = dest_stratum
                            && *dest_stratum == i
                        {
                            return Err("program cannot be stratified".to_string());
                        }
                    }
                }
            }
        }
        dep.sort_result(&mut strata, pred_to_stratum);
        let stratified = StratifiedProgram {
            program: self,
            strata: strata.into_iter().collect(),
        };
        Ok(stratified)
    }
}

impl<'p> StratifiedProgram<'p> {
    /// Returns the AST Arena containing the program data.
    pub fn arena(&'p self) -> &'p ast::Arena {
        self.program.arena()
    }

    /// Returns predicates for extensional DB (stored facts).
    pub fn extensional_preds(&'p self) -> PredicateSet {
        self.program.extensional_preds()
    }

    /// Returns predicates for intensional DB (derived rules).
    pub fn intensional_preds(&'p self) -> PredicateSet {
        self.program.intensional_preds()
    }

    /// Maps predicates of intensional DB to their defining rules.
    pub fn rules(&'p self, sym: ast::PredicateIndex) -> impl Iterator<Item = &'p ast::Clause<'p>> {
        self.program.rules(sym)
    }

    /// Returns an iterator of strata, in dependency order.
    /// Each stratum is a set of mutually recursive predicates that can be evaluated together.
    pub fn strata(&'p self) -> Vec<PredicateSet> {
        self.strata.to_vec()
    }

    /// Returns the stratum index for a given predicate symbol.
    /// Returns `None` if the predicate is not part of the stratified program (e.g. it's EDB).
    pub fn pred_to_index(&'p self, sym: ast::PredicateIndex) -> Option<usize> {
        self.strata.iter().position(|x| x.contains(&sym))
    }
}

fn make_dep_graph<'p>(program: &Program<'p>) -> DepGraph {
    let mut dep: DepGraph = FxHashMap::default();

    for (s, rule) in program.rules.iter() {
        dep.init_node(*s);
        for clause in rule.iter() {
            for premise in clause.premises.iter() {
                match premise {
                    ast::Term::Atom(atom_pred) => {
                        if !program.extensional_preds().contains(&atom_pred.sym) {
                            if clause.transform.is_empty() || clause.transform[0].var.is_some() {
                                dep.add_edge(*s, atom_pred.sym, false);
                            } else {
                                dep.add_edge(*s, atom_pred.sym, true);
                            }
                        }
                    }
                    ast::Term::NegAtom(atom_pred) => {
                        if !program.extensional_preds().contains(&atom_pred.sym) {
                            dep.add_edge(*s, atom_pred.sym, true);
                        }
                    }
                    _ => {}
                }
            }
        }
    }
    dep
}

fn apply_permutation_cycle_rotate<T: Default>(arr: &mut [T], permutation: &[usize]) {
    let n = arr.len();
    if n == 0 {
        return;
    }
    let mut visited = vec![false; n];
    for i in 0..n {
        if !visited[i] {
            let mut current_idx = i;
            if permutation[current_idx] == i {
                visited[i] = true;
                continue;
            }
            let mut current_val = std::mem::take(&mut arr[i]);
            loop {
                let target_idx = permutation[current_idx];
                visited[current_idx] = true;
                let next_val = std::mem::replace(&mut arr[target_idx], current_val);
                current_val = next_val;
                current_idx = target_idx;
                if current_idx == i {
                    break;
                }
            }
        }
    }
}

trait DepGraphExt {
    fn init_node(&mut self, src: ast::PredicateIndex);
    fn add_edge(&mut self, src: ast::PredicateIndex, dest: ast::PredicateIndex, negated: bool);
    fn transpose(&self) -> DepGraph;
    fn sccs(&self) -> Vec<Nodeset>;
    fn sort_result(
        &self,
        strata: &mut Vec<Nodeset>,
        pred_to_stratum_map: FxHashMap<ast::PredicateIndex, usize>,
    ) -> FxHashMap<ast::PredicateIndex, usize>;
}

impl DepGraphExt for DepGraph {
    fn init_node(&mut self, src: ast::PredicateIndex) {
        self.entry(src).or_default();
    }

    fn add_edge(&mut self, src: ast::PredicateIndex, dest: ast::PredicateIndex, negated: bool) {
        let edges = self.entry(src).or_default();
        if negated {
            edges.insert(dest, negated);
            return;
        }
        if edges.get(&dest).is_none() || !edges[&dest] {
            edges.insert(dest, false);
        }
    }

    fn transpose(&self) -> DepGraph {
        let mut rev: DepGraph = FxHashMap::default();
        for (src, edges) in self.iter() {
            for (dest, negated) in edges.iter() {
                rev.init_node(*dest);
                rev.add_edge(*dest, *src, *negated);
            }
        }
        rev
    }

    fn sccs(&self) -> Vec<Nodeset> {
        let mut s: Vec<ast::PredicateIndex> = Vec::new();
        let mut seen: Nodeset = FxHashSet::default();

        fn visit(
            node: ast::PredicateIndex,
            graph: &DepGraph,
            s: &mut Vec<ast::PredicateIndex>,
            seen: &mut Nodeset,
        ) {
            if !seen.contains(&node) {
                seen.insert(node);
                if let Some(edges) = graph.get(&node) {
                    for &neighbor in edges.keys() {
                        visit(neighbor, graph, s, seen);
                    }
                }
                s.push(node);
            }
        }

        for (node, _) in self.iter() {
            visit(*node, self, &mut s, &mut seen);
        }

        let rev = self.transpose();
        let mut seen: Nodeset = FxHashSet::default();
        fn rvisit(
            node: ast::PredicateIndex,
            rev: &DepGraph,
            scc: &mut Nodeset,
            seen: &mut Nodeset,
        ) {
            if !seen.contains(&node) {
                seen.insert(node);
                scc.insert(node);
                if let Some(edges) = rev.get(&node) {
                    for &e in edges.keys() {
                        rvisit(e, rev, scc, seen);
                    }
                }
            }
        }
        let mut sccs: Vec<Nodeset> = Vec::new();
        while let Some(top) = s.pop() {
            if !seen.contains(&top) {
                let mut scc: Nodeset = FxHashSet::default();
                rvisit(top, &rev, &mut scc, &mut seen);
                if !scc.is_empty() {
                    sccs.push(scc);
                }
            }
        }
        sccs
    }

    fn sort_result(
        &self,
        strata: &mut Vec<Nodeset>,
        pred_to_stratum_map: FxHashMap<ast::PredicateIndex, usize>,
    ) -> FxHashMap<ast::PredicateIndex, usize> {
        let mut sorted_indices: Vec<usize> = Vec::new();
        let mut seen: FxHashSet<usize> = FxHashSet::default();
        let num_strata = strata.len();

        fn visit_stratum(
            index: usize,
            dep: &DepGraph,
            strata: &Vec<Nodeset>,
            pred_to_stratum_map: &FxHashMap<ast::PredicateIndex, usize>,
            seen: &mut FxHashSet<usize>,
            sorted_indices: &mut Vec<usize>,
        ) {
            if seen.contains(&index) {
                return;
            }
            seen.insert(index);

            if let Some(scc) = strata.get(index) {
                for sym in scc {
                    if let Some(edges) = dep.get(sym) {
                        for d in edges.keys() {
                            if let Some(&dep_stratum_index) = pred_to_stratum_map.get(d) {
                                visit_stratum(
                                    dep_stratum_index,
                                    dep,
                                    strata,
                                    pred_to_stratum_map,
                                    seen,
                                    sorted_indices,
                                );
                            }
                        }
                    }
                }
            }
            sorted_indices.push(index);
        }

        for i in 0..num_strata {
            visit_stratum(
                i,
                self,
                strata,
                &pred_to_stratum_map,
                &mut seen,
                &mut sorted_indices,
            );
        }

        let mut permutation = vec![0; num_strata];
        let mut old_to_new_map: FxHashMap<usize, usize> = FxHashMap::default();
        for new_idx in 0..num_strata {
            let old_idx = sorted_indices[new_idx];
            permutation[old_idx] = new_idx;
            old_to_new_map.insert(old_idx, new_idx);
        }

        apply_permutation_cycle_rotate(strata, &permutation);

        let mut new_pred_to_stratum_map: FxHashMap<ast::PredicateIndex, usize> =
            FxHashMap::default();
        for (sym, &old_idx) in pred_to_stratum_map.iter() {
            if let Some(&new_idx) = old_to_new_map.get(&old_idx) {
                new_pred_to_stratum_map.insert(*sym, new_idx);
            }
        }
        new_pred_to_stratum_map
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use mangle_parse::Parser;

    #[test]
    fn test_stratification_success() {
        let arena = Arena::new_with_global_interner();
        let source = r#"
            p(1).
            q(X) :- p(X).
            r(X) :- q(X), !s(X).
            s(2).
        "#;
        let mut parser = Parser::new(&arena, source.as_bytes(), "test");
        parser.next_token().unwrap();
        let unit = parser.parse_unit().unwrap();

        let mut program = Program::new(&arena);
        for clause in unit.clauses {
            program.add_clause(&arena, clause);
        }

        let stratified = program.stratify().expect("should be stratifiable");

        // Helper to check relative order
        let get_stratum = |name: &str| -> Option<usize> {
            let name_idx = arena.lookup_opt(name)?;
            let pred_idx = arena.lookup_predicate_sym(name_idx)?;
            stratified.pred_to_index(pred_idx)
        };

        let s_idx = get_stratum("s");
        let r_idx = get_stratum("r");
        let q_idx = get_stratum("q");
        let p_idx = get_stratum("p");

        assert!(s_idx.is_some());
        assert!(r_idx.is_some());
        assert!(q_idx.is_some());
        assert!(p_idx.is_some());

        // r depends negatively on s, so r > s
        assert!(r_idx.unwrap() > s_idx.unwrap(), "r should be higher than s");

        // q depends on p, so q >= p
        assert!(q_idx.unwrap() >= p_idx.unwrap(), "q should be >= p");

        // r depends on q, so r >= q
        assert!(r_idx.unwrap() >= q_idx.unwrap(), "r should be >= q");
    }

    #[test]
    fn test_stratification_cycle() {
        let arena = Arena::new_with_global_interner();
        let source = "p(X) :- !p(X).";
        let mut parser = Parser::new(&arena, source.as_bytes(), "test");
        parser.next_token().unwrap();
        let unit = parser.parse_unit().unwrap();

        let mut program = Program::new(&arena);
        for clause in unit.clauses {
            program.add_clause(&arena, clause);
        }

        let res = program.stratify();
        assert!(res.is_err(), "should detect negation cycle");
    }
}
