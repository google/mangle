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

use fxhash::{FxHashMap, FxHashSet};
use mangle_ast as ast;

/// Rewrites a unit by prefixing local predicates with the package name.
pub fn rewrite_unit<'a>(arena: &'a ast::Arena, unit: &'a ast::Unit<'a>) -> ast::Unit<'a> {
    let (pkg_name, used_pkgs) = find_package_info(arena, unit);

    if pkg_name.is_empty() {
        return ast::Unit {
            decls: unit.decls,
            clauses: unit.clauses,
        };
    }

    let defined_preds = find_defined_preds(unit);
    let mut renamer = Renamer {
        arena,
        pkg_name,
        used_pkgs,
        defined_preds,
        cache: FxHashMap::default(),
    };

    let mut new_decls = Vec::with_capacity(unit.decls.len());
    for &decl in unit.decls {
        if let Some(new_decl) = renamer.rewrite_decl(decl) {
            new_decls.push(new_decl);
        }
    }

    let mut new_clauses = Vec::with_capacity(unit.clauses.len());
    for &clause in unit.clauses {
        if let Some(new_clause) = renamer.rewrite_clause(clause) {
            new_clauses.push(new_clause);
        }
    }

    ast::Unit {
        decls: arena.alloc_slice_copy(&new_decls),
        clauses: arena.alloc_slice_copy(&new_clauses),
    }
}

fn find_package_info<'a>(
    arena: &'a ast::Arena,
    unit: &'a ast::Unit<'a>,
) -> (&'a str, FxHashSet<&'a str>) {
    let mut pkg_name = "";
    let mut used_pkgs = FxHashSet::default();

    for &decl in unit.decls {
        let pred_name = arena.predicate_name(decl.atom.sym).unwrap_or("");
        if pred_name == "Package" {
            if let Some(desc) = find_name_desc(arena, decl.descr) {
                pkg_name = desc;
            }
        } else if pred_name == "Use"
            && let Some(desc) = find_name_desc(arena, decl.descr)
        {
            used_pkgs.insert(desc);
        }
    }
    (pkg_name, used_pkgs)
}

fn find_name_desc<'a>(arena: &'a ast::Arena, descr: &'a [&'a ast::Atom<'a>]) -> Option<&'a str> {
    for &atom in descr {
        if arena.predicate_name(atom.sym).unwrap_or("") == "name"
            && let Some(&ast::BaseTerm::Const(ast::Const::String(s))) = atom.args.first().copied()
        {
            return Some(s);
        }
    }
    None
}

fn find_defined_preds(unit: &ast::Unit) -> FxHashSet<ast::PredicateIndex> {
    let mut defined = FxHashSet::default();
    for &decl in unit.decls {
        defined.insert(decl.atom.sym);
    }
    for &clause in unit.clauses {
        defined.insert(clause.head.sym);
    }
    defined
}

struct Renamer<'a> {
    arena: &'a ast::Arena,
    pkg_name: &'a str,
    used_pkgs: FxHashSet<&'a str>,
    defined_preds: FxHashSet<ast::PredicateIndex>,
    cache: FxHashMap<ast::PredicateIndex, ast::PredicateIndex>,
}

impl<'a> Renamer<'a> {
    fn rename_pred(&mut self, sym: ast::PredicateIndex) -> Option<ast::PredicateIndex> {
        if let Some(&new_sym) = self.cache.get(&sym) {
            return Some(new_sym);
        }

        let name = self.arena.predicate_name(sym)?;

        // Don't rename Package and Use predicates themselves
        if name == "Package" || name == "Use" {
            self.cache.insert(sym, sym);
            return Some(sym);
        }

        if self.defined_preds.contains(&sym) {
            let new_name = format!("{}.{}", self.pkg_name, name);
            let new_sym = self
                .arena
                .predicate_sym(&new_name, self.arena.predicate_arity(sym));
            self.cache.insert(sym, new_sym);
            return Some(new_sym);
        }

        // Check for cross-package references (e.g. `other.foo`)
        if let Some(dot_idx) = name.rfind('.') {
            let prefix = &name[..dot_idx];
            if self.used_pkgs.contains(prefix) {
                // It's a valid external reference, keep it as is.
                self.cache.insert(sym, sym);
                return Some(sym);
            }
        }

        // Not defined locally, no dot. Must be a builtin or global?
        self.cache.insert(sym, sym);
        Some(sym)
    }

