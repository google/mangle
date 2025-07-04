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
use std::fmt;

/// An implementation of the `Program` trait.
#[derive(Clone)]
pub struct SimpleProgram<'p> {
    pub arena: &'p Arena,
    pub ext_preds: Vec<ast::PredicateIndex>,
    pub rules: FxHashMap<ast::PredicateIndex, Vec<&'p ast::Clause<'p>>>,
}

impl<'p> fmt::Debug for SimpleProgram<'p> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("SimpleProgram")
            .field("ext_preds", &self.ext_preds)
            .field("rules", &self.rules)
            .finish()
    }
}

/// An implementation of the `StratifiedProgram` trait.
/// This can be obtained through SimpleProgram::stratify.
#[derive(Clone)]
pub struct SimpleStratifiedProgram<'p> {
    program: SimpleProgram<'p>,
    strata: Vec<PredicateSet>,
}

impl<'p> fmt::Debug for SimpleStratifiedProgram<'p> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("SimpleStratifiedProgram")
            .field("program", &self.program)
            .field("strata", &self.strata)
            .finish()
    }
}

// edgeMap represents the dependencies, i.e. those IDB predicate symbols q that
// appear in the body of a rule p :- ... q ..., possibly negated.
// The boolean indicates whether the target appears negated. If there is
// both a positive and negated dependency, we only keep the negative one.
type EdgeMap = FxHashMap<ast::PredicateIndex, bool>;

// depGraph maps each predicate symbol p to its edge map.
type DepGraph = FxHashMap<ast::PredicateIndex, EdgeMap>;

// Nodeset represents a set of nodes in the dependency graph.
type Nodeset = FxHashSet<ast::PredicateIndex>;

impl<'p> Program<'p> for SimpleProgram<'p> {
    fn arena(&'p self) -> &'p Arena {
        self.arena
    }

    fn extensional_preds(&'p self) -> PredicateSet {
        let mut set = FxHashSet::default();
        set.extend(&self.ext_preds);
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

    /// Stratify checks whether a program can be stratified. It returns a
    /// SimpleStratifiedProgram in the affirmative case, an error otherwise.
    /// The result list of strata is topologically sorted.
    /// Stratification means separating a list of intensional predicate symbols with rules
    /// into strata (layers) such that each layer only depends on the
    /// evaluation of negated atoms for idb predicates in lower layers.
    pub fn stratify(self) -> Result<SimpleStratifiedProgram<'p>, String> {
        let dep = make_dep_graph(&self);
        let mut strata = dep.sccs();

        let mut pred_to_stratum: FxHashMap<ast::PredicateIndex, usize> = FxHashMap::default();

        for (i, c) in strata.iter().enumerate() {
            for sym in c {
                pred_to_stratum.insert(*sym, i);
            }
            for sym in c {
                if let Some(edges) = dep.get(&sym) {
                    for (dest, negated) in edges {
                        if !*negated {
                            continue;
                        }
                        // "Negative" dependency in same stratum indicates recursion through negation.
                        let dest_stratum = pred_to_stratum.get(dest);
                        if let Some(dest_stratum) = dest_stratum {
                            if *dest_stratum == i {
                                // TODO: iansharkey - Improve error message to include predicate names.
                                return Err("program cannot be stratified".to_string());
                            }
                        }
                    }
                }
            }
        }
        dep.sort_result(&mut strata, pred_to_stratum); // Pass strata as mutable, only get the map back
        let stratified =
            SimpleStratifiedProgram { program: self, strata: strata.into_iter().collect() }; // strata is now owned Vec
        Ok(stratified)
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

fn make_dep_graph<'p>(program: &SimpleProgram<'p>) -> DepGraph {
    let mut dep: DepGraph = FxHashMap::default();

    for (s, rule) in program.rules.iter() {
        dep.init_node(*s);
        for clause in rule.iter() {
            for premise in clause.premises.iter() {
                match premise {
                    ast::Term::Atom(atom_pred) => {
                        // TODO: iansharkey - Add support for builtins.
                        if !program.extensional_preds().contains(&atom_pred.sym) {
                            // TODO: iansharkey - Add support for multiple transforms.
                            if clause.transform.len() == 0 || clause.transform[0].var != None {
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

fn apply_permutation_cycle_rotate<T: Default>(arr: &mut Vec<T>, permutation: &[usize]) {
    let n = arr.len();
    if n == 0 {
        return;
    }
    if permutation.len() != n {
        panic!("Permutation length must match array length.");
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
        self.entry(src).or_insert_with(EdgeMap::default);
    }

    fn add_edge(&mut self, src: ast::PredicateIndex, dest: ast::PredicateIndex, negated: bool) {
        let edges = self.entry(src).or_insert_with(EdgeMap::default);
        // If the new edge is negated, it always takes precedence.
        if negated {
            edges.insert(dest, negated);
            return;
        }

        // Add a positive edge only if a negative one doesn't already exist.
        if edges.get(&dest) == None || !edges[&dest] {
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
        // Kosaraju's algorithm
        let mut s: Vec<ast::PredicateIndex> = Vec::new(); // postorder stack
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

        // Reverse pass.
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
                let mut scc: Nodeset = FxHashSet::default(); // Initialize scc here
                rvisit(top, &rev, &mut scc, &mut seen);
                if !scc.is_empty() {
                    sccs.push(scc);
                }
            }
        }
        sccs
    }

    // Sorts the strata topologically.
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
                // FxHash is deterministic, so the iteration order will always be the same.
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
            visit_stratum(i, self, &strata, &pred_to_stratum_map, &mut seen, &mut sorted_indices);
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
mod test {
    use super::*;
    use googletest::matchers::{elements_are, eq, err};
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
        let stratified = simple.stratify().unwrap();

        verify_that!(stratified.pred_to_index(bar), eq(Some(0)))?;
        verify_that!(stratified.strata(), elements_are![&single_layer])
    }

    #[test]
    fn test_stratification_neg_direct_negation() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        // foo() :- !foo().
        let foo = arena.predicate_sym("foo", Some(0));
        let clause = ast::Clause {
            head: arena.atom(foo, &[]),
            premises: &[&ast::Term::NegAtom(arena.atom(foo, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause);

        let result = simple.stratify();
        verify_that!(result, err(eq("program cannot be stratified")))
    }

    #[test]
    fn test_stratification_neg_mutual_negation() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let bar = arena.predicate_sym("bar", Some(1));
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![bar], rules: FxHashMap::default() };

        // foo(X) :- !sna(X), bar(X).
        // sna(X) :- !foo(X), bar(X).
        let foo = arena.predicate_sym("foo", Some(1));
        let sna = arena.predicate_sym("sna", Some(1));
        let var_x = arena.variable("X");
        let bar_x = arena.atom(bar, &[var_x]);
        let foo_x = arena.atom(foo, &[var_x]);
        let sna_x = arena.atom(sna, &[var_x]);

        let clause1 = ast::Clause {
            head: foo_x,
            premises: &[&ast::Term::NegAtom(sna_x), &ast::Term::Atom(bar_x)],
            transform: &[],
        };
        simple.add_clause(&arena, &clause1);

        let clause2 = ast::Clause {
            head: sna_x,
            premises: &[&ast::Term::NegAtom(foo_x), &ast::Term::Atom(bar_x)],
            transform: &[],
        };
        simple.add_clause(&arena, &clause2);

        let result = simple.stratify();
        verify_that!(result, err(eq("program cannot be stratified")))
    }

    #[test]
    fn test_stratification_pos() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        // c().
        let c = arena.predicate_sym("c", Some(0));
        let clause1 = ast::Clause { head: arena.atom(c, &[]), premises: &[], transform: &[] };
        simple.add_clause(&arena, &clause1);

        // b() :- c().
        let b = arena.predicate_sym("b", Some(0));
        let clause2 = ast::Clause {
            head: arena.atom(b, &[]),
            premises: &[&ast::Term::Atom(arena.atom(c, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause2);

        // a() :- b().
        let a = arena.predicate_sym("a", Some(0));
        let clause3 = ast::Clause {
            head: arena.atom(a, &[]),
            premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause3);

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;

        let stratified = result.unwrap();
        verify_that!(stratified.pred_to_index(c), eq(Some(0)))?;
        verify_that!(stratified.pred_to_index(b), eq(Some(1)))?;
        verify_that!(stratified.pred_to_index(a), eq(Some(2)))?;

        let mut strata0 = FxHashSet::default();
        strata0.insert(c);
        let mut strata1 = FxHashSet::default();
        strata1.insert(b);
        let mut strata2 = FxHashSet::default();
        strata2.insert(a);

        verify_that!(stratified.strata(), elements_are![&strata0, &strata1, &strata2])
    }

    #[test]
    fn test_stratification_negation_lower_strata() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        // b().
        let b = arena.predicate_sym("b", Some(0));
        let clause1 = ast::Clause { head: arena.atom(b, &[]), premises: &[], transform: &[] };
        simple.add_clause(&arena, &clause1);

        // a() :- !b().
        let a = arena.predicate_sym("a", Some(0));
        let clause2 = ast::Clause {
            head: arena.atom(a, &[]),
            premises: &[&ast::Term::NegAtom(arena.atom(b, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause2);

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;

        let stratified = result.unwrap();
        verify_that!(stratified.pred_to_index(b), eq(Some(0)))?;
        verify_that!(stratified.pred_to_index(a), eq(Some(1)))?;

        let mut strata0 = FxHashSet::default();
        strata0.insert(b);
        let mut strata1 = FxHashSet::default();
        strata1.insert(a);

        verify_that!(stratified.strata(), elements_are![&strata0, &strata1])
    }

    #[test]
    fn test_stratification_mutual_recursion_no_negation() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        // a() :- b().
        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));
        let clause1 = ast::Clause {
            head: arena.atom(a, &[]),
            premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause1);

        // b() :- a().
        let clause2 = ast::Clause {
            head: arena.atom(b, &[]),
            premises: &[&ast::Term::Atom(arena.atom(a, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause2);

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;

        let stratified = result.unwrap();
        verify_that!(stratified.strata().len(), eq(1))?;
        verify_that!(stratified.pred_to_index(a), eq(Some(0)))?;
        verify_that!(stratified.pred_to_index(b), eq(Some(0)))?;

        let mut expected_stratum = FxHashSet::default();
        expected_stratum.insert(a);
        expected_stratum.insert(b);
        verify_that!(stratified.strata(), elements_are![&expected_stratum])
    }

    #[test]
    fn test_stratification_multiple_strata_dependency() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let c = arena.predicate_sym("c", Some(0));
        let d = arena.predicate_sym("d", Some(0));
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![c, d], rules: FxHashMap::default() };

        // b() :- c().
        let b = arena.predicate_sym("b", Some(0));
        let clause2 = ast::Clause {
            head: arena.atom(b, &[]),
            premises: &[&ast::Term::Atom(arena.atom(c, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause2);

        // a() :- b(), d().
        let a = arena.predicate_sym("a", Some(0));
        let clause4 = ast::Clause {
            head: arena.atom(a, &[]),
            premises: &[&ast::Term::Atom(arena.atom(b, &[])), &ast::Term::Atom(arena.atom(d, &[]))],
            transform: &[],
        };
        simple.add_clause(&arena, &clause4);

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;

        let stratified = result.unwrap();
        verify_that!(stratified.pred_to_index(c), eq(None))?;
        verify_that!(stratified.pred_to_index(d), eq(None))?;
        verify_that!(stratified.pred_to_index(b), eq(Some(0)))?;
        verify_that!(stratified.pred_to_index(a), eq(Some(1)))?;

        let mut strata0 = FxHashSet::default();
        strata0.insert(b);
        let mut strata1 = FxHashSet::default();
        strata1.insert(a);

        verify_that!(stratified.strata(), elements_are![&strata0, &strata1])
    }

    #[test]
    fn test_stratification_scc_internal_negation() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let p1 = arena.predicate_sym("p1", Some(0));
        let p2 = arena.predicate_sym("p2", Some(0));
        let p3 = arena.predicate_sym("p3", Some(0));

        // p1 :- p2.
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p1, &[]),
                premises: &[&ast::Term::Atom(arena.atom(p2, &[]))],
                transform: &[],
            },
        );
        // p2 :- p3.
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p2, &[]),
                premises: &[&ast::Term::Atom(arena.atom(p3, &[]))],
                transform: &[],
            },
        );
        // p3 :- !p1.  -- This creates an unstratifiable cycle
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p3, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(p1, &[]))],
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result, err(eq("program cannot be stratified")))
    }

    #[test]
    fn test_stratification_negation_from_higher() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        // a() :- !b().
        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(b, &[]))],
                transform: &[],
            },
        );

        // b() :- a().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(b, &[]),
                premises: &[&ast::Term::Atom(arena.atom(a, &[]))],
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result, err(eq("program cannot be stratified")))
    }

    #[test]
    fn test_stratification_extensional_deps() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let ext = arena.predicate_sym("ext", Some(0));
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![ext], rules: FxHashMap::default() };

        // a() :- !ext().
        let a = arena.predicate_sym("a", Some(0));
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(ext, &[]))],
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;
        let stratified = result.unwrap();
        verify_that!(stratified.pred_to_index(a), eq(Some(0)))
    }

    #[test]
    fn test_stratification_empty() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };
        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;
        verify_that!(result.unwrap().strata().len(), eq(0))
    }

    #[test]
    fn test_stratification_unused_preds() -> googletest::Result<()> {
        let arena = Arena::new_global();
        // B is never a head
        let b = arena.predicate_sym("b", Some(0));
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![b], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));

        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))], // Depends on B
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;
        let stratified = result.unwrap();
        // B is not intensional, so it won't be in any stratum.
        verify_that!(stratified.pred_to_index(b), eq(None))?;
        verify_that!(stratified.pred_to_index(a), eq(Some(0)))?;
        verify_that!(stratified.strata().len(), eq(1))
    }

    #[test]
    fn test_stratification_negation_across_strata() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        // d().
        let d = arena.predicate_sym("d", Some(0));
        simple.add_clause(
            &arena,
            &ast::Clause { head: arena.atom(d, &[]), premises: &[], transform: &[] },
        );

        // c() :- d.
        let c = arena.predicate_sym("c", Some(0));
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(c, &[]),
                premises: &[&ast::Term::Atom(arena.atom(d, &[]))],
                transform: &[],
            },
        );

        // b() :- !c().
        let b = arena.predicate_sym("b", Some(0));
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(b, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(c, &[]))],
                transform: &[],
            },
        );

        // a() :- !d(), b().
        let a = arena.predicate_sym("a", Some(0));
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[
                    &ast::Term::NegAtom(arena.atom(d, &[])),
                    &ast::Term::Atom(arena.atom(b, &[])),
                ],
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;
        let stratified = result.unwrap();

        verify_that!(stratified.pred_to_index(d), eq(Some(0)))?;
        verify_that!(stratified.pred_to_index(c), eq(Some(1)))?;
        verify_that!(stratified.pred_to_index(b), eq(Some(2)))?;
        verify_that!(stratified.pred_to_index(a), eq(Some(3)))?;

        verify_that!(stratified.strata().len(), eq(4))
    }

    #[test]
    fn test_stratification_negation_precedence() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let p1 = arena.predicate_sym("p1", Some(0));
        let p2 = arena.predicate_sym("p2", Some(0));

        // p1 :- p2. (Positive dependency)
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p1, &[]),
                premises: &[&ast::Term::Atom(arena.atom(p2, &[]))],
                transform: &[],
            },
        );

        // p1 :- !p2. (Negative dependency - should override)
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p1, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(p2, &[]))],
                transform: &[],
            },
        );

        // p2 :- p1.
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p2, &[]),
                premises: &[&ast::Term::Atom(arena.atom(p1, &[]))],
                transform: &[],
            },
        );

        let result = simple.stratify();
        // Because of p1 :- !p2, p1 depends negatively on p2.
        // Because of p2 :- p1, p2 depends on p1.
        // This creates a cycle with a negation.
        verify_that!(result, err(eq("program cannot be stratified")))
    }

    #[test]
    fn test_sort_result_multiple_strata() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));

        // b().
        simple.add_clause(
            &arena,
            &ast::Clause { head: arena.atom(b, &[]), premises: &[], transform: &[] },
        );

        // a() :- b().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
                transform: &[],
            },
        );

        let dep_graph = make_dep_graph(&simple);
        let mut sccs = dep_graph.sccs();

        let mut pred_to_stratum: FxHashMap<ast::PredicateIndex, usize> = FxHashMap::default();
        for (i, c) in sccs.iter().enumerate() {
            for sym in c {
                pred_to_stratum.insert(*sym, i);
            }
        }

        dep_graph.sort_result(&mut sccs, pred_to_stratum);

        verify_that!(sccs.len(), eq(2))?;

        let mut strata0 = FxHashSet::default();
        strata0.insert(b);
        let mut strata1 = FxHashSet::default();
        strata1.insert(a);

        verify_that!(sccs, elements_are![&strata0, &strata1])
    }

    #[test]
    fn test_stratification_initial_queue_order() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let p1 = arena.predicate_sym("p1", Some(0));
        let p2 = arena.predicate_sym("p2", Some(0));
        let p3 = arena.predicate_sym("p3", Some(0));
        let p4 = arena.predicate_sym("p4", Some(0));

        // p1, p2, p3 are independent initially - All are IDB
        simple.add_clause(
            &arena,
            &ast::Clause { head: arena.atom(p1, &[]), premises: &[], transform: &[] },
        );
        simple.add_clause(
            &arena,
            &ast::Clause { head: arena.atom(p2, &[]), premises: &[], transform: &[] },
        );
        simple.add_clause(
            &arena,
            &ast::Clause { head: arena.atom(p3, &[]), premises: &[], transform: &[] },
        );

        // p4 depends on p1, p2, p3 - IDB
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(p4, &[]),
                premises: &[
                    &ast::Term::Atom(arena.atom(p1, &[])),
                    &ast::Term::Atom(arena.atom(p2, &[])),
                    &ast::Term::Atom(arena.atom(p3, &[])),
                ],
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;
        let stratified = result.unwrap();

        verify_that!(stratified.strata().len(), eq(4))?;

        let p1_idx = stratified.pred_to_index(p1).unwrap();
        let p2_idx = stratified.pred_to_index(p2).unwrap();
        let p3_idx = stratified.pred_to_index(p3).unwrap();
        let p4_idx = stratified.pred_to_index(p4).unwrap();

        // p1, p2, p3 should be in strata before p4
        verify_that!(p1_idx < p4_idx, eq(true))?;
        verify_that!(p2_idx < p4_idx, eq(true))?;
        verify_that!(p3_idx < p4_idx, eq(true))?;

        // p4 should be in the last stratum index
        verify_that!(p4_idx, eq(3))?;

        // p1, p2, p3 are in different strata
        verify_that!(p1_idx != p2_idx, eq(true))?;
        verify_that!(p1_idx != p3_idx, eq(true))?;
        verify_that!(p2_idx != p3_idx, eq(true))?;

        let mut s_p1 = FxHashSet::default();
        s_p1.insert(p1);
        let mut s_p2 = FxHashSet::default();
        s_p2.insert(p2);
        let mut s_p3 = FxHashSet::default();
        s_p3.insert(p3);
        let mut s_p4 = FxHashSet::default();
        s_p4.insert(p4);

        let strata = stratified.strata();
        verify_that!(strata.contains(&s_p1), eq(true))?;
        verify_that!(strata.contains(&s_p2), eq(true))?;
        verify_that!(strata.contains(&s_p3), eq(true))?;
        verify_that!(strata.contains(&s_p4), eq(true))?;

        Ok(())
    }

    #[test]
    fn test_make_dep_graph_simple() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));

        // a() :- b().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
                transform: &[],
            },
        );

        let dep_graph = make_dep_graph(&simple);

        verify_that!(dep_graph.len(), eq(1))?;
        verify_that!(dep_graph.contains_key(&a), eq(true))?;
        verify_that!(dep_graph.get(&a).unwrap().len(), eq(1))?;
        verify_that!(dep_graph.get(&a).unwrap().contains_key(&b), eq(true))?;
        verify_that!(dep_graph.get(&a).unwrap().get(&b), eq(Some(&false)))
    }

    #[test]
    fn test_make_dep_graph_negation() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));

        // a() :- !b().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(b, &[]))],
                transform: &[],
            },
        );

        let dep_graph = make_dep_graph(&simple);

        verify_that!(dep_graph.get(&a).unwrap().get(&b), eq(Some(&true)))
    }

    #[test]
    fn test_make_dep_graph_negation_precedence() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));

        // a() :- b().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
                transform: &[],
            },
        );

        // a() :- !b().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::NegAtom(arena.atom(b, &[]))],
                transform: &[],
            },
        );

        let dep_graph = make_dep_graph(&simple);

        verify_that!(dep_graph.get(&a).unwrap().get(&b), eq(Some(&true)))
    }

    #[test]
    fn test_make_dep_graph_ignore_extensional() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let ext = arena.predicate_sym("ext", Some(0));
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![ext], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));

        // a() :- ext().
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(ext, &[]))],
                transform: &[],
            },
        );

        let dep_graph = make_dep_graph(&simple);
        verify_that!(dep_graph.get(&a).unwrap().contains_key(&ext), eq(false))
    }

    #[test]
    fn test_make_dep_graph_do_transform() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));

        // a() :- b() do { ... }.
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
                transform: &[arena.alloc(ast::TransformStmt {
                    var: None,
                    app: arena.const_(ast::Const::Number(0)),
                })],
            },
        );

        let dep_graph = make_dep_graph(&simple);

        // Expect a negative dependency due to the do transform.
        verify_that!(dep_graph.get(&a).unwrap().get(&b), eq(Some(&true)))
    }

    #[test]
    fn test_sccs_dfs_order() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let mut simple =
            SimpleProgram { arena: &arena, ext_preds: vec![], rules: FxHashMap::default() };

        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));
        let c = arena.predicate_sym("c", Some(0));

        // Build a graph: A -> B, B -> C, C -> A (SCC: {A, B, C})
        // Also, A -> C
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
                transform: &[],
            },
        );
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(b, &[]),
                premises: &[&ast::Term::Atom(arena.atom(c, &[]))],
                transform: &[],
            },
        );
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(c, &[]),
                premises: &[&ast::Term::Atom(arena.atom(a, &[]))],
                transform: &[],
            },
        );
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(c, &[]))],
                transform: &[],
            },
        );

        let dep_graph = make_dep_graph(&simple);
        let sccs = dep_graph.sccs();

        verify_that!(sccs.len(), eq(1))?;
        let scc = &sccs[0];
        verify_that!(scc.len(), eq(3))?;
        verify_that!(scc.contains(&a), eq(true))?;
        verify_that!(scc.contains(&b), eq(true))?;
        verify_that!(scc.contains(&c), eq(true))
    }

    #[test]
    fn test_sccs_visit_order() -> googletest::Result<()> {
        let arena = Arena::new_global();
        let a = arena.predicate_sym("a", Some(0));
        let b = arena.predicate_sym("b", Some(0));
        let c = arena.predicate_sym("c", Some(0));
        let d = arena.predicate_sym("d", Some(0));

        let mut simple = SimpleProgram {
            arena: &arena,
            ext_preds: vec![d], // Add d to extensional predicates
            rules: FxHashMap::default(),
        };

        // Graph:
        // A -> B
        // A -> C
        // B -> D (D is EDB)
        // C -> D (D is EDB)

        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(b, &[]))],
                transform: &[],
            },
        );
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(a, &[]),
                premises: &[&ast::Term::Atom(arena.atom(c, &[]))],
                transform: &[],
            },
        );
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(b, &[]),
                premises: &[&ast::Term::Atom(arena.atom(d, &[]))],
                transform: &[],
            },
        );
        simple.add_clause(
            &arena,
            &ast::Clause {
                head: arena.atom(c, &[]),
                premises: &[&ast::Term::Atom(arena.atom(d, &[]))],
                transform: &[],
            },
        );

        let result = simple.stratify();
        verify_that!(result.is_ok(), eq(true))?;
        let stratified = result.unwrap();

        // Expected strata (D is extensional):
        // Strata 0: {B}
        // Strata 1: {C}
        // Strata 2: {A}
        // Or Strata 0: {C}, Strata 1: {B}, Strata 2: {A}

        verify_that!(stratified.strata().len(), eq(3))?;

        verify_that!(stratified.pred_to_index(d), eq(None))?; // D not in any stratum
        let a_idx = stratified.pred_to_index(a).unwrap();
        let b_idx = stratified.pred_to_index(b).unwrap();
        let c_idx = stratified.pred_to_index(c).unwrap();

        verify_that!(a_idx, eq(2))?; // A should be in the last stratum
        verify_that!(b_idx == c_idx, eq(false))?; // B and C should be in different strata
        verify_that!(b_idx < a_idx, eq(true))?;
        verify_that!(c_idx < a_idx, eq(true))?;

        let mut strata0 = FxHashSet::default();
        strata0.insert(b);
        let mut strata1 = FxHashSet::default();
        strata1.insert(c);
        let mut strata2 = FxHashSet::default();
        strata2.insert(a);

        // We cannot guarantee the order of {B} and {C}, so check contents
        let strata = stratified.strata();
        verify_that!(
            strata.contains(&strata0) && strata.contains(&strata1) && strata.contains(&strata2),
            eq(true)
        )?;
        Ok(())
    }
}