    fn rewrite_decl(&mut self, decl: &'a ast::Decl<'a>) -> Option<&'a ast::Decl<'a>> {
        let pred_name = self.arena.predicate_name(decl.atom.sym).unwrap_or("");
        if pred_name == "Package" || pred_name == "Use" {
            // Remove Package and Use declarations from the rewritten unit
            return None;
        }

        let new_atom = self.rewrite_atom(decl.atom)?;

        // Rewrite bounds if necessary
        let bounds = if let Some(bs) = decl.bounds {
            let mut new_bounds = Vec::new();
            for &b in bs {
                let new_base_terms: Vec<&ast::BaseTerm> = b
                    .base_terms
                    .iter()
                    .map(|&t| {
                        // Check if it's a name constant that needs rewriting (e.g. type names like /foo -> /pkg.foo)
                        if let ast::BaseTerm::Const(ast::Const::Name(name_idx)) = t {
                            let name = self.arena.lookup_name(*name_idx).unwrap_or("");
                            // Check if name corresponds to a defined predicate
                            if let Some(pred_idx) = self.arena.lookup_predicate_sym(*name_idx)
                                && self.defined_preds.contains(&pred_idx)
                            {
                                let new_name = format!("{}.{}", self.pkg_name, name);
                                let new_name_idx = self.arena.intern(&new_name);
                                return &*self
                                    .arena
                                    .alloc(ast::BaseTerm::Const(ast::Const::Name(new_name_idx)));
                            }
                        }
                        t
                    })
                    .collect();
                new_bounds.push(&*self.arena.alloc(ast::BoundDecl {
                    base_terms: self.arena.alloc_slice_copy(&new_base_terms),
                }));
            }
            Some(self.arena.alloc_slice_copy(&new_bounds) as &'a [&'a ast::BoundDecl<'a>])
        } else {
            None
        };

        Some(self.arena.alloc(ast::Decl {
            atom: new_atom,
            descr: decl.descr, // Descriptions usually don't contain predicates to rename?
            bounds,
            constraints: decl.constraints,
        }))
    }

    fn rewrite_clause(&mut self, clause: &'a ast::Clause<'a>) -> Option<&'a ast::Clause<'a>> {
        let head = self.rewrite_atom(clause.head)?;
        let mut premises = Vec::new();
        for &premise in clause.premises {
            match premise {
                ast::Term::Atom(a) => {
                    premises.push(&*self.arena.alloc(ast::Term::Atom(self.rewrite_atom(a)?)));
                }
                ast::Term::NegAtom(a) => {
                    premises.push(&*self.arena.alloc(ast::Term::NegAtom(self.rewrite_atom(a)?)));
                }
                _ => premises.push(premise),
            }
        }

        Some(self.arena.alloc(ast::Clause {
            head,
            premises: self.arena.alloc_slice_copy(&premises),
            transform: clause.transform,
        }))
    }

    fn rewrite_atom(&mut self, atom: &'a ast::Atom<'a>) -> Option<&'a ast::Atom<'a>> {
        let new_sym = self.rename_pred(atom.sym)?;
        Some(self.arena.atom(new_sym, atom.args))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use googletest::prelude::*;

    #[test]
    fn test_rename_simple() {
        let arena = ast::Arena::new_with_global_interner();

        // Package pkg!
        // foo(X) :- bar(X).
        // bar(1).

        let pkg_sym = arena.predicate_sym("Package", Some(0));
        let name_sym = arena.predicate_sym("name", Some(1));
        let pkg_name = arena.alloc(ast::BaseTerm::Const(ast::Const::String("pkg")));
        let pkg_decl = arena.alloc(ast::Decl {
            atom: arena.atom(pkg_sym, &[]),
            descr: arena.alloc_slice_copy(&[arena.atom(name_sym, &[pkg_name])]),
            bounds: None,
            constraints: None,
        });

        let foo_sym = arena.predicate_sym("foo", Some(1));
        let bar_sym = arena.predicate_sym("bar", Some(1));
        let var_x = arena.variable("X");
        let const_1 = arena.const_(ast::Const::Number(1));

        let clause1 = arena.alloc(ast::Clause {
            head: arena.atom(foo_sym, &[var_x]),
            premises: arena
                .alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(bar_sym, &[var_x])))]),
            transform: &[],
        });

        let clause2 = arena.alloc(ast::Clause {
            head: arena.atom(bar_sym, &[const_1]),
            premises: &[],
            transform: &[],
        });

        let unit = ast::Unit {
            decls: arena.alloc_slice_copy(&[pkg_decl]),
            clauses: arena.alloc_slice_copy(&[clause1, clause2]),
        };

        let new_unit = rewrite_unit(&arena, &unit);

        // Package decl should be removed
        assert_that!(new_unit.decls.len(), eq(0));
        assert_that!(new_unit.clauses.len(), eq(2));

        let c1 = new_unit.clauses[0];
        let c2 = new_unit.clauses[1];

        // Check head of c1: foo -> pkg.foo
        let head_name = arena.predicate_name(c1.head.sym).unwrap();
        assert_that!(head_name, eq("pkg.foo"));

        // Check premise of c1: bar -> pkg.bar
        if let ast::Term::Atom(a) = c1.premises[0] {
            let p_name = arena.predicate_name(a.sym).unwrap();
            assert_that!(p_name, eq("pkg.bar"));
        } else {
            panic!("Expected Atom premise");
        }

        // Check head of c2: bar -> pkg.bar
        let head_name_2 = arena.predicate_name(c2.head.sym).unwrap();
        assert_that!(head_name_2, eq("pkg.bar"));
    }

    #[test]
    fn test_rename_with_use() {
        let arena = ast::Arena::new_with_global_interner();

        // Package pkg!
        // Use other!
        // foo(X) :- other.bar(X).

        let pkg_sym = arena.predicate_sym("Package", Some(0));
        let use_sym = arena.predicate_sym("Use", Some(0));
        let name_sym = arena.predicate_sym("name", Some(1));

        let pkg_name = arena.alloc(ast::BaseTerm::Const(ast::Const::String("pkg")));
        let pkg_decl = arena.alloc(ast::Decl {
            atom: arena.atom(pkg_sym, &[]),
            descr: arena.alloc_slice_copy(&[arena.atom(name_sym, &[pkg_name])]),
            bounds: None,
            constraints: None,
        });

        let other_name = arena.alloc(ast::BaseTerm::Const(ast::Const::String("other")));
        let use_decl = arena.alloc(ast::Decl {
            atom: arena.atom(use_sym, &[]),
            descr: arena.alloc_slice_copy(&[arena.atom(name_sym, &[other_name])]),
            bounds: None,
            constraints: None,
        });

        let foo_sym = arena.predicate_sym("foo", Some(1));
        let other_bar_sym = arena.predicate_sym("other.bar", Some(1));
        let var_x = arena.variable("X");

        let clause1 = arena.alloc(ast::Clause {
            head: arena.atom(foo_sym, &[var_x]),
            premises: arena.alloc_slice_copy(&[
                arena.alloc(ast::Term::Atom(arena.atom(other_bar_sym, &[var_x])))
            ]),
            transform: &[],
        });

        let unit = ast::Unit {
            decls: arena.alloc_slice_copy(&[pkg_decl, use_decl]),
            clauses: arena.alloc_slice_copy(&[clause1]),
        };

        let new_unit = rewrite_unit(&arena, &unit);

        assert_that!(new_unit.clauses.len(), eq(1));
        let c1 = new_unit.clauses[0];

        // foo -> pkg.foo
        let head_name = arena.predicate_name(c1.head.sym).unwrap();
        assert_that!(head_name, eq("pkg.foo"));

        // other.bar -> other.bar (unchanged because 'other' is Used)
        if let ast::Term::Atom(a) = c1.premises[0] {
            let p_name = arena.predicate_name(a.sym).unwrap();
            assert_that!(p_name, eq("other.bar"));
        } else {
            panic!("Expected Atom premise");
        }
    }

    fn make_pkg_decl<'a>(arena: &'a ast::Arena, name: &str) -> &'a ast::Decl<'a> {
        let pkg_sym = arena.predicate_sym("Package", Some(0));
        let name_sym = arena.predicate_sym("name", Some(1));
        let pkg_name = arena.alloc(ast::BaseTerm::Const(ast::Const::String(
            arena.alloc_str(name),
        )));
        arena.alloc(ast::Decl {
            atom: arena.atom(pkg_sym, &[]),
            descr: arena.alloc_slice_copy(&[arena.atom(name_sym, &[pkg_name])]),
            bounds: None,
            constraints: None,
        })
    }

    #[test]
    fn test_go_case_no_package() {
        // "no package name, clauses are not rewritten"
        let arena = ast::Arena::new_with_global_interner();
        let clause = arena.alloc(ast::Clause {
            head: arena.atom(arena.predicate_sym("clause_defined_here", None), &[]),
            premises: arena.alloc_slice_copy(&[arena.alloc(ast::Term::Atom(
                arena.atom(arena.predicate_sym("other_clause", None), &[]),
            ))]),
            transform: &[],
        });
        let unit = ast::Unit {
            decls: &[],
            clauses: arena.alloc_slice_copy(&[clause]),
        };
        let new_unit = rewrite_unit(&arena, &unit);
        let head = arena.predicate_name(new_unit.clauses[0].head.sym).unwrap();
        assert_that!(head, eq("clause_defined_here"));
        if let ast::Term::Atom(a) = new_unit.clauses[0].premises[0] {
            let p = arena.predicate_name(a.sym).unwrap();
            assert_that!(p, eq("other_clause"));
        }
    }

    #[test]
    fn test_go_case_external_refs() {
        // "references to predicates outside the package are left as-is"
        let arena = ast::Arena::new_with_global_interner();
        // Package foo.bar!
        let pkg_decl = make_pkg_decl(&arena, "foo.bar");

        // clause_defined_here :- other_clause.
        // (other_clause is NOT defined locally)
        let clause = arena.alloc(ast::Clause {
            head: arena.atom(arena.predicate_sym("clause_defined_here", None), &[]),
            premises: arena.alloc_slice_copy(&[arena.alloc(ast::Term::Atom(
                arena.atom(arena.predicate_sym("other_clause", None), &[]),
            ))]),
            transform: &[],
        });

        let unit = ast::Unit {
            decls: arena.alloc_slice_copy(&[pkg_decl]),
            clauses: arena.alloc_slice_copy(&[clause]),
        };
        let new_unit = rewrite_unit(&arena, &unit);

        // head -> foo.bar.clause_defined_here
        let head = arena.predicate_name(new_unit.clauses[0].head.sym).unwrap();
        assert_that!(head, eq("foo.bar.clause_defined_here"));

        // premise -> other_clause (unchanged)
        if let ast::Term::Atom(a) = new_unit.clauses[0].premises[0] {
            let p = arena.predicate_name(a.sym).unwrap();
            assert_that!(p, eq("other_clause"));
        }
    }

    #[test]
    fn test_go_case_rewritten_local() {
        // "clauses defined in this package are rewritten"
        let arena = ast::Arena::new_with_global_interner();
        let pkg_decl = make_pkg_decl(&arena, "foo.bar");

        let defined_sym = arena.predicate_sym("clause_defined_here", None);
        let other_sym = arena.predicate_sym("other_clause", None);

        // other_clause().
        let clause1 = arena.alloc(ast::Clause {
            head: arena.atom(other_sym, &[]),
            premises: &[],
            transform: &[],
        });

        // clause_defined_here() :- other_clause().
        let clause2 = arena.alloc(ast::Clause {
            head: arena.atom(defined_sym, &[]),
            premises: arena
                .alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(other_sym, &[])))]),
            transform: &[],
        });

        let unit = ast::Unit {
            decls: arena.alloc_slice_copy(&[pkg_decl]),
            clauses: arena.alloc_slice_copy(&[clause1, clause2]),
        };
        let new_unit = rewrite_unit(&arena, &unit);

        // clause1 head: foo.bar.other_clause
        let h1 = arena.predicate_name(new_unit.clauses[0].head.sym).unwrap();
        assert_that!(h1, eq("foo.bar.other_clause"));

        // clause2 head: foo.bar.clause_defined_here
        let h2 = arena.predicate_name(new_unit.clauses[1].head.sym).unwrap();
        assert_that!(h2, eq("foo.bar.clause_defined_here"));

        // clause2 premise: foo.bar.other_clause
        if let ast::Term::Atom(a) = new_unit.clauses[1].premises[0] {
            let p = arena.predicate_name(a.sym).unwrap();
            assert_that!(p, eq("foo.bar.other_clause"));
        }
    }

    #[test]
    fn test_go_case_negation() {
        // "clause with a negation is rewritten"
        let arena = ast::Arena::new_with_global_interner();
        let pkg_decl = make_pkg_decl(&arena, "foo.bar");

        let defined_sym = arena.predicate_sym("clause_defined_here", None);
        let other_sym = arena.predicate_sym("other_clause", None);

        // other_clause(). (Needs to be defined to trigger renaming)
        let clause1 = arena.alloc(ast::Clause {
            head: arena.atom(other_sym, &[]),
            premises: &[],
            transform: &[],
        });

        // clause_defined_here() :- !other_clause().
        let clause2 = arena.alloc(ast::Clause {
            head: arena.atom(defined_sym, &[]),
            premises: arena
                .alloc_slice_copy(&[arena.alloc(ast::Term::NegAtom(arena.atom(other_sym, &[])))]),
            transform: &[],
        });

        let unit = ast::Unit {
            decls: arena.alloc_slice_copy(&[pkg_decl]),
            clauses: arena.alloc_slice_copy(&[clause1, clause2]),
        };
        let new_unit = rewrite_unit(&arena, &unit);

        if let ast::Term::NegAtom(a) = new_unit.clauses[1].premises[0] {
            let p = arena.predicate_name(a.sym).unwrap();
            assert_that!(p, eq("foo.bar.other_clause"));
        } else {
            panic!("Expected NegAtom");
        }
    }

    #[test]
    fn test_go_case_decl_only() {
        // "clauses are also rewritten if the decl was declared in this package"
        let arena = ast::Arena::new_with_global_interner();
        let pkg_decl = make_pkg_decl(&arena, "foo.bar");

        let clause_sym = arena.predicate_sym("clause", None);
        let decl_sym = arena.predicate_sym("from_decl", None);

        // Decl from_decl.
        let decl = arena.alloc(ast::Decl {
            atom: arena.atom(decl_sym, &[]),
            descr: &[],
            bounds: None,
            constraints: None,
        });

        // clause() :- from_decl().
        let clause = arena.alloc(ast::Clause {
            head: arena.atom(clause_sym, &[]),
            premises: arena
                .alloc_slice_copy(&[arena.alloc(ast::Term::Atom(arena.atom(decl_sym, &[])))]),
            transform: &[],
        });

        let unit = ast::Unit {
            decls: arena.alloc_slice_copy(&[pkg_decl, decl]),
            clauses: arena.alloc_slice_copy(&[clause]),
        };
        let new_unit = rewrite_unit(&arena, &unit);

        // clause -> foo.bar.clause
        let h = arena.predicate_name(new_unit.clauses[0].head.sym).unwrap();
        assert_that!(h, eq("foo.bar.clause"));

        // from_decl -> foo.bar.from_decl
        if let ast::Term::Atom(a) = new_unit.clauses[0].premises[0] {
            let p = arena.predicate_name(a.sym).unwrap();
            assert_that!(p, eq("foo.bar.from_decl"));
        }

        // Decl atom also rewritten
        let d_name = arena.predicate_name(new_unit.decls[0].atom.sym).unwrap();
        assert_that!(d_name, eq("foo.bar.from_decl"));
    }
}
